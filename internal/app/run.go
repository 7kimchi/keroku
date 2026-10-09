package app

import (
	"context"
	"errors"
	"slices"
	"sync"

	"github.com/7kimchi/keroku/internal/gateway"
)

// Run connects the gateway and blocks until ctx ends, then shuts down in order.
func (a *App) Run(ctx context.Context) error {
	bgCtx, stopBackground := context.WithCancel(context.WithoutCancel(ctx))
	var bg sync.WaitGroup
	if a.cfg.MetricsAddr != "" {
		a.guard.Go(&bg, "metricsServer", func() {
			if err := a.metrics.Serve(bgCtx, a.cfg.MetricsAddr, a.guard); err != nil {
				a.log.Error("metrics server stopped", "err", err)
			}
		})
	}
	a.guard.Go(&bg, "housekeeping", func() { a.housekeeping(bgCtx) })

	gw := gateway.New(gateway.Config{
		Token: a.cfg.Token.Reveal(), ShardCount: a.cfg.ShardCount, ShardIDs: a.cfg.ShardIDs,
		MessageContent: a.cfg.MessageContent, HTTPClient: a.gatewayHTTP,
	}, a.router, a.guard, a.log)
	err := a.start(ctx, gw)
	if err == nil {
		<-ctx.Done()
	}
	return errors.Join(err, a.shutdown(ctx, gw, stopBackground, &bg))
}

// start resolves the bot's identity, registers commands and opens the shards.
func (a *App) start(ctx context.Context, gw *gateway.Gateway) error {
	// Only the instance running shard 0 registers, so a fleet does not race on it.
	if a.cfg.ShardIDs == nil || slices.Contains(a.cfg.ShardIDs, 0) {
		if err := a.client.OverwriteCommands(ctx, a.appID, a.registry.Definitions()); err != nil {
			return err
		}
	}
	return gw.Open(ctx)
}
