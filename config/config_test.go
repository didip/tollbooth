package config

import (
	"testing"
	"time"
)

func TestConstructor(t *testing.T) {
	limiter := NewLimiter(1, time.Second)
	if limiter.Max != 1 {
		t.Errorf("Max field is incorrect. Value: %v", limiter.Max)
	}
	if limiter.TTL != time.Second {
		t.Errorf("TTL field is incorrect. Value: %v", limiter.TTL)
	}
	if limiter.Message != "You have reached maximum request limit." {
		t.Errorf("Message field is incorrect. Value: %v", limiter.Message)
	}
	if limiter.StatusCode != 429 {
		t.Errorf("StatusCode field is incorrect. Value: %v", limiter.StatusCode)
	}
}

func TestLimitReached(t *testing.T) {
	limiter := NewLimiter(1, time.Second)
	key := "127.0.0.1|/"

	if limiter.LimitReached(key) == true {
		t.Error("First time count should not reached the limit.")
	}

	// Off by 1 error.
	if limiter.LimitReached(key) == true {
		t.Error("Second time count should not reached the limit.")
	}

	if limiter.LimitReached(key) == false {
		t.Error("Third time count should return true because it exceeds 1 request per second.")
	}

	<-time.After(1 * time.Second)
	if limiter.LimitReached(key) == true {
		t.Error("Fourth time count should not reached the limit because the 1 second window has passed.")
	}
}
