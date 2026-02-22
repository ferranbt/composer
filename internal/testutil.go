package internal

import (
	"net/http"
	"net/http/httptest"
	"sync"
)

// MockHealthServer creates a test HTTP server that becomes healthy after a number of calls.
// Returns the server and a function to get the call count.
func MockHealthServer(healthyAfter int) (*httptest.Server, func() int) {
	var callCount int
	var mu sync.Mutex

	getCallCount := func() int {
		mu.Lock()
		defer mu.Unlock()
		return callCount
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callCount++
		healthy := callCount >= healthyAfter
		mu.Unlock()

		if healthy {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}))

	return server, getCallCount
}
