// Package config loads runtime settings from the environment and mounted secret files.
package config

import (
	"errors"
	"log/slog"
	"time"
)

// Config is the validated runtime configuration.
type Config struct {
	Token           Secret
	DatabaseURL     Secret
	MetricsAddr     string
	ShardCount      int
	ShardIDs        []int
	LogLevel        slog.Level
	Workers         int
	QueueSize       int
	DBMaxConns      int
	MessageContent  bool
	ShutdownTimeout time.Duration
}

// Source abstracts the process environment so tests can supply their own.
type Source struct {
	Getenv   func(string) string
	ReadFile func(string) ([]byte, error)
}

// Load reads and validates every setting. The first problem found is returned.
func Load(src Source) (Config, error) {
	if src.Getenv == nil || src.ReadFile == nil {
		return Config{}, errors.New("config: source is incomplete")
	}
	var c Config
	var err error
	steps := []func() error{
		func() (e error) { c.Token, e = loadToken(src); return },
		func() (e error) { c.DatabaseURL, e = loadSecret(src, "DATABASE_URL"); return },
		func() (e error) { c.MetricsAddr, e = metricsAddr(src.Getenv("METRICS_ADDR")); return },
		func() (e error) { c.ShardCount, c.ShardIDs, e = shards(src); return },
		func() (e error) { c.LogLevel, e = logLevel(src.Getenv("LOG_LEVEL")); return },
		func() (e error) { c.Workers, e = intVar(src, "WORKERS", 32, 1, 1024); return },
		func() (e error) { c.QueueSize, e = intVar(src, "QUEUE_SIZE", 256, 1, 100000); return },
		func() (e error) { c.DBMaxConns, e = intVar(src, "DB_MAX_CONNS", 40, 2, 1000); return },
		func() (e error) { c.MessageContent, e = boolVar(src, "MESSAGE_CONTENT"); return },
		func() (e error) { c.ShutdownTimeout, e = durationVar(src, "SHUTDOWN_TIMEOUT", 20*time.Second); return },
	}
	for _, step := range steps {
		if err = step(); err != nil {
			return Config{}, err
		}
	}
	if c.DBMaxConns < c.Workers+4 {
		return Config{}, errors.New("config: DB_MAX_CONNS must be at least WORKERS plus 4")
	}
	return c, nil
}
