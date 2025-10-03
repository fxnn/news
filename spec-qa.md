## Issue: Incremental Processing May Miss Emails Due to Filename Sanitization
**Mitigation:**  
- Normalise `Message-ID` by stripping angle brackets and replacing any filesystem‑unsafe characters (e.g., `/`, `\`, `:`) with safe alternatives (e.g., `_`).  
- Document the sanitisation rules and ensure they are applied consistently when generating story filenames.

## Issue: Date Formatting for Filenames Can Produce Invalid Characters
**Mitigation:**  
- Use a strict ISO‑8601 format without colons (e.g., `20060102T150405Z`) when constructing the `<date>` part of the filename.  

## Issue: Missing Story Directory Handling
**Mitigation:**  
- Although the spec says “Expect the directory to exist”, the CLI should verify the directory’s existence and exit with a clear error if it does not.  

## Issue: LLM Client Errors Not Retried
**Mitigation:**  
- Implement exponential back‑off retry logic for transient HTTP errors (e.g., 429, 5xx).  
- After a configurable number of retries, log a `WARN` and continue to the next email.

## Issue: No Validation of LLM Response Structure
**Mitigation:**  
- Validate that each story object returned by the LLM contains `headline`, `teaser`, and `url`.  
- If validation fails, log an `INFO` and retry.
- After a configurable number of retries, log a `WARN` and continue to the next email.

## Issue: Potentially Large Email Bodies Cause Memory Pressure
**Mitigation:**  
- Stream email bodies when possible, especially for large multipart messages.  
- Impose a configurable maximum body size; truncate with a warning if exceeded.

## Issue: Logging Verbosity Not Fully Controlled
**Mitigation:**  
- Use a logging library that respects log levels (e.g., `zerolog`).  
- Ensure that `--verbose` switches the logger to `DEBUG` and that all other logs default to `INFO`/`WARN`/`ERROR`.

## Issue: `--limit` Does Not Account for Skipped Emails
**Mitigation:**  
- Clarify that `--limit` counts *processed* emails (i.e., after skipping already‑processed ones).  
- Implement the limit after the incremental check so that the user gets the expected number of new stories.

## Issue: No Unit Tests for Core Components
**Mitigation:**  
- Add test files for the Maildir scanner, email parser, incremental checker, and story builder.  
- Use a mock LLM server (`httptest.Server`) to provide deterministic responses for integration tests.

## Issue: Configuration File Errors Not Gracefully Reported
**Mitigation:**  
- Validate required fields in the TOML config (e.g., `llm.provider`, `api_key`, `model`).  
- On missing or malformed entries, print a concise error to `stderr` and exit with a non‑zero status.

## Issue: Output to Stdout Lacks JSON Array Wrapper
**Mitigation:**  
- When `--storydir` is omitted, emit newline‑delimited JSON (NDJSON) with a clear documentation note.  
- Ensure the output format is machine‑parseable for downstream pipelines.

## Issue: Lack of Timeout for LLM Requests
**Mitigation:**  
- Set a reasonable HTTP client timeout (e.g., 30 seconds) to avoid hanging on slow LLM responses.  
- Log a `WARN` if a request times out and continue processing.

## Issue: No Support for Alternative LLM Providers
**Mitigation:**  
- Define an interface for the LLM client and provide a factory that selects the implementation based on `llm.provider`.  
- This makes future extensions (e.g., Anthropic, Cohere) straightforward.

## Issue: Missing Documentation for Environment Variables (e.g., API keys)
**Mitigation:**  
- Document that `api_key` can be supplied via the TOML file or an environment variable (e.g., `LLM_API_KEY`).  
- Prefer environment variables for secret management and log a warning if both are set.

## Issue: No Clear Exit Codes for Different Failure Modes
**Mitigation:**  
- Define exit codes: `0` = success, `1` = CLI argument/config error, `2` = runtime error (e.g., I/O), `3` = LLM failure.  
- Use these codes consistently throughout the program.
