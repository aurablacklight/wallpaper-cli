# S08: Reddit Source Adapter (Stretch)

**Goal:** Reddit source adapter as secondary wallpaper source

**Success Criteria:**
- Reddit HTTP API integration (OAuth2 if needed)
- r/Animewallpaper post parsing
- Image URL extraction from posts
- Cross-source deduplication (same image from different sources)
- Subreddit configuration in config file

---

## Integration Closure

Multi-source fetch (Wallhaven + Reddit) working together

## Observability Impact

Per-source success rates observable

## Proof Level

L2 - Integration complexity

---

## Dependencies

- S03: Wallhaven Source Adapter (for pattern reference)
- S06: Organization & Storage (for multi-source support)

---

## Risk

Medium - Reddit API rate limits and OAuth complexity

---

## Demo

Fetch from r/Animewallpaper, filter by resolution, no duplicates

---

## Tasks

| ID | Title | Est. | Input | Output | Files Touched |
|----|-------|------|-------|--------|---------------|
| T01 | Research Reddit API options | 30m | Spec | Research notes | Research notes |
| T02 | Implement Reddit client | 35m | T01 | internal/sources/reddit/client.go | internal/sources/reddit/client.go |
| T03 | Implement post parsing | 25m | T02 | Image URL extraction | internal/sources/reddit/parser.go |
| T04 | Add resolution filtering | 20m | T03 | Metadata extraction | internal/sources/reddit/filter.go |
| T05 | Cross-source deduplication | 20m | S05 | Source-agnostic hashing | internal/dedup/ |
| T06 | Integrate with fetch command | 15m | T02-T05 | cmd/fetch.go | cmd/fetch.go |

---

## Plan

### T01: Research Reddit API Options
**Estimate:** 30m

**Description:**
Research Reddit API options for fetching posts.

**Options:**
1. **Reddit JSON API** (Recommended for simple use)
   - No authentication required for read-only
   - URL: https://www.reddit.com/r/Animewallpaper.json
   - Rate limited (60 requests/minute per IP)
   - Simple JSON structure

2. **Reddit OAuth2 API** (if JSON API insufficient)
   - Requires app registration
   - Higher rate limits
   - More complex

3. **PRAW (Python only, not applicable)**

**Decision:** Use JSON API for simplicity, upgrade to OAuth if needed.

**Files Likely Touched:**
- Research notes only

**Expected Output:**
- API approach selected
- JSON structure understood

**Verification:**
```bash
curl -A "wallpaper-cli/1.0" https://www.reddit.com/r/Animewallpaper.json | head -100
```

---

### T02: Implement Reddit Client
**Estimate:** 35m

**Description:**
Create HTTP client for Reddit JSON API.

**Steps:**
1. Create internal/sources/reddit/client.go
2. Implement Client struct with:
   - HTTP client with timeout
   - Proper User-Agent (required by Reddit)
   - Rate limiting (respect 60 req/min)
3. Add GetPosts(subreddit, limit) method
4. Handle pagination (after parameter)
5. Add error handling

**API Structure:**
```go
type Post struct {
    Title string
    URL   string        // May be direct image or post link
    Permalink string    // Post URL
    Score int
    Created float64    // Unix timestamp
}
```

**Rate Limiting:**
- Track requests per minute
- Sleep if approaching limit
- Respect Retry-After header

**Files Likely Touched:**
- internal/sources/reddit/client.go
- internal/sources/reddit/types.go

**Expected Output:**
- Can fetch posts from r/Animewallpaper
- Respects rate limits

**Verification:**
```go
client := reddit.NewClient(nil)
posts, _ := client.GetPosts("Animewallpaper", 25)
// Returns 25 posts
```

---

### T03: Implement Post Parsing
**Estimate:** 25m

**Description:**
Parse Reddit posts to extract direct image URLs.

**Challenges:**
1. Post URLs may link to:
   - Direct image (i.redd.it, imgur.com direct)
   - Image hosting page (needs scraping)
   - Reddit gallery (multiple images)
   - External sites

2. Need to resolve to direct image URLs for downloading

**Steps:**
1. Create internal/sources/reddit/parser.go
2. Implement URL resolution:
   - i.redd.it: use directly
   - reddit galleries: extract media
   - imgur: convert page to direct link
   - others: attempt to find image
3. Return list of WallpaperInfo

**Supported Hosts:**
- i.redd.it (Reddit's image host)
- reddit.com/gallery/
- imgur.com (page → direct)
- Common direct links (jpg, png, webp)

**Files Likely Touched:**
- internal/sources/reddit/parser.go

**Expected Output:**
- Extract direct image URLs from posts
- Handle various hosting scenarios

**Verification:**
```bash
./wallpaper-cli fetch --source reddit --dry-run
# Shows resolved image URLs
```

---

### T04: Add Resolution Filtering
**Estimate:** 20m

**Description:**
Add pre-download resolution filtering for Reddit sources.

**Challenge:**
Reddit JSON API doesn't include image dimensions.

**Approaches:**
1. **HEAD request** (Recommended)
   - Request image with HEAD method
   - Check Content-Length
   - Some servers include dimensions in headers

2. **Quick download + check** (fallback)
   - Download first few KB
   - Parse image dimensions
   - Cancel if wrong resolution

3. **Skip pre-filter** (last resort)
   - Download all, filter after
   - Less efficient

**Steps:**
1. Implement HEAD request for dimension checking
2. Parse common image formats for dimensions
3. Filter before adding to download queue
4. Fallback to post-filter if HEAD fails

**Files Likely Touched:**
- internal/sources/reddit/filter.go

**Expected Output:**
- Resolution filtering works
- Efficient bandwidth usage

**Verification:**
```bash
./wallpaper-cli fetch --source reddit --resolution 4k --limit 10
# Only downloads 4K images
```

---

### T05: Cross-Source Deduplication
**Estimate:** 20m

**Description:**
Ensure deduplication works across different sources (e.g., same image on Wallhaven and Reddit).

**Already Implemented:**
- S05 deduplication uses perceptual hashing
- Should work across sources automatically

**Verification/Enhancement:**
1. Verify hash-based dedup catches cross-source duplicates
2. Add source tracking in DB (already done)
3. Add "first seen from" and "also seen on" fields

**DB Enhancement (Optional):**
```sql
-- Add sources_found field to track multiple sources
ALTER TABLE images ADD COLUMN sources_found TEXT; -- JSON array
```

**Files Likely Touched:**
- May need small updates to internal/dedup/checker.go
- internal/data/models.go (optional enhancement)

**Expected Output:**
- Same image from different sources detected as duplicate
- Source tracking in metadata

**Verification:**
```bash
# Fetch same image from both sources
./wallpaper-cli fetch --source wallhaven --limit 1
cp ~/.local/share/wallpaper-cli/wallpapers.db /tmp/db1
./wallpaper-cli fetch --source reddit --limit 1
# Second fetch should skip duplicate
```

---

### T06: Integrate with Fetch Command
**Estimate:** 15m

**Description:**
Wire up Reddit adapter to the fetch command.

**Steps:**
1. Update cmd/fetch.go:
   - Add "reddit" to --source choices
   - Handle reddit-specific config (subreddit list)
2. Add to config file:
   ```json
   "sources": {
     "reddit": {
       "enabled": true,
       "subreddits": ["Animewallpaper"]
     }
   }
   ```
3. Support multiple sources in one fetch (--source all)

**Multi-Source Fetch:**
```bash
./wallpaper-cli fetch --source all --limit 20
# Fetches from wallhaven + reddit combined
```

**Files Likely Touched:**
- cmd/fetch.go
- internal/config/config.go

**Expected Output:**
- `wallpaper-cli fetch --source reddit` works
- Multi-source fetch works

**Verification:**
```bash
./wallpaper-cli fetch --source reddit --limit 5
./wallpaper-cli fetch --source all --limit 10
./wallpaper-cli config set sources.reddit.subreddits '["Animewallpaper", "Moescape"]'
```
