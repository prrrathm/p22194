package circuitbreaker

import (
	"time"

	"github.com/sony/gobreaker"
)

// New returns a CircuitBreaker for the named upstream service.
//
// The breaker opens after 5 consecutive failures and attempts to half-open
// after a 30-second cool-down. It allows up to 3 probe requests while
// half-open before fully closing again.
func New(name string) *gobreaker.CircuitBreaker {
	return gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        name,
		MaxRequests: 3,
		Interval:    60 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
	})
}
