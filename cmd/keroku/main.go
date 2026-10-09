// Command keroku runs the Discord moderation bot.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/7kimchi/keroku/internal/app"
	"github.com/7kimchi/keroku/internal/config"
	"github.com/7kimchi/keroku/internal/gateway"
	"github.com/7kimchi/keroku/internal/logging"
)

func main() {
	os.Exit(run())
}

func run() int {
	cfg, err := config.Load(config.Source{Getenv: os.Getenv, ReadFile: os.ReadFile})
	if err != nil {
		// Config errors never echo secret values, so a plain logger is enough here.
		logging.New(os.Stderr, slog.LevelInfo, logging.NewRedactor()).Error("config invalid", "err", err)
		return 2
	}
	log := logging.New(os.Stderr, cfg.LogLevel, logging.NewRedactor(cfg.Token.Reveal(), cfg.DatabaseURL.Reveal()))
	gateway.InstallLogger(log)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	a, err := app.New(ctx, cfg, log)
	if err != nil {
		log.Error("startup failed", "err", err)
		return 1
	}
	if err := a.Run(ctx); err != nil {
		log.Error("stopped with error", "err", err)
		return 1
	}
	return 0
}
