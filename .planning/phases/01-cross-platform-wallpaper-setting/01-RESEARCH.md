# Phase 01: Cross-Platform Wallpaper Setting - Research

**Phase:** 01 - Cross-Platform Wallpaper Setting
**Research Date:** 2026-04-04
**Researcher:** gsd-phase-researcher

---

## Executive Summary

This research investigates platform-specific wallpaper setting mechanisms for macOS, Linux (GNOME/KDE/XFCE), and Windows to implement the `wallpaper-cli set` command. All three platforms have viable CLI-based approaches that don't require CGO or complex native bindings.

**Key Findings:**
- **macOS:** AppleScript via `osascript` is the most reliable, CGO-free approach
- **Linux:** DE detection via `$XDG_CURRENT_DESKTOP` with backend-specific commands
- **Windows:** PowerShell + Registry approach works on both Win10/Win11

---

## 1. macOS Wallpaper Setting

### 1.1 AppleScript Approach (Recommended per D-01)

**Command:**
```bash
osascript -e 'tell application "Finder" to set desktop picture to POSIX file "/path/to/image.jpg"'
```

**Pros:**
- No CGO required - pure Go with os/exec
- Works on all macOS versions (10.14+ through 14.x)
- Works on both Intel and Apple Silicon
- Simple error handling via exit codes

**Cons:**
- Requires Finder to be running (always true in normal macOS use)
- Can't control per-display settings without more complex AppleScript
- No direct control over wallpaper styling (fill/fit/stretch)

**Exit Codes:**
- 0: Success
- 1: File not found, invalid path, or permission denied

**Go Implementation Pattern:**
```go
func setMacOSWallpaper(path string) error {
    cmd := exec.Command("osascript", "-e",
        fmt.Sprintf(`tell application "Finder" to set desktop picture to POSIX file %q`, path))
    return cmd.Run()
}
```

### 1.2 Alternative: NSWorkspace via CGO (Deferred per D-02)

For future multi-monitor support, native API provides more control:
```objc
[[NSWorkspace sharedWorkspace] setDesktopImageURL:url forScreen:screen options:options error:&error];
```

**Why Deferred:** Requires CGO, complicates cross-compilation. AppleScript sufficient for v1.2.

---

## 2. Linux Desktop Environment Support

### 2.1 Desktop Environment Detection

**Primary Method:** Environment variable `$XDG_CURRENT_DESKTOP`

| DE | Variable Value |
|----|----------------|
| GNOME | `GNOME`, `ubuntu:GNOME` |
| KDE Plasma | `KDE` |
| XFCE | `XFCE` |
| Cinnamon | `X-Cinnamon` |
| MATE | `MATE` |
| LXDE | `LXDE` |

**Go Detection Pattern:**
```go
func detectLinuxDE() string {
    de := os.Getenv("XDG_CURRENT_DESKTOP")
    de = strings.ToUpper(de)
    // Handle ubuntu:GNOME format
    if strings.Contains(de, "GNOME") {
        return "GNOME"
    }
    return de
}
```

### 2.2 GNOME Backend (D-03)

**Command:**
```bash
gsettings set org.gnome.desktop.background picture-uri "file:///path/to/image.jpg"
```

**Key Details:**
- URI format required (file:// prefix)
- Immediate application, no refresh needed
- Available on all GNOME-based systems (Ubuntu, Fedora, Debian)

**Go Implementation:**
```go
cmd := exec.Command("gsettings", "set", "org.gnome.desktop.background", 
    "picture-uri", "file://"+path)
```

### 2.3 KDE Plasma Backend (D-03)

**Command via qdbus:**
```bash
qdbus org.kde.plasmashell /PlasmaShell evaluateScript '
    var allDesktops = desktops();
    for (var i = 0; i < allDesktops.length; i++) {
        var desktop = allDesktops[i];
        desktop.wallpaperPlugin = "org.kde.image";
        desktop.currentConfigGroup = ["Wallpaper", "org.kde.image", "General"];
        desktop.writeConfig("Image", "file:///path/to/image.jpg");
    }
'
```

**Alternative via kwriteconfig5:**
```bash
kwriteconfig5 --file plasma-org.kde.plasma.desktop-appletsrc \
    --group Containment --group 1 --group Wallpaper --group org.kde.image \
    --group General --key Image "file:///path/to/image.jpg"
```

**Note:** KDE requires more complex scripting. qdbus approach sets wallpaper on all desktops.

### 2.4 XFCE Backend (D-03)

**Command via xfconf-query:**
```bash
# Get monitor/workspace properties
xfconf-query -c xfce4-desktop -l | grep last-image

# Set wallpaper (example for monitor 0, workspace 0)
xfconf-query -c xfce4-desktop -p /backdrop/screen0/monitor0/workspace0/last-image -s "/path/to/image.jpg"
```

**Challenge:** XFCE stores per-monitor, per-workspace paths. For v1.2, setting on all monitors/workspaces is acceptable (D-18).

**Implementation Strategy:**
1. Query all backdrop properties: `xfconf-query -c xfce4-desktop -l`
2. Filter for `last-image` entries
3. Set each property to the new path

### 2.5 Fallback Strategy (D-04)

**feh:** Popular minimal wallpaper setter
```bash
feh --bg-scale /path/to/image.jpg  # or --bg-fill, --bg-center, --bg-tile
```

**nitrogen:** With explicit mode
```bash
nitrogen --set-zoom-fill /path/to/image.jpg
```

**When to Use:** When DE is unknown or unsupported. Check if command exists before attempting.

---

## 3. Windows Wallpaper Setting

### 3.1 PowerShell + Registry Approach (D-06)

**Commands:**
```powershell
# Set wallpaper path in Registry
Set-ItemProperty -Path "HKCU:\Control Panel\Desktop" -Name Wallpaper -Value "C:\path\to\image.jpg"

# Refresh desktop to apply
rundll32.exe user32.dll,UpdatePerUserSystemParameters
```

**Pros:**
- No external dependencies - uses built-in Windows tools
- Works on Windows 10 and Windows 11
- Works on Home, Pro, and Enterprise editions

**Cons:**
- `rundll32` refresh is undocumented/hacky but widely used
- Wallpaper style (fill/fit/stretch) requires additional Registry values

**Go Implementation:**
```go
func setWindowsWallpaper(path string) error {
    // PowerShell command
    psCmd := fmt.Sprintf(`Set-ItemProperty -Path "HKCU:\Control Panel\Desktop" -Name Wallpaper -Value %q`, path)
    cmd := exec.Command("powershell", "-Command", psCmd)
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("setting wallpaper registry: %w", err)
    }
    
    // Refresh
    refresh := exec.Command("rundll32.exe", "user32.dll,UpdatePerUserSystemParameters")
    return refresh.Run()
}
```

### 3.2 Alternative: Win32 API via Syscall (D-07)

```go
// Using SystemParametersInfo syscall
user32 := syscall.NewLazyDLL("user32.dll")
proc := user32.NewProc("SystemParametersInfoW")
pathPtr, _ := syscall.UTF16PtrFromString(path)
proc.Call(0x0014, 0, uintptr(unsafe.Pointer(pathPtr)), 0x01|0x02)
```

**Why Not Primary:** More complex, requires unsafe package. Keep as fallback option.

---

## 4. Platform Detection Strategy

### 4.1 OS Detection

**Use Go's built-in:**
```go
import "runtime"

switch runtime.GOOS {
case "darwin":
    // macOS
    setMacOSWallpaper(path)
case "linux":
    // Linux - detect DE
    de := detectLinuxDE()
    setLinuxWallpaper(path, de)
case "windows":
    // Windows
    setWindowsWallpaper(path)
default:
    return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
}
```

### 4.2 Build Tags (Alternative)

For compile-time separation:
```go
//go:build darwin
// +build darwin

package platform
func SetWallpaper(path string) error { /* macOS */ }
```

**Recommendation:** Use runtime.GOOS for simpler maintenance - single binary handles all platforms.

---

## 5. Error Handling Patterns

### 5.1 Validation (D-13)

**Required Checks:**
1. File exists: `os.Stat(path)`
2. Is regular file: `!info.IsDir()`
3. Supported extension: `.jpg`, `.jpeg`, `.png`, `.gif`, `.bmp`, `.webp`

### 5.2 Platform Command Errors

**Pattern:**
```go
if err := cmd.Run(); err != nil {
    if exitErr, ok := err.(*exec.ExitError); ok {
        // Command ran but failed
        return fmt.Errorf("platform wallpaper command failed (exit %d): %w", 
            exitErr.ExitCode(), err)
    }
    // Command couldn't start (not found, permission denied)
    return fmt.Errorf("could not run wallpaper command: %w", err)
}
```

### 5.3 Exit Codes (D-11, D-12)

| Exit Code | Meaning |
|-----------|---------|
| 0 | Success |
| 1 | Invalid path/file not found |
| 2 | Unsupported file format |
| 3 | Platform command not available |
| 4 | Permission denied |

---

## 6. Config Persistence (D-08, D-09)

### 6.1 Config Extensions

```go
type Config struct {
    // ... existing fields ...
    CurrentWallpaper string            `json:"current_wallpaper"`
    WallpaperHistory []WallpaperRecord `json:"wallpaper_history"`
}

type WallpaperRecord struct {
    Path      string    `json:"path"`
    Timestamp time.Time `json:"timestamp"`
    Source    string    `json:"source"` // "manual", "random", "latest"
}
```

### 6.2 History Management

**Size Limit:** 10 entries (configurable, per OpenCode's Discretion)

**Pattern:**
```go
// Add new entry
record := WallpaperRecord{Path: path, Timestamp: time.Now(), Source: source}
cfg.WallpaperHistory = append([]WallpaperRecord{record}, cfg.WallpaperHistory...)

// Trim to limit
if len(cfg.WallpaperHistory) > 10 {
    cfg.WallpaperHistory = cfg.WallpaperHistory[:10]
}
```

---

## 7. Image Discovery for --random and --latest (D-15, D-16)

### 7.1 Scanning Output Directory

```go
func findWallpapers(dir string) ([]string, error) {
    var wallpapers []string
    err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return nil // Skip inaccessible files
        }
        if info.IsDir() {
            return nil
        }
        ext := strings.ToLower(filepath.Ext(path))
        if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || 
           ext == ".gif" || ext == ".bmp" || ext == ".webp" {
            wallpapers = append(wallpapers, path)
        }
        return nil
    })
    return wallpapers, err
}
```

### 7.2 Latest Detection

**Approaches:**
1. **File Modification Time:** `info.ModTime()` - simplest, uses filesystem
2. **Metadata from download:** Check database/records if available

**Recommendation:** Use modTime for v1.2 - no dependency on download tracking.

### 7.3 Random Selection

```go
rand.Seed(time.Now().UnixNano())
randomPath := wallpapers[rand.Intn(len(wallpapers))]
```

---

## 8. Testing Strategy

### 8.1 Unit Test Approach

**Mock exec.Command:**
```go
type PlatformRunner interface {
    Run(name string, arg ...string) *exec.Cmd
}

type mockRunner struct {
    commands []string
}

func (m *mockRunner) Run(name string, arg ...string) *exec.Cmd {
    m.commands = append(m.commands, name+" "+strings.Join(arg, " "))
    return exec.Command("echo", "mock")
}
```

### 8.2 Platform-Specific Tests

**macOS:**
- Verify `osascript` command generation
- Test with invalid paths
- Test with non-image files

**Linux:**
- Mock DE detection
- Verify correct backend selected per DE
- Test fallback to feh/nitrogen

**Windows:**
- Verify PowerShell command generation
- Test path escaping for spaces

---

## 9. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| macOS AppleScript fails without Finder | Low | High | Document requirement; user always has Finder running |
| Linux DE not in supported list | Medium | Medium | Fallback to feh/nitrogen; clear error message |
| Windows rundll32 doesn't refresh | Low | Medium | Document; retry or document manual refresh |
| Image path contains spaces | High | Medium | Proper quoting in all platform commands |
| Permission denied on set | Low | High | Exit code 4, clear error, suggest checking permissions |

---

## 10. Implementation Recommendations

### 10.1 File Structure

```
internal/platform/
├── platform.go      # Interface + common code
├── detect.go        # Platform detection
├── macos.go         # macOS implementation
├── linux.go         # Linux implementation
├── windows.go       # Windows implementation
└── platform_test.go # Tests
```

### 10.2 Interface Design

```go
// Platform interface
package platform

type Setter interface {
    SetWallpaper(path string) error
    Name() string
}

// Get returns the appropriate setter for current platform
func Get() (Setter, error)
```

### 10.3 Command Design (D-14, D-15, D-16, D-17)

```
wallpaper-cli set <path>          # Set specific file
wallpaper-cli set --random        # Set random from collection
wallpaper-cli set --latest        # Set most recently downloaded
wallpaper-cli set --current       # Show current wallpaper path
wallpaper-cli set --previous      # Set previous wallpaper (future)
```

---

## 11. References

### Documentation
- Apple AppleScript Language Guide: https://developer.apple.com/library/archive/documentation/AppleScript/Conceptual/AppleScriptLangGuide/
- GNOME gsettings: https://wiki.gnome.org/Projects/dconf
- KDE Plasma Wallpaper Scripting: https://userbase.kde.org/Desktop_Scripting
- XFCE xfconf: https://docs.xfce.org/xfce/xfconf/xfconf-query
- Windows SystemParametersInfo: https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-systemparametersinfow

### Code Examples
- `reujab/wallpaper.go` - Go wallpaper library (reference only, not for copying)
- `mholt/wallpaper` - macOS wallpaper package

---

## 12. Nyquist Validation Architecture

### Validation Dimensions

| Dimension | Target | Approach |
|-----------|--------|----------|
| 1. Completeness | All platforms supported | Unit tests per platform |
| 2. Correctness | Works on real systems | CI matrix + manual verification |
| 3. Edge cases | Handles errors gracefully | Error injection tests |
| 4. Integration | CLI command works end-to-end | E2E test on each platform |
| 5. Observability | Logs actions | Structured logging |
| 6. Operations | Clear error messages | Error message validation |

### Validation Checkpoints

1. **Checkpoint 1:** Platform detection works on target OS
2. **Checkpoint 2:** Each platform backend can set wallpaper
3. **Checkpoint 3:** --random and --latest work
4. **Checkpoint 4:** Config persistence works
5. **Checkpoint 5:** Error handling is graceful

---

*Research complete. Ready for planning phase.*
