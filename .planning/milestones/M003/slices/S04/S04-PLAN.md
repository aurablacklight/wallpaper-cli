# M003-S04: Daemon & Platform Services — Background Wallpaper Rotation

**Slice:** S04 of M003  
**Goal:** Implement cross-platform background service/daemon that runs the scheduling engine  
**Estimate:** 8-10 hours  
**Dependencies:** S03 (Scheduling Engine) — daemon wraps and runs the engine

---

## Overview

Build the background service infrastructure:
- **Daemon process:** Runs persistently in background
- **Platform integration:** cron (Linux), launchd (macOS), Task Scheduler (Windows)
- **Lifecycle management:** Start, stop, restart, status
- **Logging & monitoring:** Track activity, errors, health

---

## Architecture

### High-Level Flow

```
User CLI Command → Daemon Control → Schedule Engine → Platform Setter
                                            ↓
                                    State Persistence
                                            ↓
                                    Activity Logging
```

### Component Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                         DAEMON PROCESS                               │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐            │
│  │   Control    │  │   Engine     │  │   State      │            │
│  │   Server     │◄─┤   (S03)      │◄─┤   Manager    │            │
│  │              │  │              │  │              │            │
│  │ - IPC/Signals│  │ - Tick()     │  │ - Save/Load  │            │
│  │ - Start/Stop │  │ - Schedules  │  │ - Resume     │            │
│  └──────┬───────┘  └──────┬───────┘  └──────────────┘            │
│         │                  │                                        │
│         │                  ▼                                        │
│         │           ┌──────────────┐                              │
│         │           │   Setter     │                              │
│         │           │   (M002)     │                              │
│         │           └──────┬───────┘                              │
│         │                  │                                        │
│         │                  ▼                                        │
│         │           ┌──────────────┐                              │
│         └──────────►│    Log       │                              │
│                     │   Activity   │                              │
│                     └──────────────┘                              │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
         ▲
         │
    ┌────┴────┐
    │   CLI   │
    │ Commands│
    └─────────┘
```

---

## CLI Commands

### daemon — Control the background service

```bash
# Start daemon
wallpaper-cli daemon start
wallpaper-cli daemon start --foreground  # Run in foreground (debug)

# Stop daemon
wallpaper-cli daemon stop

# Restart daemon
wallpaper-cli daemon restart

# Check status
wallpaper-cli daemon status
# Output: 
# Daemon: Running (PID 12345)
# Uptime: 2h 34m
# Active schedules: 2
# Last rotation: 14:30 (15m ago)
# Next rotation: 15:00 (in 15m)

# View logs
wallpaper-cli daemon logs [--follow] [--since 1h]

# Install as system service (platform-specific)
wallpaper-cli daemon install
# - Linux: Creates systemd user service or cron job
# - macOS: Creates launchd plist
# - Windows: Creates Task Scheduler entry

# Uninstall service
wallpaper-cli daemon uninstall

# Debug mode (verbose logging)
wallpaper-cli daemon start --debug
```

### Convenience Commands

```bash
# Quick start with default schedule
wallpaper-cli daemon start --every 30m

# Start with specific schedule
wallpaper-cli daemon start --schedule "focus"

# One-shot mode (set once and exit, no daemon)
wallpaper-cli daemon once --every 30m
```

---

## Platform Implementations

### macOS: launchd

**Service Definition:** `~/Library/LaunchAgents/com.wallpaper-cli.daemon.plist`

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.wallpaper-cli.daemon</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/wallpaper-cli</string>
        <string>daemon</string>
        <string>run</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/Users/USER/.local/share/wallpaper-cli/daemon.log</string>
    <key>StandardErrorPath</key>
    <string>/Users/USER/.local/share/wallpaper-cli/daemon.error.log</string>
</dict>
</plist>
```

**Commands:**
```bash
launchctl load ~/Library/LaunchAgents/com.wallpaper-cli.daemon.plist
launchctl unload ~/Library/LaunchAgents/com.wallpaper-cli.daemon.plist
launchctl start com.wallpaper-cli.daemon
launchctl stop com.wallpaper-cli.daemon
```

**Go Implementation:**
```go
// internal/daemon/macos.go

package daemon

import (
    "os"
    "path/filepath"
    "os/exec"
    "text/template"
)

type LaunchdManager struct {
    plistPath string
    label     string
}

func NewLaunchdManager() *LaunchdManager {
    home, _ := os.UserHomeDir()
    return &LaunchdManager{
        plistPath: filepath.Join(home, "Library", "LaunchAgents", "com.wallpaper-cli.daemon.plist"),
        label:     "com.wallpaper-cli.daemon",
    }
}

func (lm *LaunchdManager) Install() error {
    // Generate plist file
    tmpl := template.Must(template.New("plist").Parse(launchdPlistTemplate))
    
    exePath, _ := os.Executable()
    data := map[string]string{
        "Label":       lm.label,
        "ProgramPath": exePath,
        "LogPath":     filepath.Join(home, ".local", "share", "wallpaper-cli", "daemon.log"),
    }
    
    file, _ := os.Create(lm.plistPath)
    defer file.Close()
    
    if err := tmpl.Execute(file, data); err != nil {
        return err
    }
    
    // Load the service
    cmd := exec.Command("launchctl", "load", lm.plistPath)
    return cmd.Run()
}

func (lm *LaunchdManager) Uninstall() error {
    exec.Command("launchctl", "unload", lm.plistPath).Run()
    return os.Remove(lm.plistPath)
}

func (lm *LaunchdManager) Start() error {
    return exec.Command("launchctl", "start", lm.label).Run()
}

func (lm *LaunchdManager) Stop() error {
    return exec.Command("launchctl", "stop", lm.label).Run()
}

func (lm *LaunchdManager) IsRunning() (bool, error) {
    // Check if service is loaded and running
    cmd := exec.Command("launchctl", "list", lm.label)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return false, nil
    }
    // Parse output to check status
    return strings.Contains(string(output), lm.label), nil
}
```

### Linux: systemd (preferred) or cron

**Option A: systemd User Service**

Service file: `~/.config/systemd/user/wallpaper-cli.service`

```ini
[Unit]
Description=Wallpaper CLI Daemon
After=graphical-session.target

[Service]
Type=simple
ExecStart=%h/.local/bin/wallpaper-cli daemon run
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=default.target
```

**Commands:**
```bash
systemctl --user enable wallpaper-cli
systemctl --user start wallpaper-cli
systemctl --user stop wallpaper-cli
systemctl --user status wallpaper-cli
```

**Option B: Cron (Fallback for systems without systemd)**

```bash
# Add to crontab
*/5 * * * * /home/user/.local/bin/wallpaper-cli daemon tick
```

### Windows: Task Scheduler

**Task Definition:** (Created via schtasks.exe or COM API)

```powershell
# PowerShell equivalent of what we need to do
$Action = New-ScheduledTaskAction -Execute "wallpaper-cli.exe" -Argument "daemon run"
$Trigger = New-ScheduledTaskTrigger -AtLogOn
$Settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -StartWhenAvailable
$Principal = New-ScheduledTaskPrincipal -UserId "$env:USERNAME" -LogonType Interactive

Register-ScheduledTask -TaskName "WallpaperCLIDaemon" -Action $Action -Trigger $Trigger -Settings $Settings -Principal $Principal
```

**Go Implementation (using schtasks.exe):**
```go
// internal/daemon/windows.go

package daemon

import (
    "os/exec"
    "fmt"
)

type TaskSchedulerManager struct {
    taskName string
}

func NewTaskSchedulerManager() *TaskSchedulerManager {
    return &TaskSchedulerManager{
        taskName: "WallpaperCLIDaemon",
    }
}

func (tsm *TaskSchedulerManager) Install() error {
    exePath, _ := os.Executable()
    
    // Create task
    cmd := exec.Command("schtasks", "/Create",
        "/TN", tsm.taskName,
        "/TR", fmt.Sprintf(`"%s daemon run"`, exePath),
        "/SC", "ONLOGON",
        "/RL", "LIMITED",
        "/F",
    )
    return cmd.Run()
}

func (tsm *TaskSchedulerManager) Uninstall() error {
    cmd := exec.Command("schtasks", "/Delete", "/TN", tsm.taskName, "/F")
    return cmd.Run()
}

func (tsm *TaskSchedulerManager) Start() error {
    cmd := exec.Command("schtasks", "/Run", "/TN", tsm.taskName)
    return cmd.Run()
}

func (tsm *TaskSchedulerManager) Stop() error {
    cmd := exec.Command("schtasks", "/End", "/TN", tsm.taskName)
    return cmd.Run()
}

func (tsm *TaskSchedulerManager) IsRunning() (bool, error) {
    cmd := exec.Command("schtasks", "/Query", "/TN", tsm.taskName, "/FO", "CSV")
    output, err := cmd.Output()
    if err != nil {
        return false, err
    }
    // Parse CSV output to check if "Running" appears
    return strings.Contains(string(output), "Running"), nil
}
```

---

## Daemon Core Implementation

### Daemon Struct

```go
// internal/daemon/daemon.go

package daemon

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/user/wallpaper-cli/internal/schedule"
    "github.com/user/wallpaper-cli/internal/platform"
    "github.com/user/wallpaper-cli/internal/collections"
)

type Daemon struct {
    engine      *schedule.Engine
    platform    daemon.PlatformManager
    logger      *log.Logger
    
    // Control
    ctx         context.Context
    cancel      context.CancelFunc
    tickInterval time.Duration
}

func New(manager *collections.Manager, setter platform.Setter) *Daemon {
    engine := schedule.NewEngine(manager, setter)
    platform := getPlatformManager() // Returns LaunchdManager, SystemdManager, or TaskSchedulerManager
    
    ctx, cancel := context.WithCancel(context.Background())
    
    return &Daemon{
        engine:       engine,
        platform:     platform,
        logger:       setupLogger(),
        ctx:          ctx,
        cancel:       cancel,
        tickInterval: 1 * time.Minute,
    }
}

// Run is the main daemon loop (called by service/daemon run)
func (d *Daemon) Run() error {
    d.logger.Println("Daemon starting...")
    
    // Load saved state
    if err := d.engine.LoadState(); err != nil {
        d.logger.Printf("Warning: could not load state: %v", err)
    }
    
    // Start engine
    if err := d.engine.Start(); err != nil {
        return fmt.Errorf("starting engine: %w", err)
    }
    defer d.engine.Stop()
    
    // Setup signal handling
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    // Main loop
    ticker := time.NewTicker(d.tickInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-d.ctx.Done():
            d.logger.Println("Daemon shutting down...")
            return d.saveState()
            
        case sig := <-sigChan:
            d.logger.Printf("Received signal: %v", sig)
            return d.shutdown()
            
        case <-ticker.C:
            // Check and execute schedules
            if err := d.engine.Tick(); err != nil {
                d.logger.Printf("Tick error: %v", err)
            }
        }
    }
}

func (d *Daemon) shutdown() error {
    d.cancel()
    return d.saveState()
}

func (d *Daemon) saveState() error {
    if err := d.engine.SaveState(); err != nil {
        d.logger.Printf("Error saving state: %v", err)
        return err
    }
    return nil
}

// Foreground mode (for debugging)
func (d *Daemon) RunForeground() error {
    d.logger = log.New(os.Stdout, "[daemon] ", log.LstdFlags)
    return d.Run()
}

// One-shot mode (set once and exit)
func (d *Daemon) RunOnce() error {
    return d.engine.Tick()
}
```

### Platform Manager Interface

```go
// internal/daemon/platform.go

package daemon

type PlatformManager interface {
    // Service installation
    Install() error
    Uninstall() error
    
    // Runtime control
    Start() error
    Stop() error
    Restart() error
    
    // Status
    IsRunning() (bool, error)
    GetStatus() (*DaemonStatus, error)
}

type DaemonStatus struct {
    Running       bool
    PID           int
    Uptime        time.Duration
    LastRotation  time.Time
    NextRotation  time.Time
    ActiveSchedules int
}

func getPlatformManager() PlatformManager {
    switch runtime.GOOS {
    case "darwin":
        return NewLaunchdManager()
    case "linux":
        // Prefer systemd, fall back to cron
        if hasSystemd() {
            return NewSystemdManager()
        }
        return NewCronManager()
    case "windows":
        return NewTaskSchedulerManager()
    default:
        panic("unsupported platform")
    }
}
```

---

## Logging

### Log Format

```
[2026-04-04 14:30:01] [INFO] Daemon started
[2026-04-04 14:30:01] [INFO] Loaded 2 schedules
[2026-04-04 15:00:00] [INFO] Executing schedule: work-focus
[2026-04-04 15:00:01] [INFO] Set wallpaper: /Users/derek/Pictures/wallpapers/03_abc.jpg
[2026-04-04 15:00:01] [INFO] Next run: 15:30:00
[2026-04-04 18:00:00] [INFO] Executing schedule: day-cycle
[2026-04-04 18:00:00] [INFO] Theme changed: day → evening
[2026-04-04 18:00:01] [INFO] Set wallpaper: /Users/derek/Pictures/wallpapers/cozy/05_def.jpg
```

### Log Location

- **macOS:** `~/.local/share/wallpaper-cli/daemon.log`
- **Linux:** `~/.local/share/wallpaper-cli/daemon.log` (or journald if using systemd)
- **Windows:** `%APPDATA%\wallpaper-cli\daemon.log`

---

## Tasks

| ID | Title | Est. | Details |
|----|-------|------|---------|
| T01 | Create daemon core struct | 1h | Run loop, signals, state management |
| T02 | Implement macOS launchd manager | 1.5h | Plist generation, load/unload |
| T03 | Implement Linux systemd manager | 1.5h | Service file, systemctl commands |
| T04 | Implement Linux cron fallback | 1h | Crontab management |
| T05 | Implement Windows Task Scheduler | 1.5h | schtasks.exe integration |
| T06 | Create `daemon` CLI command | 1h | Start, stop, status, logs |
| T07 | Add logging infrastructure | 0.5h | File logging, rotation |
| T08 | Status command implementation | 0.5h | Show uptime, schedules, next run |
| T09 | Testing per platform | 1h | Manual testing on macOS, Linux, Windows |

**Total: 9 hours**

---

## Integration Points

### With S03 (Scheduling)
- Daemon wraps the schedule.Engine
- Calls `engine.Tick()` every minute
- Saves/loads engine state on start/stop

### With S02 (Collections)
- Engine uses collections.Manager
- Daemon runs with same permissions as user

### With S01 (TUI)
- TUI shows daemon status in status bar
- Can start/stop daemon from TUI (with confirmation)
- Shows countdown to next rotation

---

## Testing Strategy

**Per-Platform Tests:**

1. **Install service:** `daemon install` succeeds
2. **Start daemon:** `daemon start` succeeds, process visible
3. **Check status:** `daemon status` shows running
4. **Verify rotation:** Wait for interval, wallpaper changes
5. **Stop daemon:** `daemon stop` succeeds, process ends
6. **Uninstall:** `daemon uninstall` removes service
7. **Resume test:** Start daemon, verify it resumes from last position

**Edge Cases:**
- [ ] Crash recovery (save state before crash, resume after)
- [ ] Multiple start attempts (graceful handling)
- [ ] Permission issues (user vs system install)
- [ ] Log rotation (don't fill disk)
- [ ] Empty schedules (graceful no-op)

---

## Success Criteria

- [ ] Daemon starts and runs on all 3 platforms
- [ ] Service installation works (launchd/systemd/Task Scheduler)
- [ ] Automatic start on login works
- [ ] Rotation happens at scheduled intervals
- [ ] State persists across daemon restarts
- [ ] Logs are written and viewable
- [ ] Status command shows accurate info
- [ ] Graceful shutdown on signal
- [ ] Can run in foreground for debugging
- [ ] One-shot mode works (tick once and exit)

---

## Security Considerations

- **User-level only:** Daemon runs as current user (no root/admin required for wallpaper setting)
- **No network access:** Daemon doesn't need network (wallpapers already downloaded)
- **File permissions:** State and logs in user home directory
- **Minimal attack surface:** Simple ticker, no external inputs except CLI commands

---

*The daemon is what makes M003 truly "set it and forget it" — once configured, wallpapers rotate automatically without any user intervention.*
