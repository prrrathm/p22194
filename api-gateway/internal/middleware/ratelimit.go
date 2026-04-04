package middleware

import (
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

// ipLimiter holds a rate limiter per remote IP.
type ipLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	rps      rate.Limit
	burst    int
}

func newIPLimiter(rps float64, burst int) *ipLimiter {
	return &ipLimiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      rate.Limit(rps),
		burst:    burst,
	}
}

func (il *ipLimiter) get(ip string) *rate.Limiter {
	il.mu.Lock()
	defer il.mu.Unlock()
	lim, ok := il.limiters[ip]
	if !ok {
		lim = rate.NewLimiter(il.rps, il.burst)
		il.limiters[ip] = lim
	}
	return lim
}

// RateLimit returns a middleware that applies a per-IP token-bucket rate
// limiter. Requests exceeding the limit receive a 429 response immediately.
//
//   - rps:   steady-state requests per second per IP
//   - burst: maximum burst size (requests allowed before throttling kicks in)
func RateLimit(rps float64, burst int) func(http.Handler) http.Handler {
	il := newIPLimiter(rps, burst)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := realIP(r)
			if !il.get(ip).Allow() {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// realIP extracts the client IP from the request, respecting common proxy
// headers before falling back to RemoteAddr.
func realIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For may be a comma-separated list; use the first entry.
		for i := 0; i < len(ip); i++ {
			if ip[i] == ',' {
				return ip[:i]
			}
		}
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
