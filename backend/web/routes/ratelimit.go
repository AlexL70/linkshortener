package routes

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// ipLimiterEntry holds the rate limiter for a single IP address and the
// timestamp of the last request from that IP (used for idle-entry cleanup).
type ipLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// perIPStore maintains one rate.Limiter per observed client IP.
// A background goroutine prunes entries for IPs that have been idle for more
// than 2 minutes to prevent unbounded memory growth.
type perIPStore struct {
	mu    sync.Mutex
	store map[string]*ipLimiterEntry
	rpm   int
	burst int
}

// newPerIPStore creates a perIPStore for the given requests-per-minute limit
// and burst size, and starts the background cleanup goroutine.
func newPerIPStore(rpm, burst int) *perIPStore {
	s := &perIPStore{
		store: make(map[string]*ipLimiterEntry),
		rpm:   rpm,
		burst: burst,
	}
	go s.cleanupLoop()
	return s
}

// cleanupLoop runs every minute and removes entries that have been idle for
// longer than 2 minutes. It is started once by newPerIPStore and runs for the
// lifetime of the process.
func (s *perIPStore) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		cutoff := time.Now().Add(-2 * time.Minute)
		s.mu.Lock()
		for ip, entry := range s.store {
			if entry.lastSeen.Before(cutoff) {
				delete(s.store, ip)
			}
		}
		s.mu.Unlock()
	}
}

// Allow returns true when the given IP is within its rate limit and false when
// the limit has been exceeded. It creates a new limiter on first use.
func (s *perIPStore) Allow(ip string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.store[ip]
	if !exists {
		// rate.Every(time.Minute / duration(rpm)) produces a token every
		// (60/rpm) seconds, equivalent to rpm tokens replenished per minute.
		entry = &ipLimiterEntry{
			limiter: rate.NewLimiter(rate.Every(time.Minute/time.Duration(s.rpm)), s.burst),
		}
		s.store[ip] = entry
	}
	entry.lastSeen = time.Now()
	return entry.limiter.Allow()
}

// RateLimitTier associates a set of URL path prefixes with a shared per-IP
// rate limit store. All paths that share a tier share the same RPM cap.
type RateLimitTier struct {
	PathPrefixes []string
	store        *perIPStore
}

// NewRateLimitTier creates a RateLimitTier for the given path prefixes, with
// the configured requests-per-minute limit and burst size.
func NewRateLimitTier(pathPrefixes []string, rpm, burst int) RateLimitTier {
	return RateLimitTier{
		PathPrefixes: pathPrefixes,
		store:        newPerIPStore(rpm, burst),
	}
}

// RateLimitMiddleware returns a Gin middleware that enforces per-IP rate limits
// defined by the provided tiers.
//
// For each incoming request the middleware iterates the tiers and checks whether
// the request path starts with any of the tier's prefixes. The first matching
// tier is used to evaluate the limit. If the IP has exceeded its allowance the
// middleware aborts with 429 Too Many Requests. Requests that match no tier pass
// through without any rate checking.
//
// Accurate client-IP extraction requires that Gin's trusted proxies are
// configured correctly before this middleware is applied (see main.go).
func RateLimitMiddleware(tiers []RateLimitTier) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		for i := range tiers {
			for _, prefix := range tiers[i].PathPrefixes {
				if strings.HasPrefix(path, prefix) {
					if !tiers[i].store.Allow(c.ClientIP()) {
						c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
						return
					}
					c.Next()
					return
				}
			}
		}
		// No tier matched — no rate limit applied.
		c.Next()
	}
}
