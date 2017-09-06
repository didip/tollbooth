// Package limiter provides data structure to configure rate-limiter.
package limiter

import (
	"sync"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"golang.org/x/time/rate"
)

// New is a constructor for Limiter.
func New(generalExpirableOptions *ExpirableOptions) *Limiter {
	lmt := &Limiter{}

	lmt.SetMessageContentType("text/plain; charset=utf-8").
		SetMessage("You have reached maximum request limit.").
		SetStatusCode(429).
		SetRejectFunc(nil).
		SetIPLookups([]string{"RemoteAddr", "X-Forwarded-For", "X-Real-IP"}).
		SetHeaders(make(map[string][]string))

	if generalExpirableOptions != nil {
		lmt.generalExpirableOptions = generalExpirableOptions
	} else {
		lmt.generalExpirableOptions = &ExpirableOptions{}
	}

	// Default for ExpireJobInterval is every minute.
	if lmt.generalExpirableOptions.ExpireJobInterval <= 0 {
		lmt.generalExpirableOptions.ExpireJobInterval = time.Minute
	}

	// Default for DefaultExpirationTTL is 10 years.
	if lmt.generalExpirableOptions.DefaultExpirationTTL <= 0 {
		lmt.generalExpirableOptions.DefaultExpirationTTL = 87600 * time.Hour
	}

	lmt.tokenBuckets = gocache.New(
		lmt.generalExpirableOptions.DefaultExpirationTTL,
		lmt.generalExpirableOptions.ExpireJobInterval,
	)

	lmt.basicAuthUsers = gocache.New(
		lmt.generalExpirableOptions.DefaultExpirationTTL,
		lmt.generalExpirableOptions.ExpireJobInterval,
	)

	return lmt
}

// Limiter is a config struct to limit a particular request handler.
type Limiter struct {
	// Maximum number of requests to limit per duration.
	max int64

	// Duration of rate-limiter.
	ttl time.Duration

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

	// Able to configure token bucket expirations.
	generalExpirableOptions *ExpirableOptions

	// List of basic auth usernames to limit.
	basicAuthUsers *gocache.Cache

	// Map of HTTP headers to limit.
	// Empty means skip headers checking.
	headers map[string]*gocache.Cache

	// Map of limiters with TTL
	tokenBuckets *gocache.Cache

	sync.RWMutex
}

// SetMax is thread-safe way of setting maximum number of requests to limit per duration.
func (l *Limiter) SetMax(max int64) *Limiter {
	l.Lock()
	l.max = max
	l.Unlock()

	return l
}

// GetMax is thread-safe way of getting maximum number of requests to limit per duration.
func (l *Limiter) GetMax() int64 {
	l.RLock()
	defer l.RUnlock()
	return l.max
}

// SetTTL is thread-safe way of setting maximum number of requests to limit per duration.
func (l *Limiter) SetTTL(ttl time.Duration) *Limiter {
	l.Lock()
	l.ttl = ttl
	l.Unlock()

	return l
}

// GetTTL is thread-safe way of getting maximum number of requests to limit per duration.
func (l *Limiter) GetTTL() time.Duration {
	l.RLock()
	defer l.RUnlock()
	return l.ttl
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
func (l *Limiter) SetRejectFunc(fn func()) *Limiter {
	l.Lock()
	l.rejectFunc = fn
	l.Unlock()

	return l
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
	for _, basicAuthUser := range basicAuthUsers {
		l.basicAuthUsers.Set(
			basicAuthUser,
			true,
			l.generalExpirableOptions.DefaultExpirationTTL,
		)
	}

	return l
}

// GetBasicAuthUsers is thread-safe way of getting list of basic auth usernames to limit.
func (l *Limiter) GetBasicAuthUsers() []string {
	asMap := l.basicAuthUsers.Items()

	var basicAuthUsers []string
	for basicAuthUser, _ := range asMap {
		basicAuthUsers = append(basicAuthUsers, basicAuthUser)
	}

	return basicAuthUsers
}

// RemoveBasicAuthUsers is thread-safe way of removing basic auth usernames from existing list.
func (l *Limiter) RemoveBasicAuthUsers(basicAuthUsers []string) *Limiter {
	for _, toBeRemoved := range basicAuthUsers {
		l.basicAuthUsers.Delete(toBeRemoved)
	}

	return l
}

// SetHeaders is thread-safe way of setting map of HTTP headers to limit.
func (l *Limiter) SetHeaders(headers map[string][]string) *Limiter {
	if l.headers == nil {
		l.headers = make(map[string]*gocache.Cache)
	}

	for header, entries := range headers {
		l.SetHeader(header, entries)
	}

	return l
}

// GetHeaders is thread-safe way of getting map of HTTP headers to limit.
func (l *Limiter) GetHeaders() map[string][]string {
	results := make(map[string][]string)

	l.RLock()
	defer l.RUnlock()

	for header, entriesAsGoCache := range l.headers {
		entries := make([]string, 0)

		for entry, _ := range entriesAsGoCache.Items() {
			entries = append(entries, entry)
		}

		results[header] = entries
	}

	return results
}

// SetHeader is thread-safe way of setting entries of 1 HTTP header.
func (l *Limiter) SetHeader(header string, entries []string) *Limiter {
	l.RLock()
	existing, found := l.headers[header]
	l.RUnlock()

	if !found {
		existing = gocache.New(
			l.generalExpirableOptions.DefaultExpirationTTL,
			l.generalExpirableOptions.ExpireJobInterval,
		)
	}

	for _, entry := range entries {
		existing.Set(
			entry,
			true,
			l.generalExpirableOptions.DefaultExpirationTTL,
		)
	}

	l.Lock()
	l.headers[header] = existing
	l.Unlock()

	return l
}

// GetHeader is thread-safe way of getting entries of 1 HTTP header.
func (l *Limiter) GetHeader(header string) []string {
	l.RLock()
	entriesAsGoCache := l.headers[header]
	l.RUnlock()

	entriesAsMap := entriesAsGoCache.Items()
	entries := make([]string, 0)

	for entry, _ := range entriesAsMap {
		entries = append(entries, entry)
	}

	return entries
}

// RemoveHeader is thread-safe way of removing entries of 1 HTTP header.
func (l *Limiter) RemoveHeader(header string) *Limiter {
	l.Lock()
	l.headers[header] = gocache.New(
		l.generalExpirableOptions.DefaultExpirationTTL,
		l.generalExpirableOptions.ExpireJobInterval,
	)
	l.Unlock()

	return l
}

// RemoveHeaderEntries is thread-safe way of adding new entries to 1 HTTP header rule.
func (l *Limiter) RemoveHeaderEntries(header string, entriesForRemoval []string) *Limiter {
	l.RLock()
	entries, found := l.headers[header]
	l.RUnlock()

	if !found {
		return l
	}

	for _, toBeRemoved := range entriesForRemoval {
		entries.Delete(toBeRemoved)
	}

	return l
}

func (l *Limiter) limitReachedWithTokenBucketTTL(key string, tokenBucketTTL time.Duration) bool {
	lmtMax := l.GetMax()
	lmtTTL := l.GetTTL()

	l.Lock()
	defer l.Unlock()

	if _, found := l.tokenBuckets.Get(key); !found {
		l.tokenBuckets.Set(
			key,
			rate.NewLimiter(rate.Every(lmtTTL), int(lmtMax)),
			tokenBucketTTL,
		)
	}

	expiringMap, found := l.tokenBuckets.Get(key)
	if !found {
		return false
	}

	return !expiringMap.(*rate.Limiter).AllowN(time.Now(), 1)
}

// LimitReached returns a bool indicating if the Bucket identified by key ran out of tokens.
func (l *Limiter) LimitReached(key string) bool {
	return l.limitReachedWithTokenBucketTTL(key, l.generalExpirableOptions.DefaultExpirationTTL)
}

// LimitReachedWithCustomTokenBucketTTL returns a bool indicating if the Bucket identified by key ran out of tokens.
// This public API allows user to define custom expiration TTL on the key.
func (l *Limiter) LimitReachedWithTokenBucketTTL(key string, tokenBucketTTL time.Duration) bool {
	return l.limitReachedWithTokenBucketTTL(key, tokenBucketTTL)
}
