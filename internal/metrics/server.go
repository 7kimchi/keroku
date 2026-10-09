package metrics

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/7kimchi/keroku/internal/safe"
)

// Serve exposes /metrics on addr until ctx ends. addr must be on 127.0.0.1.
func (m *Metrics) Serve(ctx context.Context, addr string, g *safe.Guard) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil || host != "127.0.0.1" {
		return fmt.Errorf("metrics: refusing to listen on %q", addr)
	}
	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("metrics: listen: %w", err)
	}
	return m.serve(ctx, ln, g)
}

func (m *Metrics) serve(ctx context.Context, ln net.Listener, g *safe.Guard) error {
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.HandlerFor(m.Registry, promhttp.HandlerOpts{
		MaxRequestsInFlight: 4,
		Timeout:             5 * time.Second,
	}))
	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    8 << 10,
	}
	done := make(chan error, 1)
	var wg sync.WaitGroup
	defer wg.Wait()
	g.Go(&wg, "metricsServer", func() { done <- srv.Serve(ln) })
	select {
	case err := <-done:
		return fmt.Errorf("metrics: serve: %w", err)
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	err := srv.Shutdown(shutdownCtx)
	if serveErr := <-done; !errors.Is(serveErr, http.ErrServerClosed) && err == nil {
		err = serveErr
	}
	return err
}
