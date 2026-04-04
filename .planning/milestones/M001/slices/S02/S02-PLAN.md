# S02: CLI Interface & Config System

**Goal:** Complete CLI command structure with all commands and flags from spec

**Success Criteria:**
- fetch, update, list, search, config, clean commands defined
- All flags from spec implemented (--source, --resolution, --aspect-ratio, --tags, --limit, --output, --organize-by, --format, --dedup)
- Config file read/write (JSON format)
- Path expansion (~ → home directory)
- Input validation with helpful error messages

---

## Integration Closure

CLI interface complete and usable, config file persistence working

## Observability Impact

CLI --help and --version provide surface for health checks

## Proof Level

L1 - Core foundations

---

## Dependencies

- S01: Project Foundation & CLI Scaffold

---

## Risk

Low - standard CLI flag handling

---

## Demo

wallpaper-cli fetch --help shows all flags, config file creation works

---

## Tasks

| ID | Title | Est. | Input | Output | Files Touched |
|----|-------|------|-------|--------|---------------|
| T01 | Implement fetch command | 45m | S01 | cmd/fetch.go | cmd/fetch.go |
| T02 | Implement config command | 30m | T01 | cmd/config.go, internal/config/config.go | internal/config/config.go |
| T03 | Implement update, list, search, clean commands | 30m | T01 | cmd/*.go stubs | cmd/update.go, cmd/list.go, cmd/search.go, cmd/clean.go |
| T04 | Add input validation | 20m | T01-T03 | Validation functions | internal/validate/*.go |
| T05 | Add path expansion | 15m | T02 | Path utils | internal/utils/path.go |

---

## Plan

### T01: Implement Fetch Command
**Estimate:** 45m

**Description:**
Implement the main fetch command with all flags from the spec. This is the primary command users will use.

**Steps:**
1. Create cmd/fetch.go
2. Define all flags:
   - --source (wallhaven, reddit, zerochan, all)
   - --resolution (1080p, 1440p, 4k, 8k, or WxH)
   - --aspect-ratio (16:9, 21:9, 32:9)
   - --tags (comma-separated)
   - --limit (max images)
   - --output (directory)
   - --organize-by (source, character, style, date)
   - --format (webp, jpg, png, original)
   - --dedup (boolean)
   - --concurrent (number)
3. Add --anime shorthand flag (sets anime defaults)
4. Bind flags to viper/cobra
5. Create fetch run function stub

**Files Likely Touched:**
- cmd/fetch.go

**Expected Output:**
- `wallpaper-cli fetch --help` shows all flags
- Flag parsing works correctly

**Verification:**
```bash
go build -o wallpaper-cli .
./wallpaper-cli fetch --help
./wallpaper-cli fetch --source wallhaven --resolution 4k --limit 10 --dry-run
```

---

### T02: Implement Config Command
**Estimate:** 30m

**Description:**
Implement the config command for managing user preferences. Config is stored in JSON format.

**Steps:**
1. Create internal/config/config.go for config management
2. Define Config struct matching spec JSON schema
3. Implement config file discovery (~/.config/wallpaper-cli/config.json)
4. Create cmd/config.go with subcommands:
   - config init (create default config)
   - config get <key>
   - config set <key> <value>
   - config list (show all)
5. Add path expansion for ~/ in output directory

**Files Likely Touched:**
- internal/config/config.go
- cmd/config.go

**Expected Output:**
- Config file can be created, read, and modified
- Path expansion works (~/Pictures → /home/user/Pictures)

**Verification:**
```bash
./wallpaper-cli config init
./wallpaper-cli config list
./wallpaper-cli config set default_resolution 4k
./wallpaper-cli config get default_resolution
```

---

### T03: Implement Update, List, Search, Clean Commands
**Estimate:** 30m

**Description:**
Create stub implementations for remaining CLI commands.

**Steps:**
1. cmd/update.go - Incremental update of existing collection
2. cmd/list.go - List downloaded wallpapers with filters
3. cmd/search.go - Search sources without downloading
4. cmd/clean.go - Remove duplicates and invalid files

Each should:
- Define appropriate flags
- Have help text explaining purpose
- Create stub run function that prints "not implemented"

**Files Likely Touched:**
- cmd/update.go
- cmd/list.go
- cmd/search.go
- cmd/clean.go

**Expected Output:**
- All commands appear in help
- Each has proper flag definitions
- Stub implementation in place

**Verification:**
```bash
./wallpaper-cli --help
./wallpaper-cli update --help
./wallpaper-cli list --help
```

---

### T04: Add Input Validation
**Estimate:** 20m

**Description:**
Add validation functions for user inputs to catch errors early with helpful messages.

**Steps:**
1. Create internal/validate/validate.go
2. Implement validators:
   - Resolution format (e.g., "3840x2160" or "4k")
   - Aspect ratio (e.g., "16:9")
   - Tags format
   - Directory path existence/writability
3. Integrate validation into command pre-run

**Files Likely Touched:**
- internal/validate/validate.go
- cmd/fetch.go (add pre-run validation)

**Expected Output:**
- Invalid inputs caught before execution
- Helpful error messages

**Verification:**
```bash
./wallpaper-cli fetch --resolution invalid
# Should show error: invalid resolution format
```

---

### T05: Add Path Expansion
**Estimate:** 15m

**Description:**
Implement tilde (~) expansion for home directory in paths.

**Steps:**
1. Create internal/utils/path.go
2. Implement ExpandPath function
3. Handle edge cases (empty path, already absolute, etc.)
4. Use in config file handling and flag parsing

**Files Likely Touched:**
- internal/utils/path.go
- internal/config/config.go (use ExpandPath)
- cmd/fetch.go (use ExpandPath for --output)

**Expected Output:**
- ~/Pictures expands correctly on all platforms
- Cross-platform (macOS, Linux, Windows)

**Verification:**
```bash
./wallpaper-cli config set output_directory "~/Pictures/Wallpapers"
./wallpaper-cli config get output_directory
# Should show expanded path
```
