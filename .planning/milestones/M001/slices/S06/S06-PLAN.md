# S06: Organization & Storage

**Goal:** File organization with directory structure and metadata storage

**Success Criteria:**
- Output directory creation with proper permissions
- Organize-by modes: source, tags, date
- File naming with source ID and resolution
- Metadata storage in SQLite (source, tags, resolution, downloaded_at)
- Tag-based directory structure (if organize-by=tags)

---

## Integration Closure

End-to-end fetch → download → dedup → organize flow working

## Observability Impact

Storage usage, file counts per category observable in DB

## Proof Level

L1 - Core foundations

---

## Dependencies

- S05: Deduplication System

---

## Risk

Low - straightforward file operations

---

## Demo

Images organized by source in ~/Pictures/wallpapers/, metadata in SQLite

---

## Tasks

| ID | Title | Est. | Input | Output | Files Touched |
|----|-------|------|-------|--------|---------------|
| T01 | Design organization strategies | 20m | Spec | internal/organize/strategy.go | internal/organize/strategy.go |
| T02 | Implement source-based organization | 20m | T01 | By source | internal/organize/source.go |
| T03 | Implement tag-based organization | 25m | T01 | By tags | internal/organize/tags.go |
| T04 | Implement date-based organization | 15m | T01 | By date | internal/organize/date.go |
| T05 | File naming conventions | 15m | Spec | Naming logic | internal/organize/naming.go |
| T06 | Integrate with download manager | 15m | T02-T05 | Integration | internal/download/manager.go |

---

## Plan

### T01: Design Organization Strategies
**Estimate:** 20m

**Description:**
Design the organization strategy pattern for flexible file placement.

**Strategy Pattern:**
```go
type Strategy interface {
    GetPath(wallpaper WallpaperInfo, baseDir string) (string, error)
}
```

**Implementations:**
- SourceStrategy: organize by source (wallhaven/, reddit/)
- TagStrategy: organize by primary tag (landscape/, night/, anime/)
- DateStrategy: organize by download date (2024/01/, 2024/02/)

**Files Likely Touched:**
- internal/organize/strategy.go (interface)

**Expected Output:**
- Strategy interface defined
- Factory function for strategy selection

---

### T02: Implement Source-Based Organization
**Estimate:** 20m

**Description:**
Implement organization by source (wallhaven, reddit, etc.).

**Structure:**
```
~/Pictures/wallpapers/
├── wallhaven/
│   ├── landscape_12345_3840x2160.jpg
│   └── anime_67890_1920x1080.jpg
└── reddit/
    ├── r-Animewallpaper_abc123_3840x2160.jpg
```

**Steps:**
1. Create SourceStrategy struct
2. Implement GetPath method
3. Sanitize source names for filesystem

**Files Likely Touched:**
- internal/organize/source.go

**Expected Output:**
- Files organized in source subdirectories
- Clean folder names

**Verification:**
```bash
./wallpaper-cli fetch --source wallhaven --organize-by source
ls ~/Pictures/wallpapers/wallhaven/
```

---

### T03: Implement Tag-Based Organization
**Estimate:** 25m

**Description:**
Implement organization by primary/most relevant tag.

**Structure:**
```
~/Pictures/wallpapers/
├── landscape/
│   ├── wallhaven_12345_3840x2160.jpg
│   └── wallhaven_67890_3840x2160.jpg
├── night/
│   └── wallhaven_54321_1920x1080.jpg
└── unsorted/
    └── wallhaven_99999_1920x1080.jpg  # no tags
```

**Steps:**
1. Create TagStrategy struct
2. Implement tag selection logic:
   - Prefer first tag, or
   - Most specific tag, or
   - User-configurable priority
3. Handle missing/empty tags
4. Sanitize tag names for filesystem

**Tag Sanitization:**
- Remove special characters
- Handle spaces (replace with _ or -)
- Limit length
- Lowercase for consistency

**Files Likely Touched:**
- internal/organize/tags.go

**Expected Output:**
- Files in tag-based folders
- Fallback for untagged images

**Verification:**
```bash
./wallpaper-cli fetch --tags "landscape" --organize-by tags
ls ~/Pictures/wallpapers/landscape/
```

---

### T04: Implement Date-Based Organization
**Estimate:** 15m

**Description:**
Implement organization by download date.

**Structure:**
```
~/Pictures/wallpapers/
├── 2024/
│   ├── 01-January/
│   │   └── wallhaven_12345_3840x2160.jpg
│   └── 02-February/
│       └── wallhaven_67890_1920x1080.jpg
```

**Steps:**
1. Create DateStrategy struct
2. Get current date at download time
3. Format path as YYYY/MM/ or YYYY/MM-Month/
4. Handle timezone (use local time)

**Files Likely Touched:**
- internal/organize/date.go

**Expected Output:**
- Files organized by year/month
- Predictable structure

**Verification:**
```bash
./wallpaper-cli fetch --organize-by date
ls ~/Pictures/wallpapers/$(date +%Y)/$(date +%m)/
```

---

### T05: File Naming Conventions
**Estimate:** 15m

**Description:**
Implement consistent file naming with source ID and resolution.

**Naming Pattern:**
```
{source}_{source_id}_{width}x{height}.{ext}
# Example: wallhaven_12345_3840x2160.jpg
```

**Steps:**
1. Create internal/organize/naming.go
2. Implement GenerateFilename function
3. Handle filename collisions (add _1, _2 suffix)
4. Preserve original extension
5. Sanitize all components

**Collision Handling:**
```go
// If wallhaven_12345_3840x2160.jpg exists:
// Try wallhaven_12345_3840x2160_1.jpg
// Then wallhaven_12345_3840x2160_2.jpg
```

**Files Likely Touched:**
- internal/organize/naming.go

**Expected Output:**
- Consistent, informative filenames
- No collisions
- Safe for all filesystems

**Verification:**
```bash
ls ~/Pictures/wallpapers/
# Should see: wallhaven_12345_3840x2160.jpg
```

---

### T06: Integrate with Download Manager
**Estimate:** 15m

**Description:**
Wire up organization to the download pipeline.

**Steps:**
1. Update download manager to:
   - Accept organize-by setting
   - Select appropriate strategy
   - Calculate final path before download
   - Create parent directories
2. Ensure atomic operations (temp → final)
3. Update DB with final path

**Integration Point:**
```go
strategy := organize.NewStrategy(organizeBy) // "source", "tags", "date"
finalPath, _ := strategy.GetPath(wallpaper, outputDir)
os.MkdirAll(filepath.Dir(finalPath), 0755)
// Download to temp
// Move to finalPath
```

**Files Likely Touched:**
- internal/download/manager.go (integrate organization)

**Expected Output:**
- Files saved in organized structure
- Organization mode respected
- Parent directories created

**Verification:**
```bash
./wallpaper-cli fetch --limit 5 --organize-by source
ls -la ~/Pictures/wallpapers/wallhaven/
# Should see 5 organized files
```
