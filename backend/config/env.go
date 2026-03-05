package config

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// AppEnv represents the runtime mode of the application.
type AppEnv string

const (
	EnvDev  AppEnv = "dev"
	EnvProd AppEnv = "prod"
)

// GetAppEnv returns the current runtime mode from LINKSHORTENER_ENV.
// Must be called after LoadEnv.
func GetAppEnv() AppEnv {
	return AppEnv(os.Getenv("LINKSHORTENER_ENV"))
}

// requiredVars lists variables that must be non-empty after loading.
// LINKSHORTENER_ENV is validated by LoadEnv itself, so it is excluded here.
var requiredVars = []string{
	"DATABASE_URL",
	"JWT_SECRET",
	"SESSION_SECRET",
	"GOOGLE_CLIENT_ID",
	"GOOGLE_CLIENT_SECRET",
	"MICROSOFT_CLIENT_ID",
	"MICROSOFT_CLIENT_SECRET",
	"FACEBOOK_CLIENT_ID",
	"FACEBOOK_CLIENT_SECRET",
}

// optionalDefaults maps every optional variable to its documented default value.
var optionalDefaults = map[string]string{
	"PORT":                        "8080",
	"GIN_MODE":                    "release",
	"REQUEST_TIMEOUT_SECONDS":     "30",
	"MAX_CONCURRENT_REQUESTS":     "100",
	"DB_MAX_OPEN_CONNS":           "25",
	"DB_MAX_IDLE_CONNS":           "5",
	"DB_CONN_MAX_LIFETIME":        "3600",
	"DB_CONN_MAX_IDLE_TIME":       "600",
	"CLICK_BATCH_SIZE":            "1000",
	"CLICK_BATCH_TIMEOUT_SECONDS": "5",
	"DEFAULT_PAGE_SIZE":           "20",
}

// numericOptional is the subset of optionalDefaults whose values must be valid integers.
var numericOptional = map[string]bool{
	"PORT":                        true,
	"REQUEST_TIMEOUT_SECONDS":     true,
	"MAX_CONCURRENT_REQUESTS":     true,
	"DB_MAX_OPEN_CONNS":           true,
	"DB_MAX_IDLE_CONNS":           true,
	"DB_CONN_MAX_LIFETIME":        true,
	"DB_CONN_MAX_IDLE_TIME":       true,
	"CLICK_BATCH_SIZE":            true,
	"CLICK_BATCH_TIMEOUT_SECONDS": true,
	"DEFAULT_PAGE_SIZE":           true,
}

// secretVars lists variables whose values must be masked in prod mode logs.
var secretVars = map[string]bool{
	"JWT_SECRET":              true,
	"SESSION_SECRET":          true,
	"GOOGLE_CLIENT_SECRET":    true,
	"MICROSOFT_CLIENT_SECRET": true,
	"FACEBOOK_CLIENT_SECRET":  true,
}

// LoadEnv validates LINKSHORTENER_ENV and loads configuration from files:
//   - "dev"  mode: reads ../.env then ../.env.dev; OS values always win.
//   - "prod" mode: skips all file reading; uses OS values only.
//
// LINKSHORTENER_ENV itself is always taken from the OS; .env files cannot override it.
func LoadEnv() error {
	return loadEnv("..")
}

// loadEnv is the testable core of LoadEnv; base is the directory that contains
// the .env files (normally "..").
func loadEnv(base string) error {
	envName := os.Getenv("LINKSHORTENER_ENV")
	if envName != string(EnvDev) && envName != string(EnvProd) {
		return fmt.Errorf("LINKSHORTENER_ENV must be \"dev\" or \"prod\", got %q", envName)
	}

	if envName == string(EnvProd) {
		// prod mode: skip all file reading entirely.
		return nil
	}

	// dev mode: read base/.env, then base/.env.dev.
	fileVars, err := parseEnvFile(base + "/.env")
	if err != nil {
		return fmt.Errorf("loading .env: %w", err)
	}

	devVars, err := parseEnvFile(base + "/.env.dev")
	if err != nil {
		return fmt.Errorf("loading .env.dev: %w", err)
	}

	// .env.dev overrides .env.
	for k, v := range devVars {
		fileVars[k] = v
	}

	// Apply collected file values, but OS non-empty values always win.
	for key, fileVal := range fileVars {
		if key == "LINKSHORTENER_ENV" {
			// Never let files override this variable.
			continue
		}
		if osVal := os.Getenv(key); osVal != "" {
			// OS value is non-empty — it takes precedence.
			continue
		}
		if err := os.Setenv(key, fileVal); err != nil {
			return fmt.Errorf("setting env var %s: %w", key, err)
		}
	}

	return nil
}

// Validate checks that all required variables are set and applies documented
// defaults for optional variables that are unset. Returns an error if any
// required variable is missing or any numeric variable holds a non-integer value.
func Validate() error {
	for _, key := range requiredVars {
		if os.Getenv(key) == "" {
			return fmt.Errorf("required environment variable %s is not set", key)
		}
	}

	for key, defaultVal := range optionalDefaults {
		if os.Getenv(key) != "" {
			continue
		}
		if err := os.Setenv(key, defaultVal); err != nil {
			return fmt.Errorf("setting default for %s: %w", key, err)
		}
	}

	for key := range numericOptional {
		val := os.Getenv(key)
		if _, err := strconv.Atoi(val); err != nil {
			return fmt.Errorf("environment variable %s must be an integer, got %q", key, val)
		}
	}

	return nil
}

// LogConfig emits a startup configuration summary via slog at INFO level.
// Required variables (§4.1) are always logged. Optional variables (§4.2) are
// logged only when their resolved value differs from the documented default.
// In prod mode secret values are masked; in dev mode all values are printed as-is.
func LogConfig() {
	isProd := GetAppEnv() == EnvProd

	// LINKSHORTENER_ENV is a required variable and is always logged.
	slog.Info("config", "key", "LINKSHORTENER_ENV", "value", os.Getenv("LINKSHORTENER_ENV"))

	for _, key := range requiredVars {
		slog.Info("config", "key", key, "value", maskedValue(key, os.Getenv(key), isProd))
	}

	for key, defaultVal := range optionalDefaults {
		val := os.Getenv(key)
		if val == defaultVal {
			continue
		}
		slog.Info("config", "key", key, "value", maskedValue(key, val, isProd))
	}
}

// maskedValue returns the log-safe representation of val for the given key.
// In dev mode no masking is applied. In prod mode secrets are replaced with "****"
// and DATABASE_URL has only its password component masked.
func maskedValue(key, val string, isProd bool) string {
	if !isProd {
		return val
	}
	if secretVars[key] {
		return "****"
	}
	if key == "DATABASE_URL" {
		return maskDSN(val)
	}
	return val
}

// maskDSN masks the password component of a PostgreSQL connection string.
// It handles URI format (postgres://user:pass@host/db) and
// key=value format (host=... password=... dbname=...).
func maskDSN(dsn string) string {
	if strings.Contains(dsn, "://") {
		// Locate the userinfo block between "://" and the first "@".
		schemeEnd := strings.Index(dsn, "://") + 3
		atIdx := strings.Index(dsn[schemeEnd:], "@")
		if atIdx < 0 {
			// No userinfo block — nothing to mask.
			return dsn
		}
		atIdx += schemeEnd

		userInfo := dsn[schemeEnd:atIdx]
		colonIdx := strings.LastIndex(userInfo, ":")
		if colonIdx < 0 {
			// No colon in userinfo — no password present.
			return dsn
		}

		// Replace everything after the colon in userinfo with "****".
		return dsn[:schemeEnd] + userInfo[:colonIdx+1] + "****" + dsn[atIdx:]
	}

	// key=value format: mask the password= token.
	parts := strings.Fields(dsn)
	for i, part := range parts {
		if strings.HasPrefix(strings.ToLower(part), "password=") {
			eq := strings.IndexByte(part, '=')
			parts[i] = part[:eq+1] + "****"
		}
	}
	return strings.Join(parts, " ")
}

// parseEnvFile reads a .env-style file and returns its key/value pairs.
// Missing files are silently ignored (returns empty map, nil error).
// Each non-blank, non-comment line must follow the format: KEY=VALUE
// Values may optionally be surrounded by single or double quotes.
func parseEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path) //nolint:gosec // path is constructed from safe constants
	if os.IsNotExist(err) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close() //nolint:errcheck // read-only file, close error is not actionable

	vars := make(map[string]string)
	scanner := bufio.NewScanner(f)
	lineNo := 0

	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		eq := strings.IndexByte(line, '=')
		if eq < 0 {
			return nil, fmt.Errorf("%s line %d: missing '='", path, lineNo)
		}

		key := strings.TrimSpace(line[:eq])
		val := strings.TrimSpace(line[eq+1:])

		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') ||
				(val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}

		if key != "" {
			vars[key] = val
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	return vars, nil
}
