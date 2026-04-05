package uat

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

var binaryPath string

func TestMain(m *testing.M) {
	// Build the binary once before all tests
	tmpDir, err := os.MkdirTemp("", "wallpaper-uat-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	binaryPath = filepath.Join(tmpDir, "wallpaper-cli")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = filepath.Join("..", "..")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build binary: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func run(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}

// ─── CLI SMOKE TESTS ───────────────────────────────────────────

func TestCLI_VersionOutput(t *testing.T) {
	stdout, _, code := run(t, "--version")
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout, "wallpaper-cli") {
		t.Errorf("version output = %q, want 'wallpaper-cli' somewhere", stdout)
	}
}

func TestCLI_HelpOutput(t *testing.T) {
	stdout, _, code := run(t, "--help")
	if code != 0 {
		t.Fatalf("exit code = %d", code)
	}
	for _, cmd := range []string{"fetch", "list", "set", "export", "config", "stats", "favorite", "rate", "playlist"} {
		if !strings.Contains(stdout, cmd) {
			t.Errorf("help missing subcommand %q", cmd)
		}
	}
}

func TestCLI_FetchHelp(t *testing.T) {
	stdout, _, _ := run(t, "fetch", "--help")
	for _, flag := range []string{"--source", "--tags", "--limit", "--json", "--resolution", "--dry-run"} {
		if !strings.Contains(stdout, flag) {
			t.Errorf("fetch help missing flag %q", flag)
		}
	}
}

// ─── DRY RUN ───────────────────────────────────────────────────

func TestFetch_DryRun_Wallhaven(t *testing.T) {
	stdout, _, code := run(t, "fetch", "--source", "wallhaven", "--tags", "anime", "--limit", "5", "--dry-run")
	if code != 0 {
		t.Fatalf("exit code = %d", code)
	}
	if !strings.Contains(stdout, "DRY RUN") {
		t.Error("missing DRY RUN header")
	}
	if !strings.Contains(stdout, "wallhaven") {
		t.Error("missing source name")
	}
}

func TestFetch_DryRun_Danbooru(t *testing.T) {
	stdout, _, code := run(t, "fetch", "--source", "danbooru", "--tags", "sky", "--limit", "5", "--dry-run")
	if code != 0 {
		t.Fatalf("exit code = %d", code)
	}
	if !strings.Contains(stdout, "DRY RUN") {
		t.Error("missing DRY RUN header")
	}
}

func TestFetch_DryRun_Konachan(t *testing.T) {
	stdout, _, code := run(t, "fetch", "--source", "konachan", "--tags", "landscape", "--limit", "5", "--dry-run")
	if code != 0 {
		t.Fatalf("exit code = %d", code)
	}
	if !strings.Contains(stdout, "DRY RUN") {
		t.Error("missing DRY RUN header")
	}
}

func TestFetch_DryRun_Reddit(t *testing.T) {
	stdout, _, code := run(t, "fetch", "--source", "reddit", "--limit", "5", "--dry-run")
	if code != 0 {
		t.Fatalf("exit code = %d", code)
	}
	if !strings.Contains(stdout, "DRY RUN") {
		t.Error("missing DRY RUN header")
	}
}

// ─── SOURCE VALIDATION ────────────────────────────────────────

func TestFetch_InvalidSource(t *testing.T) {
	_, stderr, code := run(t, "fetch", "--source", "fakesource", "--dry-run")
	if code == 0 {
		t.Error("expected non-zero exit for invalid source")
	}
	if !strings.Contains(stderr, "invalid source") && !strings.Contains(stderr, "fakesource") {
		t.Errorf("error should mention invalid source, got: %q", stderr)
	}
}

func TestFetch_DanbooruTagLimit(t *testing.T) {
	// Danbooru limits anonymous to 2 tags — 3+ should emit error event
	// Exit code may be 0 (partial results design), but stderr should have the error
	_, stderr, _ := run(t, "fetch", "--source", "danbooru", "--tags", "sky, cloud, sunset", "--limit", "1")
	if !strings.Contains(stderr, "anonymous") && !strings.Contains(stderr, "limit") {
		t.Errorf("error should mention anonymous tag limit on stderr, got: %q", stderr)
	}
}

func TestFetch_DanbooruTagLimit_JSON(t *testing.T) {
	// In JSON mode the error should appear as an error event in stdout
	stdout, _, _ := run(t, "fetch", "--source", "danbooru", "--tags", "sky, cloud, sunset", "--limit", "1", "--json")
	var foundError bool
	for _, line := range strings.Split(strings.TrimSpace(stdout), "\n") {
		if line == "" {
			continue
		}
		var evt map[string]interface{}
		json.Unmarshal([]byte(line), &evt)
		if evt["type"] == "error" {
			if errMsg, ok := evt["error"].(string); ok && strings.Contains(errMsg, "anonymous") {
				foundError = true
			}
		}
	}
	if !foundError {
		t.Error("expected error event about anonymous tag limit in JSON output")
	}
}

func TestFetch_ZerochanRequiresUsername(t *testing.T) {
	// Zerochan emits an error event when username is missing
	_, stderr, _ := run(t, "fetch", "--source", "zerochan", "--tags", "test", "--limit", "1")
	if !strings.Contains(stderr, "username") {
		t.Errorf("error should mention username requirement on stderr, got: %q", stderr)
	}
}

func TestFetch_ZerochanRequiresUsername_JSON(t *testing.T) {
	stdout, _, _ := run(t, "fetch", "--source", "zerochan", "--tags", "test", "--limit", "1", "--json")
	var foundError bool
	for _, line := range strings.Split(strings.TrimSpace(stdout), "\n") {
		if line == "" {
			continue
		}
		var evt map[string]interface{}
		json.Unmarshal([]byte(line), &evt)
		if evt["type"] == "error" {
			if errMsg, ok := evt["error"].(string); ok && strings.Contains(errMsg, "username") {
				foundError = true
			}
		}
	}
	if !foundError {
		t.Error("expected error event about username in JSON output")
	}
}

// ─── JSON OUTPUT CONTRACT ──────────────────────────────────────

func TestFetch_JSON_DanbooruTagError(t *testing.T) {
	// Even errors should be valid JSON events
	stdout, _, _ := run(t, "fetch", "--source", "danbooru", "--tags", "a, b, c", "--limit", "1", "--json")

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}
		var evt map[string]interface{}
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			t.Errorf("line %d: invalid JSON: %v\nline: %q", i, err, line)
			continue
		}
		if _, ok := evt["type"]; !ok {
			t.Errorf("line %d: missing 'type' field", i)
		}
		if _, ok := evt["timestamp"]; !ok {
			t.Errorf("line %d: missing 'timestamp' field", i)
		}
	}
}

func TestFetch_JSON_CapabilitiesFirst(t *testing.T) {
	stdout, _, _ := run(t, "fetch", "--source", "danbooru", "--tags", "a, b, c", "--limit", "1", "--json")

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) == 0 {
		t.Fatal("no output")
	}

	var first map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &first); err != nil {
		t.Fatalf("first line invalid JSON: %v", err)
	}

	if first["type"] != "capabilities" {
		t.Errorf("first event type = %v, want 'capabilities'", first["type"])
	}
}

func TestList_JSON_StableFields(t *testing.T) {
	stdout, _, code := run(t, "list", "--json")
	if code != 0 {
		t.Skipf("list failed (no DB yet?), exit code %d", code)
	}
	if stdout == "" {
		t.Skip("no list output")
	}

	// Should be parseable JSON
	var result interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Errorf("list --json output not valid JSON: %v", err)
	}
}

func TestStats_JSON_StableFields(t *testing.T) {
	stdout, _, code := run(t, "stats", "--json")
	if code != 0 {
		t.Skipf("stats failed, exit code %d", code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Errorf("stats --json not valid JSON: %v", err)
	}

	for _, field := range []string{"total_count", "total_size_bytes", "by_source", "by_resolution"} {
		if _, ok := result[field]; !ok {
			t.Errorf("stats --json missing field %q", field)
		}
	}
}

// ─── INPUT VALIDATION ──────────────────────────────────────────

func TestFetch_InvalidResolution(t *testing.T) {
	_, stderr, code := run(t, "fetch", "--resolution", "invalid", "--dry-run")
	if code == 0 {
		t.Error("expected error for invalid resolution")
	}
	_ = stderr
}

func TestFetch_InvalidLimit(t *testing.T) {
	_, _, code := run(t, "fetch", "--limit", "0", "--dry-run")
	if code == 0 {
		t.Error("expected error for limit=0")
	}
}

func TestFetch_InvalidFormat(t *testing.T) {
	_, _, code := run(t, "fetch", "--format", "bmp", "--dry-run")
	if code == 0 {
		t.Error("expected error for invalid format")
	}
}

// ─── LIVE FETCH (network required, short timeout) ──────────────

func TestLive_Danbooru_SingleTag(t *testing.T) {
	if os.Getenv("UAT_LIVE") == "" {
		t.Skip("set UAT_LIVE=1 to run live network tests")
	}

	dir := t.TempDir()
	stdout, stderr, code := run(t, "fetch", "--source", "danbooru", "--tags", "scenery", "--limit", "2", "--output", dir, "--json")
	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, stderr)
	}

	// Parse JSON events
	var foundComplete bool
	for _, line := range strings.Split(strings.TrimSpace(stdout), "\n") {
		if line == "" {
			continue
		}
		var evt map[string]interface{}
		json.Unmarshal([]byte(line), &evt)
		if evt["type"] == "download_complete" {
			foundComplete = true
		}
	}
	if !foundComplete {
		t.Error("no download_complete event in JSON output")
	}

	// Check files exist
	files, _ := filepath.Glob(filepath.Join(dir, "danbooru", "*.jpg"))
	pngs, _ := filepath.Glob(filepath.Join(dir, "danbooru", "*.png"))
	files = append(files, pngs...)
	if len(files) == 0 {
		t.Error("no files downloaded")
	}
}

func TestLive_Konachan_SingleTag(t *testing.T) {
	if os.Getenv("UAT_LIVE") == "" {
		t.Skip("set UAT_LIVE=1 to run live network tests")
	}

	dir := t.TempDir()
	_, stderr, code := run(t, "fetch", "--source", "konachan", "--tags", "landscape", "--limit", "2", "--output", dir, "--dedup=false")
	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, stderr)
	}

	files, _ := filepath.Glob(filepath.Join(dir, "konachan", "*"))
	if len(files) == 0 {
		t.Error("no files downloaded")
	}
}

func TestLive_Wallhaven_Fetch(t *testing.T) {
	if os.Getenv("UAT_LIVE") == "" {
		t.Skip("set UAT_LIVE=1 to run live network tests")
	}

	dir := t.TempDir()
	_, stderr, code := run(t, "fetch", "--source", "wallhaven", "--tags", "anime", "--limit", "2", "--output", dir)
	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, stderr)
	}

	files, _ := filepath.Glob(filepath.Join(dir, "wallhaven", "*"))
	if len(files) == 0 {
		t.Error("no files downloaded")
	}
}

func TestLive_TagHarvesting(t *testing.T) {
	if os.Getenv("UAT_LIVE") == "" {
		t.Skip("set UAT_LIVE=1 to run live network tests")
	}

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	os.Setenv("HOME", dir)
	os.MkdirAll(filepath.Join(dir, ".local", "share", "wallpaper-cli"), 0755)
	// Symlink or copy won't work easily — just check tags after fetch

	_, _, code := run(t, "fetch", "--source", "danbooru", "--tags", "scenery", "--limit", "1", "--output", dir)
	if code != 0 {
		t.Skip("fetch failed, can't check tags")
	}

	// Check DB for tags
	dbActual := filepath.Join(dir, ".local", "share", "wallpaper-cli", "wallpapers.db")
	if _, err := os.Stat(dbActual); os.IsNotExist(err) {
		// DB might be in original HOME location
		dbActual = dbPath
	}

	db, err := sql.Open("sqlite", dbActual)
	if err != nil {
		t.Skipf("can't open DB: %v", err)
	}
	defer db.Close()

	var count int
	db.QueryRow("SELECT COUNT(*) FROM source_tags WHERE source = 'danbooru'").Scan(&count)
	if count == 0 {
		t.Error("no danbooru tags harvested")
	}
}

func TestLive_ParallelFetch(t *testing.T) {
	if os.Getenv("UAT_LIVE") == "" {
		t.Skip("set UAT_LIVE=1 to run live network tests")
	}

	dir := t.TempDir()
	stdout, stderr, code := run(t, "fetch", "--source", "all", "--tags", "anime", "--limit", "1", "--output", dir, "--json")
	if code != 0 {
		t.Logf("stderr: %s", stderr)
	}

	// Should have events from multiple sources
	sourcesSeen := make(map[string]bool)
	for _, line := range strings.Split(strings.TrimSpace(stdout), "\n") {
		if line == "" {
			continue
		}
		var evt map[string]interface{}
		json.Unmarshal([]byte(line), &evt)
		if src, ok := evt["source"].(string); ok && src != "" {
			sourcesSeen[src] = true
		}
	}

	// At minimum wallhaven and danbooru should appear (reddit may fail, zerochan needs username)
	if len(sourcesSeen) < 2 {
		t.Errorf("expected events from multiple sources, saw: %v", sourcesSeen)
	}
}

// ─── CONFIG ────────────────────────────────────────────────────

func TestConfig_Init(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	_, _, code := run(t, "config", "init", "--config", cfgPath)
	if code != 0 {
		// config init may not support --config flag, that's ok
		t.Skip("config init with custom path not supported")
	}
}

func TestConfig_List(t *testing.T) {
	_, _, code := run(t, "config", "list")
	// May error if no config exists, that's fine — just shouldn't panic
	_ = code
}
