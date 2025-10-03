## What is the target Go version for this CLI application?
## Which LLM service or library will be used for story extraction?
## How should the program authenticate with the chosen LLM service?
## What is the expected input format of the Maildir emails (e.g., plain text, HTML, MIME parts)?
## Which email fields are required for processing (e.g., Subject, From, Body)?
## How many stories can be extracted from a single email, and how should multiple stories be handled?
## What is the desired format for the extracted story metadata (headline, teaser, URL) – JSON, YAML, plain text?
## Where should the story output directory be created if it does not exist, and what permissions are needed?
## Should the program support incremental processing (skip already processed emails) and how to track state?
## How will errors be logged or reported to the user (stderr, log file, structured logs)?
## Do we need a configuration file (e.g., TOML, YAML) for LLM settings, directories, and other options?
## What command‑line flags are required (e.g., --maildir, --outdir, --config, --verbose)?
## Should the CLI provide a sub‑command for testing the LLM prompt on a sample email?
## Are there any performance or concurrency requirements (e.g., parallel email processing)?
## How should the program handle rate‑limits or throttling from the LLM API?
## What is the expected behavior when an email cannot be parsed or the LLM fails to extract a story?
## Should the program support dry‑run mode to preview extracted stories without writing files?
## Are there any security considerations for handling potentially malicious email content?
## What licensing or attribution is required for the generated stories and the tool itself?
