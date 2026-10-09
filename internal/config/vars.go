package config

import (
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/keroku/keroku/internal/validate"
)

func intVar(src Source, name string, def, minimum, maximum int) (int, error) {
	raw := strings.TrimSpace(src.Getenv(name))
	if raw == "" {
		return def, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < minimum || n > maximum {
		return 0, fmt.Errorf("config: %s must be a whole number from %d to %d", name, minimum, maximum)
	}
	return n, nil
}

func boolVar(src Source, name string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(src.Getenv(name))) {
	case "", "false", "0":
		return false, nil
	case "true", "1":
		return true, nil
	}
	return false, fmt.Errorf("config: %s must be true or false", name)
}

func durationVar(src Source, name string, def time.Duration) (time.Duration, error) {
	raw := src.Getenv(name)
	if strings.TrimSpace(raw) == "" {
		return def, nil
	}
	d, err := validate.Duration(raw, time.Second, 5*time.Minute)
	if err != nil {
		return 0, fmt.Errorf("config: %s must be between 1s and 5m, like 20s", name)
	}
	return d, nil
}

func logLevel(raw string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	}
	return 0, fmt.Errorf("config: LOG_LEVEL must be debug, info, warn or error")
}

// metricsAddr only allows the loopback interface. "off" disables the listener.
func metricsAddr(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	switch raw {
	case "":
		return "127.0.0.1:9100", nil
	case "off":
		return "", nil
	}
	host, port, err := net.SplitHostPort(raw)
	if err != nil || host != "127.0.0.1" {
		return "", fmt.Errorf("config: METRICS_ADDR must be 127.0.0.1:port or off")
	}
	p, err := strconv.Atoi(port)
	if err != nil || p < 1 || p > 65535 {
		return "", fmt.Errorf("config: METRICS_ADDR port must be 1 to 65535")
	}
	return raw, nil
}
