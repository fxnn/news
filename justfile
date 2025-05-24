help: # first target is the default when running `just`
  @just --list

make: fmt vet build test

build:
  go build .

fmt:
  go fmt .

vet:
  go vet .

test:
  go test .

