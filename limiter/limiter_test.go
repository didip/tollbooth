package limiter

import (
	"testing"
	"time"
)

func TestConstructor(t *testing.T) {
	lmt := New(1, time.Second)
	if lmt.Max != 1 {
		t.Errorf("Max field is incorrect. Value: %v", lmt.Max)
	}
	if lmt.TTL != time.Second {
		t.Errorf("TTL field is incorrect. Value: %v", lmt.TTL)
	}
	if lmt.Message != "You have reached maximum request limit." {
		t.Errorf("Message field is incorrect. Value: %v", lmt.Message)
	}
	if lmt.StatusCode != 429 {
		t.Errorf("StatusCode field is incorrect. Value: %v", lmt.StatusCode)
	}
}

func TestConstructorExpiringBuckets(t *testing.T) {
	lmt := NewExpiringBuckets(1, time.Second, time.Second, 0)
	if lmt.Max != 1 {
		t.Errorf("Max field is incorrect. Value: %v", lmt.Max)
	}
	if lmt.TTL != time.Second {
		t.Errorf("TTL field is incorrect. Value: %v", lmt.TTL)
	}
	if lmt.TokenBuckets.DefaultExpirationTTL != time.Second {
		t.Errorf("DefaultExpirationTTL field for TokenBuckets is incorrect. Value: %v", lmt.TokenBuckets.DefaultExpirationTTL)
	}
	if lmt.Message != "You have reached maximum request limit." {
		t.Errorf("Message field is incorrect. Value: %v", lmt.Message)
	}
	if lmt.StatusCode != 429 {
		t.Errorf("StatusCode field is incorrect. Value: %v", lmt.StatusCode)
	}
}

func TestLimitReached(t *testing.T) {
	lmt := New(1, time.Second)
	key := "127.0.0.1|/"

	if lmt.LimitReached(key) == true {
		t.Error("First time count should not reached the limit.")
	}

	if lmt.LimitReached(key) == false {
		t.Error("Second time count should return true because it exceeds 1 request per second.")
	}

	<-time.After(1 * time.Second)
	if lmt.LimitReached(key) == true {
		t.Error("Third time count should not reached the limit because the 1 second window has passed.")
	}
}

func TestLimitReachedWithCustomTokenBucketTTL(t *testing.T) {
	lmt := NewExpiringBuckets(1, time.Second, time.Second, 0)
	key := "127.0.0.1|/"

	if lmt.LimitReached(key) == true {
		t.Error("First time count should not reached the limit.")
	}

	if lmt.LimitReached(key) == false {
		t.Error("Second time count should return true because it exceeds 1 request per second.")
	}

	<-time.After(1 * time.Second)
	if lmt.LimitReached(key) == true {
		t.Error("Third time count should not reached the limit because the 1 second window has passed.")
	}
}

func TestMuchHigherMaxRequests(t *testing.T) {
	numRequests := 1000
	lmt := New(int64(numRequests), time.Second)
	key := "127.0.0.1|/"

	for i := 0; i < numRequests; i++ {
		if lmt.LimitReached(key) == true {
			t.Errorf("N(%v) limit should not be reached.", i)
		}
	}

	if lmt.LimitReached(key) == false {
		t.Errorf("N(%v) limit should be reached because it exceeds %v request per second.", numRequests+2, numRequests)
	}

}

func TestMuchHigherMaxRequestsWithCustomTokenBucketTTL(t *testing.T) {
	numRequests := 1000
	lmt := NewExpiringBuckets(int64(numRequests), time.Second, time.Minute, time.Minute)
	key := "127.0.0.1|/"

	for i := 0; i < numRequests; i++ {
		if lmt.LimitReached(key) == true {
			t.Errorf("N(%v) limit should not be reached.", i)
		}
	}

	if lmt.LimitReached(key) == false {
		t.Errorf("N(%v) limit should be reached because it exceeds %v request per second.", numRequests+2, numRequests)
	}

}
