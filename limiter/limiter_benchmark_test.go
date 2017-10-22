package limiter

import (
	"testing"
	"time"
)

func BenchmarkLimitReached(b *testing.B) {
	lmt := New(nil).SetMax(1)
	key := "127.0.0.1|/"

	for i := 0; i < b.N; i++ {
		lmt.LimitReached(key)
	}
}

func BenchmarkLimitReachedWithExpiringBuckets(b *testing.B) {
	lmt := New(&ExpirableOptions{DefaultExpirationTTL: time.Minute, ExpireJobInterval: 30 * time.Second}).SetMax(1)
	key := "127.0.0.1|/"

	for i := 0; i < b.N; i++ {
		lmt.LimitReached(key)
	}
}
