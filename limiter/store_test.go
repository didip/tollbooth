package limiter

import (
	"sync"
	"testing"
	"time"
)

// countingStore is a minimal in-memory Store fake: a fixed-window counter
// with an atomic increment, enough to observe whether multiple Limiter
// instances share one budget through a single store.
type countingStore struct {
	mu     sync.Mutex
	counts map[string]int
}

func newCountingStore() *countingStore {
	return &countingStore{counts: make(map[string]int)}
}

func (c *countingStore) Allow(key string, max float64, burst int, ttl time.Duration) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.counts[key]++
	return c.counts[key] <= int(max)+burst
}

func (c *countingStore) count(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.counts[key]
}

// TestSharedStoreEnforcesOneBudget proves that two Limiter instances wired to
// the same Store enforce ONE shared budget: with max=3, three requests are
// allowed in total even though each Limiter only sees two of them.
func TestSharedStoreEnforcesOneBudget(t *testing.T) {
	shared := newCountingStore()

	lmtOne := New(nil).SetMax(3).SetBurst(0).SetStore(shared)
	lmtTwo := New(nil).SetMax(3).SetBurst(0).SetStore(shared)

	allowed := 0
	for i := 0; i < 2; i++ {
		if !lmtOne.LimitReached("shared-key") {
			allowed++
		}
		if !lmtTwo.LimitReached("shared-key") {
			allowed++
		}
	}

	if allowed != 3 {
		t.Fatalf("allowed = %d, want 3 (two limiters must share one budget of max=3)", allowed)
	}
	if !lmtOne.LimitReached("shared-key") || !lmtTwo.LimitReached("shared-key") {
		t.Fatal("4th request was allowed, want blocked")
	}
	if shared.count("shared-key") != 6 {
		t.Fatalf("store saw %d Allow calls, want 6", shared.count("shared-key"))
	}
}

// TestSetStoreNilRestoresDefaultPath proves that after SetStore(nil) the
// Limiter goes back to its local token buckets: a fresh budget per instance,
// unaffected by requests counted while the shared store was attached.
func TestSetStoreNilRestoresDefaultPath(t *testing.T) {
	shared := newCountingStore()

	// max=1, burst=1: shared budget is max+burst = 2.
	lmt := New(nil).SetMax(1).SetBurst(1).SetStore(shared)

	if lmt.LimitReached("k") {
		t.Fatal("first request through shared store was blocked")
	}
	if lmt.LimitReached("k") {
		t.Fatal("second request through shared store was blocked")
	}
	if !lmt.LimitReached("k") {
		t.Fatal("third request through shared store was allowed, want blocked over shared budget")
	}

	lmt.SetStore(nil)
	if lmt.GetStore() != nil {
		t.Fatal("GetStore() != nil after SetStore(nil)")
	}

	// Default path: the local token bucket is untouched (burst=1 available),
	// so the next request is allowed even though the shared counter for "k"
	// is already exhausted.
	if lmt.LimitReached("k") {
		t.Fatal("first request after SetStore(nil) was blocked, want default token-bucket path")
	}
	if !lmt.LimitReached("k") {
		t.Fatal("second request after SetStore(nil) was allowed, want default token-bucket path")
	}
}
