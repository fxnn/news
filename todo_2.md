# TODO-List for Task 2
## Objective
Implement CLI flag parsing and structured logging for the email-story-extractor.

## Steps
### Step 1
- [x] write failing test for flag parsing (e.g., missing required flags cause exit code 1)
- [x] run test and check it fails
- [x] implement flag parsing using flag or cobra
- [x] run test again and check it succeeds

### Step 2
- [x] write failing test for logger default level INFO
- [x] run test and check it fails
- [x] implement logger initialization with zerolog, default INFO
- [x] run test again and check it succeeds

### Step 3
- [x] write failing test for --verbose flag setting DEBUG level
- [x] run test and check it fails
- [x] modify logger to respect verbose flag
- [x] run test again and check it succeeds

## Outcome
CLI parses required flags, optional flags, exits with proper codes on errors, and logs at INFO or DEBUG based on --verbose.
