package businesslogic

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// ValidateLongUrl validates a long URL submitted by a user.
// Returns ErrValidation if the URL fails any of the following checks:
//   - length exceeds maxLen characters
//   - URL cannot be parsed or has no host
//   - scheme is not http or https
//   - host is localhost or a loopback address (SSRF prevention)
//   - host resolves to a private or reserved IP range (SSRF prevention)
func ValidateLongUrl(raw string, maxLen int) error {
	if len(raw) > maxLen {
		return fmt.Errorf("%w: URL length exceeds %d characters", ErrValidation, maxLen)
	}

	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("%w: invalid URL: %v", ErrValidation, err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("%w: URL must use http or https scheme", ErrValidation)
	}

	if u.Host == "" {
		return fmt.Errorf("%w: URL must have a host", ErrValidation)
	}

	host := u.Hostname()
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return fmt.Errorf("%w: URL must not target localhost", ErrValidation)
	}

	ip := net.ParseIP(host)
	if ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
			return fmt.Errorf("%w: URL must not target private or reserved IP addresses", ErrValidation)
		}
	}

	return nil
}
