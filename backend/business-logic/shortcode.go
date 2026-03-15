package businesslogic

import (
	"crypto/rand"
	"fmt"
	"strings"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// reservedShortcodes contains values that must not be used as user-facing shortcodes
// because they conflict with application routes or well-known URL patterns.
var reservedShortcodes = map[string]struct{}{
	"auth":   {},
	"login":  {},
	"logout": {},
	"admin":  {},
	"user":   {},
	"api":    {},
	"docs":   {},
	"static": {},
	"public": {},
	"assets": {},
	"css":    {},
	"js":     {},
	"img":    {},
	"health": {},
	"status": {},
	"ping":   {},
	"robots": {},
}

// Base62Generator generates unique shortcodes using random Base62 encoding.
type Base62Generator struct {
	length int
}

// NewShortcodeGenerator returns a Base62Generator that produces random base62 shortcodes
// of the specified length.
func NewShortcodeGenerator(length int) *Base62Generator {
	return &Base62Generator{length: length}
}

// GenerateShortcode returns a randomly generated base62 shortcode of the configured length.
func (g *Base62Generator) GenerateShortcode() (string, error) {
	buf := make([]byte, g.length)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("GenerateShortcode: %w", err)
	}
	result := make([]byte, g.length)
	for i, b := range buf {
		result[i] = base62Chars[int(b)%len(base62Chars)]
	}
	return string(result), nil
}

// ValidateCustomShortcode validates a user-supplied shortcode against the given length limits.
// Returns ErrValidation if the shortcode length is outside [minLen, maxLen],
// contains disallowed characters (anything other than alphanumeric and hyphens),
// or matches a reserved word.
func ValidateCustomShortcode(shortcode string, minLen, maxLen int) error {
	if len(shortcode) < minLen || len(shortcode) > maxLen {
		if minLen == maxLen {
			return fmt.Errorf("%w: shortcode must be exactly %d characters", ErrValidation, minLen)
		}
		return fmt.Errorf("%w: shortcode length must be between %d and %d characters", ErrValidation, minLen, maxLen)
	}
	for _, c := range shortcode {
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '-') {
			return fmt.Errorf("%w: shortcode must contain only alphanumeric characters and hyphens", ErrValidation)
		}
	}
	lower := strings.ToLower(shortcode)
	for word := range reservedShortcodes {
		if strings.HasPrefix(lower, word) {
			return fmt.Errorf("%w: shortcode is reserved", ErrValidation)
		}
	}
	return nil
}
