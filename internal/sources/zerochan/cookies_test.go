package zerochan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseCookiesFile(t *testing.T) {
	content := `# Netscape HTTP Cookie File
# https://curl.se/docs/http-cookies.html
.zerochan.net	TRUE	/	TRUE	1783140448	z_theme	11
.zerochan.net	TRUE	/	TRUE	1783140448	guest_id	2414468
www.zerochan.net	FALSE	/	FALSE	0	PHPSESSID	abc123session
.zerochan.net	TRUE	/	TRUE	0	xbotcheck	secrettoken
.example.com	TRUE	/	FALSE	0	other	ignored
`
	dir := t.TempDir()
	path := filepath.Join(dir, "cookies.txt")
	os.WriteFile(path, []byte(content), 0644)

	cookies, err := ParseCookiesFile(path, "www.zerochan.net")
	if err != nil {
		t.Fatalf("ParseCookiesFile: %v", err)
	}

	// Should include zerochan cookies but not example.com
	if !contains(cookies, "PHPSESSID=abc123session") {
		t.Errorf("missing PHPSESSID, got: %q", cookies)
	}
	if !contains(cookies, "xbotcheck=secrettoken") {
		t.Errorf("missing xbotcheck, got: %q", cookies)
	}
	if contains(cookies, "other=ignored") {
		t.Errorf("should not include example.com cookies, got: %q", cookies)
	}
}

func TestParseCookiesFile_MissingFile(t *testing.T) {
	_, err := ParseCookiesFile("/nonexistent/cookies.txt", "zerochan.net")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestParseCookiesFile_NoCookiesForDomain(t *testing.T) {
	content := `.example.com	TRUE	/	FALSE	0	name	value
`
	dir := t.TempDir()
	path := filepath.Join(dir, "cookies.txt")
	os.WriteFile(path, []byte(content), 0644)

	_, err := ParseCookiesFile(path, "www.zerochan.net")
	if err == nil {
		t.Error("expected error when no cookies match domain")
	}
}

func TestMatchDomain(t *testing.T) {
	tests := []struct {
		cookie, target string
		want           bool
	}{
		{".zerochan.net", "www.zerochan.net", true},
		{"zerochan.net", "www.zerochan.net", true},
		{"www.zerochan.net", "www.zerochan.net", true},
		{".example.com", "www.zerochan.net", false},
		{".zerochan.net", "zerochan.net", true},
	}
	for _, tt := range tests {
		got := matchDomain(tt.cookie, tt.target)
		if got != tt.want {
			t.Errorf("matchDomain(%q, %q) = %v, want %v", tt.cookie, tt.target, got, tt.want)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > len(sub) && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
