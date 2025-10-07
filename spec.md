# Specification for Email Story Extraction CLI

## Overview
A Go command‑line application that reads emails from a Maildir, extracts one or more story objects from each email using a configurable Large Language Model (LLM), and outputs the story metadata as JSON. The tool supports incremental processing, configurable via a TOML file, and provides detailed structured logging to stderr.

## Functional Requirements

1. **Email Ingestion**
   - Accept a Maildir path (`--maildir`) and recursively read all files.
   - Support plain‑text, HTML, and multipart MIME messages.
   - Parse the `Subject`, `From`, `Date`, `Message‑ID`, and body content.

2. **Story Extraction**
   - For each email, send the `Subject` and body to the configured LLM.
   - Receive zero or more story objects, each containing:
     - `headline` (string)
     - `teaser` (string)
     - `url` (string)
   - Enrich each story with email metadata:
     - `from_email` (email address from `From`)
     - `from_name` (display name from `From`, if present)
     - `date` (timestamp from `Date`, ISO‑8601)

3. **Output**
   - When `--storydir` is provided, write each story as an individual JSON file.
   - File name format: `<date>_<message-id>_<index>.json` where `index` starts at `1`.
   - When `--storydir` is omitted, emit newline‑delimited JSON (NDJSON), one object per line.

4. **Incremental Processing**
   - Before processing an email, check the story directory for any file whose name starts with the sanitized `<date>_<message-id>` prefix. The `<date>` uses a strict ISO‑8601 format without colons (e.g., `20060102T150405Z`), and the `Message‑ID` is normalized by stripping angle brackets and replacing filesystem‑unsafe characters with underscores.
   - If such files exist, skip the email (already processed).

5. **Configuration**
   - Load LLM settings, API keys, model name, temperature, etc. from a TOML file supplied via `--config`.
   - Example TOML structure:
     ```toml
     [llm]
     provider = "openai"
     api_key = "YOUR_API_KEY"
     model = "gpt-4o-mini"
     temperature = 0.7
     ```

6. **CLI Flags**
   - `--maildir` (required): path to the Maildir.
   - `--storydir` (optional): directory for story JSON files.
   - `--config` (required): path to the TOML configuration file.
   - `--limit` (optional): maximum number of emails to process.
   - `--limit` counts only emails that are actually processed (i.e., after skipping already‑processed ones).
   - `--verbose` (optional): enable DEBUG‑level logs.
   - `--help`: display usage information.

   ## Exit Codes

   - `0` = success
   - `1` = CLI argument/config error
   - `2` = runtime error (e.g., I/O)
   - `3` = LLM failure

7. **Logging**
   - Emit logs to **stderr** in logfmt format (`key=value` pairs) using a structured logger such as `zerolog`.
   - Levels: `INFO`, `WARN`, `ERROR`, `DEBUG` (when `--verbose`).
   - Log key events: start/end of processing, email parsing failures, LLM request/response, file writes, skipped emails.

8. **Error Handling**
   - Unparsable emails → log `WARN` with email path, continue.
   - LLM request failures → log `WARN` with error details, continue.
   - Fatal configuration or flag errors → print error to stderr and exit with a defined non‑zero status (see Exit Codes).

## Non‑Functional Requirements

- **Performance**: Process emails sequentially; no concurrency required.
- **Timeouts**: HTTP client requests (e.g., to the LLM) use a configurable timeout (default 30 seconds) to avoid hangs.
- **Body Streaming**: Large email bodies are streamed and truncated to a configurable maximum size, with a warning logged if truncation occurs.
- **Reliability**: Must not crash on malformed emails or LLM errors.
- **Portability**: Buildable on Linux, macOS, and Windows (pure Go).
- **Security**: No special handling; treat email content as untrusted but do not execute it.
- **Licensing**: No external licensing constraints for the tool itself.

## Architecture Overview

```
+-------------------+      +-------------------+      +-------------------+
| CLI & Flag Parser | ---> | Config Loader     | ---> | LLM Client        |
+-------------------+      +-------------------+      +-------------------+
          |                         |                         |
          v                         v                         v
+-------------------+      +-------------------+      +-------------------+
| Maildir Scanner   | ---> | Email Parser      | ---> | Story Builder     |
+-------------------+      +-------------------+      +-------------------+
          |                         |                         |
          v                         v                         v
+---------------------------------------------------------------+
| Incremental Checker (storydir)                               |
+---------------------------------------------------------------+
          |
          v
+-------------------+
| Output Writer     |
| (JSON files or   |
|  stdout)          |
+-------------------+
```

- **LLM Client Interface**: Defined as an interface with implementations per provider (e.g., OpenAI). Allows future extensions to other providers.
- **CLI & Flag Parser**: Uses `flag` or `cobra` to parse command‑line arguments.
- **Config Loader**: Reads TOML via `BurntSushi/toml` (or `pelletier/go-toml`).
- **Maildir Scanner**: Walks the directory tree, yields file paths.
- **Email Parser**: Uses `net/mail` and `mime/multipart` to extract headers and body (plain text fallback if HTML only).
- **LLM Client**: Abstract interface; concrete implementation for the chosen provider (e.g., OpenAI HTTP API). Sends a prompt containing `Subject` and body, receives JSON‑encoded story list.
- **Story Builder**: Merges LLM output with email metadata, validates required fields.
- **Incremental Checker**: Scans `--storydir` for existing files matching `<date>_<message-id>_*`.
- **Output Writer**: Serializes each story to JSON (`encoding/json`) and writes to file or stdout.

## Data Models

```go
type Story struct {
    Headline   string `json:"headline"`
    Teaser     string `json:"teaser"`
    URL        string `json:"url"`
    FromEmail  string `json:"from_email"`
    FromName   string `json:"from_name,omitempty"`
    Date       string `json:"date"` // ISO‑8601
}
```

- LLM response must be a JSON array of objects matching the first three fields; the tool adds the remaining fields.

## Configuration File (`config.toml`)

```toml
[llm]
provider = "openai"          # future providers can be added
api_key = "YOUR_API_KEY"
# Can also be supplied via the environment variable `LLM_API_KEY`. If both are set, the environment variable takes precedence.
model = "gpt-4o-mini"
temperature = 0.7
prompt_template = """
You are given an email with Subject and Body.
Extract each story as a JSON object with fields:
headline, teaser, url.
Return a JSON array of story objects.
"""
```

## Command‑Line Interface

```
email-story-extractor \
  --maildir /path/to/Maildir \
  --storydir /path/to/stories \
  --config config.toml \
  [--limit N] \
  [--verbose] \
  [--help]
```

- `--limit` stops after processing N emails (useful for tests).
- `--verbose` sets log level to DEBUG.

## Testing Plan

1. **Unit Tests**
   - **Maildir Scanner**: Verify recursive file discovery, respect of `--limit`.
   - **Email Parser**: Test plain‑text, HTML‑only, multipart with attachments, malformed headers.
   - **Incremental Checker**: Ensure detection of existing story files based on date and Message‑ID.
   - **Story Builder**: Validate merging of LLM output with email metadata; reject missing required fields.
   - **LLM Client (Mock)**: Provide deterministic responses for given prompts.

2. **Integration Tests**
   - End‑to‑end run with a temporary Maildir containing a few crafted emails and a mock LLM server.
   - Verify JSON files are created with correct naming and content.
   - Test `--storydir` omitted → stdout contains valid JSON lines.
   - Test `--limit` truncates processing.
   - Test `--verbose` produces DEBUG logs.

3. **Error‑Handling Tests**
   - Corrupt email file → WARN log, continue.
   - LLM returns HTTP error → WARN log, continue.
   - Missing required CLI flags → program exits with non‑zero status and prints usage.

4. **Performance / Regression**
   - Process a Maildir with 1000 emails; ensure no panics and reasonable runtime (sequential processing).

All tests should be runnable with `go test ./...` and use the standard Go testing framework. Mocking can be done with `httptest.Server` for the LLM endpoint.

## Acceptance Criteria (derived from requirements)

- **AC1**: Valid Maildir → one JSON file per extracted story, named `<date>_<message-id>_<index>.json`.
- **AC2**: Without `--storydir`, stories are printed to stdout as JSON and program exits with status 0.
- **AC3**: `--limit N` processes exactly N emails (or fewer if not enough).
- **AC4**: Re‑running on the same Maildir does not duplicate story files.
- **AC5**: Unparsable emails generate a WARN log entry but do not abort the run.
- **AC6**: Missing required flags cause an error message and non‑zero exit status.
- **AC7**: All logs are emitted to stderr in logfmt format; `--verbose` enables DEBUG logs.

## Dependencies (Go)

- `golang.org/x/net/html` – HTML parsing fallback.
- `github.com/BurntSushi/toml` – TOML configuration.
- `github.com/rs/zerolog` or a simple logfmt writer for structured logs.
- `net/mail` – RFC‑5322 email parsing.
- `mime/multipart` – MIME handling.
- `encoding/json` – JSON serialization.
- HTTP client (standard library) for LLM API calls.

The name of the project itself shall be `github.com/fxnn/news`.

## Build & Run

```bash
go build -o email-story-extractor ./cmd
./email-story-extractor --maildir ./maildir --storydir ./stories --config ./config.toml
```

The binary should be portable across supported OSes without external C dependencies.
