# Requirements Document

## Overview
The tool processes emails stored in a Maildir, extracts one or more stories from each email using a Large Language Model (LLM), and outputs the extracted story metadata in JSON format. It supports incremental processing, configurable via a TOML file, and provides a command‑line interface (CLI) with appropriate flags.

## Functional Requirements

1. **Email Ingestion**
   - Accept a Maildir path (`--maildir`) and recursively read all email files.
   - Support plain‑text, HTML, and multipart MIME messages.

2. **Story Extraction**
   - For each email, send the `Subject` and `Body` to the configured LLM.
   - Receive one or more story objects, each containing:
     - `headline` (string)
     - `teaser` (string)
     - `url` (string)
   - Include the following email metadata in each story object:
     - `from_email` (email address from the `From` header)
     - `from_name` (display name from the `From` header, if present)
     - `date` (timestamp from the `Date` header, ISO‑8601)

3. **Output**
   - Write each story as a separate JSON file in the story directory (`--storydir`).
   - File name format: `<date>_<message-id>_<index>.json` where `index` starts at `1`.
   - If `--storydir` is omitted, write all story JSON objects to **stdout**.

4. **Incremental Processing**
   - Before processing an email, check the story directory for files prefixed with the email’s date and message‑ID.
   - Skip processing if matching story files already exist.

5. **Configuration**
   - Load LLM settings, API keys, and other options from a TOML file (`--config`).

6. **CLI Flags**
   - `--maildir` (required): path to the Maildir.
   - `--storydir` (optional): output directory for story files.
   - `--config` (required): path to the TOML configuration file.
   - `--limit` (optional): maximum number of emails to process.
   - `--verbose` (optional): enable detailed log output.
   - `--help`: display usage information.

7. **Logging**
   - Emit logs to **stderr** in logfmt format.
   - Log levels: INFO, WARN, ERROR.
   - Log a warning and continue when an email cannot be parsed or the LLM fails.

## Non‑Functional Requirements

- **Performance**: Process emails sequentially; no concurrency requirements.
- **Reliability**: Gracefully handle malformed emails and LLM errors without crashing.
- **Portability**: Buildable and runnable on Linux, macOS, and Windows.
- **Security**: No special handling required for malicious content.
- **Licensing**: No external licensing constraints for the tool itself.

## Usage Scenarios

| Scenario | Description |
|----------|-------------|
| **Basic processing** | User runs the CLI with `--maildir` and `--storydir` to extract stories from all emails. |
| **Stdout only** | User omits `--storydir`; stories are printed to stdout for piping or inspection. |
| **Limited run** | User supplies `--limit 10` to process only the first ten emails (useful for testing). |
| **Verbose mode** | User adds `--verbose` to see detailed processing logs. |
| **Incremental run** | After an initial run, the user re‑executes the CLI; already‑processed emails are skipped. |

## Acceptance Criteria

- **AC1**: Given a valid Maildir with mixed‑type emails, the tool creates a JSON file per extracted story with the correct naming convention.
- **AC2**: When `--storydir` is omitted, the tool outputs valid JSON to stdout and exits with status 0.
- **AC3**: Providing `--limit N` results in processing exactly *N* emails (or fewer if the Maildir contains fewer).
- **AC4**: Re‑running the tool on the same Maildir does not duplicate story files (incremental behavior).
- **AC5**: Invalid or unparsable emails generate a WARN log entry but do not abort the run.
- **AC6**: Missing required flags (`--maildir`, `--config`) cause the CLI to display an error message and exit with a non‑zero status.
- **AC7**: All logs are emitted to stderr in logfmt format and respect the `--verbose` flag.
