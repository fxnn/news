# HINT: `just` is a `make` replacement,
# see https://just.systems/
#

help: # first target is the default when running `just`
  @just --list

# includes format, lint, build, test
make: fmt lint build test

build:
  go build .

fmt:
  go fmt .

lint:
  go vet .

test:
  go test .

