package tollbooth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/didip/tollbooth/limiter"
)

func TestLimitByKeys(t *testing.T) {
	lmt := NewLimiter(1, nil) // Only 1 request per second is allowed.

	exceed, _ := LimitByKeys(lmt, []string{"127.0.0.1", "/"})
	if exceed {
		t.Errorf("First time count should return false")
	}

	exceed, _ = LimitByKeys(lmt, []string{"127.0.0.1", "/"})
	if !exceed {
		t.Errorf("Second time count should return true because it exceeds 1 request per second.")
	}

	<-time.After(1 * time.Second)
	exceed, _ = LimitByKeys(lmt, []string{"127.0.0.1", "/"})
	if exceed {
		t.Errorf("Third time count should return false because the 1 second window has passed.")
	}
}

func TestDefaultBuildKeys(t *testing.T) {
	lmt := NewLimiter(1, nil)
	lmt.SetIPLookups([]string{"X-Forwarded-For", "X-Real-IP", "RemoteAddr"})

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
		for i, keyChunk := range keys {
			if i == 0 && keyChunk != request.Header.Get("X-Real-IP") {
				t.Errorf("The first chunk should be remote IP. KeyChunk: %v", keyChunk)
			}
			if i == 1 && keyChunk != request.URL.Path {
				t.Errorf("The second chunk should be request path. KeyChunk: %v", keyChunk)
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

	for _, keys := range BuildKeys(lmt, request) {
		if len(keys) != 3 {
			t.Error("Keys should be made of 3 parts.")
		}
		for i, keyChunk := range keys {
			if i == 0 && keyChunk != request.Header.Get("X-Real-IP") {
				t.Errorf("The (%v) chunk should be remote IP. KeyChunk: %v", i+1, keyChunk)
			}
			if i == 1 && keyChunk != request.URL.Path {
				t.Errorf("The (%v) chunk should be request path. KeyChunk: %v", i+1, keyChunk)
			}
			if i == 2 && keyChunk != "bro" {
				t.Errorf("The (%v) chunk should be request username. KeyChunk: %v", i+1, keyChunk)
			}
		}
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

	for _, keys := range BuildKeys(lmt, request) {
		if len(keys) != 4 {
			t.Errorf("Keys should be made of 4 parts. Keys: %v", keys)
		}
		for i, keyChunk := range keys {
			if i == 0 && keyChunk != request.Header.Get("X-Real-IP") {
				t.Errorf("The (%v) chunk should be remote IP. KeyChunk: %v", i+1, keyChunk)
			}
			if i == 1 && keyChunk != request.URL.Path {
				t.Errorf("The (%v) chunk should be request path. KeyChunk: %v", i+1, keyChunk)
			}
			if i == 2 && keyChunk != "X-Auth-Token" {
				t.Errorf("The (%v) chunk should be request header. KeyChunk: %v", i+1, keyChunk)
			}
			if i == 3 && (keyChunk != "totally-top-secret" && keyChunk != "another-secret") {
				t.Errorf("The (%v) chunk should be request path. KeyChunk: %v", i+1, keyChunk)
			}
		}
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

	for _, keys := range BuildKeys(lmt, request) {
		if len(keys) != 3 {
			t.Errorf("Keys should be made of 3 parts. Keys: %v", keys)
		}
		for i, keyChunk := range keys {
			if i == 0 && keyChunk != request.Header.Get("X-Real-IP") {
				t.Errorf("The (%v) chunk should be remote IP. KeyChunk: %v", i+1, keyChunk)
			}
			if i == 1 && keyChunk != request.URL.Path {
				t.Errorf("The (%v) chunk should be request path. KeyChunk: %v", i+1, keyChunk)
			}
			if i == 2 && keyChunk != "GET" {
				t.Errorf("The (%v) chunk should be request method. KeyChunk: %v", i+1, keyChunk)
			}
		}
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

	for _, keys := range BuildKeys(lmt, request) {
		if len(keys) != 5 {
			t.Errorf("Keys should be made of 4 parts. Keys: %v", keys)
		}
		for i, keyChunk := range keys {
			if i == 0 && keyChunk != request.Header.Get("X-Real-IP") {
				t.Errorf("The (%v) chunk should be remote IP. KeyChunk: %v", i+1, keyChunk)
			}
			if i == 1 && keyChunk != request.URL.Path {
				t.Errorf("The (%v) chunk should be request path. KeyChunk: %v", i+1, keyChunk)
			}
			if i == 2 && keyChunk != "GET" {
				t.Errorf("The (%v) chunk should be request method. KeyChunk: %v", i+1, keyChunk)
			}
			if i == 3 && keyChunk != "X-Auth-Token" {
				t.Errorf("The (%v) chunk should be request header. KeyChunk: %v", i+1, keyChunk)
			}
			if i == 4 && (keyChunk != "totally-top-secret" && keyChunk != "another-secret") {
				t.Errorf("The (%v) chunk should be request path. KeyChunk: %v", i+1, keyChunk)
			}
		}
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

	for _, keys := range BuildKeys(lmt, request) {
		if len(keys) != 4 {
			t.Errorf("Keys should be made of 4 parts. Keys: %v", keys)
		}
		for i, keyChunk := range keys {
			if i == 0 && keyChunk != request.Header.Get("X-Real-IP") {
				t.Errorf("The (%v) chunk should be remote IP. KeyChunk: %v", i+1, keyChunk)
			}
			if i == 1 && keyChunk != request.URL.Path {
				t.Errorf("The (%v) chunk should be request path. KeyChunk: %v", i+1, keyChunk)
			}
			if i == 2 && keyChunk != "GET" {
				t.Errorf("The (%v) chunk should be request method. KeyChunk: %v", i+1, keyChunk)
			}
			if i == 3 && keyChunk != "bro" {
				t.Errorf("The (%v) chunk should be basic auth user. KeyChunk: %v", i+1, keyChunk)
			}
		}
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

	for _, keys := range BuildKeys(lmt, request) {
		if len(keys) != 6 {
			t.Errorf("Keys should be made of 4 parts. Keys: %v", keys)
		}
		for i, keyChunk := range keys {
			if i == 0 && keyChunk != request.Header.Get("X-Real-IP") {
				t.Errorf("The (%v) chunk should be remote IP. KeyChunk: %v", i+1, keyChunk)
			}
			if i == 1 && keyChunk != request.URL.Path {
				t.Errorf("The (%v) chunk should be request path. KeyChunk: %v", i+1, keyChunk)
			}
			if i == 2 && keyChunk != "GET" {
				t.Errorf("The (%v) chunk should be request method. KeyChunk: %v", i+1, keyChunk)
			}
			if i == 3 && keyChunk != "X-Auth-Token" {
				t.Errorf("The (%v) chunk should be request header. KeyChunk: %v", i+1, keyChunk)
			}
			if i == 4 && (keyChunk != "totally-top-secret" && keyChunk != "another-secret") {
				t.Errorf("The (%v) chunk should be request path. KeyChunk: %v", i+1, keyChunk)
			}
			if i == 5 && keyChunk != "bro" {
				t.Errorf("The (%v) chunk should be basic auth user. KeyChunk: %v", i+1, keyChunk)
			}
		}
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
	//Should not be limited
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
