// Package redis provides an optional shared counter store backed by Redis,
// implementing [github.com/didip/tollbooth/v8/limiter/store.Store].
//
// Multiple instances of an application (for example, replicas behind a load
// balancer) can each construct a tollbooth [Limiter] wired to the same Redis
// server via [github.com/didip/tollbooth/v8/limiter.Limiter.SetStore]; the
// instances then enforce one shared rate-limit budget instead of one budget
// per process.
//
// # Semantics
//
// The store implements a fixed-window counter, not a token bucket. Within
// each window a key gets a budget derived from the Limiter's max and burst
// settings (see [Store.Allow]); once the budget is consumed, further requests
// are blocked until the window expires. The local token bucket smooths
// bursts over time (a burst does not immediately exhaust a long window),
// while a fixed window resets the full budget at every window boundary, so
// the effective burst behavior differs: worst case, a client may spend the
// entire budget at the start of one window and again right after the reset.
//
// # Failure behavior
//
// If Redis returns an error, the store fails open by default: the request is
// allowed and no counter is updated. This keeps tollbooth's availability
// when the shared store is down, at the cost of temporarily unthrottled
// traffic. Pass [WithFailOpen](false) to fail closed instead: on error, all
// requests are blocked until Redis responds again.
//
// # Interoperability
//
// The state is plain Redis counters with no hidden encoding, so other
// processes (in any language) can read, write, or seed the same keys and
// participate in the same budget.
package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/didip/tollbooth/v8/limiter/store"
)

// Lua script implementing an atomic fixed-window counter.
//
// KEYS[1] is the counter key. ARGV[1] is the per-window budget (integer),
// ARGV[2] is the window TTL in milliseconds.
//
// INCR the counter; when the counter is new (value becomes 1) set its expiry
// with PEXPIRE. The budget is passed in from Go so the script itself stays
// independent of the max/burst formula. Returns the incremented count.
var incrScript = redis.NewScript(`
local current = redis.call('INCR', KEYS[1])
if current == 1 then
    redis.call('PEXPIRE', KEYS[1], ARGV[2])
end
return current
`)

// Store is a fixed-window rate-limit counter stored in Redis.
// It is safe for concurrent use.
type Store struct {
	client   redis.Scripter
	failOpen bool
	// defaultTTL is used when a caller passes ttl <= 0 to Allow.
	defaultTTL time.Duration
}

// Option configures a [Store].
type Option func(*Store)

// WithFailOpen controls what happens when Redis returns an error.
// By default (fail-open, the zero configuration) a request is allowed when
// Redis cannot be reached, trading strict enforcement for availability.
// WithFailOpen(false) makes the store fail closed: on error every request is
// blocked until Redis responds again.
func WithFailOpen(failOpen bool) Option {
	return func(s *Store) { s.failOpen = failOpen }
}

// WithWindowTTLDefault sets the window TTL used by [Store.Allow] when the
// caller passes ttl <= 0. The default is one hour.
func WithWindowTTLDefault(ttl time.Duration) Option {
	return func(s *Store) { s.defaultTTL = ttl }
}

// New creates a Redis-backed shared counter store.
//
// The client must not be nil; any *redis.Client, *redis.ClusterClient or
// *redis.UniversalClient satisfies the accepted interface. By default the
// store fails open on Redis errors; see [WithFailOpen].
func New(client redis.Scripter, opts ...Option) *Store {
	s := &Store{
		client:     client,
		failOpen:   true,
		defaultTTL: time.Hour,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Allow atomically increments the fixed-window counter for key and reports
// whether the request is under the limit.
//
// The per-window budget is max + burst: max is the sustained rate the
// Limiter is configured with (requests per second) and burst is the extra
// allowance a local token bucket would start with.
//
// ttl is the window duration: a counter is created with this expiry when the
// key is new, and the budget resets when it expires. If ttl <= 0, the value
// set with [WithWindowTTLDefault] applies.
//
// On a Redis error, the store fails open (returns true) or closed (returns
// false) depending on [WithFailOpen].
func (s *Store) Allow(key string, max float64, burst int, ttl time.Duration) bool {
	if ttl <= 0 {
		ttl = s.defaultTTL
	}

	budget := int(max) + burst
	if budget < 1 {
		budget = 1
	}

	count, err := incrScript.Run(context.Background(), s.client, []string{key}, budget, ttl.Milliseconds()).Int()
	if err != nil {
		// Redis is unreachable or errored: honor the configured failure mode.
		return s.failOpen
	}

	return count <= budget
}

// Compile-time check that *Store satisfies the tollbooth store interface.
var _ store.Store = (*Store)(nil)
