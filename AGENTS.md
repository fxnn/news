# AI Coding Conventions

We implement a newsletter reading app in Golang.

## Coding Workflow

- We are writing the code like Kent Beck would write it.
- Before modifying code we consider whether tidying first would make the change
  easier.
- Commits will be separated into commits that change the behavior of the code
  and commits that only change the structure of the code.
- Write the code one test at a time. Write the test. Get it to compile. Get it
  to pass. Tidy after if appropriate.
- Always start a change with the tests. Add a new test case, or change an
  existing one.
- Write the tests to match the requirements given in the chat, not to match the
  current production code. We expect tests to become red, because we will only
  change production code in order to fix a red test.
- Only implement enough code to make the test you just wrote pass, along with
  all the previous tests.
- Under no circumstances should you erase or alter tests just to get a commit to
  pass. If there is a genuine bug in a test, fix the test.
- Always keep documentation (Readme, etc.) in sync with code changes, especially
  when modifying CLI usage, configuration, or setup steps.

## Testing Practices

- Ad-hoc tests using the Bash tool are acceptable for bootstrapping and build
  verification (does it compile? do files exist?).
- For testing code behavior, always write proper unit tests instead of ad-hoc
  manual tests.
- Examples of behavior that requires unit tests:
  - Does --help produce the correct output?
  - Does the program exit with the right error when required arguments are
    missing?
  - Does the parser correctly handle different input formats?
  - Does the function return the expected result for given inputs?
- Unit tests provide documentation, prevent regressions, and enable confident
  refactoring.

## Code Requirements

- Don't leave commented-out code in the files. No dead code.
- Extract code into functions often. A function should do one thing.
  If you need to separate different aspects using comments, split it up.
- Watch the placement of functions and types. Things that belong together should
  go into the same file. But one file should do **one** thing. If there is no
  common purpose within the file, split it up.
- Comment on the intention: why is that code there? What are purpose, context,
  hidden dependencies, reasons?
- Do not comment on the history. Do not comment "changed from struct to list" or
  "add for new b/w color scheme".
- **Separation of Concerns**: Maintain strict separation between configurations for
  different tools. Avoid shared packages (like `internal/cli`) if they couple
  independent tools.
- **Configuration Naming**: Use application-specific configuration filenames
  (e.g., `app-name.toml`) instead of generic names like `config.toml` to avoid
  ambiguity.
- **Environment Variables**: Use distinct, application-specific prefixes
  (e.g., `APP_NAME_`) for environment variables to prevent collisions.
- **TOML Syntax**: Ensure configuration files respect syntax rules (e.g., global
  keys must be defined *before* any table definitions).
- **Error Handling**: Always handle error results from functions. Never discard
  errors with `_`. If an error is not actionable, log it or return it to the
  caller.
