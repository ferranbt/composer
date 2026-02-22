package internal

import (
	"context"
	"net/http"
	"time"
)

type HealthCheck struct {
	HTTP         *HTTPCheck
	InitialDelay time.Duration
	Interval     time.Duration
	Timeout      time.Duration
}

type HTTPCheck struct {
	URL            string
	ExpectedStatus int
}

type Prober struct {
	check   *HealthCheck
	updater func(healthy bool)
	stop    chan struct{}
}

func NewProber(check *HealthCheck, updater func(healthy bool)) *Prober {
	return &Prober{
		check:   check,
		updater: updater,
		stop:    make(chan struct{}),
	}
}

func (p *Prober) Start() {
	go p.run()
}

func (p *Prober) Stop() {
	close(p.stop)
}

func (p *Prober) run() {
	// Wait for initial delay
	select {
	case <-time.After(p.check.InitialDelay):
	case <-p.stop:
		return
	}

	// Start probing
	ticker := time.NewTicker(p.check.Interval)
	defer ticker.Stop()

	for {
		healthy := p.probe()
		p.updater(healthy)

		if healthy {
			return
		}

		select {
		case <-ticker.C:
		case <-p.stop:
			return
		}
	}
}

func (p *Prober) probe() bool {
	if p.check.HTTP != nil {
		return p.probeHTTP(p.check.HTTP)
	}
	return false
}

func (p *Prober) probeHTTP(check *HTTPCheck) bool {
	ctx, cancel := context.WithTimeout(context.Background(), p.check.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", check.URL, nil)
	if err != nil {
		return false
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == check.ExpectedStatus
}
