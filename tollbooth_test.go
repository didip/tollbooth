package tollbooth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/didip/tollbooth/v8/limiter"
)

func TestLimitByKeys(t *testing.T) {
	lmt := NewLimiter(1, nil) // Only 1 request per second is allowed.

	httperror := LimitByKeys(lmt, []string{"127.0.0.1", "/"})
	if httperror != nil {
		t.Errorf("First time count should not return error. Error: %v", httperror.Error())
	}

	httperror = LimitByKeys(lmt, []string{"127.0.0.1", "/"})
	if httperror == nil {
		t.Errorf("Second time count should return error because it exceeds 1 request per second.")
	}

	<-time.After(1 * time.Second)
	httperror = LimitByKeys(lmt, []string{"127.0.0.1", "/"})
	if httperror != nil {
		t.Errorf("Third time count should not return error because the 1 second window has passed.")
	}
}

func TestDefaultBuildKeys(t *testing.T) {
	lmt := NewLimiter(1, nil).SetIPLookup(limiter.IPLookup{
		Name:           "X-Real-IP",
		IndexFromRight: 0,
	})

	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Real-IP", "2601:7:1c82:4097:59a0:a80b:2841:b8c8")
	expectedIP := "2601:7:1c82:4097::"

	sliceKeys := BuildKeys(lmt, request)
	if len(sliceKeys) == 0 {
		t.Fatal("Length of sliceKeys should never be empty.")
	}

	for _, keys := range sliceKeys {
		expectedKeys := [][]string{
			{expectedIP},
			{request.URL.Path},
		}

		checkKeys(t, keys, expectedKeys)
	}
}

func TestIgnoreURLBuildKeys(t *testing.T) {
	lmt := NewLimiter(1, nil).
		SetIPLookup(limiter.IPLookup{
			Name:           "X-Real-IP",
			IndexFromRight: 0,
		}).
		SetIgnoreURL(true)

	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Real-IP", "172.217.0.46")

	for _, keys := range BuildKeys(lmt, request) {
		for i, keyChunk := range keys {
			if i == 0 && keyChunk != request.Header.Get("X-Real-IP") {
				t.Errorf("The (%v) chunk should be remote IP. KeyChunk: %v", i+1, keyChunk)
			}
		}
	}
}

func TestBasicAuthBuildKeys(t *testing.T) {
	lmt := NewLimiter(1, nil).
		SetIPLookup(limiter.IPLookup{
			Name:           "X-Real-IP",
			IndexFromRight: 0,
		}).
		SetBasicAuthUsers([]string{"bro"})

	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Real-IP", "2601:7:1c82:4097:59a0:a80b:2841:b8c8")
	expectedIP := "2601:7:1c82:4097::"

	request.SetBasicAuth("bro", "tato")

	sliceKeys := BuildKeys(lmt, request)
	if len(sliceKeys) == 0 {
		t.Fatal("Length of sliceKeys should never be empty.")
	}

	for _, keys := range sliceKeys {
		expectedKeys := [][]string{
			{expectedIP},
			{request.URL.Path},
			{"bro"},
		}

		checkKeys(t, keys, expectedKeys)
	}
}

func TestCustomHeadersBuildKeys(t *testing.T) {
	lmt := NewLimiter(1, nil).
		SetIPLookup(limiter.IPLookup{
			Name:           "X-Real-IP",
			IndexFromRight: 0,
		}).
		SetHeader("X-Auth-Token", []string{"totally-top-secret", "another-secret"})

	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Real-IP", "172.217.0.46")
	request.Header.Set("X-Auth-Token", "totally-top-secret")

	sliceKeys := BuildKeys(lmt, request)
	if len(sliceKeys) == 0 {
		t.Fatal("Length of sliceKeys should never be empty.")
	}

	for _, keys := range sliceKeys {
		expectedKeys := [][]string{
			{request.Header.Get("X-Real-IP")},
			{request.URL.Path},
			{"X-Auth-Token"},
			{"totally-top-secret", "another-secret"},
		}

		checkKeys(t, keys, expectedKeys)
	}
}

func TestRequestMethodBuildKeys(t *testing.T) {
	lmt := NewLimiter(1, nil).
		SetIPLookup(limiter.IPLookup{
			Name:           "X-Real-IP",
			IndexFromRight: 0,
		}).
		SetMethods([]string{"GET"})

	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Real-IP", "2601:7:1c82:4097:59a0:a80b:2841:b8c8")
	expectedIP := "2601:7:1c82:4097::"

	sliceKeys := BuildKeys(lmt, request)
	if len(sliceKeys) == 0 {
		t.Fatal("Length of sliceKeys should never be empty.")
	}

	for _, keys := range sliceKeys {
		expectedKeys := [][]string{
			{expectedIP},
			{request.URL.Path},
			{"GET"},
		}

		checkKeys(t, keys, expectedKeys)
	}
}

func TestContextValueBuildKeys(t *testing.T) {
	lmt := NewLimiter(1, nil).
		SetIPLookup(limiter.IPLookup{
			Name:           "X-Real-IP",
			IndexFromRight: 0,
		}).
		SetContextValue("API-access-level", []string{"basic"})

	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Real-IP", "172.217.0.46")
	//nolint:revive,staticcheck // limiter.SetContextValue requires string as a key, so we have to live with that
	request = request.WithContext(context.WithValue(request.Context(), "API-access-level", "basic"))

	sliceKeys := BuildKeys(lmt, request)
	if len(sliceKeys) == 0 {
		t.Fatal("Length of sliceKeys should never be empty.")
	}

	for _, keys := range sliceKeys {
		expectedKeys := [][]string{
			{request.Header.Get("X-Real-IP")},
			{request.URL.Path},
			{"API-access-level"},
			{"basic"},
		}

		checkKeys(t, keys, expectedKeys)
	}
}

func TestRequestMethodAndCustomHeadersBuildKeys(t *testing.T) {
	lmt := NewLimiter(1, nil).
		SetIPLookup(limiter.IPLookup{
			Name:           "X-Real-IP",
			IndexFromRight: 0,
		}).
		SetMethods([]string{"GET"}).
		SetHeader("X-Auth-Token", []string{"totally-top-secret", "another-secret"})

	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Real-IP", "2601:7:1c82:4097:59a0:a80b:2841:b8c8")
	expectedIP := "2601:7:1c82:4097::"
	request.Header.Set("X-Auth-Token", "totally-top-secret")

	sliceKeys := BuildKeys(lmt, request)
	if len(sliceKeys) == 0 {
		t.Fatal("Length of sliceKeys should never be empty.")
	}

	for _, keys := range sliceKeys {
		expectedKeys := [][]string{
			{expectedIP},
			{request.URL.Path},
			{"GET"},
			{"X-Auth-Token"},
			{"totally-top-secret", "another-secret"},
		}

		checkKeys(t, keys, expectedKeys)
	}
}

func TestRequestMethodAndBasicAuthUsersBuildKeys(t *testing.T) {
	lmt := NewLimiter(1, nil).
		SetIPLookup(limiter.IPLookup{
			Name:           "X-Real-IP",
			IndexFromRight: 0,
		}).
		SetMethods([]string{"GET"}).
		SetBasicAuthUsers([]string{"bro"})

	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Real-IP", "172.217.0.46")
	request.SetBasicAuth("bro", "tato")

	sliceKeys := BuildKeys(lmt, request)
	if len(sliceKeys) == 0 {
		t.Fatal("Length of sliceKeys should never be empty.")
	}

	for _, keys := range sliceKeys {
		expectedKeys := [][]string{
			{request.Header.Get("X-Real-IP")},
			{request.URL.Path},
			{"GET"},
			{"bro"},
		}

		checkKeys(t, keys, expectedKeys)
	}
}

func TestRequestMethodCustomHeadersAndBasicAuthUsersBuildKeys(t *testing.T) {
	lmt := NewLimiter(1, nil).
		SetIPLookup(limiter.IPLookup{
			Name:           "X-Real-IP",
			IndexFromRight: 0,
		}).
		SetMethods([]string{"GET"}).
		SetHeader("X-Auth-Token", []string{"totally-top-secret", "another-secret"}).
		SetBasicAuthUsers([]string{"bro"})

	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Real-IP", "2601:7:1c82:4097:59a0:a80b:2841:b8c8")
	expectedIP := "2601:7:1c82:4097::"
	request.Header.Set("X-Auth-Token", "totally-top-secret")
	request.SetBasicAuth("bro", "tato")

	sliceKeys := BuildKeys(lmt, request)
	if len(sliceKeys) == 0 {
		t.Fatal("Length of sliceKeys should never be empty.")
	}

	for _, keys := range sliceKeys {
		expectedKeys := [][]string{
			{expectedIP},
			{request.URL.Path},
			{"GET"},
			{"X-Auth-Token"},
			{"totally-top-secret", "another-secret"},
			{"bro"},
		}

		checkKeys(t, keys, expectedKeys)
	}
}

func TestRequestMethodCustomHeadersAndBasicAuthUsersAndContextValuesBuildKeys(t *testing.T) {
	lmt := NewLimiter(1, nil).
		SetIPLookup(limiter.IPLookup{
			Name:           "X-Real-IP",
			IndexFromRight: 0,
		}).
		SetMethods([]string{"GET"}).
		SetHeader("X-Auth-Token", []string{"totally-top-secret", "another-secret"}).
		SetContextValue("API-access-level", []string{"basic"}).
		SetBasicAuthUsers([]string{"bro"})

	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Real-IP", "172.217.0.46")
	request.Header.Set("X-Auth-Token", "totally-top-secret")
	request.SetBasicAuth("bro", "tato")
	//nolint:revive,staticcheck // limiter.SetContextValue requires string as a key, so we have to live with that
	request = request.WithContext(context.WithValue(request.Context(), "API-access-level", "basic"))

	sliceKeys := BuildKeys(lmt, request)
	if len(sliceKeys) == 0 {
		t.Fatal("Length of sliceKeys should never be empty.")
	}

	for _, keys := range sliceKeys {
		expectedKeys := [][]string{
			{request.Header.Get("X-Real-IP")},
			{request.URL.Path},
			{"GET"},
			{"X-Auth-Token"},
			{"totally-top-secret", "another-secret"},
			{"API-access-level"},
			{"basic"},
			{"bro"},
		}

		checkKeys(t, keys, expectedKeys)
	}
}

func TestLimitHandler(t *testing.T) {
	lmt := limiter.New(nil).SetMax(1).SetBurst(1).
		SetIPLookup(limiter.IPLookup{
			Name:           "X-Real-IP",
			IndexFromRight: 0,
		}).
		SetMethods([]string{"POST"})

	counter := 0
	lmt.SetOnLimitReached(func(http.ResponseWriter, *http.Request) { counter++ })

	handler := LimitHandler(lmt, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`hello world`))
	}))

	req, err := http.NewRequest("POST", "/doesntmatter", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("X-Real-IP", "2601:7:1c82:4097:59a0:a80b:2841:b8c8")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	// Should not be limited
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	// check RateLimit headers
	if value := rr.Result().Header[http.CanonicalHeaderKey("RateLimit-Limit")]; len(value) < 1 || value[0] != "1" {
		t.Errorf("handler returned wrong value: got %s want %s", value, "1")
	}
	if value := rr.Result().Header[http.CanonicalHeaderKey("RateLimit-Reset")]; len(value) < 1 || value[0] != "1" {
		t.Errorf("handler returned wrong value: got %s want %s", value, "1")
	}
	if value := rr.Result().Header[http.CanonicalHeaderKey("RateLimit-Remaining")]; len(value) < 1 || value[0] != "0" {
		t.Errorf("handler returned wrong value: got %s want %s", value, "0")
	}

	ch := make(chan int)
	go func() {
		// Different address, same /64 prefix
		req.Header.Set("X-Real-IP", "2601:7:1c82:4097:59a0:a80b:2841:b8c9")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		// Should be limited
		if status := rr.Code; status != http.StatusTooManyRequests {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusTooManyRequests)
		}
		// check X-Rate-Limit headers
		if value := rr.Result().Header[http.CanonicalHeaderKey("X-Rate-Limit-Limit")]; len(value) < 1 || value[0] != "1.00" {
			t.Errorf("X-Rate-Limit-Limit has wrong value: got %s want %v", value, "1.00")
		}
		if value := rr.Result().Header[http.CanonicalHeaderKey("X-Rate-Limit-Duration")]; len(value) < 1 || value[0] != "1" {
			t.Errorf("X-Rate-Limit-Duration has wrong value: got %s want %v", value, "1")
		}
		// check RateLimit headers
		if value := rr.Result().Header[http.CanonicalHeaderKey("RateLimit-Limit")]; len(value) < 1 || value[0] != "1" {
			t.Errorf("RateLimit-Limit has wrong value: got %s want %v", value, "1")
		}
		if value := rr.Result().Header[http.CanonicalHeaderKey("RateLimit-Reset")]; len(value) < 1 || value[0] != "1" {
			t.Errorf("RateLimit-Reset has wrong value: got %s want %v", value, "1")
		}
		if value := rr.Result().Header[http.CanonicalHeaderKey("RateLimit-Remaining")]; len(value) < 1 || value[0] != "0" {
			t.Errorf("RateLimit-Remaining has wrong value: got %s want %v", value, "0")
		}
		// OnLimitReached should be called
		if counter != 1 {
			t.Errorf("onLimitReached was not called")
		}
		close(ch)
	}()
	<-ch // Block until go func is done.
}

func TestOverrideForResponseWriter(t *testing.T) {
	lmt := limiter.New(nil).SetMax(1).SetBurst(1).
		SetIPLookup(limiter.IPLookup{
			Name:           "X-Real-IP",
			IndexFromRight: 0,
		}).
		SetMethods([]string{"POST"}).
		SetOverrideDefaultResponseWriter(true)

	counter := 0
	lmt.SetOnLimitReached(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("rejecting the large amount of requests"))
		counter++
	})

	handler := LimitHandler(lmt, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`hello world`))
	}))

	req, err := http.NewRequest("POST", "/doesntmatter", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("X-Real-IP", "172.217.0.46")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	// Should not be limited
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	ch := make(chan int)
	go func() {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		// Should be limited
		if status := rr.Code; status != http.StatusNotAcceptable {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotAcceptable)
		}
		// OnLimitReached should be called
		if counter != 1 {
			t.Errorf("onLimitReached was not called")
		}
		close(ch)
	}()
	<-ch // Block until go func is done.
}

func checkKeys(t *testing.T, keys []string, expectedKeys [][]string) {
	for i, keyChunk := range keys {
		switch {
		case i == 0 && !isInSlice(keyChunk, expectedKeys[0]):
			t.Errorf("The (%v) chunk should be remote IP. KeyChunk: %v", i+1, keyChunk)
		case i == 1 && !isInSlice(keyChunk, expectedKeys[1]):
			t.Errorf("The (%v) chunk should be request path. KeyChunk: %v", i+1, keyChunk)
		}
	}

	for _, ekeys := range expectedKeys {
		found := false
		for _, ekey := range ekeys {
			for _, key := range keys {
				if ekey == key {
					found = true
					break
				}
			}
		}

		if !found {
			t.Fatalf("expectedKeys missing: %v", strings.Join(ekeys, " "))
		}
	}
}

func isInSlice(key string, keys []string) bool {
	for _, sliceKey := range keys {
		if key == sliceKey {
			return true
		}
	}
	return false
}

type LockMap struct {
	m map[string]int64
	sync.Mutex
}

func (lm *LockMap) Set(key string, value int64) {
	lm.Lock()
	lm.m[key] = value
	lm.Unlock()
}

func (lm *LockMap) Get(key string) (int64, bool) {
	lm.Lock()
	value, ok := lm.m[key]
	lm.Unlock()
	return value, ok
}

func (lm *LockMap) Add(key string, incr int64) {
	lm.Lock()
	if val, ok := lm.m[key]; ok {
		lm.m[key] = val + incr
	} else {
		lm.m[key] = incr
	}
	lm.Unlock()
}

func TestLimitHandlerEmptyHeader(t *testing.T) {
	lmt := limiter.New(nil).SetMax(1).SetBurst(1)
	lmt.SetIPLookup(limiter.IPLookup{
		Name:           "X-Real-IP",
		IndexFromRight: 0,
	})
	lmt.SetMethods([]string{"POST"})
	lmt.SetHeader("user_id", []string{})

	counterMap := &LockMap{m: map[string]int64{}}
	lmt.SetOnLimitReached(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w, r
		counterMap.Add(r.Header.Get("user_id"), 1)
	})

	handler := LimitHandler(lmt, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r
		w.Write([]byte(`hello world`))
	}))

	req, err := http.NewRequest("POST", "/doesntmatter", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("X-Real-IP", "2601:7:1c82:4097:59a0:a80b:2841:b8c8")
	req.Header.Set("user_id", "0")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	{ // Should not be limited
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
		// check RateLimit headers
		if value := rr.Result().Header[http.CanonicalHeaderKey("RateLimit-Limit")]; len(value) < 1 || value[0] != "1" {
			t.Errorf("handler returned wrong value: got %s want %s", value, "1")
		}
		if value := rr.Result().Header[http.CanonicalHeaderKey("RateLimit-Reset")]; len(value) < 1 || value[0] != "1" {
			t.Errorf("handler returned wrong value: got %s want %s", value, "1")
		}
		if value := rr.Result().Header[http.CanonicalHeaderKey("RateLimit-Remaining")]; len(value) < 1 || value[0] != "0" {
			t.Errorf("handler returned wrong value: got %s want %s", value, "0")
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	// same user_id, should be limited
	go func() {
		defer wg.Done()

		req1, _ := http.NewRequest("POST", "/doesntmatter", nil)
		req1.Header.Set("X-Real-IP", "2601:7:1c82:4097:59a0:a80b:2841:b8c8")
		req1.Header.Set("user_id", "0")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req1)
		// Should be limited
		{
			if status := rr.Code; status != http.StatusTooManyRequests {
				t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusTooManyRequests)
			}
			// check X-Rate-Limit headers
			if value := rr.Result().Header[http.CanonicalHeaderKey("X-Rate-Limit-Limit")]; len(value) < 1 || value[0] != "1.00" {
				t.Errorf("X-Rate-Limit-Limit has wrong value: got %s want %v", value, "1.00")
			}
			if value := rr.Result().Header[http.CanonicalHeaderKey("X-Rate-Limit-Duration")]; len(value) < 1 || value[0] != "1" {
				t.Errorf("X-Rate-Limit-Duration has wrong value: got %s want %v", value, "1")
			}
			// check RateLimit headers
			if value := rr.Result().Header[http.CanonicalHeaderKey("RateLimit-Limit")]; len(value) < 1 || value[0] != "1" {
				t.Errorf("RateLimit-Limit has wrong value: got %s want %v", value, "1")
			}
			if value := rr.Result().Header[http.CanonicalHeaderKey("RateLimit-Reset")]; len(value) < 1 || value[0] != "1" {
				t.Errorf("RateLimit-Reset has wrong value: got %s want %v", value, "1")
			}
			if value := rr.Result().Header[http.CanonicalHeaderKey("RateLimit-Remaining")]; len(value) < 1 || value[0] != "0" {
				t.Errorf("RateLimit-Remaining has wrong value: got %s want %v", value, "0")
			}
			// OnLimitReached should be called
			if aint, ok := counterMap.Get(req1.Header.Get("user_id")); ok {
				if aint == 0 {
					t.Errorf("onLimitReached was not called")
				}
			}
		}
	}()

	wg.Wait() // Block until go func is done.
}

func TestHTTPMiddleware(t *testing.T) {
	t.Run("basic request", func(t *testing.T) {
		lmt := NewLimiter(1, nil)
		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		wrapped := HTTPMiddleware(lmt)(handler)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/test", nil)
		r.RemoteAddr = "127.0.0.1:12345"
		wrapped.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("rate limit exceeded", func(t *testing.T) {
		lmt := NewLimiter(0.1, nil) // only allow one request per 10 seconds
		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		wrapped := HTTPMiddleware(lmt)(handler)

		// first request
		w1 := httptest.NewRecorder()
		r1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		r1.RemoteAddr = "127.0.0.1:12345"
		wrapped.ServeHTTP(w1, r1)
		if w1.Code != http.StatusOK {
			t.Errorf("first request: expected status %d, got %d", http.StatusOK, w1.Code)
		}

		// immediate second request should fail
		w2 := httptest.NewRecorder()
		r2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		r2.RemoteAddr = "127.0.0.1:12345"
		wrapped.ServeHTTP(w2, r2)
		if w2.Code != http.StatusTooManyRequests {
			t.Errorf("second request: expected status %d, got %d", http.StatusTooManyRequests, w2.Code)
		}
		if !strings.Contains(w2.Body.String(), "maximum request limit") {
			t.Errorf("expected error message containing 'maximum request limit', got %q", w2.Body.String())
		}
	})

	t.Run("context cancelled", func(t *testing.T) {
		lmt := NewLimiter(1, nil)
		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		wrapped := HTTPMiddleware(lmt)(handler)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/test", nil)
		ctx, cancel := context.WithCancel(r.Context())
		cancel()
		r = r.WithContext(ctx)
		wrapped.ServeHTTP(w, r)
		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
		}
		if !strings.Contains(w.Body.String(), "Context was canceled") {
			t.Errorf("expected error message containing 'Context was canceled', got %q", w.Body.String())
		}
	})

	t.Run("custom error handler", func(t *testing.T) {
		lmt := NewLimiter(0.1, nil) // only allow one request per 10 seconds
		customMsg := "custom limit reached"
		lmt.SetMessage(customMsg)

		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		wrapped := HTTPMiddleware(lmt)(handler)

		// first request
		w1 := httptest.NewRecorder()
		r1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		r1.RemoteAddr = "127.0.0.1:12345"
		wrapped.ServeHTTP(w1, r1)
		if w1.Code != http.StatusOK {
			t.Errorf("first request: expected status %d, got %d", http.StatusOK, w1.Code)
		}

		// immediate second request should fail
		w2 := httptest.NewRecorder()
		r2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		r2.RemoteAddr = "127.0.0.1:12345"
		wrapped.ServeHTTP(w2, r2)
		if w2.Code != http.StatusTooManyRequests {
			t.Errorf("second request: expected status %d, got %d", http.StatusTooManyRequests, w2.Code)
		}
		if !strings.Contains(w2.Body.String(), customMsg) {
			t.Errorf("expected error message containing %q, got %q", customMsg, w2.Body.String())
		}
	})

	t.Run("custom IP lookup", func(t *testing.T) {
		lmt := NewLimiter(0.1, nil)
		lmt.SetIPLookup(limiter.IPLookup{Name: "X-Real-IP"})
		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		wrapped := HTTPMiddleware(lmt)(handler)

		// first request with IP1
		w1 := httptest.NewRecorder()
		r1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		r1.Header.Set("X-Real-IP", "5.5.5.5")
		wrapped.ServeHTTP(w1, r1)
		if w1.Code != http.StatusOK {
			t.Errorf("first request: expected status %d, got %d", http.StatusOK, w1.Code)
		}

		// second request with IP1 should fail
		w2 := httptest.NewRecorder()
		r2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		r2.Header.Set("X-Real-IP", "5.5.5.5")
		wrapped.ServeHTTP(w2, r2)
		if w2.Code != http.StatusTooManyRequests {
			t.Errorf("second request: expected status %d, got %d", http.StatusTooManyRequests, w2.Code)
		}

		// request with IP2 should pass
		w3 := httptest.NewRecorder()
		r3 := httptest.NewRequest(http.MethodGet, "/test", nil)
		r3.Header.Set("X-Real-IP", "6.6.6.6")
		wrapped.ServeHTTP(w3, r3)
		if w3.Code != http.StatusOK {
			t.Errorf("third request: expected status %d, got %d", http.StatusOK, w3.Code)
		}
	})
}
