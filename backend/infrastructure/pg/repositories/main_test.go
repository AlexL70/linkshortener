package repositories_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/AlexL70/linkshortener/backend/testutil"
)

func TestMain(m *testing.M) {
	if err := testutil.LoadTestEnv(); err != nil {
		// Warn but do not fatal — DATABASE_URL may already be set in CI via OS env.
		slog.Warn("TestMain: could not load dev env files", "err", err)
	}
	os.Exit(m.Run())
}
