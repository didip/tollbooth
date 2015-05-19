package libstring

import (
	"testing"
)

func TestStringInSlice(t *testing.T) {
	if StringInSlice([]string{"alice", "dan", "didip", "jason", "karl"}, "brotato") {
		t.Error("brotato should not be in slice.")
	}
}
