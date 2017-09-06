// Package tollbooth provides rate-limiting logic to HTTP request handler.
package tollbooth

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/didip/tollbooth/errors"
	"github.com/didip/tollbooth/libstring"
	"github.com/didip/tollbooth/limiter"
)

// NewLimiter is a convenience function to limiter.New.
func NewLimiter(max int64, ttl time.Duration, tbOptions *limiter.TokenBucketOptions) *limiter.Limiter {
	return limiter.New(max, ttl, tbOptions)
}

// LimitByKeys keeps track number of request made by keys separated by pipe.
// It returns HTTPError when limit is exceeded.
func LimitByKeys(limiter *limiter.Limiter, keys []string) *errors.HTTPError {
	if limiter.LimitReached(strings.Join(keys, "|")) {
		return &errors.HTTPError{Message: limiter.GetMessage(), StatusCode: limiter.GetStatusCode()}
	}

	return nil
}

// LimitByKeysWithCustomTokenBucketTTL keeps track number of request made by keys separated by pipe.
// It returns HTTPError when limit is exceeded.
// User can define a TTL for the key to expire
func LimitByKeysWithCustomTokenBucketTTL(limiter *limiter.Limiter, keys []string, bucketExpireTTL time.Duration) *errors.HTTPError {
	if limiter.LimitReachedWithCustomTokenBucketTTL(strings.Join(keys, "|"), bucketExpireTTL) {
		return &errors.HTTPError{Message: limiter.GetMessage(), StatusCode: limiter.GetStatusCode()}
	}

	return nil
}

// LimitByRequest builds keys based on http.Request struct,
// loops through all the keys, and check if any one of them returns HTTPError.
func LimitByRequest(limiter *limiter.Limiter, r *http.Request) *errors.HTTPError {
	sliceKeys := BuildKeys(limiter, r)

	// Loop sliceKeys and check if one of them has error.
	for _, keys := range sliceKeys {
		httpError := LimitByKeys(limiter, keys)
		if httpError != nil {
			return httpError
		}
	}

	return nil
}

// BuildKeys generates a slice of keys to rate-limit by given limiter and request structs.
func BuildKeys(limiter *limiter.Limiter, r *http.Request) [][]string {
	remoteIP := libstring.RemoteIP(limiter.GetIPLookups(), r)
	path := r.URL.Path
	sliceKeys := make([][]string, 0)

	// Don't BuildKeys if remoteIP is blank.
	if remoteIP == "" {
		return sliceKeys
	}

	limiterMethods := limiter.GetMethods()
	limiterHeaders := limiter.GetHeaders()
	limiterBasicAuthUsers := limiter.GetBasicAuthUsers()

	limiterHeadersIsSet := limiterHeaders != nil && len(limiterHeaders) > 0
	limiterBasicAuthUsersIsSet := limiterBasicAuthUsers != nil && len(limiterBasicAuthUsers) > 0

	if limiterMethods != nil && limiterHeadersIsSet && limiterBasicAuthUsersIsSet {
		// Limit by HTTP methods and HTTP headers+values and Basic Auth credentials.
		if libstring.StringInSlice(limiterMethods, r.Method) {
			for headerKey, headerValues := range limiterHeaders {
				if (headerValues == nil || len(headerValues) <= 0) && r.Header.Get(headerKey) != "" {
					// If header values are empty, rate-limit all request with headerKey.
					username, _, ok := r.BasicAuth()
					if ok && libstring.StringInSlice(limiterBasicAuthUsers, username) {
						sliceKeys = append(sliceKeys, []string{remoteIP, path, r.Method, headerKey, username})
					}

				} else if len(headerValues) > 0 && r.Header.Get(headerKey) != "" {
					// If header values are not empty, rate-limit all request with headerKey and headerValues.
					for _, headerValue := range headerValues {
						username, _, ok := r.BasicAuth()
						if ok && libstring.StringInSlice(limiterBasicAuthUsers, username) {
							sliceKeys = append(sliceKeys, []string{remoteIP, path, r.Method, headerKey, headerValue, username})
						}
					}
				}
			}
		}

	} else if limiterMethods != nil && limiterHeadersIsSet {
		// Limit by HTTP methods and HTTP headers+values.
		if libstring.StringInSlice(limiterMethods, r.Method) {
			for headerKey, headerValues := range limiterHeaders {
				if (headerValues == nil || len(headerValues) <= 0) && r.Header.Get(headerKey) != "" {
					// If header values are empty, rate-limit all request with headerKey.
					sliceKeys = append(sliceKeys, []string{remoteIP, path, r.Method, headerKey})

				} else if len(headerValues) > 0 && r.Header.Get(headerKey) != "" {
					// If header values are not empty, rate-limit all request with headerKey and headerValues.
					for _, headerValue := range headerValues {
						sliceKeys = append(sliceKeys, []string{remoteIP, path, r.Method, headerKey, headerValue})
					}
				}
			}
		}

	} else if limiterMethods != nil && limiterBasicAuthUsersIsSet {
		// Limit by HTTP methods and Basic Auth credentials.
		if libstring.StringInSlice(limiterMethods, r.Method) {
			username, _, ok := r.BasicAuth()
			if ok && libstring.StringInSlice(limiterBasicAuthUsers, username) {
				sliceKeys = append(sliceKeys, []string{remoteIP, path, r.Method, username})
			}
		}

	} else if limiterMethods != nil {
		// Limit by HTTP methods.
		if libstring.StringInSlice(limiterMethods, r.Method) {
			sliceKeys = append(sliceKeys, []string{remoteIP, path, r.Method})
		}

	} else if limiterHeadersIsSet {
		// Limit by HTTP headers+values.
		for headerKey, headerValues := range limiterHeaders {
			if (headerValues == nil || len(headerValues) <= 0) && r.Header.Get(headerKey) != "" {
				// If header values are empty, rate-limit all request with headerKey.
				sliceKeys = append(sliceKeys, []string{remoteIP, path, headerKey})

			} else if len(headerValues) > 0 && r.Header.Get(headerKey) != "" {
				// If header values are not empty, rate-limit all request with headerKey and headerValues.
				for _, headerValue := range headerValues {
					sliceKeys = append(sliceKeys, []string{remoteIP, path, headerKey, headerValue})
				}
			}
		}

	} else if limiterBasicAuthUsersIsSet {
		// Limit by Basic Auth credentials.
		username, _, ok := r.BasicAuth()
		if ok && libstring.StringInSlice(limiterBasicAuthUsers, username) {
			sliceKeys = append(sliceKeys, []string{remoteIP, path, username})
		}
	} else {
		// Default: Limit by remoteIP and path.
		sliceKeys = append(sliceKeys, []string{remoteIP, path})
	}

	return sliceKeys
}

// SetResponseHeaders configures X-Rate-Limit-Limit and X-Rate-Limit-Duration
func SetResponseHeaders(lmt *limiter.Limiter, w http.ResponseWriter) {
	w.Header().Add("X-Rate-Limit-Limit", strconv.FormatInt(lmt.Max, 10))
	w.Header().Add("X-Rate-Limit-Duration", lmt.TTL.String())
}

// LimitHandler is a middleware that performs rate-limiting given http.Handler struct.
func LimitHandler(lmt *limiter.Limiter, next http.Handler) http.Handler {
	middle := func(w http.ResponseWriter, r *http.Request) {
		SetResponseHeaders(lmt, w)

		httpError := LimitByRequest(lmt, r)
		if httpError != nil {
			w.Header().Add("Content-Type", lmt.GetMessageContentType())
			w.WriteHeader(httpError.StatusCode)
			w.Write([]byte(httpError.Message))

			lmt.ExecRejectFunc()
			return
		}

		// There's no rate-limit error, serve the next handler.
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(middle)
}

// LimitFuncHandler is a middleware that performs rate-limiting given request handler function.
func LimitFuncHandler(lmt *limiter.Limiter, nextFunc func(http.ResponseWriter, *http.Request)) http.Handler {
	return LimitHandler(lmt, http.HandlerFunc(nextFunc))
}
