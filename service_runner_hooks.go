package composer

import (
	"context"
	"log/slog"
	"time"

	"github.com/ferranbt/composer/hooks"
)

func (t *serviceRunner) preStart() error {
	if t.logger.Enabled(context.TODO(), slog.LevelDebug) {
		start := time.Now()
		t.logger.Debug("running prestart hooks", "start", start)
		defer func() {
			end := time.Now()
			t.logger.Debug("finished prestart hooks", "end", end, "duration", end.Sub(start))
		}()
	}

	for _, hook := range t.runnerHooks {
		pre, ok := hook.(hooks.ServicePrestartHook)
		if !ok {
			continue
		}

		name := pre.Name()

		// Build the request
		req := hooks.ServicePrestartHookRequest{
			Service: t.service,
		}

		var start time.Time
		if t.logger.Enabled(context.TODO(), slog.LevelDebug) {
			start = time.Now()
			t.logger.Debug("running prestart hook", "name", name, "start", start)
		}

		// Run the prestart hook
		if err := pre.Prestart(t.killCtx, &req); err != nil {
			return nil
		}

		if t.logger.Enabled(context.TODO(), slog.LevelDebug) {
			end := time.Now()
			t.logger.Debug("finished prestart hook", "name", name, "end", end, "duration", end.Sub(start))
		}
	}

	return nil
}

func (t *serviceRunner) postStart() error {
	if t.logger.Enabled(context.TODO(), slog.LevelDebug) {
		start := time.Now()
		t.logger.Debug("running poststart hooks", "start", start)
		defer func() {
			end := time.Now()
			t.logger.Debug("finished poststart hooks", "end", end, "duration", end.Sub(start))
		}()
	}

	for _, hook := range t.runnerHooks {
		post, ok := hook.(hooks.ServicePoststartHook)
		if !ok {
			continue
		}

		name := post.Name()

		// Build the request
		req := hooks.ServicePoststartHookRequest{
			Ip: t.handle.Ip,
		}

		var start time.Time
		if t.logger.Enabled(context.TODO(), slog.LevelDebug) {
			start = time.Now()
			t.logger.Debug("running poststart hook", "name", name, "start", start)
		}

		// Run the poststart hook
		if err := post.Poststart(t.killCtx, &req); err != nil {
			return nil
		}

		if t.logger.Enabled(context.TODO(), slog.LevelDebug) {
			end := time.Now()
			t.logger.Debug("finished poststart hook", "name", name, "end", end, "duration", end.Sub(start))
		}
	}

	return nil
}

func (t *serviceRunner) stop() error {
	if t.logger.Enabled(context.TODO(), slog.LevelDebug) {
		start := time.Now()
		t.logger.Debug("running stop hooks", "start", start)
		defer func() {
			end := time.Now()
			t.logger.Debug("finished stop hooks", "end", end, "duration", end.Sub(start))
		}()
	}

	for _, hook := range t.runnerHooks {
		stop, ok := hook.(hooks.ServiceStopHook)
		if !ok {
			continue
		}

		name := stop.Name()

		// Build the request
		req := hooks.ServiceStopRequest{}

		var start time.Time
		if t.logger.Enabled(context.TODO(), slog.LevelDebug) {
			start = time.Now()
			t.logger.Debug("running stop hook", "name", name, "start", start)
		}

		// Run the stop hook
		if err := stop.Stop(t.killCtx, &req); err != nil {
			return nil
		}

		if t.logger.Enabled(context.TODO(), slog.LevelDebug) {
			end := time.Now()
			t.logger.Debug("finished stop hook", "name", name, "end", end, "duration", end.Sub(start))
		}
	}

	return nil
}
