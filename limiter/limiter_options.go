package limiter

import (
	"time"
)

// ExpirableOptions are options used for new limiter creation
type ExpirableOptions struct {
	DefaultExpirationTTL time.Duration

	// How frequently expire job triggers
	ExpireJobInterval time.Duration
}
