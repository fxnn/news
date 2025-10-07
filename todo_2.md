# TODO-List for Task 2
## Objective
Implement CLI flag parsing and structured logging for the email-story-extractor.

## Steps
### Step 1
- [ ] write failing test for flag parsing (e.g., missing required flags cause exit code 1)
- [ ] run test and check it fails
- [ ] implement flag parsing using flag or cobra
- [ ] run test again and check it succeeds

### Step 2
- [ ] write failing test for logger default level INFO
- [ ] run test and check it fails
- [ ] implement logger initialization with zerolog, default INFO
- [ ] run test again and check it succeeds

### Step 3
- [ ] write failing test for --verbose flag setting DEBUG level
- [ ] run test and check it fails
- [ ] modify logger to respect verbose flag
- [ ] run test again and check it succeeds

## Outcome
CLI parses required flags, optional flags, exits with proper codes on errors, and logs at INFO or DEBUG based on --verbose.
