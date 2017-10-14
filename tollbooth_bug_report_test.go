// +build slow

package tollbooth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/didip/tollbooth/limiter"
)

// See: https://github.com/didip/tollbooth/issues/48
// How to run: go test -tags=slow
func Test_Issue48_RequestTerminatedEvenOnLowVolume(t *testing.T) {
	lmt := limiter.New(nil).SetMax(20).SetTTL(time.Second)
	lmt.SetMethods([]string{"GET"})

	limitReachedCounter := 0
	lmt.SetOnLimitReached(func(w http.ResponseWriter, r *http.Request) { limitReachedCounter++ })

	handler := LimitHandler(lmt, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`hello world`))
	}))

	req, err := http.NewRequest("GET", "/doesntmatter", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Why 11 times?
	// Because 11 * 2 = 22, and our limit is 20.
	// If the bug report is as what I understood, then this test is expected to break.
	tries := 11
	for i := 0; i < tries; i++ {
		// Twice per second should not be limited
		for j := 0; j < 2; j++ {
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			//Should not be limited
			if status := rr.Code; status != http.StatusOK {
				t.Fatalf("Should be able to handle 20 reqs/second. HTTP status: %v. Expected HTTP status: %v", status, http.StatusOK)
			}
		}

		time.Sleep(time.Second)
	}

	if limitReachedCounter > 0 {
		t.Fatalf("We should never reached the limit, the counter should be 0. limitReachedCounter: %v", limitReachedCounter)
	}
}
