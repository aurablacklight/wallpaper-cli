# S01: Cross-Platform Wallpaper Setting

**Goal:** Implement the `wallpaper-cli set` command with platform-specific backends for macOS, Linux, and Windows.

**Success Criteria:**
- `wallpaper-cli set <path>` sets wallpaper on all 3 platforms
- `wallpaper-cli set --random` sets random wallpaper from collection
- `wallpaper-cli set --latest` sets most recent download
- Platform auto-detection works correctly
- Graceful error handling for unsupported configurations

---

## Integration Closure

Set command works end-to-end on macOS, Linux, and Windows with unified CLI interface.

## Observability Impact

Platform detection logged, wallpaper path stored in config, set operations logged.

## Proof Level

L2 - Platform-specific integration complexity

---

## Tasks

| ID | Title | Est. | Input | Output | Files Touched |
|----|-------|------|-------|--------|---------------|
| T01 | Platform detection utility | 30m | - | internal/platform/detect.go | detect.go |
| T02 | macOS wallpaper backend | 45m | detect.go | internal/platform/macos.go | macos.go |
| T03 | Linux wallpaper backend | 60m | detect.go | internal/platform/linux.go | linux.go |
| T04 | Windows wallpaper backend | 45m | detect.go | internal/platform/windows.go | windows.go |
| T05 | Set command implementation | 45m | All backends | cmd/set.go | set.go |
| T06 | Random/Latest set options | 30m | set.go | cmd/set.go enhancements | set.go |
| T07 | Cross-platform testing | 60m | All code | Test results | - |

---

## Dependencies

None - this is the foundational v1.2 slice

---

## Risk

High - Platform APIs vary significantly, Linux DE fragmentation

---

## Demo

```bash
./wallpaper-cli set ~/Pictures/wallpapers/wallhaven/01_abc.jpg
./wallpaper-cli set --random
./wallpaper-cli set --latest
```

---

## Plan

### T01: Platform Detection Utility
**Estimate:** 30m

**Description:**
Create utility to detect the current operating system and Linux desktop environment.

**Steps:**
1. Create `internal/platform/detect.go`
2. Implement `DetectPlatform() (OS, DE)`
3. OS enum: macOS, Linux, Windows
4. DE enum for Linux: GNOME, KDE, XFCE, Unknown
5. Use runtime.GOOS for OS detection
6. Use environment variables for DE detection ($XDG_CURRENT_DESKTOP)

**Files Likely Touched:**
- internal/platform/detect.go

**Expected Output:**
- Platform detection works on all 3 OSes
- Linux DE detection covers major variants

**Verification:**
```go
os, de := platform.Detect()
// macOS: OS=macOS, DE=""
// Ubuntu GNOME: OS=Linux, DE=GNOME
// Windows: OS=Windows, DE=""
```

---

### T02: macOS Wallpaper Backend
**Estimate:** 45m

**Description:**
Implement macOS wallpaper setting using AppleScript or native API.

**Steps:**
1. Create `internal/platform/macos.go`
2. Research: `osascript` vs NSWorkspace API
3. Implement `SetWallpaper(path string) error`
4. Support both desktop images (all spaces) and current space
5. Handle multiple monitors (stretch vs fill)
6. Test on Intel and Apple Silicon

**Implementation Options:**
```bash
# AppleScript approach (preferred for simplicity)
tell application "Finder" to set desktop picture to POSIX file "/path/to/image.jpg"

# Or native API via cgo (more complex, avoid for now)
```

**Files Likely Touched:**
- internal/platform/macos.go

**Expected Output:**
- Can set wallpaper on macOS
- Works on both Intel and ARM
- Handles errors gracefully

**Verification:**
```bash
./wallpaper-cli set ~/test.jpg
# Desktop background changes
```

---

### T03: Linux Wallpaper Backend
**Estimate:** 60m

**Description:**
Implement Linux wallpaper setting supporting GNOME, KDE, and XFCE.

**Steps:**
1. Create `internal/platform/linux.go`
2. Implement backends for each DE:
   - GNOME: `gsettings set org.gnome.desktop.background picture-uri`
   - KDE: `qdbus org.kde.plasmashell /PlasmaShell evaluateScript`
   - XFCE: `xfconf-query -c xfce4-desktop -p /backdrop`
3. Auto-detect DE and use appropriate backend
4. Fallback to `feh` or `nitrogen` if DE unknown
5. Handle file:// URI conversion

**Files Likely Touched:**
- internal/platform/linux.go

**Expected Output:**
- Works on GNOME (Ubuntu default)
- Works on KDE
- Works on XFCE
- Graceful fallback for unsupported DEs

**Verification:**
```bash
# On Ubuntu GNOME
./wallpaper-cli set ~/test.jpg
# Desktop background changes
```

---

### T04: Windows Wallpaper Backend
**Estimate:** 45m

**Description:**
Implement Windows wallpaper setting using PowerShell or Registry.

**Steps:**
1. Create `internal/platform/windows.go`
2. Implement using PowerShell:
   ```powershell
   Set-ItemProperty -Path "HKCU:\Control Panel\Desktop" -Name Wallpaper -Value "path"
   rundll32.exe user32.dll,UpdatePerUserSystemParameters
   ```
3. Or use SystemParametersInfo Win32 API via syscall
4. Handle Windows 10/11 differences
5. Support different wallpaper styles (fill, fit, stretch, tile, center)

**Files Likely Touched:**
- internal/platform/windows.go

**Expected Output:**
- Works on Windows 10
- Works on Windows 11
- Updates wallpaper immediately

**Verification:**
```powershell
.\wallpaper-cli.exe set C:\Users\me\test.jpg
# Desktop background changes
```

---

### T05: Set Command Implementation
**Estimate:** 45m

**Description:**
Create the `set` CLI command with path argument.

**Steps:**
1. Create `cmd/set.go`
2. Add `set` subcommand to root
3. Implement path validation (file exists, is image)
4. Call platform-specific SetWallpaper
5. Store current wallpaper in config
6. Add success/error messages

**CLI Design:**
```bash
wallpaper-cli set <path>           # Set specific wallpaper
wallpaper-cli set --config         # Show current wallpaper path
```

**Files Likely Touched:**
- cmd/set.go

**Expected Output:**
- `set` command available
- Validates input
- Sets wallpaper on current platform

**Verification:**
```bash
./wallpaper-cli set --help
./wallpaper-cli set ~/wallpapers/test.jpg
```

---

### T06: Random/Latest Set Options
**Estimate:** 30m

**Description:**
Add `--random` and `--latest` flags to set command.

**Steps:**
1. Add `--random` flag: pick random from output directory
2. Add `--latest` flag: pick most recently downloaded
3. Read output directory from config
4. Scan directory for images
5. Use metadata or mtime for "latest"

**Files Likely Touched:**
- cmd/set.go

**Expected Output:**
- `set --random` works
- `set --latest` works
- Help text explains options

**Verification:**
```bash
./wallpaper-cli set --random
./wallpaper-cli set --latest
```

---

### T07: Cross-Platform Testing
**Estimate:** 60m

**Description:**
Test set command on all target platforms.

**Test Matrix:**
| Platform | Variant | Test |
|----------|---------|------|
| macOS | Intel | ✓ |
| macOS | Apple Silicon | ✓ |
| Linux | Ubuntu GNOME | ✓ |
| Linux | KDE Plasma | ✓ |
| Linux | XFCE | ✓ |
| Windows | Windows 10 | ✓ |
| Windows | Windows 11 | ✓ |

**Steps:**
1. Build for each platform
2. Test set with path
3. Test set --random
4. Test set --latest
5. Document any issues

**Files Likely Touched:**
- Test notes only

**Expected Output:**
- All platforms working
- Test results documented
- Any platform-specific notes added to README

**Verification:**
```bash
# On each platform
./wallpaper-cli set --latest
# Verify wallpaper changed visually
```
