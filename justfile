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

lint target='': # go vet doesn't accept individual filenames, but aider wants to give them
  go vet .

test:
  go test .

