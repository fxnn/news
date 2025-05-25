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

run-server:
  @echo "Starting server with arguments from environment variables (e.g., FLAG_SERVER, FLAG_PORT, FLAG_USERNAME, FLAG_PASSWORD, etc.)..."
  go run . --mode server \
    $(if [ -n "$FLAG_SERVER" ]; then echo -n " --server \"$FLAG_SERVER\""; fi) \
    $(if [ -n "$FLAG_PORT" ]; then echo -n " --port $FLAG_PORT"; fi) \
    $(if [ -n "$FLAG_USERNAME" ]; then echo -n " --username \"$FLAG_USERNAME\""; fi) \
    $(if [ -n "$FLAG_PASSWORD" ]; then echo -n " --password \"$FLAG_PASSWORD\""; fi) \
    $(if [ -n "$FLAG_FOLDER" ]; then echo -n " --folder \"$FLAG_FOLDER\""; fi) \
    $(if [ -n "$FLAG_DAYS" ]; then echo -n " --days $FLAG_DAYS"; fi) \
    $(if [ -n "$FLAG_LIMIT" ]; then echo -n " --limit $FLAG_LIMIT"; fi) \
    $(if [ -n "$FLAG_SUMMARIZER" ]; then echo -n " --summarizer \"$FLAG_SUMMARIZER\""; fi) \
    $(if [ -n "$FLAG_HTTP_PORT" ]; then echo -n " --http-port $FLAG_HTTP_PORT"; fi) \
    $(if [ -n "$FLAG_HTML_FILE" ]; then echo -n " --html-file \"$FLAG_HTML_FILE\""; fi)

