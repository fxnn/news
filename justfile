# HINT: `just` is a `make` replacement,
# see https://just.systems/
#

set dotenv-load := true

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

# Run the application in server mode, sourcing arguments from NEWS_* environment variables
run-server:
  go run . --mode server \
    --server "$NEWS_SERVER" \
    --port $NEWS_PORT \
    --username "$NEWS_USERNAME" \
    --password "$NEWS_PASSWORD" \
    --folder "$NEWS_FOLDER" \
    --days $NEWS_DAYS \
    --limit $NEWS_LIMIT \
    --summarizer "$NEWS_SUMMARIZER" \
    --http-port $NEWS_HTTP_PORT \
    --html-file "$NEWS_HTML_FILE"

