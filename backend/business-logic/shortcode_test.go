//go:build ignore

package businesslogic_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	businesslogic "github.com/AlexL70/linkshortener/backend/business-logic"
)

func TestBase62Generator_GenerateShortcode_Length(t *testing.T) {
	gen := businesslogic.NewShortcodeGenerator()
	sc, err := gen.GenerateShortcode()
	require.NoError(t, err)
	assert.Len(t, sc, 6, "generated shortcode must be exactly 6 characters")
}

func TestBase62Generator_GenerateShortcode_CharSet(t *testing.T) {
	gen := businesslogic.NewShortcodeGenerator()
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
	gen := businesslogic.NewShortcodeGenerator()
	seen := make(map[string]struct{}, 100)
	for i := 0; i < 100; i++ {
		sc, err := gen.GenerateShortcode()
		require.NoError(t, err)
		seen[sc] = struct{}{}
	}
	// With 6-char base62, collision probability over 100 draws is negligible.
	assert.Equal(t, 100, len(seen))
}

// ── ValidateCustomShortcode ───────────────────────────────────────────────────

func TestValidateCustomShortcode_Valid(t *testing.T) {
	cases := []string{"abc123", "ABC123", "a1b2c3", "ab-123", "------"}
	for _, sc := range cases {
		assert.NoError(t, businesslogic.ValidateCustomShortcode(sc), "shortcode %q", sc)
	}
}

func TestValidateCustomShortcode_TooShort(t *testing.T) {
	err := businesslogic.ValidateCustomShortcode("abc")
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrValidation)
}

func TestValidateCustomShortcode_TooLong(t *testing.T) {
	err := businesslogic.ValidateCustomShortcode("abcdefg")
	require.Error(t, err)
	assert.ErrorIs(t, err, businesslogic.ErrValidation)
}

func TestValidateCustomShortcode_InvalidChar(t *testing.T) {
	for _, sc := range []string{"abc!23", "abc 12", "abc@12"} {
		err := businesslogic.ValidateCustomShortcode(sc)
		require.Error(t, err, "shortcode %q", sc)
		assert.ErrorIs(t, err, businesslogic.ErrValidation, "shortcode %q", sc)
	}
}

func TestValidateCustomShortcode_Reserved(t *testing.T) {
	for _, sc := range []string{"admin1", "login1", "logout", "auth12", "static"} {
		err := businesslogic.ValidateCustomShortcode(sc)
		require.Error(t, err, "shortcode %q", sc)
		assert.ErrorIs(t, err, businesslogic.ErrValidation, "shortcode %q", sc)
	}
}











































































}	}		assert.ErrorIs(t, err, businesslogic.ErrValidation, "shortcode %q", sc)		require.Error(t, err, "shortcode %q", sc)		err := businesslogic.ValidateCustomShortcode(sc)	for _, sc := range []string{"admin1", "login1", "logout", "auth12", "static"} {func TestValidateCustomShortcode_Reserved(t *testing.T) {}	}		assert.ErrorIs(t, err, businesslogic.ErrValidation, "shortcode %q", sc)		require.Error(t, err, "shortcode %q", sc)		err := businesslogic.ValidateCustomShortcode(sc)	for _, sc := range []string{"abc!23", "abc 12", "abc@12"} {func TestValidateCustomShortcode_InvalidChar(t *testing.T) {}	assert.ErrorIs(t, err, businesslogic.ErrValidation)	require.Error(t, err)	err := businesslogic.ValidateCustomShortcode("abcdefg")func TestValidateCustomShortcode_TooLong(t *testing.T) {}	assert.ErrorIs(t, err, businesslogic.ErrValidation)	require.Error(t, err)	err := businesslogic.ValidateCustomShortcode("abc")func TestValidateCustomShortcode_TooShort(t *testing.T) {}	}		assert.NoError(t, businesslogic.ValidateCustomShortcode(sc), "shortcode %q", sc)	for _, sc := range cases {	cases := []string{"abc123", "ABC123", "a1b2c3", "ab-123", "------"}func TestValidateCustomShortcode_Valid(t *testing.T) {// ── ValidateCustomShortcode ───────────────────────────────────────────────────}	assert.Equal(t, 100, len(seen))	// With 6-char base62, collision probability over 100 draws is negligible.	}		seen[sc] = struct{}{}		require.NoError(t, err)		sc, err := gen.GenerateShortcode()	for i := 0; i < 100; i++ {	seen := make(map[string]struct{}, 100)	gen := businesslogic.NewShortcodeGenerator()func TestBase62Generator_GenerateShortcode_ReturnsUniqueValues(t *testing.T) {}	}		}			assert.True(t, strings.ContainsRune(base62, c), "character %q is not in base62 charset", c)		for _, c := range sc {		require.NoError(t, err)		sc, err := gen.GenerateShortcode()	for i := 0; i < 100; i++ {	const base62 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"	gen := businesslogic.NewShortcodeGenerator()func TestBase62Generator_GenerateShortcode_CharSet(t *testing.T) {}	assert.Len(t, sc, 6, "generated shortcode must be exactly 6 characters")	require.NoError(t, err)	sc, err := gen.GenerateShortcode()	gen := businesslogic.NewShortcodeGenerator()func TestBase62Generator_GenerateShortcode_Length(t *testing.T) {)	businesslogic "github.com/AlexL70/linkshortener/backend/business-logic"	"github.com/stretchr/testify/require"	"github.com/stretchr/testify/assert"	"testing"	"strings"import (