package app

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/7kimchi/keroku/internal/gateway"
)

// shutdown stops intake first, then drains queued commands, then background jobs, then the
// modlog queue they feed, then the pool.
func (a *App) shutdown(parent context.Context, gw *gateway.Gateway, stopBackground context.CancelFunc, bg *sync.WaitGroup) error {
	a.log.Info("shutting down", "timeout", a.cfg.ShutdownTimeout.String())
	// parent is already cancelled when shutdown starts, so only its values are kept.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parent), a.cfg.ShutdownTimeout)
	defer cancel()
	gw.Close()
	// Ack jobs feed the lanes, so the ack pool drains first.
	err := errors.Join(a.ack.Close(ctx), a.lanes.Close(ctx))
	stopBackground()
	bg.Wait()
	err = errors.Join(err, a.modlog.Close(ctx))
	a.store.Close()
	a.log.Info("shutdown complete")
	return err
}

// housekeeping sweeps idle limiter and dedupe entries and reports queue depth.
func (a *App) housekeeping(ctx context.Context) {
	t := time.NewTicker(30 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			a.userLimit.Sweep(2000)
			a.guildLimit.Sweep(2000)
			a.seen.Sweep(2000)
			a.settings.Sweep()
			a.metrics.QueueDepth.WithLabelValues("modlog").Set(float64(a.modlog.Depth()))
			a.metrics.QueueDepth.WithLabelValues("ack").Set(float64(a.ack.Depth()))
			a.metrics.QueueDepth.WithLabelValues("lanes").Set(float64(a.lanes.Depth()))
		}
	}
}
