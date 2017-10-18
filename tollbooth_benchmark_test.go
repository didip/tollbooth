package tollbooth

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/didip/tollbooth/limiter"
)

func BenchmarkLimitByKeys(b *testing.B) {
	lmt := limiter.New(nil).SetMax(1) // Only 1 request per second is allowed.

	for i := 0; i < b.N; i++ {
		LimitByKeys(lmt, []string{"127.0.0.1", "/"})
	}
}

func BenchmarkLimitByKeysWithExpiringBuckets(b *testing.B) {
	lmt := limiter.New(
		&limiter.ExpirableOptions{DefaultExpirationTTL: time.Minute, ExpireJobInterval: time.Minute},
	).SetMax(1) // Only 1 request per second is allowed.

	for i := 0; i < b.N; i++ {
		LimitByKeys(lmt, []string{"127.0.0.1", "/"})
	}
}

func BenchmarkBuildKeys(b *testing.B) {
	lmt := limiter.New(nil).SetMax(1) // Only 1 request per second is allowed.
	lmt.SetIPLookups([]string{"X-Real-IP", "RemoteAddr", "X-Forwarded-For"})
	lmt.SetHeaders(make(map[string][]string))
	lmt.SetHeader("X-Real-IP", []string{"2601:7:1c82:4097:59a0:a80b:2841:b8c8"})

	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		fmt.Printf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Real-IP", lmt.GetHeader("X-Real-IP")[0])
	for i := 0; i < b.N; i++ {
		sliceKeys := BuildKeys(lmt, request)
		if len(sliceKeys) == 0 {
			fmt.Print("Length of sliceKeys should never be empty.")
		}
	}
}
