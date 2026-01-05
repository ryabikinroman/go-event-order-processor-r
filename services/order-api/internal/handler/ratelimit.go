package handler

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter implements per-IP rate limiting using token bucket algorithm
type RateLimiter struct {
	limiters    map[string]*rate.Limiter
	mu          sync.RWMutex
	rate        rate.Limit
	burst       int
	cleanupOnce sync.Once
}

// NewRateLimiter creates a new rate limiter
// ratePerSecond: maximum requests per second per IP
// burst: maximum burst size (allows short bursts above the rate)
func NewRateLimiter(ratePerSecond float64, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(ratePerSecond),
		burst:    burst,
	}
}

// getLimiter returns the rate limiter for the given IP address
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[ip] = limiter
	}

	return limiter
}

// Cleanup removes old limiters to prevent memory leaks
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		// Remove limiters that haven't been used (simple cleanup)
		// In production, you'd want more sophisticated tracking
		if len(rl.limiters) > 10000 {
			rl.limiters = make(map[string]*rate.Limiter)
		}
		rl.mu.Unlock()
	}
}

// Limit is a middleware that applies rate limiting per IP
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	// Start cleanup goroutine only once
	rl.cleanupOnce.Do(func() {
		go rl.cleanup()
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-RateLimit-Limit", "100")
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"rate limit exceeded","retry_after_seconds":60}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getIP extracts the real IP address from the request
func getIP(r *http.Request) string {
	// Check X-Forwarded-For header (set by proxies/load balancers)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP in the list
		if ip, _, err := net.SplitHostPort(forwarded); err == nil {
			return ip
		}
		return forwarded
	}

	// Check X-Real-IP header (set by some proxies)
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}
