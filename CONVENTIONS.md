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

## Code Requirements

- Don't leave commented-out code in the files. No dead code.
- Extract code into functions often. A function should do one thing.
  If you need to separate different aspects using comments, split it up.
