# TODO-List for Task 1
## Objective
Set up a minimal Go module, directory layout, and CI pipeline to enable building and testing the project.

## Steps
### Step 1: Initialize Go module
- [x] write failing test that checks `go.mod` exists or that `go list ./...` succeeds
- [x] run test and confirm it fails because the module is not initialized
- [x] implement by running `go mod init github.com/fxnn/news` (choose appropriate module path)
- [x] run test again and confirm it succeeds

### Step 2: Create main package scaffold
- [x] write failing test that verifies `cmd/email-story-extractor/main.go` builds
- [x] run test and confirm it fails due to the missing file
- [x] implement by creating `cmd/email-story-extractor/main.go` with a placeholder `package main` and a `func main(){}` that prints “TODO”
- [x] run test again and confirm it succeeds

### Step 3: Add CI workflow
- [x] write failing test that checks the file `.github/workflows/go.yml` exists
- [x] run test and confirm it fails because the workflow file is missing
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
