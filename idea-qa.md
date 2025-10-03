## What is the expected input format of the Maildir emails (e.g., plain text, HTML, MIME parts)?
Any kind of e-mail shall be supported, particularly plaintext and HTML e-mails, including such with MIME.

## Which email fields are required for processing (e.g., Subject, From, Body)?
The output for each story must include the `From` and `Date` fields.
The LLM analysis shall be based on the `Subject` and `Body`.

## How many stories can be extracted from a single email, and how should multiple stories be handled?
No limits on the number of stories. It's expected that an e-mail contains one or multiple stories.

## What is the desired format for the extracted story metadata (headline, teaser, URL) – JSON, YAML, plain text?
JSON

## Which information shall be stored per story?
- headline
- teaser
- URL
- e-mail address (from the `From` field)
- sender name (from the `From` field)
- timestamp (from the `Date` field)

## Where should the story output directory be created if it does not exist, and what permissions are needed?
Expect the directory to exist

## Should the program support incremental processing (skip already processed emails) and how to track state?
It must support incremental processing.
It should be able to check for a given e-mail the contents of the story directory to find out whether
stories for this e-mail already exist or not.
For this, the names of the story files shall be prefixed with the date and the message ID.

## How will errors be logged or reported to the user (stderr, log file, structured logs)?
stderr, logfmt

## Do we need a configuration file (e.g., TOML, YAML) for LLM settings, directories, and other options?
Yes, TOML

## What command‑line flags are required (e.g., --maildir, --outdir, --config, --verbose)?
Yes, at least --maildir, --storydir, --config, --verbose, --help

## Should the CLI provide a sub‑command for testing the LLM prompt on a sample email?
Add a --limit parameter to limit the number of processed e-mails and make --storydir optional,
so that stories are just written to stdout

## Are there any performance or concurrency requirements (e.g., parallel email processing)?
None so far

## How should the program handle rate‑limits or throttling from the LLM API?
Shouldn't right now

## What is the expected behavior when an email cannot be parsed or the LLM fails to extract a story?
Log a warning and continue

## Are there any security considerations for handling potentially malicious email content?
No

## What licensing or attribution is required for the generated stories and the tool itself?
Irrelevant for now

