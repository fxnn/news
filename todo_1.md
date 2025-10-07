# TODO-List for Task 1
## Objective
Set up a minimal Go module, directory layout, and CI pipeline to enable building and testing the project.

## Steps
### Step 1: Initialize Go module
- [x] write failing test that checks `go.mod` exists or that `go list ./...` succeeds
- [ ] run test and confirm it fails because the module is not initialized
- [ ] implement by running `go mod init example.com/email-story-extractor` (choose appropriate module path)
- [ ] run test again and confirm it succeeds

### Step 2: Create main package scaffold
- [ ] write failing test that verifies `cmd/email-story-extractor/main.go` builds
- [ ] run test and confirm it fails due to the missing file
- [ ] implement by creating `cmd/email-story-extractor/main.go` with a placeholder `package main` and a `func main(){}` that prints “TODO”
- [ ] run test again and confirm it succeeds

### Step 3: Add CI workflow
- [ ] write failing test that checks the file `.github/workflows/go.yml` exists
- [ ] run test and confirm it fails because the workflow file is missing
- [ ] implement by adding the minimal GitHub Actions workflow (the file is now present)
- [ ] run test again and confirm it succeeds

### Step 4: Verify build passes locally
- [ ] write failing test that runs `go build ./cmd/email-story-extractor` and expects success
- [ ] run test and confirm it fails before the implementation is complete
- [ ] implement any missing build steps (e.g., ensure `main.go` compiles)
- [ ] run test again and confirm it succeeds

## Outcome
- `go.mod` is initialized with the correct module path.
- Directory `cmd/email-story-extractor` contains a placeholder `main.go`.
- CI workflow `.github/workflows/go.yml` is present and functional.
- All tests for Task 1 pass, confirming the scaffold is operational.
