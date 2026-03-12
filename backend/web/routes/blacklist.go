package routes

import (
	"sync"
	"time"
)

// TokenBlacklist tracks JWT IDs (jti) for tokens that have been explicitly
// invalidated (e.g., via logout). Entries are kept until the token's natural
// expiry so that the map does not grow unboundedly.
//
// The blacklist is in-memory only and is cleared on server restart. Because
// full JWTs expire after 24 hours, the maximum window of exposure after a
// restart is bounded by the token lifetime.
type TokenBlacklist struct {
	mu      sync.RWMutex
	entries map[string]time.Time // jti → token expiry time
}

// NewTokenBlacklist creates an empty blacklist and starts a background
// goroutine that evicts stale entries once per hour.
func NewTokenBlacklist() *TokenBlacklist {
	bl := &TokenBlacklist{
		entries: make(map[string]time.Time),
	}
	go bl.cleanupLoop()
	return bl
}

// Add marks the given JWT ID as invalidated. It will be removed automatically
// once expiresAt is in the past.
func (bl *TokenBlacklist) Add(jti string, expiresAt time.Time) {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	bl.entries[jti] = expiresAt
}

// IsBlacklisted reports whether jti is currently on the blacklist.
// It performs lazy eviction: if the entry exists but has already expired it is
// removed and false is returned (the token would be rejected by expiry checking
// in ParseJWT anyway).
func (bl *TokenBlacklist) IsBlacklisted(jti string) bool {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	exp, exists := bl.entries[jti]
	if !exists {
		return false
	}
	if time.Now().After(exp) {
		delete(bl.entries, jti)
		return false
	}
	return true
}

// cleanupLoop runs in the background and removes expired entries once per hour.
func (bl *TokenBlacklist) cleanupLoop() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		bl.evictExpired()
	}
}

// evictExpired removes all entries whose expiry is in the past.
func (bl *TokenBlacklist) evictExpired() {
	now := time.Now()
	bl.mu.Lock()
	defer bl.mu.Unlock()
	for jti, exp := range bl.entries {
		if now.After(exp) {
			delete(bl.entries, jti)
		}
	}
}
