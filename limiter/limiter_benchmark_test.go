package limiter

import (
	"testing"
	"time"
)

func BenchmarkLimitReached(b *testing.B) {
	lmt := New(1, time.Second)
	key := "127.0.0.1|/"

	for i := 0; i < b.N; i++ {
		lmt.LimitReached(key)
	}
}

func BenchmarkLimitReachedWithExpiringBuckets(b *testing.B) {
	lmt := NewWithExpiringBuckets(1, time.Second, time.Minute, 30*time.Second)
	key := "127.0.0.1|/"

	for i := 0; i < b.N; i++ {
		lmt.LimitReached(key)
	}
}
