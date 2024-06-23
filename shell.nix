{ pkgs ? import (builtins.fetchGit {
  url = "https://github.com/nixos/nixpkgs/";
  # `git ls-remote https://github.com/nixos/nixpkgs nixos-unstable`
  ref = "refs/heads/nixos-unstable";
  rev = "d603719ec6e294f034936c0d0dc06f689d91b6c3";
}) {} }:

pkgs.mkShell rec {
  nativeBuildInputs = [
    
  ];
  buildInputs = with pkgs; [
    go
  ];
}
