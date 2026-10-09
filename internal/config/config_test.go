package config

import (
	"log/slog"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	c, err := Load(env(base(), nil))
	if err != nil {
		t.Fatal(err)
	}
	if c.Token.Reveal() != fakeToken || c.DatabaseURL.Reveal() != "postgres://u:p@db/keroku" {
		t.Fatal("secrets not loaded")
	}
	if c.MetricsAddr != "127.0.0.1:9100" || c.ShardCount != 0 || c.ShardIDs != nil {
		t.Fatalf("bad defaults %+v", c)
	}
	if c.LogLevel != slog.LevelInfo || c.Workers != 32 || c.QueueSize != 256 || c.DBMaxConns != 40 {
		t.Fatalf("bad defaults %+v", c)
	}
	if c.MessageContent || c.ShutdownTimeout != 20*time.Second {
		t.Fatalf("bad defaults %+v", c)
	}
}

func TestLoadOverrides(t *testing.T) {
	m := base()
	for k, v := range map[string]string{
		"METRICS_ADDR": "127.0.0.1:9999", "SHARD_COUNT": "8", "SHARD_IDS": "0-2, 5",
		"LOG_LEVEL": "DEBUG", "WORKERS": "64", "QUEUE_SIZE": "10", "DB_MAX_CONNS": "100",
		"MESSAGE_CONTENT": "true", "SHUTDOWN_TIMEOUT": "45s",
	} {
		m[k] = v
	}
	c, err := Load(env(m, nil))
	if err != nil {
		t.Fatal(err)
	}
	if c.MetricsAddr != "127.0.0.1:9999" || c.ShardCount != 8 || len(c.ShardIDs) != 4 || c.ShardIDs[3] != 5 {
		t.Fatalf("got %+v", c)
	}
	if c.LogLevel != slog.LevelDebug || c.Workers != 64 || c.QueueSize != 10 || c.DBMaxConns != 100 {
		t.Fatalf("got %+v", c)
	}
	if !c.MessageContent || c.ShutdownTimeout != 45*time.Second {
		t.Fatalf("got %+v", c)
	}
}

func TestLoadRejectsBadValues(t *testing.T) {
	for k, v := range map[string]string{
		"WORKERS": "0", "QUEUE_SIZE": "-1", "DB_MAX_CONNS": "1", "LOG_LEVEL": "trace",
		"MESSAGE_CONTENT": "yes", "SHUTDOWN_TIMEOUT": "10m", "SHARD_COUNT": "4097",
		"METRICS_ADDR": "0.0.0.0:9100",
	} {
		mustFail(t, with(k, v), nil)
	}
	mustFail(t, with("WORKERS", "99999999999999999999"), nil)
	mustFail(t, with("WORKERS", "100"), nil)
}

func TestMetricsAddr(t *testing.T) {
	if c, err := Load(env(with("METRICS_ADDR", "off"), nil)); err != nil || c.MetricsAddr != "" {
		t.Fatalf("off: %q %v", c.MetricsAddr, err)
	}
	for _, bad := range []string{
		":9100", "localhost:9100", "[::]:9100", "10.0.0.1:9100", "127.0.0.1", "127.0.0.1:0",
		"127.0.0.1:65536", "127.0.0.1:http", "127.0.0.2:9100",
	} {
		mustFail(t, with("METRICS_ADDR", bad), nil)
	}
}
