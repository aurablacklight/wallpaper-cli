# S05: Deduplication System

**Goal:** Deduplication system with perceptual hashing and SQLite storage

**Success Criteria:**
- pHash library integrated (goimagehash or similar)
- Hash calculation for downloaded images
- SQLite schema for image metadata and hashes
- Hash lookup before download decision
- Similarity threshold for near-duplicates
- Cross-session persistence

---

## Integration Closure

Download pipeline prevents duplicates across sessions

## Observability Impact

Hash calculation time, DB query performance, hit rate observable

## Proof Level

L2 - Integration complexity

---

## Dependencies

- S04: Download Manager

---

## Risk

High - pHash implementation complexity, similarity threshold tuning

---

## Demo

Duplicate image detected via pHash, skipped on second fetch

---

## Tasks

| ID | Title | Est. | Input | Output | Files Touched |
|----|-------|------|-------|--------|---------------|
| T01 | Research pHash libraries | 30m | Spec | Library selection | Research notes |
| T02 | Integrate pHash library | 30m | T01 | internal/dedup/hash.go | internal/dedup/hash.go |
| T03 | Implement SQLite storage | 30m | Spec | internal/data/db.go | internal/data/db.go |
| T04 | Implement dedup checker | 25m | T02, T03 | internal/dedup/checker.go | internal/dedup/checker.go |
| T05 | Add pre-download hash check | 20m | T04 | Integration | internal/download/manager.go |
| T06 | Add dedup to CLI flags | 15m | T05 | Flag support | cmd/fetch.go |

---

## Plan

### T01: Research pHash Libraries
**Estimate:** 30m

**Description:**
Research and select a Go perceptual hashing library.

**Options to Evaluate:**
1. **goimagehash** (most popular)
   - GitHub: https://github.com/corona10/goimagehash
   - Implements aHash, dHash, pHash, wHash
   - Pure Go, no CGO
2. **imagehash** (alternative)
   - Different implementations
3. **Custom implementation**
   - High effort, not recommended

**Criteria:**
- Speed: <100ms per 4K image
- Accuracy: catches resized/cropped versions
- Dependencies: prefer pure Go
- License: compatible

**Files Likely Touched:**
- Research notes only

**Expected Output:**
- Selected library documented
- Proof of concept tested

**Verification:**
```bash
# Test with sample images
go run cmd/poc/main.go
# Time the hashing
```

---

### T02: Integrate pHash Library
**Estimate:** 30m

**Description:**
Integrate the selected pHash library and implement image hashing.

**Steps:**
1. Add dependency: `go get github.com/corona10/goimagehash`
2. Create internal/dedup/hash.go
3. Implement functions:
   - CalculateHash(imagePath) (hash, error)
   - CalculateHashFromReader(r) (hash, error) // for streaming
   - HashToString(hash) string // for storage
   - HashFromString(s) (hash, error)
4. Add distance/similarity calculation

**Hash Types:**
- pHash (perceptual) - best for visual similarity
- aHash (average) - faster, less accurate
- dHash (difference) - good for gradients

**Files Likely Touched:**
- internal/dedup/hash.go

**Expected Output:**
- Can hash images efficiently
- Similar images have similar hashes

**Verification:**
```go
// Unit test
hash1, _ := dedup.CalculateHash("test1.jpg")
hash2, _ := dedup.CalculateHash("test1_resized.jpg")
distance, _ := hash1.Distance(hash2)
// distance should be small (<10)
```

---

### T03: Implement SQLite Storage
**Estimate:** 30m

**Description:**
Implement SQLite storage for image metadata and hashes.

**Steps:**
1. Add dependency: `go get github.com/mattn/go-sqlite3` or `modernc.org/sqlite` (CGO-free)
2. Create internal/data/db.go
3. Define schema (from SPEC.md):
   ```sql
   CREATE TABLE images (
       id INTEGER PRIMARY KEY,
       hash TEXT UNIQUE NOT NULL,
       source TEXT NOT NULL,
       source_id TEXT,
       url TEXT NOT NULL,
       local_path TEXT,
       resolution TEXT,
       aspect_ratio TEXT,
       tags TEXT,  -- JSON array
       downloaded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
       file_size INTEGER
   );
   CREATE INDEX idx_hash ON images(hash);
   ```
4. Implement CRUD operations
5. Add database initialization/migration

**Database Location:**
- `~/.local/share/wallpaper-cli/wallpapers.db`
- Or `~/.config/wallpaper-cli/wallpapers.db`

**Files Likely Touched:**
- internal/data/db.go
- internal/data/schema.go
- internal/data/models.go

**Expected Output:**
- Database created automatically
- CRUD operations working
- Schema migrations in place

**Verification:**
```go
db, _ := data.OpenDB("~/.local/share/wallpaper-cli/wallpapers.db")
db.SaveImage(img)
img, _ := db.GetImageByHash("a1b2c3...")
```

---

### T04: Implement Deduplication Checker
**Estimate:** 25m

**Description:**
Implement the deduplication logic with similarity threshold.

**Steps:**
1. Create internal/dedup/checker.go
2. Implement Checker struct with DB reference
3. Add IsDuplicate(hash) (bool, existingPath, error)
4. Add FindSimilar(hash, threshold) ([]matches, error)
5. Add AddImage(img) error
6. Configure similarity threshold (default: 10)

**Algorithm:**
```go
func (c *Checker) IsDuplicate(hash ImageHash) (bool, string, error) {
    // 1. Check exact hash match
    // 2. Check similar hashes within threshold
    // 3. Return match info
}
```

**Files Likely Touched:**
- internal/dedup/checker.go
- internal/dedup/config.go

**Expected Output:**
- Can check if image is duplicate
- Configurable similarity threshold
- Returns existing file path for reference

**Verification:**
```go
checker := dedup.NewChecker(db, 10) // threshold 10
isDup, path, _ := checker.IsDuplicate(newHash)
// isDup = true if similar image exists
```

---

### T05: Add Pre-Download Hash Check
**Estimate:** 20m

**Description:**
Integrate deduplication check into the download pipeline.

**Steps:**
1. Modify download manager to:
   - Check hash before downloading
   - Skip if duplicate (with log message)
   - Hash after downloading (if not skipped)
   - Store hash in DB after successful download
2. Add "checking for duplicates" progress message
3. Handle hash calculation errors gracefully

**Download Flow:**
```
1. Get wallpaper from source
2. Check DB by source_id (fast path)
3. Download to temp file
4. Calculate hash
5. Check if hash exists in DB
6. If duplicate: delete temp, log skip
7. If new: move to final location, save to DB
```

**Files Likely Touched:**
- internal/download/manager.go (integrate dedup)
- internal/download/job.go (add hash check)

**Expected Output:**
- Duplicate images are not re-downloaded
- Log shows "skipping duplicate: <path>"
- New images saved and recorded

**Verification:**
```bash
# First fetch - downloads all
./wallpaper-cli fetch --limit 5 --output ~/test

# Second fetch - should skip all as duplicates
./wallpaper-cli fetch --limit 5 --output ~/test
# Output: "5/5 skipped (duplicates)"
```

---

### T06: Add Dedup to CLI Flags
**Estimate:** 15m

**Description:**
Wire up the --dedup flag to control deduplication behavior.

**Steps:**
1. Update cmd/fetch.go:
   - --dedup flag (default: true)
   - --dedup-threshold flag (default: 10)
2. Pass dedup config to download manager
3. Support --no-dedup (explicit disable)
4. Add config file options for dedup

**Files Likely Touched:**
- cmd/fetch.go
- internal/config/config.go (add dedup settings)

**Expected Output:**
- --dedup flag works
- Configurable threshold
- Can disable with --no-dedup

**Verification:**
```bash
./wallpaper-cli fetch --help | grep dedup
./wallpaper-cli fetch --no-dedup --limit 5
# Should re-download even if duplicates exist
```
