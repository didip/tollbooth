package tollbooth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/didip/tollbooth/v6/limiter"
)

// See: https://github.com/didip/tollbooth/issues/48
func Test_Issue48_RequestTerminatedEvenOnLowVolumeOnSameIP(t *testing.T) {
	// The issue seen by the reporter is that the limiter slowly "leaks", causing requests
	// to fail after a prolonged period of continuous usage. Try to model that here.
	//
	// Report stated that a constant 2 requests per second over several minutes would cause
	// a limit of 2/req/sec to start returning 429.

	requestsPerSecond := float64(2)

	lmt := NewLimiter(requestsPerSecond, nil)
	lmt.SetMethods([]string{"GET"})

	limitReachedCounter := 0
	lmt.SetOnLimitReached(func(w http.ResponseWriter, r *http.Request) { limitReachedCounter++ })

	handler := LimitHandler(lmt, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`hello world`))
	}))

	timeout := time.After(1 * time.Second)
	start := time.Now()

	// Create the HTTP request
	req, _ := http.NewRequest("GET", "/doesntmatter", nil)
	req.RemoteAddr = "127.0.0.1"

Top:
	for {
		select {
		case <-timeout:
			break Top
		case <-time.After(500 * time.Millisecond):
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("Should be able to handle %v reqs/second. HTTP status: %v. Expected HTTP status: %v. Failed in %v seconds. limitReachedCounter: %d", requestsPerSecond, rr.Code, http.StatusOK, time.Since(start).Seconds(), limitReachedCounter)
			}
		}
	}

	if limitReachedCounter > 0 {
		t.Fatalf("We should never reached the limit, the counter should be 0. limitReachedCounter: %v", limitReachedCounter)
	}
}

var issue66HeaderKey = "X-Customer-ID"

func issue66RateLimiter(h http.HandlerFunc, customerIDs []string) (http.HandlerFunc, *limiter.Limiter) {
	allocationLimiter := NewLimiter(1, nil).SetMethods([]string{"POST"})

	handler := func(w http.ResponseWriter, r *http.Request) {
		allocationLimiter.SetHeader(issue66HeaderKey, customerIDs)
		LimitFuncHandler(allocationLimiter, h).ServeHTTP(w, r)
	}

	return handler, allocationLimiter
}

// See: https://github.com/didip/tollbooth/issues/66
func Test_Issue66_CustomRateLimitByHeaderValues(t *testing.T) {
	customerID1 := "1234"
	customerID2 := "5678"

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	h, allocationLimiter := issue66RateLimiter(h, []string{customerID1, customerID2})
	testServer := httptest.NewServer(h)
	defer testServer.Close()

	client := &http.Client{}

	// subtest 1:
	// There are 2 different customer ids,
	// both should pass rate limiter the 1st time and failed the second time.
	request1, _ := http.NewRequest("POST", testServer.URL, nil)
	request1.Header.Add(issue66HeaderKey, customerID1)

	request2, _ := http.NewRequest("POST", testServer.URL, nil)
	request2.Header.Add(issue66HeaderKey, customerID2)

	for _, request := range []*http.Request{request1} {
		// 1st, 200
		response, _ := client.Do(request)
		if response.StatusCode != http.StatusOK {
			t.Fatalf(`
Customer %v must pass rate limiter the first time.
Expected to receive: %v status code. Got: %v.
limiter.headers: %v`,
				request.Header.Get(issue66HeaderKey),
				http.StatusOK, response.StatusCode,
				allocationLimiter.GetHeaders())
		}

		// 2nd, 429
		response, _ = client.Do(request)
		if response.StatusCode != http.StatusTooManyRequests {
			t.Fatalf(`Both customer must fail rate limiter.
Expected to receive: %v status code. Got: %v`,
				http.StatusTooManyRequests, response.StatusCode)
		}
	}

	// subtest 2:
	// There is 1 customer not defined in rate limiter.
	// S/he should not be rate limited.
	request3, _ := http.NewRequest("POST", testServer.URL, nil)
	request3.Header.Add(issue66HeaderKey, "777")

	for i := 0; i < 2; i++ {
		response, _ := client.Do(request3)

		if response.StatusCode != http.StatusOK {
			t.Fatalf(`
Customer %v must always pass rate limiter.
Expected to receive: %v status code. Got: %v`,
				request3.Header.Get(issue66HeaderKey),
				http.StatusOK, response.StatusCode)
		}
	}
}

func Test_Issue91_BrokenSetMethod_DontBlockGet(t *testing.T) {
	requestsPerSecond := float64(1)

	lmt := NewLimiter(requestsPerSecond, nil)
	lmt.SetMethods([]string{"POST"})

	methods := lmt.GetMethods()
	if methods[0] != "POST" {
		t.Fatalf("Failed to set methods correctly. Expected: POST Got: %v", methods[0])
	}

	// -------------------------------------------------------------------

	handler := LimitHandler(lmt, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`hello world`))
	}))

	// Create GET HTTP request
	req, _ := http.NewRequest("GET", "/doesntmatter", nil)
	req.RemoteAddr = "127.0.0.1"

	// We should never reach the limit because we are sending 10 GET requests and
	// we are only limiting POST requests.
	for i := 0; i < 10; i++ {
		start := time.Now()

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Should be able to handle %v reqs/second. HTTP status: %v. Expected HTTP status: %v. Failed in %v microseconds", requestsPerSecond, rr.Code, http.StatusOK, time.Since(start).Microseconds())
		}
	}
}

func Test_Issue91_BrokenSetMethod_BlockPost(t *testing.T) {
	requestsPerSecond := float64(1)

	lmt := NewLimiter(requestsPerSecond, nil)
	lmt.SetMethods([]string{"POST"})

	limitReachedCounter := 0
	lmt.SetOnLimitReached(func(w http.ResponseWriter, r *http.Request) {
		limitReachedCounter++
	})

	methods := lmt.GetMethods()
	if methods[0] != "POST" {
		t.Fatalf("Failed to set methods correctly. Expected: POST Got: %v", methods[0])
	}

	// -------------------------------------------------------------------

	handler := LimitHandler(lmt, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`hello world`))
	}))

	// Create POST HTTP request
	req, _ := http.NewRequest("POST", "/blockmeafter2", nil)
	req.RemoteAddr = "127.0.0.1"

	// We should reach the limit because we are sending 2 POST requests and
	// our limiter is 1 POST per second.
	for i := 0; i < 2; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	if limitReachedCounter == 0 {
		t.Fatalf("Should have reached limit. Limit reached counter: %d", limitReachedCounter)
	}
}
