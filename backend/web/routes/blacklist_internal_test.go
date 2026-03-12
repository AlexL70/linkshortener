package routes

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestEvictExpired_RemovesOnlyExpiredEntries exercises the background cleanup
// logic directly without waiting for the hourly ticker to fire.
func TestEvictExpired_RemovesOnlyExpiredEntries(t *testing.T) {
	bl := &TokenBlacklist{entries: make(map[string]time.Time)}

	future := time.Now().Add(time.Hour)
	past := time.Now().Add(-time.Second)

	bl.entries["live"] = future
	bl.entries["dead"] = past

	bl.evictExpired()

	bl.mu.RLock()
	defer bl.mu.RUnlock()
	_, livePresentAfter := bl.entries["live"]
	_, deadPresentAfter := bl.entries["dead"]
	assert.True(t, livePresentAfter, "non-expired entry must remain after eviction")
	assert.False(t, deadPresentAfter, "expired entry must be removed by eviction")
}
