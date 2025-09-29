# Specification for Maildir Story Extraction CLI

## Overview
The **Maildir Story Extraction** tool is a command‑line application written in Go.  
It scans a Maildir directory, processes each email with a Large Language Model (LLM) to extract one or more *stories*, and writes the extracted stories to a designated output directory. Each story consists of:

| Field      | Description                                            |
|------------|--------------------------------------------------------|
| `headline` | A concise, human‑readable title for the story.        |
| `teaser`   | A short summary (1‑2 sentences) giving context.       |
| `url`      | A link that can be opened in a browser to read the full story. |

The tool is intended for developers to integrate into automated pipelines or personal workflows.

---

## Functional Requirements
1. **CLI Arguments**
   - `--maildir <path>` (required): Path to the Maildir directory containing email files.
   - `--outdir <path>` (required): Directory where extracted story files will be written.
   - `--config <path>` (optional): Path to a JSON/YAML configuration file for LLM settings.
   - `--prompt <path>` (optional): Path to a text file containing the prompt template used for story extraction.
   - `--concurrency <n>` (optional, default: 4): Number of emails processed in parallel.

2. **Email Processing**
   - Recursively walk the `new` and `cur` sub‑directories of the Maildir.
   - For each email file:
     - Parse the RFC‑822 headers and body (plain‑text or HTML).
     - Strip signatures and quoted replies where possible.
     - Send the cleaned body to the LLM with the configured prompt.

3. **Story Extraction**
   - The LLM must return a JSON array of story objects matching the schema above.
   - Validate each story (non‑empty headline, valid URL format, etc.).
   - Write each story to a separate file in `outdir` using the filename pattern:
     ```
     <timestamp>_<hash>.json
     ```
     where `<timestamp>` is the email’s `Date` header (or processing time) and `<hash>` is a SHA‑256 of the story content.

4. **Configuration**
   - Support at least one LLM provider (e.g., OpenAI, Anthropic) via a pluggable interface.
   - Configuration file defines:
     - Provider name
     - API key (environment variable fallback)
     - Model name
     - Temperature, max tokens, etc.
   - Prompt file may contain placeholders (`{{email_body}}`) that are replaced at runtime.

5. **Logging & Metrics**
   - Structured JSON logs to stdout (level: INFO, WARN, ERROR).
   - Emit metrics: number of emails processed, stories extracted, processing latency.

---

## Non‑Functional Requirements
- **Performance**: Process at least 10 emails per second with default concurrency.
- **Reliability**: Continue processing remaining emails if a single email or LLM call fails.
- **Security**: Do not write raw email contents to logs; mask API keys.
- **Portability**: Buildable on Linux, macOS, and Windows with Go 1.22+.

---

## Architecture Overview
```
+-------------------+      +-------------------+      +-------------------+
|   CLI Entrypoint  | ---> |   Email Walker    | ---> |   Worker Pool     |
+-------------------+      +-------------------+      +-------------------+
                                 |                         |
                                 v                         v
                        +-------------------+   +-------------------+
                        |   Email Parser    |   |   LLM Client      |
                        +-------------------+   +-------------------+
                                 |                         |
                                 v                         v
                        +-------------------+   +-------------------+
                        |   Prompt Builder  |   |   Response Parser |
                        +-------------------+   +-------------------+
                                 \_______________________/
                                              |
                                              v
                                    +-------------------+
                                    |   Story Writer    |
                                    +-------------------+
```

- **CLI Entrypoint**: Parses flags, loads configuration, starts processing.
- **Email Walker**: Walks the Maildir, streams file paths to a channel.
- **Worker Pool**: Fixed‑size pool of goroutines that read from the channel.
- **Email Parser**: Uses `net/mail` to extract headers/body, handles multipart.
- **Prompt Builder**: Inserts the cleaned body into the prompt template.
- **LLM Client**: Abstract interface; concrete implementation for chosen provider.
- **Response Parser**: Decodes JSON, validates schema.
- **Story Writer**: Serialises each story to a file, ensures atomic write.

---

## Data Flow Details
1. **Input**: Email file → `net/mail` → `Email` struct.
2. **Cleaning**: Strip signatures (`-- `) and quoted blocks (`>`) using heuristics.
3. **Prompt Generation**: `template.Execute` with `{{email_body}}`.
4. **LLM Call**: HTTP POST with JSON payload; timeout configurable (default 30s).
5. **Response**: JSON array → `[]Story` → validation.
6. **Output**: For each `Story`, write `outdir/<timestamp>_<hash>.json`.

---

## Error Handling Strategy
| Error Type                     | Handling Approach                                   |
|--------------------------------|-----------------------------------------------------|
| Invalid CLI args               | Print usage, exit with code 2.                      |
| Maildir not found / unreadable | Log error, exit with code 1.                        |
| Email parse failure            | Log WARN with file path, skip email.                |
| LLM request failure            | Retry up to 3 times with exponential backoff; if still failing, log ERROR and skip email. |
| Invalid LLM response           | Log WARN, count as “story extraction failure”.     |
| File write error               | Log ERROR, abort processing and exit with code 1.   |
| Unexpected panic                | Recover in worker goroutine, log stack trace, continue. |

All errors are emitted as structured JSON logs with fields: `timestamp`, `level`, `msg`, `error`, `file`, `email_id`.

---

## Testing Plan
### Unit Tests
- **Email Parser**: Verify correct extraction of headers/body for plain‑text, HTML, multipart emails.
- **Prompt Builder**: Ensure placeholders are replaced and special characters are escaped.
- **Response Parser**: Test valid and malformed JSON responses, schema validation.
- **Story Writer**: Confirm deterministic filename generation and atomic write.

### Integration Tests
- **End‑to‑End**: Use a temporary Maildir with a set of fixture emails, a mock LLM server returning predefined JSON, and verify that the correct story files are created.
- **Concurrency**: Run with varying `--concurrency` values and assert no race conditions (use `-race` flag).

### Mocking
- Provide an interface `LLMClient` with a `Call(prompt string) (string, error)` method.
- Implement a `MockLLMClient` for tests that returns canned responses.

### CI Pipeline
- Run `go test ./...` with `-race`.
- Lint with `golangci-lint run`.
- Build binary for Linux, macOS, Windows.

---

## Future Extensions (Optional)
- Support additional LLM providers via plugin system.
- Add a `--dry-run` mode that only logs extracted stories without writing files.
- Implement a small web UI to browse extracted stories.
- Provide a configuration schema validator (e.g., using `gojsonschema`).

---

*Prepared on `{{date}}` by the brainstorming team.*
