package dedup

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/corona10/goimagehash"
	"github.com/user/wallpaper-cli/internal/data"
)

// Checker handles image deduplication
type Checker struct {
	db        *data.DB
	threshold int // Hamming distance threshold
}

// NewChecker creates a new deduplication checker
func NewChecker(db *data.DB, threshold int) *Checker {
	if threshold <= 0 {
		threshold = 10 // Default threshold
	}
	return &Checker{
		db:        db,
		threshold: threshold,
	}
}

// CheckResult represents the result of a deduplication check
type CheckResult struct {
	IsDuplicate bool
	Existing    *data.ImageRecord
	Hash        string
}

// CheckFile checks if a file is a duplicate by computing its hash
func (c *Checker) CheckFile(path string) (*CheckResult, error) {
	hash, err := ComputeHash(path)
	if err != nil {
		return nil, fmt.Errorf("computing hash: %w", err)
	}

	return c.CheckHash(hash)
}

// CheckHash checks if a hash already exists in the database
func (c *Checker) CheckHash(hash string) (*CheckResult, error) {
	// Check for exact match first
	exists, err := c.db.ImageExists(hash)
	if err != nil {
		return nil, fmt.Errorf("checking database: %w", err)
	}

	if exists {
		record, err := c.db.GetImageByHash(hash)
		if err != nil {
			return nil, err
		}
		return &CheckResult{
			IsDuplicate: true,
			Existing:    record,
			Hash:        hash,
		}, nil
	}

	// TODO: Check for similar hashes within threshold
	// This requires loading all hashes and comparing, which is expensive
	// For now, we only do exact matching

	return &CheckResult{
		IsDuplicate: false,
		Existing:    nil,
		Hash:        hash,
	}, nil
}

// ComputeHash computes the perceptual hash of an image file
func ComputeHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("opening image: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("decoding image: %w", err)
	}

	// Compute pHash (perceptual hash)
	hash, err := goimagehash.PerceptionHash(img)
	if err != nil {
		return "", fmt.Errorf("computing pHash: %w", err)
	}

	return hash.ToString(), nil
}

// ComputeHashFromImage computes hash from an image.Image
func ComputeHashFromImage(img image.Image) (string, error) {
	hash, err := goimagehash.PerceptionHash(img)
	if err != nil {
		return "", fmt.Errorf("computing pHash: %w", err)
	}
	return hash.ToString(), nil
}

// CompareHashes compares two hashes and returns the Hamming distance
func CompareHashes(hash1, hash2 string) (int, error) {
	h1, err := goimagehash.ImageHashFromString(hash1)
	if err != nil {
		return 0, fmt.Errorf("parsing hash1: %w", err)
	}

	h2, err := goimagehash.ImageHashFromString(hash2)
	if err != nil {
		return 0, fmt.Errorf("parsing hash2: %w", err)
	}

	distance, err := h1.Distance(h2)
	if err != nil {
		return 0, fmt.Errorf("computing distance: %w", err)
	}

	return distance, nil
}

// IsSimilar returns true if the distance is within the threshold
func (c *Checker) IsSimilar(distance int) bool {
	return distance <= c.threshold
}
