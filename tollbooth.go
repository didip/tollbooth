// Package tollbooth provides core rate-limit logic.
package tollbooth

import (
	"fmt"
	"github.com/didip/tollbooth/storages"
	"net/http"
	"time"
)

func NewRequestLimit(max int64, ttl time.Duration) *RequestLimit {
	return &RequestLimit{Max: max, TTL: ttl}
}

type RequestLimit struct {
	Max int64
	TTL time.Duration
}

type HTTPError struct {
	Message    string
	StatusCode int
}

func (httperror *HTTPError) Error() string {
	return fmt.Sprintf("%v: %v", httperror.StatusCode, httperror.Message)
}

func LimitByRemoteIP(storage storages.ICounterStorage, reqLimit *RequestLimit, r *http.Request) *HTTPError {
	remoteIP := r.Header.Get("REMOTE_ADDR")
	path := r.URL.Path
	key := path + "|" + remoteIP

	storage.IncrBy(key, int64(1), reqLimit.TTL)
	currentCount, _ := storage.Get(key)

	// Check if the returned counter exceeds our limit
	if currentCount > reqLimit.Max {
		return &HTTPError{Message: "You have reached maximum request limit.", StatusCode: 429}
	}
	return nil
}

// RemoteIPLimiterHandler is a middleware that limits by RemoteIP.
func RemoteIPLimiterHandler(storage storages.ICounterStorage, reqLimit *RequestLimit, next http.Handler) http.Handler {
	middle := func(w http.ResponseWriter, r *http.Request) {
		httpError := LimitByRemoteIP(storage, reqLimit, r)

		if httpError != nil {
			http.Error(w, httpError.Message, httpError.StatusCode)
			return
		} else {
			next.ServeHTTP(w, r)
		}
	}
	return http.HandlerFunc(middle)
}

func RemoteIPLimiterFuncHandler(storage storages.ICounterStorage, reqLimit *RequestLimit, nextFunc func(http.ResponseWriter, *http.Request)) http.Handler {
	next := http.HandlerFunc(nextFunc)
	return RemoteIPLimiterHandler(storage, reqLimit, next)
}
