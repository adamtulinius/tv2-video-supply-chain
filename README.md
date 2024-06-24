# TV2 video supply chain assignment

Each part of the assignment has a dedicated branch. I've split part 2 into several steps - please refer to the terminology section of this readme for details.


## Running it

Running it from inside the project root:

```go run . < data/tv2-video-metadata-ingest.txt```

By default it will bind to port `8080` to expose prometheus metrics. To change the port set the `PORT` environment variable to something other than `8080`, e.g.

```PORT=8081 go run . < data/tv2-video-metadata-ingest.txt```

The metrics are exposed on `/metrics`. To view the metrics I would suggest opening a second terminal and running something like:

```watch -n1 "curl -sSL localhost:8080/metrics | grep -i tv2"```

This way the metrics are updated every second, and all default metrics are filtered away.

It's a design requirement that the system stops when processing finishes, but to keep it running anyways (beneficial for still having access to the prometheus metrics) set the `KEEP_RUNNING` environment variable to `yes`, e.g.:

```KEEP_RUNNING=yes go run . < data/tv2-video-metadata-ingest.txt```

Likewise it's possible to change the number of encoders to something other than 10 using the environment var `ENCODERS`, e.g. `ENCODERS=1`, to throttle the output a bit.


Run tests:

```go test```


Run tests with coverage:

```go test -coverprofile=coverage.out```

.. and see coverage report:
```go tool cover -html=coverage.out```


I used the following version of go:
```go version go1.22.3 linux/amd64```


## Terminology

I've chosen to split the implementation into more parts than in the assignment

- Part 1: The original part 1 in the assignment
- Part 2a: Part 2 where the publisher is started at the start of encoding
- Part 2b: Part 2 where the encoding sends publishing requests to a queue
- Part 2c: Part 2 with the "free style" optimizations to the system


## Assumptions

- Since encoding video is very demanding, and the system encodes multiple videos in parallel, I assume that the system is meant to be distributed, and _not_ a batch processing system running on a single host.

- The system as described is split into a client and a server. The client is responsible for submitting encoding requests to the server, and is therefore also in charge of parsing the metadata. I have chosen not to write tests for the client, since (to me) it's not really part of the encoding and publishing system, but more of a necessity to show that the system is working (maybe the client should have been written as an integration test instead).





## Design

My solution is split into two parts:
- a client
- a server

Both client and server are started from `main.go`, that also sets up a single communications channel between the client and server. In the real production system this channel would be replaced by some kind of HTTP-based API (or maybe an external queue like Kafka or redis).

I've implemented some very basic prometheus metrics to get a better overview of the current status of the pipeline.

My design is also using channels as interface between encoders and publishers (for part 2). These would be replaced by an external queue like Kafka or redis in a real system.

Two structs are used to represent encoding jobs through the entire system (see `types.go`):

1. `IngestObject`: This is the parsed metadata as described in the description of the challange.
2. `PublishingObject`: This is passed from encoder to publisher. Since `publicationTimeout` starts ticking when the encoder starts, this struct wraps the `IngestObject` and adds the time encoding started, so that the total publication time can be tracked across encoder and publisher.


### Client (`client.go`)

The client reads the metadata from `STDIN` (one JSON-object per line) and sends encoding requests to the server.

The client is responsible for submitting encoding requests at _up to_ 2 requests/s. In the real system the server would be responsible for telling the client to retry later. Implementing the throttling in this PoC would require an extra channel between the client and the encoders, since there's 10 encoders running at the same time, so I chose to do it this way for simplicity. The throttling is implicit since writing to the shared channel blocks until an encoder is ready. 


### Server (`server.go`)

The starts by spawning a web server for the prometheus metrics, and a helper goroutine that updates the length of the encoding queue. It then spawns 10 encoders.

Each encoder reads encoding jobs from a shared queue.

#### Part 1

Each encoder spawns a publisher after encoding finishes. This publisher is also running as a goroutine, and will receive the encoding job through a dedicated and temporary channel. This channel is strictly not necessary for part 1, but makes the difference to part 2a simple. 


#### Part 2a

The encoder now starts the publisher as a goroutine before simulating encoding (making the publisher start in parallel with the encoding).


#### Part 2b

Publishers are now started as soon as encoders make PublishingObjects available on a queue.
Since publishers are no longer started when encoding starts, the rate of failed publications are now the same as in Part 1. I assume this was not the intention, so I will improve on this in part 2c.


#### Part 2c

Let's start out by answering the question about whether all video metadata objects are equal:

They are not.

1) Objects where the publicationTimeout is 6 seconds longer than the encodingTime _should_ always publish, even under the rules of part 1. Random jitter and bad luck might still cause them to fail, e.g. if the publisher takes the full maximum 6 seconds to start. This could be somewhat alleviated by always rounding the time taken towards 0 (to avoid timeouts that are only a few ms above the threshold), but it also feels a bit like cheating.

2) Objects where the publicationTimeout is 6 seconds (maximum publisher startup time) or longer _should_ always publish using the rules of part 2a (e.g. publisher starting up at the same time as the encoder), but with similar caveats as in (1) above. One could assume that objects with publicationTimeout of 7 seconds and above would always publish using the part 2a-rules, and it would be fairly safe to do so, except that random hangs or freezes (maybe caused by swapping or high load) could delay startup of the publisher.

With that knowledge it makes sense to implement pre-warmed publishers, that are ready to publish at a moments notice. If the system maintains a certain number of pre-warmed publishers it's possible to choose which objects should always use a pre-warmed publisher, and which objects can wait long enough for a new publisher to start up in case the number of pre-warmed publishers are low.

Implementation: For this part the system now tries to keep 10 publishers ready in the background. This is done by spawning 10 publishers immediately at startup, that are all reading from the publishing queue. When a publisher retrieves an object to publish, it announces on a separate channel (`new_publisher_queue`) that it is now in use. Another goroutine reads from that channel, and spawns a new publisher to replace the one that is now busy publishing.

Based on 4 runs publish failures are now down to 0-5 failures each time, and since I'm running out off time that will be the last optimization for now.


## Other design choices

Lack of error handling (e.g. encoders and publishers could catch panics and write an error message to a channel on panic, making it possible to start a new worker on errors.)

JSON log output would have been nice, but not sure it improves readability for the purpose of the assignment.

There's some double (ac)counting going on: I wanted to add prometheus metrics, but it's non-trivial to get the value of prometheus counters/gauges, and thus it's easier to just count the things needed to show in the terminal twice.

The requirement that the system must terminate after processing has caused quite a few headaches, and I'm not very happy with how I solved them. To make goroutines shut down nicely, and not too early, I'm relying a lot on closing channels that e.g. the encoders are reading from. This has caused certain components to not be as decoupled as I wanted to, and certain changes have also been a bit error prone to implement. Removing that requirement would make it possible to clean up some code, and -- after all -- the assignment does specificy that the system is supposed to be a (soft, I asssume) real-time system, and not a batch processing system.


## Diagrams



## Statistics

Each line represents different runs.

```
part 1:

stats: client_parse_errors=0 client_submitted=154 encoded=154 published=54 publish_failed=100
stats: client_parse_errors=0 client_submitted=154 encoded=154 published=66 publish_failed=88
stats: client_parse_errors=0 client_submitted=154 encoded=154 published=61 publish_failed=93


part 2a:

stats: client_parse_errors=0 client_submitted=154 encoded=154 published=128 publish_failed=26
stats: client_parse_errors=0 client_submitted=154 encoded=154 published=126 publish_failed=28
stats: client_parse_errors=0 client_submitted=154 encoded=154 published=123 publish_failed=31


part 2b:

stats: client_parse_errors=0 client_submitted=154 encoded=154 published=63 publish_failed=91
stats: client_parse_errors=0 client_submitted=154 encoded=154 published=59 publish_failed=95


### part 2c

stats: client_parse_errors=0 client_submitted=154 encoded=154 published=154 publish_failed=0
stats: client_parse_errors=0 client_submitted=154 encoded=154 published=153 publish_failed=1
stats: client_parse_errors=0 client_submitted=154 encoded=154 published=153 publish_failed=1
stats: client_parse_errors=0 client_submitted=154 encoded=154 published=151 publish_failed=3