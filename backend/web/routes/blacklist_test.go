package routes_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/AlexL70/linkshortener/backend/web/routes"
)

func TestBlacklist_AddAndIsBlacklisted(t *testing.T) {
	bl := routes.NewTokenBlacklist()
	future := time.Now().Add(time.Hour)

	assert.False(t, bl.IsBlacklisted("jti-1"), "new blacklist should not contain any jti")

	bl.Add("jti-1", future)
	assert.True(t, bl.IsBlacklisted("jti-1"), "jti should be blacklisted after Add")
}

func TestBlacklist_ExpiredEntryIsNotBlacklisted(t *testing.T) {
	bl := routes.NewTokenBlacklist()
	// Add an entry that is already expired.
	past := time.Now().Add(-time.Second)
	bl.Add("jti-expired", past)

	assert.False(t, bl.IsBlacklisted("jti-expired"),
		"expired entry should not be reported as blacklisted")
}

func TestBlacklist_ExpiredEntryIsEvictedLazily(t *testing.T) {
	bl := routes.NewTokenBlacklist()
	past := time.Now().Add(-time.Second)
	bl.Add("jti-lazy", past)

	// First call evicts the expired entry.
	bl.IsBlacklisted("jti-lazy")
	// A second call should also return false without panic.
	assert.False(t, bl.IsBlacklisted("jti-lazy"))
}

func TestBlacklist_UnknownJTI(t *testing.T) {
	bl := routes.NewTokenBlacklist()
	assert.False(t, bl.IsBlacklisted("unknown-jti"))
}

func TestBlacklist_ConcurrentSafety(t *testing.T) {
	bl := routes.NewTokenBlacklist()
	future := time.Now().Add(time.Hour)
	const goroutines = 50

	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	// Concurrent writers.
	for i := range goroutines {
		go func(n int) {
			defer wg.Done()
			jti := fmt.Sprintf("jti-%d", n)
			bl.Add(jti, future)
		}(i)
	}

	// Concurrent readers.
	for i := range goroutines {
		go func(n int) {
			defer wg.Done()
			jti := fmt.Sprintf("jti-%d", n)
			bl.IsBlacklisted(jti) // result may be true or false depending on timing
		}(i)
	}

	wg.Wait() // must not panic or race
}
