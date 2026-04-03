package middleware

import (
	"net/http"
	"sync"
	"time"
)

type userTracker struct {
	count  int
	window time.Time
}

var (
	userRateLimits = make(map[string]*userTracker)
	rateMu         sync.RWMutex

	RequestsPerMinute = 60
	BurstLimit        = 10
	RateLimitWindow   = time.Minute
)

func UserRateLimitMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := GetUserID(r.Context())
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			userKey := userID.String()

			rateMu.Lock()
			tracker, exists := userRateLimits[userKey]
			now := time.Now()

			if !exists || now.Sub(tracker.window) > RateLimitWindow {
				userRateLimits[userKey] = &userTracker{
					count:  1,
					window: now,
				}
				rateMu.Unlock()
				next.ServeHTTP(w, r)
				return
			}

			if tracker.count >= RequestsPerMinute {
				rateMu.Unlock()
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "60")
				http.Error(w, `{"error":"rate limit exceeded","code":"RATE_LIMIT_EXCEEDED","details":"Too many requests. Please try again later."}`, http.StatusTooManyRequests)
				return
			}

			tracker.count++
			rateMu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

func ClearUserRateLimits() {
	rateMu.Lock()
	defer rateMu.Unlock()
	userRateLimits = make(map[string]*userTracker)
}
