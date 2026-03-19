package businesslogic_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	businesslogic "github.com/AlexL70/linkshortener/backend/business-logic"
)

func TestValidateLongUrl_ValidHttp(t *testing.T) {
	assert.NoError(t, businesslogic.ValidateLongUrl("http://example.com/path?q=1", 2048, nil, false))
}

func TestValidateLongUrl_ValidHttps(t *testing.T) {
	assert.NoError(t, businesslogic.ValidateLongUrl("https://www.example.com", 2048, nil, false))
}

func TestValidateLongUrl_TooLong(t *testing.T) {
	raw := "https://example.com/" + strings.Repeat("a", 2048)
	err := businesslogic.ValidateLongUrl(raw, 2048, nil, false)
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrValidation)
}

func TestValidateLongUrl_InvalidScheme(t *testing.T) {
	for _, raw := range []string{"ftp://example.com", "file:///etc/passwd", "data:text/plain,hello"} {
		err := businesslogic.ValidateLongUrl(raw, 2048, nil, false)
		require.Error(t, err, "URL: %s", raw)
		assert.ErrorIs(t, err, businesslogic.ErrValidation, "URL: %s", raw)
	}
}

func TestValidateLongUrl_NoHost(t *testing.T) {
	err := businesslogic.ValidateLongUrl("https://", 2048, nil, false)
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrValidation)
}

func TestValidateLongUrl_Localhost(t *testing.T) {
	for _, raw := range []string{
		"http://localhost",
		"http://localhost:8080/path",
		"https://sub.localhost",
	} {
		err := businesslogic.ValidateLongUrl(raw, 2048, nil, false)
		require.Error(t, err, "URL: %s", raw)
		assert.ErrorIs(t, err, businesslogic.ErrValidation, "URL: %s", raw)
	}
}

func TestValidateLongUrl_LoopbackIP(t *testing.T) {
	err := businesslogic.ValidateLongUrl("http://127.0.0.1/evil", 2048, nil, false)
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrValidation)
}

func TestValidateLongUrl_IPv6Loopback(t *testing.T) {
	err := businesslogic.ValidateLongUrl("http://[::1]/evil", 2048, nil, false)
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrValidation)
}

func TestValidateLongUrl_PrivateIP_RFC1918(t *testing.T) {
	for _, raw := range []string{
		"http://10.0.0.1",
		"http://192.168.1.1",
		"http://172.16.0.1",
		"http://172.31.255.255",
	} {
		err := businesslogic.ValidateLongUrl(raw, 2048, nil, false)
		require.Error(t, err, "URL: %s", raw)
		assert.ErrorIs(t, err, businesslogic.ErrValidation, "URL: %s", raw)
	}
}

func TestValidateLongUrl_PublicIP_Valid(t *testing.T) {
	// 8.8.8.8 is a public IP address (Google DNS)
	assert.NoError(t, businesslogic.ValidateLongUrl("https://8.8.8.8", 2048, nil, false))
}

// ── ValidateCustomShortcode ───────────────────────────────────────────────────

func TestBase62Generator_GenerateShortcode_Length(t *testing.T) {
	gen := businesslogic.NewShortcodeGenerator(6)
	sc, err := gen.GenerateShortcode()
	require.NoError(t, err)
	assert.Len(t, sc, 6, "generated shortcode must be exactly 6 characters")
}

func TestBase62Generator_GenerateShortcode_CharSet(t *testing.T) {
	gen := businesslogic.NewShortcodeGenerator(6)
	const base62 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	for i := 0; i < 100; i++ {
		sc, err := gen.GenerateShortcode()
		require.NoError(t, err)
		for _, c := range sc {
			assert.True(t, strings.ContainsRune(base62, c), "character %q is not in base62 charset", c)
		}
	}
}

func TestBase62Generator_GenerateShortcode_ReturnsUniqueValues(t *testing.T) {
	gen := businesslogic.NewShortcodeGenerator(6)
	seen := make(map[string]struct{}, 100)
	for i := 0; i < 100; i++ {
		sc, err := gen.GenerateShortcode()
		require.NoError(t, err)
		seen[sc] = struct{}{}
	}
	// With 6-char base62, collision probability over 100 draws is negligible.
	assert.Equal(t, 100, len(seen))
}

func TestValidateCustomShortcode_Valid(t *testing.T) {
	cases := []string{"abc123", "ABC123", "a1b2c3", "ab-123", "------"}
	for _, sc := range cases {
		assert.NoError(t, businesslogic.ValidateCustomShortcode(sc, 6, 6), "shortcode %q", sc)
	}
}

func TestValidateCustomShortcode_TooShort(t *testing.T) {
	err := businesslogic.ValidateCustomShortcode("abc", 6, 6)
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrValidation)
}

func TestValidateCustomShortcode_TooShort_RangeMessage(t *testing.T) {
	// When minLen != maxLen the error message uses "between … and …" wording.
	err := businesslogic.ValidateCustomShortcode("ab", 4, 8)
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrValidation)
	assert.ErrorContains(t, err, "between")
}

func TestValidateCustomShortcode_TooLong(t *testing.T) {
	err := businesslogic.ValidateCustomShortcode("abcdefg", 6, 6)
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrValidation)
}

func TestValidateCustomShortcode_InvalidChar(t *testing.T) {
	for _, sc := range []string{"abc!23", "abc 12", "abc@12"} {
		err := businesslogic.ValidateCustomShortcode(sc, 6, 6)
		require.Error(t, err, "shortcode %q", sc)
		assert.ErrorIs(t, err, businesslogic.ErrValidation, "shortcode %q", sc)
	}
}

func TestValidateCustomShortcode_Reserved(t *testing.T) {
	for _, sc := range []string{"admin1", "login1", "logout", "auth12", "static"} {
		err := businesslogic.ValidateCustomShortcode(sc, 6, 6)
		require.Error(t, err, "shortcode %q", sc)
		assert.ErrorIs(t, err, businesslogic.ErrValidation, "shortcode %q", sc)
	}
}

// ── DNS-based SSRF Prevention ─────────────────────────────────────────────────

func TestValidateLongUrl_DNS_NilLookupHost_Skipped(t *testing.T) {
	// With nil lookupHost (dev mode) no DNS resolution is attempted — the URL passes.
	assert.NoError(t, businesslogic.ValidateLongUrl("https://internal.evil.com", 2048, nil, false))
}

func TestValidateLongUrl_DNS_PublicHostname_Allowed(t *testing.T) {
	stub := func(_ string) ([]string, error) { return []string{"8.8.8.8"}, nil }
	assert.NoError(t, businesslogic.ValidateLongUrl("https://dns.google.com", 2048, stub, false))
}

func TestValidateLongUrl_DNS_PrivateHostname_Rejected(t *testing.T) {
	stub := func(_ string) ([]string, error) { return []string{"192.168.1.1"}, nil }
	err := businesslogic.ValidateLongUrl("https://internal.evil.com", 2048, stub, false)
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrValidation)
}

func TestValidateLongUrl_DNS_LoopbackHostname_Rejected(t *testing.T) {
	stub := func(_ string) ([]string, error) { return []string{"127.0.0.1"}, nil }
	err := businesslogic.ValidateLongUrl("https://loopback.example.com", 2048, stub, false)
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrValidation)
}

func TestValidateLongUrl_DNS_LookupError_FailClosed(t *testing.T) {
	stub := func(_ string) ([]string, error) { return nil, errors.New("no such host") }
	err := businesslogic.ValidateLongUrl("https://nonexistent.example.com", 2048, stub, false)
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrValidation)
}

func TestValidateLongUrl_DNS_LookupError_FailOpen(t *testing.T) {
	stub := func(_ string) ([]string, error) { return nil, errors.New("no such host") }
	// With dnsFailOpen=true DNS errors are non-fatal; the URL is allowed through.
	assert.NoError(t, businesslogic.ValidateLongUrl("https://nonexistent.example.com", 2048, stub, true))
}

func TestValidateLongUrl_DNS_LiteralPublicIP_NoDnsCall(t *testing.T) {
	// Literal public IPs must not trigger DNS lookup.
	called := false
	stub := func(_ string) ([]string, error) { called = true; return nil, nil }
	err := businesslogic.ValidateLongUrl("https://8.8.8.8", 2048, stub, false)
	assert.NoError(t, err)
	assert.False(t, called, "DNS lookup must not be called for literal IP addresses")
}
