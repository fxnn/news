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
