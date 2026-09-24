package redis

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// newTestStore starts a miniredis server and returns it together with a
// Store backed by it. Skips the test when miniredis cannot start.
func newTestStore(t *testing.T, opts ...Option) (*miniredis.Miniredis, *Store) {
	t.Helper()

	mr, err := miniredis.Run()
	if err != nil {
		t.Skipf("miniredis unavailable: %v", err)
	}
	t.Cleanup(mr.Close)

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	return mr, New(client, opts...)
}

func TestAllowWithinLimit(t *testing.T) {
	_, store := newTestStore(t)

	// max=3 per second, burst=0 -> budget of 3 per window.
	for i := 1; i <= 3; i++ {
		if !store.Allow("k", 3, 0, time.Second) {
			t.Fatalf("request %d within limit was blocked", i)
		}
	}
}

func TestBlockOverLimit(t *testing.T) {
	_, store := newTestStore(t)

	for i := 1; i <= 3; i++ {
		store.Allow("k", 3, 0, time.Second)
	}
	if store.Allow("k", 3, 0, time.Second) {
		t.Fatal("4th request in window was allowed, want blocked")
	}

	// A different key has its own budget.
	if !store.Allow("other", 3, 0, time.Second) {
		t.Fatal("different key was blocked by another key's budget")
	}
}

func TestTTLExpiryResetsWindow(t *testing.T) {
	mr, store := newTestStore(t)

	for i := 1; i <= 3; i++ {
		store.Allow("k", 3, 0, time.Second)
	}
	if store.Allow("k", 3, 0, time.Second) {
		t.Fatal("request over limit was allowed before window expiry")
	}

	// Fast-forward past the window: the counter expires and the budget resets.
	mr.FastForward(2 * time.Second)
	if !store.Allow("k", 3, 0, time.Second) {
		t.Fatal("request after window expiry was blocked, want allowed")
	}
}

func TestFailOpenWhenServerClosed(t *testing.T) {
	mr, store := newTestStore(t) // default is fail-open
	mr.Close()

	if !store.Allow("k", 3, 0, time.Second) {
		t.Fatal("default store blocked request on Redis error, want fail-open (allowed)")
	}
}

func TestFailClosedWhenConfigured(t *testing.T) {
	mr, _ := newTestStore(t)
	addr := mr.Addr()
	mr.Close()

	client := redis.NewClient(&redis.Options{Addr: addr})
	defer client.Close() //nolint:errcheck

	closed := New(client, WithFailOpen(false))
	if closed.Allow("k", 3, 0, time.Second) {
		t.Fatal("fail-closed store allowed request on Redis error, want blocked")
	}
}

func TestWindowTTLDefault(t *testing.T) {
	mr, store := newTestStore(t, WithWindowTTLDefault(50*time.Millisecond))

	store.Allow("k", 1, 0, 0) // ttl <= 0 -> default window applies
	mr.FastForward(100 * time.Millisecond)

	if !store.Allow("k", 1, 0, 0) {
		t.Fatal("counter did not expire with the default window TTL")
	}
}
