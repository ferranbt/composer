package internal

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestProberHTTPSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	var healthy bool
	var mu sync.Mutex
	done := make(chan struct{})

	check := &HealthCheck{
		HTTP: &HTTPCheck{
			URL:            server.URL,
			ExpectedStatus: http.StatusOK,
		},
		InitialDelay: 10 * time.Millisecond,
		Interval:     50 * time.Millisecond,
		Timeout:      1 * time.Second,
	}

	prober := NewProber(check, func(h bool) {
		mu.Lock()
		healthy = h
		mu.Unlock()
		close(done)
	})

	prober.Start()
	defer prober.Stop()

	select {
	case <-done:
		mu.Lock()
		require.True(t, healthy)
		mu.Unlock()
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for health check")
	}
}

func TestProberHTTPFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	var updates []bool
	var mu sync.Mutex

	check := &HealthCheck{
		HTTP: &HTTPCheck{
			URL:            server.URL,
			ExpectedStatus: http.StatusOK,
		},
		InitialDelay: 10 * time.Millisecond,
		Interval:     50 * time.Millisecond,
		Timeout:      1 * time.Second,
	}

	prober := NewProber(check, func(h bool) {
		mu.Lock()
		updates = append(updates, h)
		mu.Unlock()
	})

	prober.Start()

	time.Sleep(200 * time.Millisecond)
	prober.Stop()

	mu.Lock()
	defer mu.Unlock()
	require.NotEmpty(t, updates)
	for _, u := range updates {
		require.False(t, u)
	}
}

func TestProberEventuallyHealthy(t *testing.T) {
	server, _ := MockHealthServer(3)
	defer server.Close()

	var healthy bool
	var healthMu sync.Mutex
	done := make(chan struct{})

	check := &HealthCheck{
		HTTP: &HTTPCheck{
			URL:            server.URL,
			ExpectedStatus: http.StatusOK,
		},
		InitialDelay: 10 * time.Millisecond,
		Interval:     50 * time.Millisecond,
		Timeout:      1 * time.Second,
	}

	prober := NewProber(check, func(h bool) {
		if h {
			healthMu.Lock()
			healthy = h
			healthMu.Unlock()
			close(done)
		}
	})

	prober.Start()
	defer prober.Stop()

	select {
	case <-done:
		healthMu.Lock()
		require.True(t, healthy)
		healthMu.Unlock()
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for health check")
	}
}
