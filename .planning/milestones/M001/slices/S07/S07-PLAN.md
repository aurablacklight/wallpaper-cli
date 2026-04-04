# S07: Cross-Platform Builds & Optimization

**Goal:** Cross-platform build pipeline and resource efficiency verification

**Success Criteria:**
- GoReleaser configuration for multi-platform builds
- macOS (amd64, arm64), Linux (amd64, arm64), Windows (amd64)
- Binary size < 20MB for all targets
- Memory usage test < 10MB at idle
- Static linking (no CGO dependencies)
- Release workflow with artifact upload

---

## Integration Closure

Release-ready binaries for all target platforms

## Observability Impact

Binary size, memory usage verifiable via profiling

## Proof Level

L2 - Integration complexity

---

## Dependencies

- S06: Organization & Storage

---

## Risk

Medium - CGO dependencies (SQLite), platform-specific paths

---

## Demo

Cross-compiled binaries for macOS, Linux, Windows all under 20MB

---

## Tasks

| ID | Title | Est. | Input | Output | Files Touched |
|----|-------|------|-------|--------|---------------|
| T01 | Research CGO-free SQLite | 20m | Spec | Library selection | Research notes |
| T02 | Switch to CGO-free SQLite | 20m | T01 | Updated deps | go.mod, internal/data/db.go |
| T03 | Configure GoReleaser | 30m | Spec | .goreleaser.yml | .goreleaser.yml |
| T04 | Test cross-compilation | 25m | T03 | Test binaries | scripts/ |
| T05 | Verify binary sizes | 15m | T04 | Size report | Build output |
| T06 | Add memory profiling | 30m | T05 | Benchmark | internal/debug/profile.go |
| T07 | Create release workflow | 20m | T03 | GitHub Actions | .github/workflows/release.yml |

---

## Plan

### T01: Research CGO-free SQLite
**Estimate:** 20m

**Description:**
Research CGO-free SQLite options for static cross-compilation.

**Options:**
1. **modernc.org/sqlite** (Recommended)
   - Pure Go implementation of SQLite
   - No CGO required
   - Slower but acceptable for this use case
2. **sqlite with CGO** (current)
   - Requires cross-compilation toolchain
   - Complex for Windows builds
3. **Alternative DBs**
   - BoltDB (pure Go)
   - JSON file (simpler but slower)

**Decision:** Use modernc.org/sqlite for zero-CGO builds.

**Files Likely Touched:**
- Research notes only

**Expected Output:**
- Library selected and tested
- Migration path documented

---

### T02: Switch to CGO-free SQLite
**Estimate:** 20m

**Description:**
Replace mattn/go-sqlite3 with modernc.org/sqlite.

**Steps:**
1. Update go.mod:
   ```
   go get modernc.org/sqlite
   go mod tidy
   ```
2. Update imports in internal/data/db.go
3. Test locally to ensure compatibility
4. Verify all tests pass

**API Compatibility:**
- Both implement database/sql interface
- Should be drop-in replacement

**Files Likely Touched:**
- go.mod
- internal/data/db.go (imports only)

**Expected Output:**
- No CGO dependencies
- Cross-compilation works
- All functionality preserved

**Verification:**
```bash
go build -o wallpaper-cli .
./wallpaper-cli --version
CGO_ENABLED=0 go build -o wallpaper-cli-static .
# Should build without CGO
```

---

### T03: Configure GoReleaser
**Estimate:** 30m

**Description:**
Set up GoReleaser for automated multi-platform releases.

**Steps:**
1. Install GoReleaser locally (optional)
2. Create .goreleaser.yml:
   - Define builds for all targets
   - Set CGO_ENABLED=0
   - Add ldflags for version
   - Configure archives (tar.gz, zip)
   - Add checksums
3. Configure changelog generation
4. Add release notes template

**Build Matrix:**
| GOOS | GOARCH | Notes |
|------|--------|-------|
| darwin | amd64 | macOS Intel |
| darwin | arm64 | macOS Apple Silicon |
| linux | amd64 | Linux x86_64 |
| linux | arm64 | Linux ARM |
| windows | amd64 | Windows x64 |

**Files Likely Touched:**
- .goreleaser.yml

**Expected Output:**
- goreleaser build works locally
- All target binaries produced

**Verification:**
```bash
goreleaser build --snapshot --clean
ls -la dist/
# Should see binaries for all platforms
```

---

### T04: Test Cross-Compilation
**Estimate:** 25m

**Description:**
Test binaries on target platforms (manual or CI).

**Steps:**
1. Build for all targets
2. Test locally (matching platform)
3. Create test script for smoke tests:
   - --version works
   - --help works
   - config init works
4. Document any platform-specific issues

**Smoke Test:**
```bash
./wallpaper-cli --version
./wallpaper-cli --help
./wallpaper-cli config init
./wallpaper-cli config list
```

**Files Likely Touched:**
- scripts/test-release.sh

**Expected Output:**
- All binaries pass smoke test
- No platform-specific crashes

**Verification:**
```bash
# Test each binary
for binary in dist/*/wallpaper-cli*; do
    echo "Testing $binary"
    $binary --version
done
```

---

### T05: Verify Binary Sizes
**Estimate:** 15m

**Description:**
Verify all binaries are under 20MB target.

**Steps:**
1. Check sizes of all produced binaries
2. If over 20MB, investigate:
   - Strip symbols: -ldflags "-s -w"
   - UPX compression (optional)
   - Remove unnecessary dependencies
3. Document actual sizes

**Target:**
- All binaries < 20MB
- Ideally < 15MB with optimizations

**Files Likely Touched:**
- .goreleaser.yml (add ldflags)

**Expected Output:**
- Size report
- All targets under 20MB

**Verification:**
```bash
ls -lh dist/*/wallpaper-cli* | awk '{print $5, $9}'
# All sizes should be < 20M
```

---

### T06: Add Memory Profiling
**Estimate:** 30m

**Description:**
Add memory profiling to verify < 10MB at idle.

**Steps:**
1. Create internal/debug/profile.go (optional)
2. Add runtime.MemStats reporting
3. Create memory benchmark test
4. Run profile during idle state
5. Document baseline memory usage

**Memory Check:**
```go
var m runtime.MemStats
runtime.ReadMemStats(&m)
fmt.Printf("Alloc = %v MB", m.Alloc / 1024 / 1024)
```

**Target:**
- Idle memory < 10MB
- Peak during download < 50MB

**Files Likely Touched:**
- internal/debug/profile.go
- cmd/root.go (add debug flag)

**Expected Output:**
- Memory baseline established
- Profile data available

**Verification:**
```bash
./wallpaper-cli --debug-memory
# Shows memory usage and exits
# Or use runtime profiling during fetch
```

---

### T07: Create Release Workflow
**Estimate:** 20m

**Description:**
Create GitHub Actions workflow for automated releases.

**Steps:**
1. Create .github/workflows/release.yml
2. Trigger on tag push (v*)
3. Use goreleaser/goreleaser-action
4. Add artifact upload to GitHub releases
5. Add Homebrew tap (optional stretch)

**Workflow Steps:**
1. Checkout code
2. Setup Go
3. Run tests
4. Run GoReleaser
5. Upload artifacts

**Files Likely Touched:**
- .github/workflows/release.yml

**Expected Output:**
- Automated releases on tag
- All artifacts attached
- Release notes generated

**Verification:**
```bash
# Push a test tag
git tag -a v0.1.0-test -m "Test release"
git push origin v0.1.0-test
# Check GitHub Actions runs
```
