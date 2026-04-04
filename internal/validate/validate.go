package validate

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ValidResolutions maps common resolution names to dimensions
var ValidResolutions = map[string]string{
	"1080p": "1920x1080",
	"1440p": "2560x1440",
	"4k":    "3840x2160",
	"8k":    "7680x4320",
}

// ValidAspectRatios maps common aspect ratios
var ValidAspectRatios = map[string]string{
	"16:9":  "16x9",
	"21:9":  "21x9",
	"32:9":  "32x9",
	"4:3":   "4x3",
	"1:1":   "1x1",
}

// ResolutionPattern matches WxH format
var ResolutionPattern = regexp.MustCompile(`^(\d+)x(\d+)$`)

// ValidateResolution checks if a resolution string is valid
func ValidateResolution(res string) error {
	if res == "" {
		return nil
	}

	// Check predefined names
	res = strings.ToLower(res)
	if _, ok := ValidResolutions[res]; ok {
		return nil
	}

	// Check WxH format
	if ResolutionPattern.MatchString(res) {
		matches := ResolutionPattern.FindStringSubmatch(res)
		width, _ := strconv.Atoi(matches[1])
		height, _ := strconv.Atoi(matches[2])
		if width > 0 && height > 0 {
			return nil
		}
	}

	return fmt.Errorf("invalid resolution: %s (expected: 1080p, 1440p, 4k, 8k, or WxH)", res)
}

// ValidateAspectRatio checks if an aspect ratio string is valid
func ValidateAspectRatio(ratio string) error {
	if ratio == "" {
		return nil
	}

	ratio = strings.ToLower(ratio)
	if _, ok := ValidAspectRatios[ratio]; ok {
		return nil
	}

	// Allow custom ratios like "21:9"
	parts := strings.Split(ratio, ":")
	if len(parts) == 2 {
		w, err1 := strconv.ParseFloat(parts[0], 64)
		h, err2 := strconv.ParseFloat(parts[1], 64)
		if err1 == nil && err2 == nil && w > 0 && h > 0 {
			return nil
		}
	}

	return fmt.Errorf("invalid aspect ratio: %s (expected: 16:9, 21:9, 32:9, 4:3, or W:H)", ratio)
}

// NormalizeResolution converts resolution name to WxH format
func NormalizeResolution(res string) string {
	res = strings.ToLower(res)
	if normalized, ok := ValidResolutions[res]; ok {
		return normalized
	}
	return res
}

// NormalizeAspectRatio converts aspect ratio to WxH format
func NormalizeAspectRatio(ratio string) string {
	ratio = strings.ToLower(ratio)
	if normalized, ok := ValidAspectRatios[ratio]; ok {
		return normalized
	}
	return strings.ReplaceAll(ratio, ":", "x")
}
