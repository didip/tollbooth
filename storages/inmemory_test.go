package storages

import (
	"testing"
	"time"
)

func TestCRUD(t *testing.T) {
	inmem := NewInMemory()

	key := "/|127.0.0.1"

	count, exists := inmem.Get(key)
	if exists || count > 0 {
		t.Errorf("Expected empty inmem to return no count")
	}

	inmem.IncrBy("/|127.0.0.1", int64(1), time.Second)
	count, exists = inmem.Get(key)
	if !exists {
		t.Errorf("Expected inmem to return count for key: %v", key)
	}
	if count != 1 {
		t.Errorf("Expected inmem to return 1 for key: %v", key)
	}

	<-time.After(2 * time.Second)
	_, exists = inmem.Get(key)
	if exists {
		t.Errorf("Expected key: %v to have expired.", key)
	}
}
