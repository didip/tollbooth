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
