// Package tollbooth provides rate limiting logic for HTTP request handler.
package tollbooth

import (
	"fmt"
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

// LimitByIPHandler is a middleware that limits by IP given http.Handler struct.
// It keeps track number of request made by REMOTE_ADDR and returns HTTPError when limit is exceeded.
func LimitByIPHandler(storage storages.ICounterStorage, reqLimit *RequestLimit, next http.Handler) http.Handler {
	middle := func(w http.ResponseWriter, r *http.Request) {
		remoteIP := r.Header.Get("REMOTE_ADDR")
		path := r.URL.Path
		defaultKeyParts := []string{remoteIP, path}

		var httpError *HTTPError

		if reqLimit.Methods != nil {
			// Limit by HTTP methods.
			for _, method := range reqLimit.Methods {
				keyParts := append(defaultKeyParts, method)
				httpError = LimitByKeyParts(storage, reqLimit, keyParts)
				if httpError != nil {
					http.Error(w, httpError.Message, httpError.StatusCode)
					return
				}
			}
		} else {
			// Default limiter.
			httpError = LimitByKeyParts(storage, reqLimit, defaultKeyParts)
			if httpError != nil {
				http.Error(w, httpError.Message, httpError.StatusCode)
				return
			}
		}

		// There's no rate-limit error, serve the next handler.
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(middle)
}

// LimitByIPFuncHandler is a middleware that limits by IP given request handler function.
func LimitByIPFuncHandler(storage storages.ICounterStorage, reqLimit *RequestLimit, nextFunc func(http.ResponseWriter, *http.Request)) http.Handler {
	return LimitByIPHandler(storage, reqLimit, http.HandlerFunc(nextFunc))
}
