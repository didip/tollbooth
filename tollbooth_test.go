package tollbooth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/didip/tollbooth/v6/limiter"
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
	lmt := NewLimiter(1, nil)

	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Real-IP", "2601:7:1c82:4097:59a0:a80b:2841:b8c8")

	sliceKeys := BuildKeys(lmt, request)
	if len(sliceKeys) == 0 {
		t.Fatal("Length of sliceKeys should never be empty.")
	}

	for _, keys := range sliceKeys {
		expectedKeys := [][]string{
			{request.Header.Get("X-Real-IP")},
			{request.URL.Path},
		}

		checkKeys(t, keys, expectedKeys)
	}
}

func TestIgnoreURLBuildKeys(t *testing.T) {
	lmt := NewLimiter(1, nil)
	lmt.SetIgnoreURL(true)

	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Real-IP", "2601:7:1c82:4097:59a0:a80b:2841:b8c8")

	for _, keys := range BuildKeys(lmt, request) {
		for i, keyChunk := range keys {
			if i == 0 && keyChunk != request.Header.Get("X-Real-IP") {
				t.Errorf("The (%v) chunk should be remote IP. KeyChunk: %v", i+1, keyChunk)
			}
		}
	}
}

func TestBasicAuthBuildKeys(t *testing.T) {
	lmt := NewLimiter(1, nil)
	lmt.SetBasicAuthUsers([]string{"bro"})

	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Real-IP", "2601:7:1c82:4097:59a0:a80b:2841:b8c8")

	request.SetBasicAuth("bro", "tato")

	sliceKeys := BuildKeys(lmt, request)
	if len(sliceKeys) == 0 {
		t.Fatal("Length of sliceKeys should never be empty.")
	}

	for _, keys := range sliceKeys {
		expectedKeys := [][]string{
			{request.Header.Get("X-Real-IP")},
			{request.URL.Path},
			{"bro"},
		}

		checkKeys(t, keys, expectedKeys)
	}
}

func TestCustomHeadersBuildKeys(t *testing.T) {
	lmt := NewLimiter(1, nil)
	lmt.SetHeader("X-Auth-Token", []string{"totally-top-secret", "another-secret"})

	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Real-IP", "2601:7:1c82:4097:59a0:a80b:2841:b8c8")
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
	lmt := NewLimiter(1, nil)
	lmt.SetMethods([]string{"GET"})

	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Real-IP", "2601:7:1c82:4097:59a0:a80b:2841:b8c8")

	sliceKeys := BuildKeys(lmt, request)
	if len(sliceKeys) == 0 {
		t.Fatal("Length of sliceKeys should never be empty.")
	}

	for _, keys := range sliceKeys {
		expectedKeys := [][]string{
			{request.Header.Get("X-Real-IP")},
			{request.URL.Path},
			{"GET"},
		}

		checkKeys(t, keys, expectedKeys)
	}
}

func TestContextValueBuildKeys(t *testing.T) {
	lmt := NewLimiter(1, nil)
	lmt.SetContextValue("API-access-level", []string{"basic"})

	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Real-IP", "2601:7:1c82:4097:59a0:a80b:2841:b8c8")
	//nolint:golint,staticcheck // limiter.SetContextValue requires string as a key, so we have to live with that
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
	lmt := NewLimiter(1, nil)
	lmt.SetMethods([]string{"GET"})
	lmt.SetHeader("X-Auth-Token", []string{"totally-top-secret", "another-secret"})

	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Real-IP", "2601:7:1c82:4097:59a0:a80b:2841:b8c8")
	request.Header.Set("X-Auth-Token", "totally-top-secret")

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
		}

		checkKeys(t, keys, expectedKeys)
	}
}

func TestRequestMethodAndBasicAuthUsersBuildKeys(t *testing.T) {
	lmt := NewLimiter(1, nil)
	lmt.SetMethods([]string{"GET"})
	lmt.SetBasicAuthUsers([]string{"bro"})

	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Real-IP", "2601:7:1c82:4097:59a0:a80b:2841:b8c8")
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
	lmt := NewLimiter(1, nil)
	lmt.SetMethods([]string{"GET"})
	lmt.SetHeader("X-Auth-Token", []string{"totally-top-secret", "another-secret"})
	lmt.SetBasicAuthUsers([]string{"bro"})

	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Real-IP", "2601:7:1c82:4097:59a0:a80b:2841:b8c8")
	request.Header.Set("X-Auth-Token", "totally-top-secret")
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
			{"X-Auth-Token"},
			{"totally-top-secret", "another-secret"},
			{"bro"},
		}

		checkKeys(t, keys, expectedKeys)
	}
}

func TestRequestMethodCustomHeadersAndBasicAuthUsersAndContextValuesBuildKeys(t *testing.T) {
	lmt := NewLimiter(1, nil)
	lmt.SetMethods([]string{"GET"})
	lmt.SetHeader("X-Auth-Token", []string{"totally-top-secret", "another-secret"})
	lmt.SetContextValue("API-access-level", []string{"basic"})
	lmt.SetBasicAuthUsers([]string{"bro"})

	request, err := http.NewRequest("GET", "/", strings.NewReader("Hello, world!"))
	if err != nil {
		t.Errorf("Unable to create new HTTP request. Error: %v", err)
	}

	request.Header.Set("X-Real-IP", "2601:7:1c82:4097:59a0:a80b:2841:b8c8")
	request.Header.Set("X-Auth-Token", "totally-top-secret")
	request.SetBasicAuth("bro", "tato")
	//nolint:golint,staticcheck // limiter.SetContextValue requires string as a key, so we have to live with that
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
	lmt := limiter.New(nil).SetMax(1).SetBurst(1)
	lmt.SetIPLookups([]string{"X-Real-IP", "RemoteAddr", "X-Forwarded-For"})
	lmt.SetMethods([]string{"POST"})

	counter := 0
	lmt.SetOnLimitReached(func(w http.ResponseWriter, r *http.Request) { counter++ })

	handler := LimitHandler(lmt, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	ch := make(chan int)
	go func() {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		// Should be limited
		if status := rr.Code; status != http.StatusTooManyRequests {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusTooManyRequests)
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
	lmt := limiter.New(nil).SetMax(1).SetBurst(1)
	lmt.SetIPLookups([]string{"X-Real-IP", "RemoteAddr", "X-Forwarded-For"})
	lmt.SetMethods([]string{"POST"})
	lmt.SetOverrideDefaultResponseWriter(true)

	counter := 0
	lmt.SetOnLimitReached(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("rejecting the large amount of requests"))
		counter++
	})

	handler := LimitHandler(lmt, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
