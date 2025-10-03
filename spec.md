# Specification for Email Story Extraction CLI Tool

## Overview
A command‑line application written in Go that processes emails stored in a Maildir directory, extracts one or more stories from each email using a Large Language Model (LLM), and outputs the story metadata (headline, teaser, URL) as JSON. The tool supports incremental processing, configurable logging, and a TOML configuration file for LLM settings.

## Functional Requirements

1. **CLI Interface**
   - `--maildir <path>` (required): Path to the Maildir containing email messages.
   - `--storydir <path>` (optional): Directory where story JSON files are written. If omitted, all stories are printed to **stdout**.
   - `--config <file>` (required): Path to a TOML file containing LLM configuration and API keys.
   - `--limit N` (optional): Process at most *N* emails.
   - `--verbose` (optional): Enable detailed logfmt output to **stderr**.
   - `--help`: Show usage information.

2. **Email Ingestion**
   - Recursively scan the Maildir for all email files.
   - Support any MIME type: plain text, HTML, multipart, and attachments.
   - Parse each email to extract:
     - `From` header (string)
     - `Date` header (RFC‑2822 date)
     - `Message-ID` header (unique identifier)
     - `Subject` header (string)
     - Body content (plain‑text version; if only HTML is present, strip HTML tags).

3. **Story Extraction**
   - For each email, construct a prompt that includes the `Subject` and body.
   - Send the prompt to the configured LLM via its HTTP API.
   - The LLM returns a JSON array of story objects, each containing:
     - `headline` (string)
     - `teaser` (string)
     - `url` (string)
   - No limit on the number of stories per email.

4. **Output**
   - When `--storydir` is provided:
     - Ensure the directory exists (error if not).
     - For each story, write a file named `<date>_<message-id>_<index>.json` where:
       - `<date>` is `YYYYMMDD` derived from the email `Date`.
       - `<message-id>` is a sanitized version of the `Message-ID` header (non‑alphanumeric characters replaced with `_`).
       - `<index>` starts at `1` for the first story of that email.
   - When `--storydir` is omitted:
     - Serialize each story as a JSON object and write it to **stdout**, one per line.

5. **Incremental Processing**
   - Before processing an email, check the `storydir` for any file whose name starts with the same `<date>_<message-id>` prefix.
   - If such files exist, skip the email (already processed).

6. **Configuration (TOML)**
   - Must contain at least:
     ```toml
     [llm]
     endpoint = "https://api.example.com/v1/completions"
     api_key = "YOUR_API_KEY"
     model = "gpt-4"
     temperature = 0.7
     ```
   - Additional optional fields (e.g., timeout, max_tokens) may be added.

7. **Logging**
   - All logs are written to **stderr** in *logfmt* (`key=value` pairs).
   - Levels: `info`, `warn`, `error`.
   - In verbose mode, include detailed context (e.g., email path, LLM request/response sizes).

8. **Error Handling**
   - If an email cannot be parsed, log a `warn` and continue.
   - If the LLM request fails or returns malformed JSON, log a `warn` and continue.
   - Fatal errors (e.g., missing required CLI flags, unreadable config file) cause the program to exit with a non‑zero status.

## Non‑Functional Requirements

- **Portability**: Must compile and run on macOS, Linux, and Windows with the Go runtime.
- **Performance**: Sequential processing is acceptable; no concurrency requirements.
- **Security**: No special handling required; treat email content as untrusted but do not execute it.
- **Reliability**: Must not crash on malformed emails or network failures.

## Architecture Overview

```
+-------------------+      +-------------------+      +-------------------+
|   CLI (main.go)   | ---> |   Config Loader   | ---> |   LLM Client      |
+-------------------+      +-------------------+      +-------------------+
          |                         |                         |
          v                         v                         v
+-------------------+      +-------------------+      +-------------------+
| Maildir Scanner   | ---> | Email Parser      | ---> | Story Writer      |
+-------------------+      +-------------------+      +-------------------+
          |                         |                         |
          +-------------------------+-------------------------+
                                    |
                                    v
                           +-------------------+
                           |   Logger (logfmt) |
                           +-------------------+
```

- **CLI** parses flags, validates required arguments, and orchestrates the workflow.
- **Config Loader** reads the TOML file into a strongly‑typed struct.
- **Maildir Scanner** walks the Maildir hierarchy and yields file paths.
- **Email Parser** uses Go’s `net/mail` and `mime/multipart` packages to extract required fields and a plain‑text body.
- **LLM Client** builds the prompt, performs an HTTP POST with appropriate headers, and unmarshals the JSON response.
- **Story Writer** handles file naming, incremental checks, and JSON serialization.
- **Logger** provides a thin wrapper around `log` that formats entries as `key=value`.

## Data Flow

1. **Start** → Parse CLI flags.
2. **Load Config** → Validate required fields.
3. **Scan Maildir** → For each email file (respecting `--limit`):
   - Parse email → Extract metadata and body.
   - Check `storydir` for existing prefix → Skip if present.
   - Build LLM prompt → Call LLM client.
   - Receive story list → For each story:
     - Serialize to JSON.
     - Write to file or stdout.
   - Log progress and any warnings.

## Testing Plan

### Unit Tests
- **Config Loader**: Verify successful parsing of valid TOML and proper error on missing fields.
- **Maildir Scanner**: Mock a directory structure and ensure correct file enumeration and limit handling.
- **Email Parser**: Test with plain‑text, HTML‑only, and multipart emails; verify extraction of `From`, `Date`, `Message-ID`, `Subject`, and body.
- **LLM Client**: Use an HTTP test server to simulate successful and error responses; validate request payload and response handling.
- **Story Writer**: Test filename generation, prefix detection, and JSON output formatting.

### Integration Tests
- End‑to‑end test using a small fixture Maildir containing a few emails and a mock LLM server; assert that story files are created with correct names and contents, and that duplicate processing is skipped.

### CI Pipeline
- Run `go test ./...` on each push.
- Lint with `golangci-lint`.
- Verify that the binary builds for all supported OS/arch (`go build ./...`).

### Manual Test Scenarios
1. **Full run** with `--storydir` set – verify files on disk.
2. **Stdout mode** (no `--storydir`) – pipe output to `jq` and inspect JSON.
3. **Incremental** – run twice on the same Maildir and confirm no duplicate files.
4. **Error cases** – corrupt email, unreachable LLM endpoint, malformed config – ensure warnings are logged and processing continues where appropriate.

## Acceptance Criteria (derived from requirements)

- All CLI flags are validated; missing required flags cause exit with usage message.
- Emails of any MIME type are parsed without panic.
- At least one JSON story file (or stdout line) is produced per email that yields stories.
- Story filenames follow `<date>_<message-id>_<index>.json`.
- Re‑processing the same email does not create duplicate files.
- Parsing or LLM failures generate `warn` logs; processing continues.
- `--limit` correctly caps the number of processed emails.
- Omitting `--storydir` prints all story JSON objects to stdout.
- Configuration is loaded from the supplied TOML file and validated.
- Verbose mode adds detailed logfmt entries to stderr.

## Future Extensions (optional)

- Parallel email processing with worker pools.
- Rate‑limit handling for LLM API.
- Support for alternative output formats (YAML, CSV).
- Configurable output filename template.
