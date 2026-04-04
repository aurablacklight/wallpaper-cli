package thumbs

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/nfnt/resize"
)

const (
	ThumbnailSizeSmall  = 128  // For list view
	ThumbnailSizeMedium = 256  // For detailed view
	CacheDirName        = "thumbs"
	MetadataFile        = "cache.json"
	MaxConcurrency      = 4
)

// CacheEntry tracks metadata for a cached thumbnail
type CacheEntry struct {
	OriginalPath string    `json:"original_path"`
	ModTime      time.Time `json:"mod_time"`
	ThumbnailPath string   `json:"thumbnail_path"`
	Size         int64     `json:"size"`
}

// Cache manages thumbnail generation and caching
type Cache struct {
	mu       sync.RWMutex
	entries  map[string]CacheEntry
	cacheDir string
}

// NewCache creates a new thumbnail cache
func NewCache() (*Cache, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting home directory: %w", err)
	}

	cacheDir := filepath.Join(home, ".cache", "wallpaper-cli", CacheDirName)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("creating cache directory: %w", err)
	}

	c := &Cache{
		entries:  make(map[string]CacheEntry),
		cacheDir: cacheDir,
	}

	// Load existing metadata
	if err := c.loadMetadata(); err != nil {
		// Log but don't fail - we'll regenerate thumbnails
		fmt.Fprintf(os.Stderr, "Warning: failed to load thumbnail metadata: %v\n", err)
	}

	return c, nil
}

// Generate creates a thumbnail for an image (default size: 256x256)
func (c *Cache) Generate(imagePath string) (string, error) {
	return c.GenerateWithSize(imagePath, ThumbnailSizeMedium)
}

// GenerateWithSize creates a thumbnail with specified size
func (c *Cache) GenerateWithSize(imagePath string, size int) (string, error) {
	// Check if valid image
	if !isImageFile(imagePath) {
		return "", fmt.Errorf("not an image file: %s", imagePath)
	}

	// Get file info
	info, err := os.Stat(imagePath)
	if err != nil {
		return "", fmt.Errorf("stat image: %w", err)
	}

	// Generate cache key from path and size
	cacheKey := generateCacheKeyWithSize(imagePath, size)
	thumbPath := filepath.Join(c.cacheDir, fmt.Sprintf("%s_%d.jpg", cacheKey, size))

	// Check if thumbnail is up to date
	c.mu.RLock()
	entry, exists := c.entries[cacheKey]
	c.mu.RUnlock()

	if exists && entry.ModTime.Equal(info.ModTime()) {
		// Thumbnail is current
		return thumbPath, nil
	}

	// Generate thumbnail with specified size
	if err := c.createThumbnailWithSize(imagePath, thumbPath, size); err != nil {
		return "", fmt.Errorf("creating thumbnail: %w", err)
	}

	// Update metadata
	c.mu.Lock()
	c.entries[cacheKey] = CacheEntry{
		OriginalPath:  imagePath,
		ModTime:       info.ModTime(),
		ThumbnailPath: thumbPath,
		Size:          info.Size(),
	}
	c.mu.Unlock()

	// Save metadata asynchronously
	go c.saveMetadata()

	return thumbPath, nil
}

// GenerateBatch generates thumbnails for multiple images concurrently
func (c *Cache) GenerateBatch(imagePaths []string, progress func(done, total int)) ([]string, error) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, MaxConcurrency)
	
	results := make([]string, len(imagePaths))
	var mu sync.Mutex
	errors := make([]error, 0)
	completed := 0

	for i, path := range imagePaths {
		wg.Add(1)
		go func(index int, imagePath string) {
			defer wg.Done()
			
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			thumbPath, err := c.Generate(imagePath)
			
			mu.Lock()
			if err != nil {
				errors = append(errors, fmt.Errorf("%s: %w", imagePath, err))
			} else {
				results[index] = thumbPath
			}
			completed++
			if progress != nil {
				progress(completed, len(imagePaths))
			}
			mu.Unlock()
		}(i, path)
	}

	wg.Wait()

	if len(errors) > 0 {
		// Return what we have, but note errors
		fmt.Fprintf(os.Stderr, "Warning: %d thumbnails failed to generate\n", len(errors))
	}

	return results, nil
}

// Get retrieves a thumbnail path if it exists and is current (default 256x256)
func (c *Cache) Get(imagePath string) (string, bool) {
	return c.GetWithSize(imagePath, ThumbnailSizeMedium)
}

// GetWithSize retrieves a thumbnail of specific size
func (c *Cache) GetWithSize(imagePath string, size int) (string, bool) {
	cacheKey := generateCacheKeyWithSize(imagePath, size)
	thumbPath := filepath.Join(c.cacheDir, fmt.Sprintf("%s_%d.jpg", cacheKey, size))
	
	c.mu.RLock()
	entry, exists := c.entries[cacheKey]
	c.mu.RUnlock()

	if !exists {
		return "", false
	}

	// Verify original hasn't changed
	info, err := os.Stat(imagePath)
	if err != nil {
		return "", false
	}

	if !entry.ModTime.Equal(info.ModTime()) {
		return "", false
	}

	// Verify thumbnail file exists
	if _, err := os.Stat(thumbPath); err != nil {
		return "", false
	}

	return thumbPath, true
}

// createThumbnail generates a resized thumbnail
func (c *Cache) createThumbnail(sourcePath, destPath string) error {
	return c.createThumbnailWithSize(sourcePath, destPath, ThumbnailSizeMedium)
}

// createThumbnailWithSize generates a thumbnail with specified dimensions
func (c *Cache) createThumbnailWithSize(sourcePath, destPath string, size int) error {
	// Open source image
	file, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Decode image
	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("decoding image: %w", err)
	}

	// Resize using Lanczos3 resampling (high quality)
	resized := resize.Resize(uint(size), uint(size), img, resize.Lanczos3)

	// Create output file
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Encode as JPEG with quality 85
	return jpeg.Encode(out, resized, &jpeg.Options{Quality: 85})
}

// generateCacheKey creates a unique key for an image path
func generateCacheKey(path string) string {
	hash := sha256.Sum256([]byte(path))
	return hex.EncodeToString(hash[:])[:16]
}

// generateCacheKeyWithSize creates a unique key including size
func generateCacheKeyWithSize(path string, size int) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", path, size)))
	return hex.EncodeToString(hash[:])[:16]
}

// isImageFile checks if a file is an image
func isImageFile(path string) bool {
	ext := filepath.Ext(path)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
		return true
	default:
		return false
	}
}

// loadMetadata loads the cache metadata from disk
func (c *Cache) loadMetadata() error {
	metadataPath := filepath.Join(c.cacheDir, MetadataFile)
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No metadata yet
		}
		return err
	}

	var entries map[string]CacheEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}

	c.mu.Lock()
	c.entries = entries
	c.mu.Unlock()

	return nil
}

// saveMetadata saves the cache metadata to disk
func (c *Cache) saveMetadata() error {
	c.mu.RLock()
	data, err := json.Marshal(c.entries)
	c.mu.RUnlock()
	
	if err != nil {
		return err
	}

	metadataPath := filepath.Join(c.cacheDir, MetadataFile)
	return os.WriteFile(metadataPath, data, 0644)
}

// Clear removes all cached thumbnails
func (c *Cache) Clear() error {
	c.mu.Lock()
	c.entries = make(map[string]CacheEntry)
	c.mu.Unlock()

	// Remove all files in cache dir
	files, err := os.ReadDir(c.cacheDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.Name() != MetadataFile {
			os.Remove(filepath.Join(c.cacheDir, file.Name()))
		}
	}

	return c.saveMetadata()
}

// CacheDir returns the cache directory path
func (c *Cache) CacheDir() string {
	return c.cacheDir
}
