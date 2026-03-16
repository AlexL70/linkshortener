package config

import (
	"os"
	"path/filepath"
	"testing"
)

// saveEnv saves the current values of the given env vars and returns a restore
// function that must be deferred by the caller.
func saveEnv(keys ...string) func() {
	saved := make(map[string]string, len(keys))
	present := make(map[string]bool, len(keys))
	for _, k := range keys {
		if v, ok := os.LookupEnv(k); ok {
			saved[k] = v
			present[k] = true
		}
	}
	return func() {
		for _, k := range keys {
			if present[k] {
				os.Setenv(k, saved[k]) //nolint:errcheck
			} else {
				os.Unsetenv(k) //nolint:errcheck
			}
		}
	}
}

// writeFile creates path with the given content, failing the test on error.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writeFile %s: %v", path, err)
	}
}

// allRequiredKeys returns all required variable names (excluding LINKSHORTENER_ENV
// which is validated by loadEnv itself).
func allRequiredKeys() []string {
	return []string{
		"DATABASE_URL",
		"JWT_SECRET",
		"SESSION_SECRET",
		"GOOGLE_CLIENT_ID",
		"GOOGLE_CLIENT_SECRET",
		"MICROSOFT_CLIENT_ID",
		"MICROSOFT_CLIENT_SECRET",
		"FACEBOOK_CLIENT_ID",
		"FACEBOOK_CLIENT_SECRET",
		"SUPER_ADMIN_EMAIL",
		"APP_BASE_URL",
	}
}

// setRequiredVars sets all required variables to placeholder values for testing.
func setRequiredVars() {
	placeholders := map[string]string{
		"DATABASE_URL":            "postgres://user:pass@host/db",
		"JWT_SECRET":              "jwt-secret-placeholder",
		"SESSION_SECRET":          "session-secret-placeholder",
		"GOOGLE_CLIENT_ID":        "google-client-id",
		"GOOGLE_CLIENT_SECRET":    "google-client-secret",
		"MICROSOFT_CLIENT_ID":     "ms-client-id",
		"MICROSOFT_CLIENT_SECRET": "ms-client-secret",
		"FACEBOOK_CLIENT_ID":      "fb-client-id",
		"FACEBOOK_CLIENT_SECRET":  "fb-client-secret",
		"SUPER_ADMIN_EMAIL":       "admin@example.com",
		"APP_BASE_URL":            "http://localhost:8080",
	}
	for k, v := range placeholders {
		os.Setenv(k, v) //nolint:errcheck
	}
}

// ── maskDSN ───────────────────────────────────────────────────────────────────

func TestMaskDSN_URIWithPassword(t *testing.T) {
	got := maskDSN("postgres://user:secret@host:5432/mydb")
	want := "postgres://user:****@host:5432/mydb"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMaskDSN_URIWithoutPassword(t *testing.T) {
	got := maskDSN("postgres://user@host/mydb")
	if got != "postgres://user@host/mydb" {
		t.Errorf("got %q, want unchanged", got)
	}
}

func TestMaskDSN_URINoUserInfo(t *testing.T) {
	got := maskDSN("postgres://host/mydb")
	if got != "postgres://host/mydb" {
		t.Errorf("got %q, want unchanged", got)
	}
}

func TestMaskDSN_KeyValueWithPassword(t *testing.T) {
	got := maskDSN("host=localhost password=secret dbname=mydb")
	want := "host=localhost password=**** dbname=mydb"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMaskDSN_KeyValuePasswordCaseInsensitive(t *testing.T) {
	got := maskDSN("host=localhost PASSWORD=secret dbname=mydb")
	want := "host=localhost PASSWORD=**** dbname=mydb"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMaskDSN_KeyValueWithoutPassword(t *testing.T) {
	got := maskDSN("host=localhost dbname=mydb")
	if got != "host=localhost dbname=mydb" {
		t.Errorf("got %q, want unchanged", got)
	}
}

// ── maskedValue ───────────────────────────────────────────────────────────────

func TestMaskedValue_DevModeNoMasking(t *testing.T) {
	keys := []string{"JWT_SECRET", "SESSION_SECRET", "DATABASE_URL", "PORT", "GOOGLE_CLIENT_SECRET"}
	for _, key := range keys {
		if got := maskedValue(key, "somevalue", false); got != "somevalue" {
			t.Errorf("dev mode: maskedValue(%q) = %q, want unmasked", key, got)
		}
	}
}

func TestMaskedValue_ProdSecretsMasked(t *testing.T) {
	secrets := []string{
		"JWT_SECRET", "SESSION_SECRET",
		"GOOGLE_CLIENT_SECRET", "MICROSOFT_CLIENT_SECRET", "FACEBOOK_CLIENT_SECRET",
	}
	for _, key := range secrets {
		if got := maskedValue(key, "supersecret", true); got != "****" {
			t.Errorf("prod mode: maskedValue(%q) = %q, want ****", key, got)
		}
	}
}

func TestMaskedValue_ProdDatabaseURLPasswordMasked(t *testing.T) {
	got := maskedValue("DATABASE_URL", "postgres://u:pass@host/db", true)
	want := "postgres://u:****@host/db"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMaskedValue_ProdNonSecretUnmasked(t *testing.T) {
	if got := maskedValue("PORT", "9090", true); got != "9090" {
		t.Errorf("prod mode: non-secret should not be masked, got %q", got)
	}
}

// ── loadEnv ───────────────────────────────────────────────────────────────────

func TestLoadEnv_InvalidEnvName(t *testing.T) {
	restore := saveEnv("LINKSHORTENER_ENV")
	defer restore()

	os.Setenv("LINKSHORTENER_ENV", "staging")
	if err := LoadEnvFrom(t.TempDir()); err == nil {
		t.Error("expected error for invalid LINKSHORTENER_ENV, got nil")
	}
}

func TestLoadEnv_EmptyEnvName(t *testing.T) {
	restore := saveEnv("LINKSHORTENER_ENV")
	defer restore()

	os.Unsetenv("LINKSHORTENER_ENV")
	if err := LoadEnvFrom(t.TempDir()); err == nil {
		t.Error("expected error for empty LINKSHORTENER_ENV, got nil")
	}
}

func TestLoadEnv_ProdSkipsFiles(t *testing.T) {
	const testVar = "LINKSHORTENER_TEST_PROD_SKIP"
	restore := saveEnv("LINKSHORTENER_ENV", testVar)
	defer restore()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".env"), testVar+"=from_file\n")

	os.Setenv("LINKSHORTENER_ENV", "prod")
	os.Unsetenv(testVar)

	if err := LoadEnvFrom(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := os.Getenv(testVar); got != "" {
		t.Errorf("prod mode should not read .env files, but %s = %q", testVar, got)
	}
}

func TestLoadEnv_DevReadsBaseAndDevFile(t *testing.T) {
	const baseVar = "LINKSHORTENER_TEST_BASE"
	const devVar = "LINKSHORTENER_TEST_DEV"
	const overrideVar = "LINKSHORTENER_TEST_OVERRIDE"
	restore := saveEnv("LINKSHORTENER_ENV", baseVar, devVar, overrideVar)
	defer restore()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".env"), baseVar+"=base_value\n"+overrideVar+"=base_override\n")
	writeFile(t, filepath.Join(dir, ".env.dev"), devVar+"=dev_value\n"+overrideVar+"=dev_override\n")

	os.Setenv("LINKSHORTENER_ENV", "dev")
	os.Unsetenv(baseVar)
	os.Unsetenv(devVar)
	os.Unsetenv(overrideVar)

	if err := LoadEnvFrom(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := os.Getenv(baseVar); got != "base_value" {
		t.Errorf("%s = %q, want %q", baseVar, got, "base_value")
	}
	if got := os.Getenv(devVar); got != "dev_value" {
		t.Errorf("%s = %q, want %q", devVar, got, "dev_value")
	}
	// .env.dev must override .env
	if got := os.Getenv(overrideVar); got != "dev_override" {
		t.Errorf("%s = %q, want %q (.env.dev should override .env)", overrideVar, got, "dev_override")
	}
}

func TestLoadEnv_DevOSValueWins(t *testing.T) {
	const testVar = "LINKSHORTENER_TEST_OS_WIN"
	restore := saveEnv("LINKSHORTENER_ENV", testVar)
	defer restore()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".env"), testVar+"=from_file\n")

	os.Setenv("LINKSHORTENER_ENV", "dev")
	os.Setenv(testVar, "from_os")

	if err := LoadEnvFrom(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := os.Getenv(testVar); got != "from_os" {
		t.Errorf("OS value should win: %s = %q, want %q", testVar, got, "from_os")
	}
}

func TestLoadEnv_DevCannotOverrideLINKSHORTENER_ENV(t *testing.T) {
	restore := saveEnv("LINKSHORTENER_ENV")
	defer restore()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".env"), "LINKSHORTENER_ENV=prod\n")

	os.Setenv("LINKSHORTENER_ENV", "dev")

	if err := LoadEnvFrom(dir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := os.Getenv("LINKSHORTENER_ENV"); got != "dev" {
		t.Errorf("LINKSHORTENER_ENV should remain %q, got %q", "dev", got)
	}
}

func TestLoadEnv_MissingFilesAreOk(t *testing.T) {
	restore := saveEnv("LINKSHORTENER_ENV")
	defer restore()

	os.Setenv("LINKSHORTENER_ENV", "dev")
	// Use an empty temp dir — no .env or .env.dev files present.
	if err := LoadEnvFrom(t.TempDir()); err != nil {
		t.Errorf("missing .env files should be silently ignored, got: %v", err)
	}
}

// ── Validate ──────────────────────────────────────────────────────────────────

func TestValidate_AllRequiredSet_DefaultsApplied(t *testing.T) {
	allOptional := make([]string, 0, len(optionalDefaults))
	for k := range optionalDefaults {
		allOptional = append(allOptional, k)
	}
	restore := saveEnv(append(allRequiredKeys(), allOptional...)...)
	defer restore()

	setRequiredVars()
	for k := range optionalDefaults {
		os.Unsetenv(k) //nolint:errcheck
	}

	if err := Validate(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	// Spot-check that defaults were applied.
	checks := map[string]string{
		"PORT":              "8080",
		"GIN_MODE":          "release",
		"DEFAULT_PAGE_SIZE": "20",
	}
	for k, want := range checks {
		if got := os.Getenv(k); got != want {
			t.Errorf("default for %s: got %q, want %q", k, got, want)
		}
	}
}

func TestValidate_MissingRequiredVariable(t *testing.T) {
	allOptional := make([]string, 0, len(optionalDefaults))
	for k := range optionalDefaults {
		allOptional = append(allOptional, k)
	}
	restore := saveEnv(append(allRequiredKeys(), allOptional...)...)
	defer restore()

	setRequiredVars()
	os.Unsetenv("DATABASE_URL")

	if err := Validate(); err == nil {
		t.Error("expected error for missing DATABASE_URL, got nil")
	}
}

func TestValidate_InvalidNumericOptional(t *testing.T) {
	allOptional := make([]string, 0, len(optionalDefaults))
	for k := range optionalDefaults {
		allOptional = append(allOptional, k)
	}
	restore := saveEnv(append(allRequiredKeys(), allOptional...)...)
	defer restore()

	setRequiredVars()
	for k := range optionalDefaults {
		os.Unsetenv(k) //nolint:errcheck
	}
	os.Setenv("PORT", "not-a-number")

	if err := Validate(); err == nil {
		t.Error("expected error for non-integer PORT, got nil")
	}
}

func TestValidate_OSValueNotOverriddenByDefault(t *testing.T) {
	allOptional := make([]string, 0, len(optionalDefaults))
	for k := range optionalDefaults {
		allOptional = append(allOptional, k)
	}
	restore := saveEnv(append(allRequiredKeys(), allOptional...)...)
	defer restore()

	setRequiredVars()
	for k := range optionalDefaults {
		os.Unsetenv(k) //nolint:errcheck
	}
	os.Setenv("PORT", "9090")

	if err := Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := os.Getenv("PORT"); got != "9090" {
		t.Errorf("Validate should not overwrite non-empty PORT: got %q", got)
	}
}

// ── SUPER_ADMIN_EMAIL validation ──────────────────────────────────────────────

func TestValidate_SuperAdminEmailMissing(t *testing.T) {
	allOptional := make([]string, 0, len(optionalDefaults))
	for k := range optionalDefaults {
		allOptional = append(allOptional, k)
	}
	restore := saveEnv(append(allRequiredKeys(), allOptional...)...)
	defer restore()

	setRequiredVars()
	os.Unsetenv("SUPER_ADMIN_EMAIL")

	if err := Validate(); err == nil {
		t.Error("expected error for missing SUPER_ADMIN_EMAIL, got nil")
	}
}

func TestValidate_SuperAdminEmailInvalid(t *testing.T) {
	allOptional := make([]string, 0, len(optionalDefaults))
	for k := range optionalDefaults {
		allOptional = append(allOptional, k)
	}
	restore := saveEnv(append(allRequiredKeys(), allOptional...)...)
	defer restore()

	setRequiredVars()
	os.Setenv("SUPER_ADMIN_EMAIL", "not-a-valid-email")

	if err := Validate(); err == nil {
		t.Error("expected error for invalid SUPER_ADMIN_EMAIL, got nil")
	}
}

func TestValidate_SuperAdminEmailValid(t *testing.T) {
	allOptional := make([]string, 0, len(optionalDefaults))
	for k := range optionalDefaults {
		allOptional = append(allOptional, k)
	}
	restore := saveEnv(append(allRequiredKeys(), allOptional...)...)
	defer restore()

	setRequiredVars()
	for k := range optionalDefaults {
		os.Unsetenv(k) //nolint:errcheck
	}
	os.Setenv("SUPER_ADMIN_EMAIL", "superadmin@mycompany.org")

	if err := Validate(); err != nil {
		t.Errorf("expected no error for valid SUPER_ADMIN_EMAIL, got: %v", err)
	}
}
