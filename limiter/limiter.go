// Package limiter provides data structure to configure rate-limiter.
package limiter

import (
	"sync"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"golang.org/x/time/rate"
)

// New is a constructor for Limiter.
func New(max int64, ttl time.Duration, tbOptions *TokenBucketOptions) *Limiter {
	limiter := &Limiter{Max: max, TTL: ttl}
	limiter.SetMessageContentType("text/plain; charset=utf-8")
	limiter.SetMessage("You have reached maximum request limit.")
	limiter.SetStatusCode(429)
	limiter.SetRejectFunc(nil)
	limiter.SetIPLookups([]string{"RemoteAddr", "X-Forwarded-For", "X-Real-IP"})
	limiter.SetHeaders(make(map[string][]string))

	limiter.tokenBucketsNoTTL = make(map[string]*rate.Limiter)

	if tbOptions != nil {
		limiter.tokenBucketOptions = tbOptions

		// Default for ExpireJobInterval is every minute.
		if limiter.tokenBucketOptions.ExpireJobInterval <= 0 {
			limiter.tokenBucketOptions.ExpireJobInterval = time.Minute
		}

		limiter.tokenBucketsWithTTL = gocache.New(
			limiter.tokenBucketOptions.DefaultExpirationTTL,
			limiter.tokenBucketOptions.ExpireJobInterval,
		)
	}

	return limiter
}

// Limiter is a config struct to limit a particular request handler.
type Limiter struct {
	// HTTP message when limit is reached.
	message string

	// Content-Type for Message
	messageContentType string

	// HTTP status code when limit is reached.
	statusCode int

	// A function to call when a request is rejected.
	rejectFunc func()

	// List of places to look up IP address.
	// Default is "RemoteAddr", "X-Forwarded-For", "X-Real-IP".
	// You can rearrange the order as you like.
	ipLookups []string

	// List of HTTP Methods to limit (GET, POST, PUT, etc.).
	// Empty means limit all methods.
	methods []string

	// List of basic auth usernames to limit.
	basicAuthUsers []string

	// Map of HTTP headers to limit.
	// Empty means skip headers checking.
	headers map[string][]string

	// Maximum number of requests to limit per duration.
	Max int64

	// Duration of rate-limiter.
	TTL time.Duration

	// Able to configure token bucket expirations.
	tokenBucketOptions *TokenBucketOptions

	// Map of limiters without TTL
	tokenBucketsNoTTL map[string]*rate.Limiter

	// Map of limiters with TTL
	tokenBucketsWithTTL *gocache.Cache

	sync.RWMutex
}

// SetMessage is thread-safe way of setting HTTP message when limit is reached.
func (l *Limiter) SetMessage(msg string) *Limiter {
	l.Lock()
	l.message = msg
	l.Unlock()

	return l
}

// GetMessage is thread-safe way of getting HTTP message when limit is reached.
func (l *Limiter) GetMessage() string {
	l.RLock()
	defer l.RUnlock()
	return l.message
}

// SetMessageContentType is thread-safe way of setting HTTP message Content-Type when limit is reached.
func (l *Limiter) SetMessageContentType(contentType string) *Limiter {
	l.Lock()
	l.messageContentType = contentType
	l.Unlock()

	return l
}

// GetMessageContentType is thread-safe way of getting HTTP message Content-Type when limit is reached.
func (l *Limiter) GetMessageContentType() string {
	l.RLock()
	defer l.RUnlock()
	return l.messageContentType
}

// SetStatusCode is thread-safe way of setting HTTP status code when limit is reached.
func (l *Limiter) SetStatusCode(statusCode int) *Limiter {
	l.Lock()
	l.statusCode = statusCode
	l.Unlock()

	return l
}

// GetStatusCode is thread-safe way of getting HTTP status code when limit is reached.
func (l *Limiter) GetStatusCode() int {
	l.RLock()
	defer l.RUnlock()
	return l.statusCode
}

// SetRejectFunc is thread-safe way of setting after-rejection function when limit is reached.
func (l *Limiter) SetRejectFunc(fn func()) {
	l.Lock()
	l.rejectFunc = fn
	l.Unlock()
}

// ExecRejectFunc is thread-safe way of executing after-rejection function when limit is reached.
func (l *Limiter) ExecRejectFunc() {
	l.RLock()
	defer l.RUnlock()

	fn := l.rejectFunc
	if fn != nil {
		fn()
	}
}

// SetIPLookups is thread-safe way of setting list of places to look up IP address.
func (l *Limiter) SetIPLookups(ipLookups []string) *Limiter {
	l.Lock()
	l.ipLookups = ipLookups
	l.Unlock()

	return l
}

// GetIPLookups is thread-safe way of getting list of places to look up IP address.
func (l *Limiter) GetIPLookups() []string {
	l.RLock()
	defer l.RUnlock()
	return l.ipLookups
}

// SetMethods is thread-safe way of setting list of HTTP Methods to limit (GET, POST, PUT, etc.).
func (l *Limiter) SetMethods(methods []string) *Limiter {
	l.Lock()
	l.methods = methods
	l.Unlock()

	return l
}

// GetMethods is thread-safe way of getting list of HTTP Methods to limit (GET, POST, PUT, etc.).
func (l *Limiter) GetMethods() []string {
	l.RLock()
	defer l.RUnlock()
	return l.methods
}

// SetBasicAuthUsers is thread-safe way of setting list of basic auth usernames to limit.
func (l *Limiter) SetBasicAuthUsers(basicAuthUsers []string) *Limiter {
	l.Lock()
	l.basicAuthUsers = basicAuthUsers
	l.Unlock()

	return l
}

// GetBasicAuthUsers is thread-safe way of getting list of basic auth usernames to limit.
func (l *Limiter) GetBasicAuthUsers() []string {
	l.RLock()
	defer l.RUnlock()
	return l.basicAuthUsers
}

// AddBasicAuthUsers is thread-safe way of adding basic auth usernames to existing list.
func (l *Limiter) AddBasicAuthUsers(basicAuthUsers []string) *Limiter {
	l.Lock()
	defer l.Unlock()

	for _, toBeAdded := range basicAuthUsers {
		alreadyExists := false
		for _, existing := range l.basicAuthUsers {
			if existing == toBeAdded {
				alreadyExists = true
				break
			}
		}

		if !alreadyExists {
			l.basicAuthUsers = append(l.basicAuthUsers, toBeAdded)
		}
	}

	return l
}

// RemoveBasicAuthUsers is thread-safe way of removing basic auth usernames from existing list.
func (l *Limiter) RemoveBasicAuthUsers(basicAuthUsers []string) *Limiter {
	newList := make([]string, 0)

	l.RLock()
	for _, existing := range l.basicAuthUsers {
		found := false
		for _, toBeRemoved := range basicAuthUsers {
			if existing == toBeRemoved {
				found = true
				break
			}
		}

		if !found {
			newList = append(newList, existing)
		}
	}
	l.RUnlock()

	l.Lock()
	l.basicAuthUsers = newList
	l.Unlock()

	return l
}

// SetHeaders is thread-safe way of setting map of HTTP headers to limit.
func (l *Limiter) SetHeaders(headers map[string][]string) *Limiter {
	l.Lock()
	l.headers = headers
	l.Unlock()

	return l
}

// GetHeaders is thread-safe way of getting map of HTTP headers to limit.
func (l *Limiter) GetHeaders() map[string][]string {
	l.RLock()
	defer l.RUnlock()
	return l.headers
}

// SetHeader is thread-safe way of setting entries of 1 HTTP header.
func (l *Limiter) SetHeader(header string, entries []string) *Limiter {
	l.Lock()
	l.headers[header] = entries
	l.Unlock()

	return l
}

// GetHeader is thread-safe way of getting entries of 1 HTTP header.
func (l *Limiter) GetHeader(header string) []string {
	l.RLock()
	defer l.RUnlock()
	return l.headers[header]
}

// RemoveHeader is thread-safe way of removing entries of 1 HTTP header.
func (l *Limiter) RemoveHeader(header string) *Limiter {
	l.Lock()
	l.headers[header] = make([]string, 0)
	l.Unlock()

	return l
}

// AddHeaderEntries is thread-safe way of adding new entries to 1 HTTP header rule.
func (l *Limiter) AddHeaderEntries(header string, newEntries []string) *Limiter {
	l.Lock()
	defer l.Unlock()

	if len(l.headers[header]) == 0 {
		l.headers[header] = newEntries
		return l
	}

	for _, newEntry := range newEntries {
		alreadyExists := false
		for _, existing := range l.headers[header] {
			if existing == newEntry {
				alreadyExists = true
				break
			}
		}

		if !alreadyExists {
			l.headers[header] = append(l.headers[header], newEntry)
		}
	}

	return l
}

// RemoveHeaderEntries is thread-safe way of adding new entries to 1 HTTP header rule.
func (l *Limiter) RemoveHeaderEntries(header string, entriesForRemoval []string) *Limiter {
	newList := make([]string, 0)

	l.RLock()
	for _, existing := range l.headers[header] {
		found := false
		for _, toBeRemoves := range entriesForRemoval {
			if existing == toBeRemoves {
				found = true
				break
			}
		}

		if !found {
			newList = append(newList, existing)
		}
	}
	l.RUnlock()

	l.Lock()
	l.headers[header] = newList
	l.Unlock()

	return l
}

func (l *Limiter) isUsingTokenBucketsWithTTL() bool {
	if l.tokenBucketOptions == nil {
		return false
	}
	return l.tokenBucketOptions.DefaultExpirationTTL > 0
}

func (l *Limiter) limitReachedNoTokenBucketTTL(key string) bool {
	l.Lock()
	defer l.Unlock()

	if _, found := l.tokenBucketsNoTTL[key]; !found {
		l.tokenBucketsNoTTL[key] = rate.NewLimiter(rate.Every(l.TTL), int(l.Max))
	}

	return !l.tokenBucketsNoTTL[key].AllowN(time.Now(), 1)
}

func (l *Limiter) limitReachedWithDefaultTokenBucketTTL(key string) bool {
	return l.limitReachedWithCustomTokenBucketTTL(key, gocache.DefaultExpiration)
}

func (l *Limiter) limitReachedWithCustomTokenBucketTTL(key string, tokenBucketTTL time.Duration) bool {
	l.Lock()
	defer l.Unlock()

	if _, found := l.tokenBucketsWithTTL.Get(key); !found {
		l.tokenBucketsWithTTL.Set(
			key,
			rate.NewLimiter(rate.Every(l.TTL), int(l.Max)),
			tokenBucketTTL,
		)
	}

	expiringMap, found := l.tokenBucketsWithTTL.Get(key)
	if !found {
		return false
	}

	return !expiringMap.(*rate.Limiter).AllowN(time.Now(), 1)
}

// LimitReached returns a bool indicating if the Bucket identified by key ran out of tokens.
func (l *Limiter) LimitReached(key string) bool {
	if l.isUsingTokenBucketsWithTTL() {
		return l.limitReachedWithDefaultTokenBucketTTL(key)

	} else {
		return l.limitReachedNoTokenBucketTTL(key)
	}

	return false
}

// LimitReachedWithCustomTokenBucketTTL returns a bool indicating if the Bucket identified by key ran out of tokens.
// This public API allows user to define custom expiration TTL on the key.
func (l *Limiter) LimitReachedWithCustomTokenBucketTTL(key string, tokenBucketTTL time.Duration) bool {
	if l.isUsingTokenBucketsWithTTL() {
		return l.limitReachedWithCustomTokenBucketTTL(key, tokenBucketTTL)

	} else {
		return l.limitReachedNoTokenBucketTTL(key)
	}

	return false
}
