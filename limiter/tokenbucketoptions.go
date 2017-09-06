package limiter

import (
	"time"
)

type TokenBucketOptions struct {
	// Default TTL to expire bucket per key basis.
	DefaultExpirationTTL time.Duration

	// How frequently tollbooth will trigger the expire job
	ExpireJobInterval time.Duration
}
