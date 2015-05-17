// Package tollbooth provides rate limiting logic for HTTP request handler.
package tollbooth

import (
	"fmt"
	"github.com/didip/tollbooth/libstring"
	"github.com/didip/tollbooth/storages"
	"net/http"
	"strings"
	"time"
)

// NewRequestLimit is a constructor for RequestLimit.
func NewRequestLimit(max int64, ttl time.Duration) *RequestLimit {
	return &RequestLimit{Max: max, TTL: ttl}
}

// RequestLimit is a config struct to limit a particular request handler.
type RequestLimit struct {
	// Maximum number of requests to limit per duration.
	Max int64

	// Duration of rate limiter.
	TTL time.Duration

	// List of HTTP Methods to limit (GET, POST, PUT, etc.).
	// Empty means limit all methods.
	Methods []string
}

// HTTPError is an error struct that returns both message and status code.
type HTTPError struct {
	Message    string
	StatusCode int
}

// Error returns error message.
func (httperror *HTTPError) Error() string {
	return fmt.Sprintf("%v: %v", httperror.StatusCode, httperror.Message)
}

// LimitByKeyParts keeps track number of request made by keyParts separated by pipe.
// It returns HTTPError when limit is exceeded.
func LimitByKeyParts(storage storages.ICounterStorage, reqLimit *RequestLimit, keyParts []string) *HTTPError {
	key := strings.Join(keyParts, "|")

	storage.IncrBy(key, int64(1), reqLimit.TTL)
	currentCount, _ := storage.Get(key)

	// Check if the returned counter exceeds our limit
	if currentCount > reqLimit.Max {
		return &HTTPError{Message: "You have reached maximum request limit.", StatusCode: 429}
	}
	return nil
}

// LimitByIP keeps track number of request made by REMOTE_ADDR and returns HTTPError when limit is exceeded.
func LimitByIP(storage storages.ICounterStorage, reqLimit *RequestLimit, r *http.Request) *HTTPError {
	remoteIP := r.Header.Get("REMOTE_ADDR")
	path := r.URL.Path
	return LimitByKeyParts(storage, reqLimit, []string{path, remoteIP})
}

// LimitByIPHandler is a middleware that limits by IP given http.Handler struct.
func LimitByIPHandler(storage storages.ICounterStorage, reqLimit *RequestLimit, next http.Handler) http.Handler {
	middle := func(w http.ResponseWriter, r *http.Request) {
		// 1. If reqLimit.Methods is not defined or empty, checks all HTTP methods.
		// 2. If request method is included in reqLimit.Methods, check it.
		if reqLimit.Methods == nil || len(reqLimit.Methods) == 0 || libstring.StringInSlice(reqLimit.Methods, r.Method) {
			httpError := LimitByIP(storage, reqLimit, r)

			if httpError != nil {
				http.Error(w, httpError.Message, httpError.StatusCode)
				return
			} else {
				next.ServeHTTP(w, r)
			}
		} else {
			next.ServeHTTP(w, r)
		}

	}
	return http.HandlerFunc(middle)
}

// LimitByIPFuncHandler is a middleware that limits by IP given request handler function.
func LimitByIPFuncHandler(storage storages.ICounterStorage, reqLimit *RequestLimit, nextFunc func(http.ResponseWriter, *http.Request)) http.Handler {
	next := http.HandlerFunc(nextFunc)
	return LimitByIPHandler(storage, reqLimit, next)
}
