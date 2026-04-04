# S01: Project Foundation & CLI Scaffold

**Goal:** Project scaffolding with Go modules, CLI framework, and build pipeline

**Success Criteria:**
- go mod init and dependency management set up
- Cobra CLI framework integrated
- Basic project structure (cmd/, internal/, pkg/)
- Build scripts for local development
- CI scaffold for future testing

---

## Integration Closure

Project structure in place with working cobra CLI, build scripts, and test infrastructure

## Observability Impact

None - foundational slice

## Proof Level

L1 - Core foundations

---

## Tasks

| ID | Title | Est. | Input | Output | Files Touched |
|----|-------|------|-------|--------|---------------|
| T01 | Initialize Go module | 15m | wallpaper-cli-spec.md | go.mod, go.sum | go.mod |
| T02 | Set up Cobra CLI framework | 30m | go.mod | cmd/root.go, main.go | cmd/root.go, main.go |
| T03 | Create project structure | 20m | spec | cmd/, internal/, pkg/ directories | (directory structure) |
| T04 | Add build scripts | 20m | spec | Makefile, scripts/build.sh | Makefile, scripts/build.sh |
| T05 | Set up testing scaffold | 15m | spec | .github/workflows/ci.yml (optional) | .github/workflows/ci.yml |

---

## Dependencies

None - this is the foundational slice

---

## Risk

Low - standard Go project scaffolding

---

## Demo

Binary builds, CLI help works, --version reports correctly

---

## Plan

### T01: Initialize Go Module
**Estimate:** 15m

**Description:**
Initialize the Go module for the project, set up proper module path, and establish dependency management.

**Steps:**
1. Run `go mod init github.com/user/wallpaper-cli` (adjust path as needed)
2. Create initial go.mod with Go 1.21+ requirement
3. Add basic dependencies (cobra will be added in T02)

**Files Likely Touched:**
- go.mod

**Expected Output:**
- go.mod file initialized
- Module path defined

**Verification:**
```bash
go mod tidy
```

---

### T02: Set Up Cobra CLI Framework
**Estimate:** 30m

**Description:**
Integrate the Cobra CLI framework for command-line interface handling. Create the root command structure.

**Steps:**
1. Add cobra dependency: `go get github.com/spf13/cobra`
2. Create cmd/root.go with root command definition
3. Create main.go as entry point
4. Add version command (--version flag)
5. Add basic help text

**Files Likely Touched:**
- cmd/root.go
- main.go
- go.mod (updated)

**Expected Output:**
- CLI binary builds
- `wallpaper-cli --help` works
- `wallpaper-cli --version` works

**Verification:**
```bash
go build -o wallpaper-cli .
./wallpaper-cli --help
./wallpaper-cli --version
```

---

### T03: Create Project Structure
**Estimate:** 20m

**Description:**
Create the directory structure for the project following Go best practices and the architecture from SPEC.md.

**Steps:**
1. Create cmd/ for CLI commands
2. Create internal/ for internal packages
3. Create internal/sources/ for source adapters
4. Create internal/download/ for download manager
5. Create internal/dedup/ for deduplication
6. Create internal/filter/ for filtering
7. Create internal/organize/ for organization
8. Create internal/data/ for SQLite storage

**Files Likely Touched:**
- Directory structure only

**Expected Output:**
- Directory hierarchy in place
- .gitkeep files where needed

**Verification:**
```bash
find . -type d -name "internal" -o -type d -name "cmd" | head -10
tree -L 3  # or ls -R
```

---

### T04: Add Build Scripts
**Estimate:** 20m

**Description:**
Create build automation scripts for local development and future CI/CD.

**Steps:**
1. Create Makefile with common targets:
   - build: build the binary
   - test: run tests
   - clean: clean build artifacts
   - fmt: format code
   - vet: run go vet
2. Create scripts/build.sh for shell-based building
3. Add .gitignore for Go projects

**Files Likely Touched:**
- Makefile
- scripts/build.sh
- .gitignore

**Expected Output:**
- `make build` produces working binary
- `make test` runs tests (empty for now)
- .gitignore properly configured

**Verification:**
```bash
make build
./wallpaper-cli --version
make clean
```

---

### T05: Set Up Testing Scaffold
**Estimate:** 15m

**Description:**
Set up the testing infrastructure and optional CI configuration.

**Steps:**
1. Create internal/ packages with _test.go stubs
2. Set up test helpers/utilities if needed
3. Create .github/workflows/ci.yml (optional, for future GitHub Actions)

**Files Likely Touched:**
- internal/*/ *_test.go (stubs)
- .github/workflows/ci.yml

**Expected Output:**
- `go test ./...` runs without errors
- Test scaffolding in place

**Verification:**
```bash
go test ./...
```
