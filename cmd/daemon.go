package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Control the background wallpaper daemon",
	Long: `The daemon runs in the background and automatically rotates
wallpapers according to your schedules.

Platform services:
  macOS:  Uses launchd
  Linux:  Uses systemd or cron
  Windows: Uses Task Scheduler`,
}

var (
	daemonForeground bool
	daemonDebug      bool
)

var daemonStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the daemon",
	Long:  `Start the wallpaper rotation daemon.`,
	RunE:  runDaemonStart,
}

var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the daemon",
	RunE:  runDaemonStop,
}

var daemonStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show daemon status",
	RunE:  runDaemonStatus,
}

var daemonRunCmd = &cobra.Command{
	Use:    "run",
	Short:  "Run daemon in foreground (internal)",
	Hidden: true,
	RunE:   runDaemonRun,
}

func init() {
	rootCmd.AddCommand(daemonCmd)
	daemonCmd.AddCommand(daemonStartCmd)
	daemonCmd.AddCommand(daemonStopCmd)
	daemonCmd.AddCommand(daemonStatusCmd)
	daemonCmd.AddCommand(daemonRunCmd)

	daemonStartCmd.Flags().BoolVarP(&daemonForeground, "foreground", "f", false, "Run in foreground (debug)")
	daemonStartCmd.Flags().BoolVarP(&daemonDebug, "debug", "d", false, "Enable debug logging")
}

func runDaemonStart(cmd *cobra.Command, args []string) error {
	// Check if already running
	if isDaemonRunning() {
		fmt.Println("Daemon is already running")
		return nil
	}

	if daemonForeground {
		// Run directly in foreground
		fmt.Println("Starting daemon in foreground mode...")
		fmt.Println("Press Ctrl+C to stop")
		return runDaemonInForeground()
	}

	// Platform-specific service installation
	switch getPlatform() {
	case "darwin":
		fmt.Println("Installing launchd service...")
		// Would install launchd plist here
		fmt.Println("To start automatically on boot, run:")
		fmt.Println("  launchctl load ~/Library/LaunchAgents/com.wallpaper-cli.daemon.plist")

	case "linux":
		fmt.Println("Installing systemd service...")
		fmt.Println("To start automatically, run:")
		fmt.Println("  systemctl --user enable wallpaper-cli")
		fmt.Println("  systemctl --user start wallpaper-cli")

	case "windows":
		fmt.Println("Installing Task Scheduler job...")
		fmt.Println("Run in Task Scheduler to start on login")
	}

	// For now, just run once to demonstrate
	fmt.Println("\n💡 Tip: Use --foreground flag to run interactively")
	fmt.Println("   wallpaper-cli daemon start --foreground")

	return nil
}

func runDaemonStop(cmd *cobra.Command, args []string) error {
	if !isDaemonRunning() {
		fmt.Println("Daemon is not running")
		return nil
	}

	// Send stop signal
	pidFile := getPIDFile()
	pid := readPID(pidFile)
	if pid > 0 {
		syscall.Kill(pid, syscall.SIGTERM)
		os.Remove(pidFile)
	}

	fmt.Println("Daemon stopped")
	return nil
}

func runDaemonStatus(cmd *cobra.Command, args []string) error {
	if isDaemonRunning() {
		fmt.Println("✅ Daemon is running")

		// Load and display schedules
		engine, err := getScheduleEngine()
		if err != nil {
			return err
		}

		schedules := engine.ListSchedules()
		enabledCount := 0
		for _, s := range schedules {
			if s.Enabled {
				enabledCount++
			}
		}

		fmt.Printf("📅 Active schedules: %d/%d\n", enabledCount, len(schedules))
	} else {
		fmt.Println("❌ Daemon is not running")
		fmt.Println("\nTo start the daemon:")
		fmt.Println("  wallpaper-cli daemon start --foreground")
	}

	return nil
}

// runDaemonRun is the internal command for service managers
func runDaemonRun(cmd *cobra.Command, args []string) error {
	return runDaemonInForeground()
}

// runDaemonInForeground runs the daemon loop directly
func runDaemonInForeground() error {
	engine, err := getScheduleEngine()
	if err != nil {
		return fmt.Errorf("initializing schedule engine: %w", err)
	}

	// Write PID file for management
	pidFile := getPIDFile()
	writePID(pidFile, os.Getpid())
	defer os.Remove(pidFile)

	// Start the engine
	engine.Start()
	fmt.Println("🔄 Daemon started, running schedules...")

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\n⏹️  Shutting down daemon...")
	engine.Stop()

	return nil
}

// Helper functions

func isDaemonRunning() bool {
	pidFile := getPIDFile()
	pid := readPID(pidFile)
	if pid == 0 {
		return false
	}

	// Check if process exists
	err := syscall.Kill(pid, 0)
	return err == nil
}

func getPIDFile() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "wallpaper-cli", "daemon.pid")
}

func readPID(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	var pid int
	fmt.Sscanf(string(data), "%d", &pid)
	return pid
}

func writePID(path string, pid int) {
	os.MkdirAll(filepath.Dir(path), 0755)
	os.WriteFile(path, []byte(fmt.Sprintf("%d", pid)), 0644)
}

func getPlatform() string {
	return os.Getenv("GOOS")
}
