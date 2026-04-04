# Phase 01: Cross-Platform Wallpaper Setting - Validation Strategy

**Phase:** 01-cross-platform-wallpaper-setting
**Created:** 2026-04-04
**Nyquist Version:** 1.0

---

## Validation Dimensions

### Dimension 1: Completeness
**Target:** All 3 platforms supported (macOS, Linux, Windows)

**Validation Methods:**
- Unit tests for each platform backend
- Platform detection coverage
- Error path coverage

**Pass Criteria:**
- [ ] Platform detection works on macOS, Linux, Windows
- [ ] Each OS has a working wallpaper backend
- [ ] Linux DE detection covers GNOME, KDE, XFCE
- [ ] Fallback mechanisms in place for unsupported Linux DEs

---

### Dimension 2: Correctness
**Target:** Commands work on real systems

**Validation Methods:**
- Integration tests with mocked system calls
- Manual verification on target platforms
- CI matrix testing (GitHub Actions with multiple OS)

**Pass Criteria:**
- [ ] macOS AppleScript command format is correct
- [ ] Linux gsettings command format is correct
- [ ] Windows PowerShell command format is correct
- [ ] File path handling escapes special characters properly

---

### Dimension 3: Edge Cases
**Target:** Handles errors gracefully

**Validation Methods:**
- Error injection tests
- Invalid input handling tests
- Missing dependency tests

**Pass Criteria:**
- [ ] Non-existent file returns clear error
- [ ] Non-image file returns clear error
- [ ] Unsupported platform returns clear error
- [ ] Missing system commands (gsettings, osascript) handled gracefully

---

### Dimension 4: Integration
**Target:** CLI command works end-to-end

**Validation Methods:**
- End-to-end CLI tests
- Flag combination tests
- Config persistence tests

**Pass Criteria:**
- [ ] `set <path>` works on all platforms
- [ ] `set --random` selects and sets random wallpaper
- [ ] `set --latest` selects and sets most recent wallpaper
- [ ] `set --current` displays current wallpaper
- [ ] Config stores current wallpaper and history

---

### Dimension 5: Observability
**Target:** Actions are logged and traceable

**Validation Methods:**
- Log output verification
- Config file inspection

**Pass Criteria:**
- [ ] Successful sets are logged
- [ ] Failed sets are logged with error details
- [ ] Current wallpaper is persisted to config
- [ ] History is updated with each set operation

---

### Dimension 6: Operations
**Target:** Clear error messages and troubleshooting

**Validation Methods:**
- Error message review
- Exit code verification
- Documentation review

**Pass Criteria:**
- [ ] All errors have actionable messages
- [ ] Exit codes follow conventions (0=success, non-zero=failure)
- [ ] README or help text documents platform requirements

---

## Validation Checkpoints

### Checkpoint 1: Platform Detection
**When:** After platform detection utility complete
**Verify:**
```bash
go test ./internal/platform -run TestDetect -v
```
**Pass:** Detection returns correct OS and DE on target systems

### Checkpoint 2: Platform Backends
**When:** After each platform backend complete
**Verify:**
```bash
# Per platform
go test ./internal/platform -run TestSetWallpaper -v
```
**Pass:** Each backend generates correct system commands

### Checkpoint 3: CLI Integration
**When:** After set command implementation
**Verify:**
```bash
go build -o wallpaper-cli
./wallpaper-cli set --help
./wallpaper-cli set --current  # Before any set
./wallpaper-cli set <test-image>
./wallpaper-cli set --current  # After set
```
**Pass:** Help shows flags, current shows path, set changes wallpaper

### Checkpoint 4: Random/Latest
**When:** After T06 complete
**Verify:**
```bash
./wallpaper-cli set --random
./wallpaper-cli set --latest
```
**Pass:** Random selects different image, latest selects most recent

### Checkpoint 5: Cross-Platform Testing
**When:** After all implementation complete
**Verify:** Manual testing matrix

| Platform | Variant | Status |
|----------|---------|--------|
| macOS | Intel | [ ] |
| macOS | Apple Silicon | [ ] |
| Linux | Ubuntu GNOME | [ ] |
| Linux | KDE Plasma | [ ] |
| Linux | XFCE | [ ] |
| Windows | Windows 10 | [ ] |
| Windows | Windows 11 | [ ] |

---

## Automated Tests

### Unit Tests Required

```go
// internal/platform/detect_test.go
func TestDetectPlatform(t *testing.T)
func TestDetectLinuxDE(t *testing.T)

// internal/platform/macos_test.go
func TestMacOSSetWallpaper(t *testing.T)
func TestMacOSSetWallpaperInvalidPath(t *testing.T)

// internal/platform/linux_test.go
func TestLinuxSetWallpaperGNOME(t *testing.T)
func TestLinuxSetWallpaperKDE(t *testing.T)
func TestLinuxSetWallpaperXFCE(t *testing.T)
func TestLinuxSetWallpaperFallback(t *testing.T)

// internal/platform/windows_test.go
func TestWindowsSetWallpaper(t *testing.T)
func TestWindowsSetWallpaperPathEscaping(t *testing.T)

// cmd/set_test.go
func TestSetCommand(t *testing.T)
func TestSetCommandRandom(t *testing.T)
func TestSetCommandLatest(t *testing.T)
func TestSetCommandCurrent(t *testing.T)
```

### Integration Tests Required

```go
// test/integration/set_test.go
func TestSetCommandE2E(t *testing.T)  // Full CLI execution
```

---

## Verification Commands

### Build Verification
```bash
# Build all platforms
go build -o wallpaper-cli
go build -o wallpaper-cli.exe GOOS=windows
go build -o wallpaper-cli-linux GOOS=linux
```

### Test Verification
```bash
# Run all tests
go test ./...

# Run platform tests only
go test ./internal/platform/...

# Run with coverage
go test -cover ./internal/platform/...
```

### Manual Verification
```bash
# Help displays
./wallpaper-cli set --help

# Set specific wallpaper
./wallpaper-cli set ~/Pictures/wallpapers/test.jpg

# Verify config updated
cat ~/.config/wallpaper-cli/config.json | grep current_wallpaper

# Random and latest
./wallpaper-cli set --random
./wallpaper-cli set --latest

# Check history
./wallpaper-cli set --current
```

---

## Success Criteria

This phase is complete when:

1. ✓ All 3 platforms have working wallpaper backends
2. ✓ Platform detection works correctly
3. ✓ CLI `set` command is functional with all flags
4. ✓ Config persists current wallpaper and history
5. ✓ Tests cover all platform backends
6. ✓ Manual verification passes on target platforms

---

*Validation strategy derived from RESEARCH.md Section 12*
