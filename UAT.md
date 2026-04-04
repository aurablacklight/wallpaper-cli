# UAT Guide: Wallpaper CLI Tool

**Version:** dev (M001 Complete)  
**Date:** 2026-04-04  
**Tester:** _______________

---

## Pre-Flight Checklist

- [ ] Binary downloaded/built for your platform
- [ ] Terminal open with `wallpaper-cli` accessible
- [ ] Test directory ready (will create `~/test-wallpapers`)
- [ ] Internet connection active
- [ ] At least 100MB free disk space

---

## Test 1: Basic Installation & Help
**Purpose:** Verify CLI is working and self-documenting

```bash
# 1.1 Check version
./wallpaper-cli --version
# Expected: wallpaper-cli version dev (commit: unknown, built: ...)

# 1.2 Check help
./wallpaper-cli --help
# Expected: Shows available commands (fetch, config, list, etc.)

# 1.3 Check fetch command help
./wallpaper-cli fetch --help
# Expected: Shows all flags (--source, --resolution, --limit, etc.)
```

**Pass Criteria:**
- [ ] Version displays correctly
- [ ] Help shows all commands
- [ ] Fetch flags are documented

---

## Test 2: Config Management
**Purpose:** Verify config file creation and management

```bash
# 2.1 Create config
./wallpaper-cli config init
# Expected: "Created default config at: /Users/[name]/.config/wallpaper-cli/config.json"

# 2.2 List config
./wallpaper-cli config list
# Expected: JSON output with default_source, output_directory, etc.

# 2.3 Get a specific value
./wallpaper-cli config get output_directory
# Expected: Shows your home path + /Pictures/wallpapers

# 2.4 Set a value
./wallpaper-cli config set default_resolution 1440p
./wallpaper-cli config get default_resolution
# Expected: "1440p"

# 2.5 Verify config file exists
cat ~/.config/wallpaper-cli/config.json
# Expected: Valid JSON with your changes
```

**Pass Criteria:**
- [ ] Config file created
- [ ] List shows valid JSON
- [ ] Get/Set commands work
- [ ] Changes persist to file

---

## Test 3: Dry Run Fetch
**Purpose:** Test filtering without downloading

```bash
# 3.1 Basic dry run
./wallpaper-cli fetch --limit 5 --dry-run
# Expected: Shows 5 random 4K wallpapers (default resolution)

# 3.2 Dry run with tags
./wallpaper-cli fetch --tags "landscape,night" --limit 3 --dry-run
# Expected: Shows 3 landscape night wallpapers

# 3.3 Dry run with resolution
./wallpaper-cli fetch --resolution 1920x1080 --limit 3 --dry-run
# Expected: Shows 3 1080p wallpapers

# 3.4 Dry run with aspect ratio
./wallpaper-cli fetch --aspect-ratio 16:9 --limit 3 --dry-run
# Expected: Shows 3 16:9 aspect ratio wallpapers

# 3.5 Verify no files created
ls ~/Pictures/wallpapers/ 2>/dev/null || echo "No directory created (correct)"
# Expected: Directory doesn't exist yet (or is empty)
```

**Pass Criteria:**
- [ ] Dry run shows wallpaper list
- [ ] Tags filter works
- [ ] Resolution filter works
- [ ] Aspect ratio filter works
- [ ] No files downloaded

---

## Test 4: Actual Downloads
**Purpose:** Verify download pipeline works

```bash
# Create test directory
mkdir -p ~/test-wallpapers

# 4.1 Download 2 wallpapers
./wallpaper-cli fetch --output ~/test-wallpapers --limit 2 --resolution 4k
# Expected:
# - Progress messages showing downloads
# - 2 files in ~/test-wallpapers/wallhaven/
# - File sizes reported (e.g., "1.87 MB")

# 4.2 Verify files exist
ls -lh ~/test-wallpapers/wallhaven/
# Expected: 2 .jpg files with wallhaven IDs in names

# 4.3 Check file integrity
file ~/test-wallpapers/wallhaven/*.jpg
# Expected: "JPEG image data" for each file

# 4.4 Open an image to verify (optional)
open ~/test-wallpapers/wallhaven/*.jpg  # macOS
# or
xdg-open ~/test-wallpapers/wallhaven/*.jpg  # Linux
```

**Pass Criteria:**
- [ ] 2 files downloaded successfully
- [ ] Files are valid JPEGs/PNGs
- [ ] Progress shown during download
- [ ] Files in correct subdirectory

---

## Test 5: Concurrent Downloads
**Purpose:** Verify parallel downloading works

```bash
# Clean test directory
rm -rf ~/test-concurrent

# 5.1 Download 10 wallpapers with 5 concurrent
./wallpaper-cli fetch --output ~/test-concurrent --limit 10 --concurrent 5 --resolution 4k
# Expected:
# - Multiple "Starting download..." messages appear quickly
# - Downloads complete in parallel (not sequentially)
# - Total time less than 10 sequential downloads

# 5.2 Verify all 10 downloaded
ls ~/test-concurrent/wallhaven/ | wc -l
# Expected: 10
```

**Pass Criteria:**
- [ ] 10 files downloaded
- [ ] Downloads happen concurrently (not one-by-one)
- [ ] No errors during parallel downloads

---

## Test 6: Deduplication
**Purpose:** Verify pHash deduplication works

```bash
# Clean test directory
rm -rf ~/test-dedup

# 6.1 Download 3 wallpapers
./wallpaper-cli fetch --output ~/test-dedup --limit 3 --resolution 4k
# Expected: 3 successful downloads

# 6.2 Check database was created
ls ~/.local/share/wallpaper-cli/wallpapers.db
# Expected: File exists

# 6.3 Try downloading same 3 again (will get different ones from API)
./wallpaper-cli fetch --output ~/test-dedup --limit 3 --resolution 4k
# Expected: 
# - May download new ones OR
# - Show "Skipped (duplicate)" if same images returned

# 6.4 Check for skip messages
grep -r "duplicate" ~/test-dedup 2>/dev/null || echo "No duplicates in this test run (API returned different images)"
# Note: This is probabilistic - API returns random results
```

**Pass Criteria:**
- [ ] Database file created
- [ ] Second run either skips or downloads new (both valid)
- [ ] No duplicate files in directory

---

## Test 7: Organization Modes
**Purpose:** Test different organization strategies

```bash
# Test by source (default)
rm -rf ~/test-org
./wallpaper-cli fetch --output ~/test-org --limit 2 --organize-by source --resolution 4k
ls ~/test-org/
# Expected: "wallhaven" subdirectory

# Test by date
rm -rf ~/test-org
./wallpaper-cli fetch --output ~/test-org --limit 2 --organize-by date --resolution 4k
ls ~/test-org/
# Expected: "2026" (year) subdirectory
ls ~/test-org/2026/*/
# Expected: "04" (month) subdirectory with images

# Test by tags
rm -rf ~/test-org
./wallpaper-cli fetch --output ~/test-org --limit 2 --organize-by tags --tags "anime" --resolution 4k
ls ~/test-org/
# Expected: Subdirectory named after first tag (e.g., "anime", "tagme", etc.)
```

**Pass Criteria:**
- [ ] Source organization works
- [ ] Date organization works (year/month)
- [ ] Tag organization works
- [ ] Files in correct subdirectories

---

## Test 8: Error Handling
**Purpose:** Verify graceful error handling

```bash
# 8.1 Invalid resolution
./wallpaper-cli fetch --resolution invalid
# Expected: Error message "invalid resolution: invalid (expected: 1080p, 1440p, 4k, 8k, or WxH)"

# 8.2 Invalid source
./wallpaper-cli fetch --source invalid
# Expected: Error message "invalid source: invalid (expected: wallhaven, reddit, all)"

# 8.3 Invalid limit (too high)
./wallpaper-cli fetch --limit 10000
# Expected: Error message "invalid limit: 10000 (expected: 1-1000)"

# 8.4 Invalid organize-by
./wallpaper-cli fetch --organize-by invalid
# Expected: Error message "invalid organize-by: invalid (expected: source, tags, date)"

# 8.5 Non-existent output directory (should be created)
rm -rf ~/deep/nested/path
./wallpaper-cli fetch --output ~/deep/nested/path --limit 1 --dry-run
# Expected: No error (directory will be created on actual download)
```

**Pass Criteria:**
- [ ] All invalid inputs caught
- [ ] Helpful error messages
- [ ] No crashes or panics

---

## Test 9: Full Integration Test
**Purpose:** End-to-end workflow

```bash
# 9.1 Clean slate
rm -rf ~/test-final
rm -f ~/.local/share/wallpaper-cli/wallpapers.db

# 9.2 Configure for this test
./wallpaper-cli config init 2>/dev/null || true
./wallpaper-cli config set default_resolution 4k
./wallpaper-cli config set output_directory ~/test-final

# 9.3 Fetch anime 4K wallpapers
./wallpaper-cli fetch --tags "anime" --limit 5 --organize-by source --anime

# 9.4 Verify results
echo "=== Files Downloaded ==="
find ~/test-final -type f -name "*.jpg" -o -name "*.png" | head -5

echo "=== Database Contents ==="
ls -lh ~/.local/share/wallpaper-cli/wallpapers.db

echo "=== File Sizes ==="
du -sh ~/test-final/*
```

**Pass Criteria:**
- [ ] 5 wallpapers downloaded
- [ ] Files organized in wallhaven/ subdirectory
- [ ] Database created with records
- [ ] All files valid images

---

## Test 10: Cross-Platform Build Verification (Optional)
**Purpose:** If you have multiple systems

| Platform | Command | Expected Result |
|----------|---------|-----------------|
| macOS (Intel) | `./wallpaper-cli-darwin-amd64 --version` | Version displays |
| macOS (Apple Silicon) | `./wallpaper-cli-darwin-arm64 --version` | Version displays |
| Linux | `./wallpaper-cli-linux-amd64 --version` | Version displays |
| Windows | `wallpaper-cli-windows-amd64.exe --version` | Version displays |

---

## Binary Size Verification

```bash
# Check binary size
ls -lh ./wallpaper-cli
# Expected: < 20MB (actual: ~11MB)

# On macOS, check it's a universal binary or specific arch
file ./wallpaper-cli
# Expected: "Mach-O 64-bit executable arm64" (Apple Silicon)
# or "Mach-O 64-bit executable x86_64" (Intel)
```

**Pass Criteria:**
- [ ] Binary under 20MB
- [ ] Correct architecture for your system

---

## Final Sign-Off

| Requirement | Status |
|-------------|--------|
| Single binary < 20MB | [ ] Pass [ ] Fail |
| Cross-platform builds work | [ ] Pass [ ] Fail |
| Wallhaven source fetches | [ ] Pass [ ] Fail |
| Resolution filtering works | [ ] Pass [ ] Fail |
| Concurrent downloads (5 parallel) | [ ] Pass [ ] Fail |
| Deduplication prevents re-downloads | [ ] Pass [ ] Fail |
| Organization by source/date/tags | [ ] Pass [ ] Fail |
| Config management works | [ ] Pass [ ] Fail |
| Error messages are helpful | [ ] Pass [ ] Fail |
| Progress reporting during downloads | [ ] Pass [ ] Fail |

**Overall UAT Result:** [ ] PASS / [ ] FAIL

**Tester Signature:** _______________  **Date:** _______________

---

## Quick Smoke Test (5 minutes)

If you only have 5 minutes, run this:

```bash
./wallpaper-cli --version
./wallpaper-cli fetch --limit 2 --output ~/smoke-test --resolution 4k
ls ~/smoke-test/wallhaven/
```

**Expected:** 2 files downloaded successfully.

---

## Troubleshooting

### Issue: "command not found: ./wallpaper-cli"
**Solution:** Make sure you're in the correct directory:
```bash
cd /Users/derek/code_projects/wallpaper-cli-tool
./wallpaper-cli --help
```

### Issue: "permission denied"
**Solution:** Make binary executable:
```bash
chmod +x ./wallpaper-cli
```

### Issue: Downloads fail / timeouts
**Solution:** Check internet connection, try with smaller limit:
```bash
./wallpaper-cli fetch --limit 1 --output ~/test --timeout 60
```

### Issue: Database errors
**Solution:** Clear database and retry:
```bash
rm -rf ~/.local/share/wallpaper-cli/
./wallpaper-cli fetch --limit 2 --output ~/test
```
