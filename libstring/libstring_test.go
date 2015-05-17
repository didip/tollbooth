package libstring

import (
	"testing"
)

func TestFlattenMapSliceString(t *testing.T) {
	headersToCheck := make(map[string][]string)
	headersToCheck["X-Auth-Token"] = []string{"abc123", "brotato!23"}

	for i, flatten := range FlattenMapSliceString(headersToCheck, "headers", "") {
		if flatten != "headers:X-Auth-Token:"+headersToCheck["X-Auth-Token"][i] {
			t.Errorf("Failed to flatten map correctly. Result: %v", flatten)
		}
	}
}
