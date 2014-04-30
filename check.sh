#!/bin/sh

set -e

EXES="cmd/https_proxy.go cmd/https_server.go"

for EXE in $EXES; do
  go build $EXE
  go fmt $EXE
  go vet $EXE
  echo $EXE
done
