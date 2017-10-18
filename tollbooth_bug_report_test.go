// +build slow
// How to run: go test -tags=slow

package tollbooth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/didip/tollbooth/limiter"
)

// See: https://github.com/didip/tollbooth/issues/48
func Test_Issue48_RequestTerminatedEvenOnLowVolumeOnSameIP(t *testing.T) {
	lmt := limiter.New(nil).SetMax(20).SetTTL(time.Second)
	lmt.SetMethods([]string{"GET"})

	limitReachedCounter := 0
	lmt.SetOnLimitReached(func(w http.ResponseWriter, r *http.Request) { limitReachedCounter++ })

	handler := LimitHandler(lmt, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`hello world`))
	}))

	// The issue seen by the reporter is that the limiter slowly "leaks", causing requests
	// to fail after a prolonged period of continuous usage. Try to model that here.
	//
	// Report stated that a constant 2 requests per second over several minutes would cause
	// a limit of 20/req/sec to start returning 429.
	timeout := time.After(10 * time.Minute)
	iterations := 0
	start := time.Now()
	for {
		select {
		case <-timeout:
			break
		case <-time.After(500 * time.Millisecond):
			req, _ := http.NewRequest("GET", "/doesntmatter", nil)
			req.RemoteAddr = "127.0.0.1"

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Fatalf("Should be able to handle 20 reqs/second. HTTP status: %v. Expected HTTP status: %v. Failed after %d iterations in %f seconds.", status, http.StatusOK, iterations, time.Since(start).Seconds())
			}
			iterations++
		}
	}

	if limitReachedCounter > 0 {
		t.Fatalf("We should never reached the limit, the counter should be 0. limitReachedCounter: %v", limitReachedCounter)
	}
}

// See: https://github.com/didip/tollbooth/issues/48
func Test_Issue48_RequestTerminatedEvenOnLowVolumeOnDifferentIPs(t *testing.T) {
	lmt := limiter.New(nil).SetMax(20).SetTTL(time.Second)
	lmt.SetMethods([]string{"GET"})

	limitReachedCounter := 0
	lmt.SetOnLimitReached(func(w http.ResponseWriter, r *http.Request) { limitReachedCounter++ })

	handler := LimitHandler(lmt, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`hello world`))
	}))

	// Why 11 times?
	// Because 11 * 2 = 22, and our limit is 20.
	// If the bug report is as what I understood, then this test is expected to break.
	tries := 11
	for i := 0; i < tries; i++ {
		// Twice per second should not be limited
		for j := 0; j < 2; j++ {
			req, _ := http.NewRequest("GET", "/doesntmatter", nil)
			req.RemoteAddr = fmt.Sprintf("127.0.0.%v", i)

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
