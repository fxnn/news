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
  @echo "Starting server with arguments from environment variables (e.g., NEWS_SERVER, NEWS_PORT, NEWS_USERNAME, NEWS_PASSWORD, etc.)..."
  go run . --mode server \
    $(if [ -n "$NEWS_SERVER" ]; then echo -n " --server \"$NEWS_SERVER\""; fi) \
    $(if [ -n "$NEWS_PORT" ]; then echo -n " --port $NEWS_PORT"; fi) \
    $(if [ -n "$NEWS_USERNAME" ]; then echo -n " --username \"$NEWS_USERNAME\""; fi) \
    $(if [ -n "$NEWS_PASSWORD" ]; then echo -n " --password \"$NEWS_PASSWORD\""; fi) \
    $(if [ -n "$NEWS_FOLDER" ]; then echo -n " --folder \"$NEWS_FOLDER\""; fi) \
    $(if [ -n "$NEWS_DAYS" ]; then echo -n " --days $NEWS_DAYS"; fi) \
    $(if [ -n "$NEWS_LIMIT" ]; then echo -n " --limit $NEWS_LIMIT"; fi) \
    $(if [ -n "$NEWS_SUMMARIZER" ]; then echo -n " --summarizer \"$NEWS_SUMMARIZER\""; fi) \
    $(if [ -n "$NEWS_HTTP_PORT" ]; then echo -n " --http-port $NEWS_HTTP_PORT"; fi) \
    $(if [ -n "$NEWS_HTML_FILE" ]; then echo -n " --html-file \"$NEWS_HTML_FILE\""; fi)

