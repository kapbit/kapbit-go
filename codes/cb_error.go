package codes

import "fmt"

// CircuitBreakerOpenError is returned when a requested service is currently
// unavailable due to an open circuit breaker.
type CircuitBreakerOpenError struct {
	service string
}

// NewCircuitBreakerOpenError creates a new CircuitBreakerOpenError for the
// specified service.
func NewCircuitBreakerOpenError(service string) error {
	return &CircuitBreakerOpenError{service: service}
}

// Error returns a formatted error message.
func (e *CircuitBreakerOpenError) Error() string {
	return fmt.Sprintf("circuit breaker open for %q", e.service)
}

// Service returns the name of the service that triggered the error.
func (e *CircuitBreakerOpenError) Service() string {
	return e.service
}
