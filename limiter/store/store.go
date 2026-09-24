// Package store defines an optional interface for sharing rate-limit
// counters across multiple instances of an application (for example,
// instances behind a load balancer).
//
// By default, tollbooth keeps its counters in the local process. When a
// Limiter is configured with a non-nil Store via
// [github.com/didip/tollbooth/v8/limiter.Limiter.SetStore], the Limiter
// delegates the limit decision to the store instead of its local token
// bucket, so that every instance wired to the same store enforces one
// shared budget.
//
// A Store implementation must be safe for concurrent use, and its Allow
// method must atomically perform the get-or-create plus decrement of the
// counter identified by key, so that concurrent callers cannot exceed the
// configured limit.
//
// Note: a counter-based store does not reproduce the local token bucket's
// smoothing of bursts. Implementations typically enforce a fixed window
// (see the store/redis submodule for one); callers should review the
// documentation of whichever implementation they choose.
package store

import "time"

// Store is an optional interface for sharing rate-limit counters
// across multiple instances of an application.
//
// Allow atomically performs the get-or-create plus decrement of the
// counter identified by key, and reports whether the request is allowed:
// it returns true if the request is under the limit (max requests per
// second with the given burst), and false once the limit is reached.
//
// ttl is the duration after which the counter for key expires and a fresh
// budget begins. Implementations must be safe for concurrent use.
type Store interface {
	Allow(key string, max float64, burst int, ttl time.Duration) bool
}
