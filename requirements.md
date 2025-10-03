# Requirements Document

## Overview
A command‑line tool written in Go that processes emails stored in a Maildir directory, extracts one or more stories from each email using a Large Language Model (LLM), and outputs the extracted story metadata (headline, teaser, URL) in JSON format.

## Functional Requirements

1. **Input Sources**
   - Accept a Maildir path (`--maildir`) containing downloaded email messages.
   - Support any email format: plain text, HTML, and multipart MIME.

2. **Story Extraction**
   - For each email, extract the `From` and `Date` fields for metadata.
   - Use the email `Subject` and body as input to the LLM prompt.
   - Allow extraction of multiple stories per email.

3. **Output**
   - Write each story as a separate JSON file in a story directory (`--storydir`).
   - File names must be prefixed with the email date and message‑ID to enable incremental processing.
   - When `--storydir` is omitted, write all story JSON objects to **stdout**.

4. **Incremental Processing**
   - Skip emails that already have corresponding story files in the output directory.
   - Detect existing story files by matching the date‑message‑ID prefix.

5. **Configuration**
   - Load LLM settings, API keys, and other options from a TOML configuration file (`--config`).

6. **CLI Flags**
   - `--maildir` (required): path to the Maildir directory.
   - `--storydir` (optional): directory for story JSON files; defaults to stdout if omitted.
   - `--config` (required): path to the TOML configuration file.
   - `--limit` (optional): maximum number of emails to process.
   - `--verbose` (optional): enable detailed log output.
   - `--help`: display usage information.

7. **Logging**
   - Log warnings and errors to **stderr** using logfmt format.
   - Continue processing subsequent emails after a warning.

8. **Error Handling**
   - If an email cannot be parsed, log a warning and skip it.
   - If the LLM fails to extract a story, log a warning and continue.

## Non‑Functional Requirements

- **Performance**: No specific concurrency requirements; processing can be sequential.
- **Security**: No special handling required for potentially malicious email content.
- **Portability**: Must run on macOS, Linux, and Windows with Go runtime.
- **Licensing**: No external licensing or attribution needed for the tool itself.

## Usage Scenarios

1. **Full processing**  
   Process all emails in a Maildir and write stories to a directory:  
   `mytool --maildir /path/to/maildir --storydir /path/to/stories --config config.toml`

2. **Limited run for testing**  
   Process only the first 5 emails and output to stdout:  
   `mytool --maildir /path/to/maildir --limit 5 --config config.toml`

3. **Verbose debugging**  
   Run with detailed logs to troubleshoot:  
   `mytool --maildir /path/to/maildir --storydir /path/to/stories --config config.toml --verbose`

## Acceptance Criteria

- **AC1**: The tool accepts the required CLI flags and validates their presence.
- **AC2**: Emails of any MIME type are parsed without crashing.
- **AC3**: For each email, at least one JSON story file is produced when stories are found.
- **AC4**: Story files are named `<date>_<message-id>_<index>.json` and placed in the specified directory.
- **AC5**: When the same email is processed again, no duplicate story files are created.
- **AC6**: Errors in email parsing or LLM calls are logged as warnings; processing continues.
- **AC7**: Providing `--limit` restricts the number of processed emails accordingly.
- **AC8**: Omitting `--storydir` results in all story JSON objects being printed to stdout.
- **AC9**: Configuration is correctly loaded from the supplied TOML file.
- **AC10**: Verbose mode outputs detailed logfmt entries to stderr.

## Command‑Line Interface Summary

```
mytool --maildir <path> [--storydir <path>] --config <file> [--limit N] [--verbose] [--help]
```

* `--maildir`   : Path to Maildir containing emails (required)  
* `--storydir`  : Directory for JSON story files (optional)  
* `--config`    : TOML configuration file for LLM settings (required)  
* `--limit`     : Maximum number of emails to process (optional)  
* `--verbose`   : Enable verbose logfmt logging (optional)  
* `--help`      : Show usage information  
