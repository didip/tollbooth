// Package config provides data structure to configure rate-limiter.
package config

import (
	"github.com/juju/ratelimit"
	"time"
)

// NewLimiter is a constructor for Limiter.
func NewLimiter(max int64, ttl time.Duration) *Limiter {
	limiter := &Limiter{Max: max, TTL: ttl}
	limiter.Message = "You have reached maximum request limit."
	limiter.StatusCode = 429
	limiter.tokenBuckets = make(map[string]*ratelimit.Bucket)

	return limiter
}

// Limiter is a config struct to limit a particular request handler.
type Limiter struct {
	// HTTP message when limit is reached.
	Message string

	// HTTP status code when limit is reached.
	StatusCode int

	// Maximum number of requests to limit per duration.
	Max int64

	// Duration of rate-limiter.
	TTL time.Duration

	// List of HTTP Methods to limit (GET, POST, PUT, etc.).
	// Empty means limit all methods.
	Methods []string

	// List of HTTP headers to limit.
	// Empty means skip headers checking.
	Headers map[string][]string

	// List of basic auth usernames to limit.
	BasicAuthUsers []string

	// Throttler struct
	tokenBuckets map[string]*ratelimit.Bucket
}

// LimitReached returns a bool indicating if the Bucket identified by key ran out of tokens.
func (l *Limiter) LimitReached(key string) bool {
	println("Inside LimitReached")

	if _, found := l.tokenBuckets[key]; !found {
		l.tokenBuckets[key] = ratelimit.NewBucket(l.TTL, l.Max)
	}

	_, isSoonerThanMaxWait := l.tokenBuckets[key].TakeMaxDuration(1, l.TTL)
	if isSoonerThanMaxWait {
		return false
	}

	return true
}
