package zerochan

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ParseCookiesFile reads a Netscape/Mozilla cookies.txt file and returns
// the cookies as a raw "Cookie" header value for the given domain.
// This is the same format used by curl, yt-dlp, gallery-dl, and browser
// cookie export extensions.
//
// Format: domain\tinclude_subdomains\tpath\tsecure\texpiry\tname\tvalue
// Lines starting with # are comments. Empty lines are skipped.
func ParseCookiesFile(path, domain string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("opening cookies file: %w", err)
	}
	defer f.Close()

	var pairs []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 7 {
			continue
		}

		cookieDomain := fields[0]
		name := fields[5]
		value := fields[6]

		// Match domain: ".zerochan.net" matches "www.zerochan.net"
		if !matchDomain(cookieDomain, domain) {
			continue
		}

		pairs = append(pairs, name+"="+value)
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading cookies file: %w", err)
	}

	if len(pairs) == 0 {
		return "", fmt.Errorf("no cookies found for domain %q in %s", domain, path)
	}

	return strings.Join(pairs, "; "), nil
}

func matchDomain(cookieDomain, targetDomain string) bool {
	cookieDomain = strings.TrimPrefix(cookieDomain, ".")
	targetDomain = strings.TrimPrefix(targetDomain, ".")

	if cookieDomain == targetDomain {
		return true
	}
	// .zerochan.net matches www.zerochan.net
	return strings.HasSuffix(targetDomain, "."+cookieDomain)
}
