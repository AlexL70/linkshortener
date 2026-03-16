// Package testutil provides shared helpers for backend integration tests.
package testutil

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlexL70/linkshortener/backend/config"
)

// FindRepoRoot walks up from the current working directory until it finds a
// directory containing AGENTS.md, which marks the repository root.
func FindRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("testutil.FindRepoRoot: getwd: %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("testutil.FindRepoRoot: AGENTS.md not found in any parent directory")
		}
		dir = parent
	}
}

// LoadTestEnv loads dev environment variables from .env / .env.dev files in the
// repository root, mirroring what the main application does at startup.
//
// If LINKSHORTENER_ENV is not already set, it is set to "dev" so that
// config.LoadEnvFrom reads the env files. OS-level values always win over
// file values — the same precedence rule as in production.
//
// It is safe to call LoadTestEnv multiple times; subsequent calls are
// effectively no-ops because OS env vars set by the first call prevent file
// values from overwriting them.
func LoadTestEnv() error {
	if os.Getenv("LINKSHORTENER_ENV") == "" {
		if err := os.Setenv("LINKSHORTENER_ENV", string(config.EnvDev)); err != nil {
			return fmt.Errorf("testutil.LoadTestEnv: set LINKSHORTENER_ENV: %w", err)
		}
	}

	root, err := FindRepoRoot()
	if err != nil {
		return fmt.Errorf("testutil.LoadTestEnv: %w", err)
	}

	if err := config.LoadEnvFrom(root); err != nil {
		return fmt.Errorf("testutil.LoadTestEnv: %w", err)
	}

	return nil
}
