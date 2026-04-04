# S03: Wallhaven Source Adapter

**Goal:** Wallhaven API adapter with search, filtering, and metadata extraction

**Success Criteria:**
- Wallhaven API v1 /search endpoint integration
- Query parameter construction (q, resolutions, ratios, sorting)
- JSON response parsing into Go structs
- Resolution pre-filtering (check before download)
- Tag extraction and filtering
- Pagination handling for large result sets

---

## Integration Closure

CLI fetch command can query Wallhaven and return filtered results

## Observability Impact

API response times and error rates observable

## Proof Level

L2 - Integration complexity

---

## Dependencies

- S02: CLI Interface & Config System

---

## Risk

Medium - API integration with query parameter complexity

---

## Demo

Fetch 10 wallpapers from Wallhaven with 4k filter, resolution verified pre-download

---

## Tasks

| ID | Title | Est. | Input | Output | Files Touched |
|----|-------|------|-------|--------|---------------|
| T01 | Define Wallhaven API types | 20m | Spec | internal/sources/wallhaven/types.go | internal/sources/wallhaven/types.go |
| T02 | Implement API client | 45m | T01 | internal/sources/wallhaven/client.go | internal/sources/wallhaven/client.go |
| T03 | Implement search/query building | 30m | T02 | Query builder | internal/sources/wallhaven/query.go |
| T04 | Add response parsing | 20m | T01 | Parser | internal/sources/wallhaven/parser.go |
| T05 | Add pagination support | 20m | T02 | Pagination logic | client.go updates |
| T06 | Integrate with CLI | 15m | T02 | cmd/fetch.go integration | cmd/fetch.go |

---

## Plan

### T01: Define Wallhaven API Types
**Estimate:** 20m

**Description:**
Define Go structs matching the Wallhaven API v1 response schema.

**Steps:**
1. Study Wallhaven API v1 documentation
2. Create types.go with:
   - SearchResponse struct
   - Wallpaper struct (id, url, resolution, tags, etc.)
   - Tag struct
   - Meta struct (pagination info)
3. Add JSON tags for unmarshaling

**Key API Fields:**
- id, url, short_url
- views, favorites, source
- purity (sfw/sketchy/nsfw)
- category (general/anime/people)
- dimension_x, dimension_y (resolution)
- resolution (string like "3840x2160")
- ratio (string like "16x9")
- tags array

**Files Likely Touched:**
- internal/sources/wallhaven/types.go

**Expected Output:**
- Go structs that match API response
- JSON unmarshaling works

**Verification:**
```bash
go build ./...
go test ./internal/sources/wallhaven/...
```

---

### T02: Implement API Client
**Estimate:** 45m

**Description:**
Create HTTP client for Wallhaven API with proper configuration.

**Steps:**
1. Create client.go with Client struct
2. Add NewClient constructor with:
   - HTTP client with timeout
   - Base URL (https://wallhaven.cc/api/v1/)
   - Optional API key support
   - User-Agent header
3. Implement Search method
4. Add error handling (HTTP status, rate limiting)
5. Add retry logic with backoff

**Files Likely Touched:**
- internal/sources/wallhaven/client.go

**Expected Output:**
- Client can make API requests
- Proper error handling
- Configurable timeouts

**Verification:**
```go
// Test in code
client := wallhaven.NewClient(nil)
resp, err := client.Search(context.Background(), "landscape")
```

---

### T03: Implement Search/Query Building
**Estimate:** 30m

**Description:**
Implement query parameter construction for search requests.

**Steps:**
1. Create query.go with SearchOptions struct
2. Map CLI flags to API parameters:
   - q → tags query
   - resolutions → exact resolution filter
   - ratios → aspect ratio filter
   - sorting → random, relevance, etc.
   - page → pagination
3. Add URL encoding
4. Validate parameters before sending

**Parameter Mapping:**
| CLI Flag | API Param | Example |
|----------|-----------|---------|
| --tags | q | "landscape night" |
| --resolution | resolutions | "3840x2160" |
| --aspect-ratio | ratios | "16x9" |
| --limit (pages) | page | 1, 2, 3... |

**Files Likely Touched:**
- internal/sources/wallhaven/query.go

**Expected Output:**
- Query parameters correctly formatted
- URL encoding handles special characters

**Verification:**
```bash
# Build and test
./wallpaper-cli fetch --source wallhaven --tags "landscape" --resolution 4k --dry-run
# Should show query URL
```

---

### T04: Add Response Parsing
**Estimate:** 20m

**Description:**
Parse API responses into usable wallpaper metadata.

**Steps:**
1. Add parsing logic to extract key fields
2. Convert API types to internal model
3. Validate required fields
4. Handle missing/optional fields gracefully

**Internal Model:**
```go
type WallpaperInfo struct {
    ID         string
    Source     string // "wallhaven"
    SourceID   string // wallhaven ID
    URL        string // direct image URL
    Resolution string // "3840x2160"
    AspectRatio string // "16:9"
    Tags       []string
    FileSize   int64  // if available
}
```

**Files Likely Touched:**
- internal/sources/wallhaven/parser.go

**Expected Output:**
- Clean internal representation
- All metadata extracted

**Verification:**
```go
// Unit test with sample JSON
```

---

### T05: Add Pagination Support
**Estimate:** 20m

**Description:**
Handle fetching multiple pages of results.

**Steps:**
1. Parse meta.last_page from response
2. Add iterator/paginator pattern
3. Implement automatic page fetching until limit reached
4. Stop early when --limit is satisfied

**Files Likely Touched:**
- internal/sources/wallhaven/client.go (pagination methods)

**Expected Output:**
- Can fetch >24 images (API page size)
- Respects --limit flag
- Efficient pagination

**Verification:**
```bash
./wallpaper-cli fetch --limit 50 --source wallhaven
# Should fetch ~3 pages and stop at 50
```

---

### T06: Integrate with CLI
**Estimate:** 15m

**Description:**
Wire up the Wallhaven adapter to the fetch command.

**Steps:**
1. Update cmd/fetch.go run function
2. Create Wallhaven client with config
3. Call Search with parsed options
4. Print results (or pass to download manager)
5. Handle errors gracefully

**Files Likely Touched:**
- cmd/fetch.go

**Expected Output:**
- `wallpaper-cli fetch --source wallhaven` returns wallpapers
- Results displayed or passed to download

**Verification:**
```bash
./wallpaper-cli fetch --source wallhaven --tags "anime" --limit 5 --dry-run
# Should show 5 wallpaper URLs that would be downloaded
```
