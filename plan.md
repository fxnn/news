# Implementation Plan: Email Story Extraction CLI

## Overview
This document outlines a step‑by‑step blueprint for building the Email Story Extraction CLI described in `spec.md`. The application reads emails from a Maildir, extracts story objects via a configurable LLM, and writes the results as JSON files or NDJSON. The plan is organized into high‑level components, their interactions, and a series of incremental tasks that can be developed and tested in isolation.

## High‑Level Blueprint
| Component                | Responsibility |
|--------------------------|-----------------|
| **CLI & Flag Parser**    | Parse command‑line arguments, configure logging, and orchestrate the workflow. |
| **Config Loader**        | Load TOML configuration, expose LLM settings and timeouts. |
| **Maildir Scanner**      | Recursively walk the Maildir, produce a stream of file paths, respect `--limit`. |
| **Email Parser**         | Open each file, parse RFC‑5322 headers, extract plain‑text body (fallback from HTML), return a structured `Email` struct. |
| **Incremental Checker**  | Given `--storydir`, determine whether an email has already been processed by checking for existing story files with the same `<date>_<msgid>` prefix. |
| **LLM Client**           | Abstract interface (`Client`) with an OpenAI implementation; send prompt containing `Subject` and body, receive JSON array of story objects. |
| **Story Builder**        | Merge LLM output with email metadata, validate required fields, produce `Story` structs. |
| **Output Writer**        | Serialize each `Story` to JSON; write to individual files in `--storydir` or emit NDJSON to stdout. |
| **Logger**               | Structured `logfmt` logger (`zerolog`) writing to stderr; supports INFO, WARN, ERROR, DEBUG (via `--verbose`). |

**Data Flow**  
`CLI` → `Config Loader` → `Maildir Scanner` → (`Incremental Checker`?) → `Email Parser` → `LLM Client` → `Story Builder` → `Output Writer` → (files/stdout). Logging occurs at each stage.

## Incremental Development Tasks

### Task 1: Project Scaffold & Build System
**Objective**: Set up a minimal Go module, directory layout, and CI pipeline.
- Initialize `go.mod` with module name.
- Create `cmd/email-story-extractor/main.go` with a placeholder `main` that prints “TODO”.
- Add a basic GitHub Actions workflow (`.github/workflows/go.yml`) if not present.
- Write a simple test that `go test ./...` passes.

### Task 2: CLI Flag Parsing & Logging
**Objective**: Implement command‑line parsing and structured logging.
- Use `flag` (or `cobra`) to define required flags (`--maildir`, `--config`) and optional ones (`--storydir`, `--limit`, `--verbose`).
- Initialise `zerolog` logger; default level INFO, switch to DEBUG when `--verbose`.
- Log start‑up parameters at INFO level.
- Add a test that verifies flag parsing errors produce exit code 1.

### Task 3: Configuration Loader
**Objective**: Load TOML configuration and expose a `Config` struct.
- Define structs matching the TOML schema (`LLMConfig` etc.).
- Use `github.com/BurntSushi/toml` to decode the file.
- Validate required fields; on failure log WARN and exit with code 1.
- Unit test with a sample TOML file (both valid and missing fields).

### Task 4: Maildir Scanner
**Objective**: Recursively enumerate email files respecting `--limit`.
- Implement `Scanner` that walks the directory tree (`filepath.WalkDir`).
- Yield file paths via a channel; stop when limit reached.
- Unit test with a temporary directory containing nested files.

### Task 5: Incremental Checker
**Objective**: Detect already‑processed emails.
- Implement helper to sanitize date (`YYYYMMDDTHHMMSSZ`) and Message‑ID.
- Scan `--storydir` for files matching the prefix; return `bool` indicating skip.
- Unit test with a temporary story directory containing matching and non‑matching files.

### Task 6: Email Parser
**Objective**: Parse RFC‑5322 headers and extract a plain‑text body.
- Use `net/mail` to read headers; parse `From` into name/email.
- For multipart messages, prefer `text/plain`; fallback to stripped HTML using `golang.org/x/net/html`.
- Return an `Email` struct with fields: `Subject`, `FromName`, `FromEmail`, `Date`, `MessageID`, `Body`.
- Unit tests covering plain‑text, HTML‑only, multipart with attachments, and malformed headers.

### Task 7: LLM Client Interface & OpenAI Implementation
**Objective**: Abstract LLM calls and provide a concrete OpenAI client.
- Define `type Client interface { ExtractStories(ctx context.Context, subject, body string) ([]Story, error) }`.
- Implement `OpenAIClient` using the HTTP API, respecting timeout from config.
- Serialize request with `prompt_template` from config; parse JSON response into slice of partial `Story` (headline, teaser, url).
- Add a mock client for tests.
- Unit test the client with an `httptest.Server` returning a known JSON payload.

### Task 8: Story Builder & Validation
**Objective**: Combine LLM output with email metadata and validate.
- For each partial story, fill `FromEmail`, `FromName`, `Date`.
- Ensure `headline`, `teaser`, `url` are non‑empty; log WARN and drop invalid entries.
- Return slice of complete `Story` structs.
- Unit test with various LLM responses (valid, missing fields, extra fields).

### Task 9: Output Writer
**Objective**: Serialize stories to JSON files or NDJSON.
- If `--storydir` is set, create files named `<date>_<msgid>_<index>.json` (sanitize characters).
- Write each story with `json.Encoder` (indent optional).
- If omitted, write each story as a line to stdout using `json.Marshal` + newline.
- Log each write at INFO level.
- Unit tests for file naming, content correctness, and NDJSON output.

### Task 10: End‑to‑End Integration
**Objective**: Wire all components together in `main.go`.
- Sequence: parse flags → load config → init logger → scanner → for each email:
  1. Incremental check → skip if needed.
  2. Parse email → on error log WARN, continue.
  3. Call LLM client → on error log WARN, continue.
  4. Build stories → filter invalid.
  5. Write output.
- Respect `--limit`.
- Log overall start/end and processing counts.
- Integration test using a temporary Maildir with a few crafted emails and a mock LLM server; verify files or stdout.

### Task 11: CI & Release Automation
**Objective**: Ensure tests run on CI and produce a binary.
- Update GitHub Actions to run `go test ./...` and `go build -o email-story-extractor ./cmd`.
- Add a `Makefile` target for `build`, `test`, and `run`.
- Tag a release (optional).

## Testing Strategy
- **Unit tests** for each component (scanner, parser, checker, builder, writer, client mock).
- **Table‑driven tests** for edge cases (missing headers, large bodies, truncation warnings).
- **Integration test** covering the full pipeline with a mock LLM.
- Use `go test -cover ./...` to ensure coverage.

## Future Extensions
- Add support for additional LLM providers via the `Client` interface.
- Implement concurrency for email processing (optional performance boost).
- Add configuration for body size limit and HTTP timeout.
- Provide a `--dry-run` mode that only logs actions without writing files.

--- 

This plan provides a clear, incremental path from an empty repository to a fully functional, tested CLI tool that satisfies all requirements in `spec.md`.
