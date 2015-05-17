// Package tollbooth provides rate limiting logic for HTTP request handler.
package tollbooth

import (
	"net/http"
	"strings"
	"time"

	"github.com/didip/tollbooth/config"
	"github.com/didip/tollbooth/errors"
	"github.com/didip/tollbooth/libstring"
	"github.com/didip/tollbooth/storages"
)

// NewLimiter is a convenience function to config.NewLimiter.
func NewLimiter(max int64, ttl time.Duration) *config.Limiter {
	return config.NewLimiter(max, ttl)
}

// LimitByKeyParts keeps track number of request made by keyParts separated by pipe.
// It keeps track number of request made by REMOTE_ADDR and returns HTTPError when limit is exceeded.
func LimitByKeyParts(storage storages.ICounterStorage, limiter *config.Limiter, keyParts []string) *errors.HTTPError {
	key := strings.Join(keyParts, "|")

	storage.IncrBy(key, int64(1), limiter.TTL)
	currentCount, _ := storage.Get(key)

	// Check if the returned counter exceeds our limit
	if currentCount > limiter.Max {
		return &errors.HTTPError{Message: limiter.Message, StatusCode: limiter.StatusCode}
	}
	return nil
}

// LimitHandler is a middleware that limits by IP given http.Handler struct.
func LimitHandler(storage storages.ICounterStorage, limiter *config.Limiter, next http.Handler) http.Handler {
	middle := func(w http.ResponseWriter, r *http.Request) {
		remoteIP := r.Header.Get("REMOTE_ADDR")
		path := r.URL.Path
		defaultKeyParts := []string{remoteIP, path}

		var httpError *errors.HTTPError

		if limiter.Methods != nil && limiter.Headers != nil {
			// Limit by HTTP methods and HTTP headers.
			for _, method := range limiter.Methods {
				keyParts := append(defaultKeyParts, method)

				for _, headerKeyParts := range libstring.FlattenMapSliceString(limiter.Headers, "headers", ":") {
					keyParts = append(keyParts, headerKeyParts)
					httpError = LimitByKeyParts(storage, limiter, keyParts)
					if httpError != nil {
						http.Error(w, httpError.Message, httpError.StatusCode)
						return
					}
				}
			}

		} else if limiter.Methods != nil {
			// Limit by HTTP methods only.
			for _, method := range limiter.Methods {
				keyParts := append(defaultKeyParts, method)
				httpError = LimitByKeyParts(storage, limiter, keyParts)
				if httpError != nil {
					http.Error(w, httpError.Message, httpError.StatusCode)
					return
				}
			}
		} else if limiter.Headers != nil {
			// Limit by HTTP headers only.
			for _, headerKeyParts := range libstring.FlattenMapSliceString(limiter.Headers, "headers", ":") {
				keyParts := append(defaultKeyParts, headerKeyParts)
				httpError = LimitByKeyParts(storage, limiter, keyParts)
				if httpError != nil {
					http.Error(w, httpError.Message, httpError.StatusCode)
					return
				}
			}

		} else {
			// Default: Limit by remote IP and request path.
			httpError = LimitByKeyParts(storage, limiter, defaultKeyParts)
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

// LimitFuncHandler is a middleware that limits by IP given request handler function.
func LimitFuncHandler(storage storages.ICounterStorage, limiter *config.Limiter, nextFunc func(http.ResponseWriter, *http.Request)) http.Handler {
	return LimitHandler(storage, limiter, http.HandlerFunc(nextFunc))
}
