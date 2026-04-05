package validate

import (
	"testing"
)

func TestValidateResolution(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"", false},           // empty is valid (no filter)
		{"1080p", false},
		{"1440p", false},
		{"4k", false},
		{"8k", false},
		{"1920x1080", false},
		{"3840x2160", false},
		{"800x600", false},
		{"banana", true},
		{"1920", true},
		{"x1080", true},
		{"0x0", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			err := ValidateResolution(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateResolution(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateAspectRatio(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"", false},
		{"16:9", false},
		{"21:9", false},
		{"4:3", false},
		{"1:1", false},
		{"3:2", false},    // custom ratio
		{"banana", true},
		{"16", true},
		{":9", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			err := ValidateAspectRatio(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAspectRatio(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestNormalizeResolution(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"4k", "3840x2160"},
		{"1080p", "1920x1080"},
		{"1920x1080", "1920x1080"}, // already normalized
		{"unknown", "unknown"},      // passthrough
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := NormalizeResolution(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeResolution(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeAspectRatio(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"16:9", "16x9"},
		{"21:9", "21x9"},
		{"3:2", "3x2"},   // custom passthrough with colon->x
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := NormalizeAspectRatio(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeAspectRatio(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
