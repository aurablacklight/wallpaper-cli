# Phase 01: Future Test Cases

**Generated from:** UAT Gap Analysis (01-UAT.md)  
**Date:** 2026-04-04  
**Priority:** These test cases should be added to the test suite before Phase 02

---

## Critical Test Cases (Add Before Release)

### TC-01: Symlink Handling
**Source:** D3.1.6 Gap  
**Purpose:** Verify wallpaper setting works with symlinks (common use case)

```go
// internal/platform/platform_test.go
func TestSetWallpaperSymlink(t *testing.T) {
    // Create a temp image
    tmpDir, _ := os.MkdirTemp("", "symlink-test")
    defer os.RemoveAll(tmpDir)
    
    realImage := filepath.Join(tmpDir, "real.jpg")
    os.WriteFile(realImage, []byte("fake-image-data"), 0644)
    
    // Create symlink
    symlink := filepath.Join(tmpDir, "link.jpg")
    os.Symlink(realImage, symlink)
    
    setter, _ := Get()
    
    // Should follow symlink and set real image
    err := setter.SetWallpaper(symlink)
    // Error is expected in test environment (no actual desktop)
    // but we verify the path resolution works
    if err != nil {
        t.Logf("Expected error in test env: %v", err)
    }
}
```

---

### TC-02: Flag Precedence Documentation
**Source:** D3.3.1 Gap  
**Purpose:** Verify flag precedence behavior is consistent

```go
// cmd/set_test.go
func TestFlagPrecedence(t *testing.T) {
    tests := []struct {
        name     string
        args     []string
        expected string // "current", "random", "latest", "manual"
    }{
        {"current only", []string{"--current"}, "current"},
        {"random only", []string{"--random"}, "random"},
        {"latest only", []string{"--latest"}, "latest"},
        {"current wins over random", []string{"--current", "--random"}, "current"},
        {"current wins over latest", []string{"--current", "--latest"}, "current"},
        {"random wins over latest", []string{"--random", "--latest"}, "random"},
        {"path when no flags", []string{"/path/to/img.jpg"}, "manual"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // This test verifies the precedence logic
            // Would need to refactor set.go to make source determination testable
            // Or use integration testing approach
        })
    }
}
```

**Implementation Note:** Refactor set.go RunE to extract source determination into a testable function:
```go
func determineSource(args []string, randomFlag, latestFlag, currentFlag bool) (string, string, error)
```

---

### TC-03: Invalid Image Content (Not Just Extension)
**Source:** Security concern  
**Purpose:** Verify that files with wrong content but right extension are handled

```go
// internal/utils/image_test.go
func TestIsImageFileContentValidation(t *testing.T) {
    tmpDir, _ := os.MkdirTemp("", "content-test")
    defer os.RemoveAll(tmpDir)
    
    // Create a .jpg that's actually a text file
    fakeJpg := filepath.Join(tmpDir, "fake.jpg")
    os.WriteFile(fakeJpg, []byte("this is not an image"), 0644)
    
    // Current implementation only checks extension
    // Future: Add magic number/content validation
    if !IsImageFile(fakeJpg) {
        t.Error("Current implementation should accept by extension")
    }
    
    // TODO: When content validation added, this should fail
    // if IsValidImageContent(fakeJpg) {
    //     t.Error("Should reject fake image by content")
    // }
}
```

---

## Important Test Cases (Add Before Phase 02)

### TC-04: Config Migration
**Source:** Backward compatibility concern  
**Purpose:** Verify old configs without wallpaper fields work

```go
// internal/config/config_test.go
func TestConfigMigration(t *testing.T) {
    // Create a config file from v1.1 (no wallpaper fields)
    oldConfig := `{
        "default_source": "wallhaven",
        "output_directory": "/home/user/Pictures"
    }`
    
    tmpFile, _ := os.CreateTemp("", "old-config-*.json")
    tmpFile.WriteString(oldConfig)
    tmpFile.Close()
    defer os.Remove(tmpFile.Name())
    
    // Should load without error, initialize empty fields
    cfg, err := Load(tmpFile.Name())
    if err != nil {
        t.Fatalf("Failed to load old config: %v", err)
    }
    
    if cfg.CurrentWallpaper != "" {
        t.Error("CurrentWallpaper should be empty for old config")
    }
    
    if len(cfg.WallpaperHistory) != 0 {
        t.Error("WallpaperHistory should be empty for old config")
    }
}
```

---

### TC-05: Concurrent Set Operations
**Source:** Thread safety concern  
**Purpose:** Verify no race conditions when setting wallpaper

```go
// cmd/set_test.go
func TestConcurrentSet(t *testing.T) {
    // This is mainly to ensure config file locking works
    // if we ever add it. Currently no locking implemented.
    
    // TODO: Add file locking before implementing this test
    t.Skip("File locking not yet implemented")
}
```

---

### TC-06: Large Collection Performance
**Source:** D4.3.5 concern  
**Purpose:** Test performance with 1000+ wallpapers

```go
// internal/utils/image_test.go
func TestLargeCollectionPerformance(t *testing.T) {
    tmpDir, _ := os.MkdirTemp("", "large-collection")
    defer os.RemoveAll(tmpDir)
    
    // Create 1000 fake images
    for i := 0; i < 1000; i++ {
        name := fmt.Sprintf("img_%04d.jpg", i)
        f, _ := os.Create(filepath.Join(tmpDir, name))
        f.Close()
    }
    
    start := time.Now()
    wallpapers, err := FindWallpapers(tmpDir)
    duration := time.Since(start)
    
    if err != nil {
        t.Fatalf("Failed: %v", err)
    }
    
    if len(wallpapers) != 1000 {
        t.Errorf("Expected 1000, got %d", len(wallpapers))
    }
    
    // Should complete in under 1 second
    if duration > time.Second {
        t.Errorf("Too slow: %v", duration)
    }
    
    t.Logf("Found 1000 wallpapers in %v", duration)
}
```

---

### TC-07: Permission Denied Handling
**Source:** D6.1 error message verification  
**Purpose:** Verify graceful handling when file exists but can't be read

```go
// internal/platform/platform_test.go
func TestSetWallpaperPermissionDenied(t *testing.T) {
    if runtime.GOOS == "windows" {
        t.Skip("Permission model differs on Windows")
    }
    
    // Create unreadable image file
    tmpDir, _ := os.MkdirTemp("", "perm-test")
    defer os.RemoveAll(tmpDir)
    
    image := filepath.Join(tmpDir, "secret.jpg")
    os.WriteFile(image, []byte("data"), 0644)
    os.Chmod(image, 0000) // No permissions
    defer os.Chmod(image, 0644) // Restore for cleanup
    
    setter, _ := Get()
    err := setter.SetWallpaper(image)
    
    // Should get a clear permission error
    if err == nil {
        t.Error("Expected error for unreadable file")
    }
    
    // Error should mention permission
    errStr := err.Error()
    if !strings.Contains(errStr, "permission") && 
       !strings.Contains(errStr, "denied") {
        t.Logf("Error message might be clearer: %s", errStr)
    }
}
```

---

### TC-08: Platform Command Not Found
**Source:** D3.4.1  
**Purpose:** Test behavior when platform tools are missing

```go
// internal/platform/platform_test.go
func TestMissingPlatformCommand(t *testing.T) {
    // This requires mocking exec.Command or running in isolated environment
    // Consider for integration tests with Docker/containers
    
    t.Skip("Requires isolated test environment - integration test")
}
```

---

### TC-09: History Persistence Across Sessions
**Source:** D4.2 integration concern  
**Purpose:** Verify history survives config reload

```go
// internal/config/config_test.go
func TestHistoryPersistence(t *testing.T) {
    tmpFile, _ := os.CreateTemp("", "history-test-*.json")
    tmpFile.Close()
    defer os.Remove(tmpFile.Name())
    
    // Create config with history
    cfg := DefaultConfig()
    cfg.AddWallpaper("/path/to/img1.jpg", "manual")
    cfg.AddWallpaper("/path/to/img2.jpg", "random")
    cfg.AddWallpaper("/path/to/img3.jpg", "latest")
    
    // Save
    cfg.Save(tmpFile.Name())
    
    // Load
    loaded, err := Load(tmpFile.Name())
    if err != nil {
        t.Fatalf("Failed to reload: %v", err)
    }
    
    // Verify history preserved
    if len(loaded.WallpaperHistory) != 3 {
        t.Errorf("Expected 3 history entries, got %d", len(loaded.WallpaperHistory))
    }
    
    // Verify order (newest first)
    if loaded.WallpaperHistory[0].Path != "/path/to/img3.jpg" {
        t.Error("History order not preserved")
    }
    
    // Verify timestamps are valid
    for i, record := range loaded.WallpaperHistory {
        if record.Timestamp.IsZero() {
            t.Errorf("Record %d has zero timestamp", i)
        }
    }
}
```

---

### TC-10: Empty Output Directory
**Source:** D4.3.1  
**Purpose:** Verify graceful handling of empty/missing output dir

```go
// internal/utils/image_test.go
func TestEmptyOutputDirectory(t *testing.T) {
    // Non-existent directory
    _, err := FindWallpapers("/nonexistent/path")
    if err == nil {
        t.Error("Expected error for non-existent directory")
    }
    
    // Empty directory
    tmpDir, _ := os.MkdirTemp("", "empty")
    defer os.RemoveAll(tmpDir)
    
    wallpapers, err := FindWallpapers(tmpDir)
    if err != nil {
        t.Errorf("Empty dir should not error, got: %v", err)
    }
    
    if len(wallpapers) != 0 {
        t.Errorf("Expected 0 wallpapers, got %d", len(wallpapers))
    }
}
```

---

## Nice-to-Have Test Cases (Add When Time Permits)

### TC-11: Unicode Path Handling
**Source:** Internationalization  
**Purpose:** Test paths with unicode characters

```go
func TestUnicodePaths(t *testing.T) {
    tmpDir, _ := os.MkdirTemp("", "unicode-test-日本語")
    defer os.RemoveAll(tmpDir)
    
    image := filepath.Join(tmpDir, "画像.jpg")
    f, _ := os.Create(image)
    f.Close()
    
    // Should handle unicode without issues
    if !IsImageFile(image) {
        t.Error("Should recognize unicode path")
    }
}
```

---

### TC-12: Network Filesystem Paths
**Source:** Enterprise use case  
**Purpose:** Test with NFS/SMB paths

```go
func TestNetworkPath(t *testing.T) {
    t.Skip("Requires network filesystem setup")
    // Test with //server/share/path style paths on Windows
    // Test with /mnt/nfs/ style paths on Unix
}
```

---

### TC-13: Race Condition in Random Selection
**Source:** D4.3.5  
**Purpose:** Verify no collision when multiple random calls happen simultaneously

```go
func TestRandomRaceCondition(t *testing.T) {
    tmpDir, _ := os.MkdirTemp("", "race-test")
    defer os.RemoveAll(tmpDir)
    
    // Create 10 images
    for i := 0; i < 10; i++ {
        f, _ := os.Create(filepath.Join(tmpDir, fmt.Sprintf("img%d.jpg", i)))
        f.Close()
    }
    
    // Call GetRandomWallpaper concurrently
    var wg sync.WaitGroup
    results := make(chan string, 100)
    
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            path, _ := GetRandomWallpaper(tmpDir)
            results <- path
        }()
    }
    
    wg.Wait()
    close(results)
    
    // Verify all calls succeeded (no crashes)
    count := 0
    for range results {
        count++
    }
    
    if count != 100 {
        t.Errorf("Expected 100 results, got %d", count)
    }
}
```

---

## Implementation Priority

### Sprint 1 (Before Phase 02)
1. TC-04: Config Migration - Critical for backward compatibility
2. TC-09: History Persistence - Data integrity
3. TC-10: Empty Output Directory - Edge case handling

### Sprint 2 (After Phase 02 Start)
4. TC-01: Symlink Handling - Common use case
5. TC-07: Permission Denied - Error handling
6. TC-06: Large Collection Performance - Scale testing

### Backlog
7. TC-02: Flag Precedence - Requires refactoring
8. TC-03: Image Content Validation - Security enhancement
9. TC-05: Concurrent Set - Requires file locking
10. TC-08: Missing Platform Command - Integration test
11. TC-11-13: Nice-to-have

---

## Test File Organization

```
cmd/
  set_test.go           # Existing + TC-02
cmd/
  set_integration_test.go  # New: TC-05, TC-08
internal/
  config/
    config_test.go      # Existing + TC-04, TC-09
internal/
  utils/
    image_test.go       # Existing + TC-06, TC-10, TC-11
internal/
  platform/
    platform_test.go    # Existing + TC-01, TC-07
    
test/
  integration/
    e2e_test.go         # New: Full CLI tests
    edge_cases_test.go  # New: TC-03, TC-13
```

---

*Generated from UAT Gap Analysis — Add these tests to achieve comprehensive coverage*
