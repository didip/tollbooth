package tollbooth

import (
	"testing"
	"time"

	"github.com/didip/tollbooth/storages"
)

func TestLimitByKeyParts(t *testing.T) {
	storage := storages.NewInMemory()
	limiter := NewLimiter(1, time.Second) // Only 1 request per second is allowed.

	httperror := LimitByKeyParts(storage, limiter, []string{"127.0.0.1", "/"})
	if httperror != nil {
		t.Errorf("First time count should not return error. Error: %v", httperror.Error())
	}

	httperror = LimitByKeyParts(storage, limiter, []string{"127.0.0.1", "/"})
	if httperror == nil {
		t.Errorf("Second time count should return error because it exceeds 1 request per second.")
	}

	<-time.After(1 * time.Second)
	httperror = LimitByKeyParts(storage, limiter, []string{"127.0.0.1", "/"})
	if httperror != nil {
		t.Errorf("Third time count should not return error because the 1 second window has passed.")
	}
}
