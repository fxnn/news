# TODO-List for Task 3
## Objective
Load TOML configuration and expose a `Config` struct with validation.

## Steps
### Step 1
- [ ] write failing test for loading a valid config file
- [ ] run test and confirm it fails
- [ ] implement `Config` structs and loading logic
- [ ] run test again and confirm it passes

### Step 2
- [ ] write failing test for missing required fields (e.g., `llm.provider`)
- [ ] run test and confirm it fails
- [ ] add validation for required fields and proper error handling
- [ ] run test again and confirm it passes

### Step 3
- [ ] write failing test for environment variable override of `api_key`
- [ ] run test and confirm it fails
- [ ] implement env‑var fallback logic
- [ ] run test again and confirm it passes

### Step 4
- [ ] write failing test for default values (e.g., timeout, temperature)
- [ ] run test and confirm it fails
- [ ] add default handling in the config loader
- [ ] run test again and confirm it succeeds

### Step 5
- [ ] commit changes and verify git log commit‑message schema
- [ ] run `git log` to inspect the message format
- [ ] amend the commit with an appropriate message detailing the Config loader implementation

## Outcome
A robust configuration loader with tests covering valid loading, validation errors, environment overrides, and defaults, ready for integration.
