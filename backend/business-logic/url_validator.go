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
//   - host is a literal private or reserved IP address (SSRF prevention)
//   - host resolves (via lookupHost) to a private or reserved IP (SSRF prevention)
//
// lookupHost is called only in prod mode (pass nil to skip DNS resolution in dev).
// When dnsFailOpen is true, DNS resolution errors are non-fatal and the URL is
// allowed through; when false, resolution errors cause the URL to be rejected.
func ValidateLongUrl(raw string, maxLen int, lookupHost func(string) ([]string, error), dnsFailOpen bool) error {
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
	} else if lookupHost != nil {
		// Host is not a literal IP; resolve it to catch SSRF via DNS rebinding.
		addrs, resolveErr := lookupHost(host)
		if resolveErr != nil {
			if !dnsFailOpen {
				return fmt.Errorf("%w: unable to resolve hostname: %v", ErrValidation, resolveErr)
			}
			// dnsFailOpen=true: treat DNS errors as non-fatal and allow the URL.
		} else {
			for _, addr := range addrs {
				resolved := net.ParseIP(addr)
				if resolved != nil && (resolved.IsLoopback() || resolved.IsPrivate() || resolved.IsLinkLocalUnicast() || resolved.IsLinkLocalMulticast() || resolved.IsUnspecified()) {
					return fmt.Errorf("%w: URL must not target private or reserved IP addresses", ErrValidation)
				}
			}
		}
	}

	return nil
}
