package logging

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
)

const secret = "abc.def.secretvalue"

func capture(t *testing.T, level slog.Level) (*slog.Logger, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	return New(&buf, level, NewRedactor(secret, "")), &buf
}

type stringer struct{}

func (stringer) String() string { return "s:" + secret }

type nested struct {
	Name  string
	Token string
}

func TestEveryPathIsRedacted(t *testing.T) {
	log, buf := capture(t, slog.LevelDebug)
	log.Info("msg with "+secret,
		"plain", "value "+secret,
		"err", fmt.Errorf("wrap: %w", errors.New(secret)),
		"str", stringer{},
		"bytes", []byte(secret),
		"struct", nested{Name: secret},
		slog.Group("g", "inner", secret),
		"token", "anything",
		"dbPassword", "anything",
		"Authorization", "Bot x",
	)
	log.With("bound", secret).WithGroup("grp").Warn("x", "k", secret)
	out := buf.String()
	if strings.Contains(out, "secretvalue") || strings.Contains(out, "anything") || strings.Contains(out, "Bot x") {
		t.Fatalf("leak: %s", out)
	}
	if strings.Count(out, redacted) < 10 {
		t.Fatalf("expected redaction markers: %s", out)
	}
}

func TestOutputIsJSONPerLine(t *testing.T) {
	log, buf := capture(t, slog.LevelInfo)
	log.Info("hello", "n", 3, "ok", true)
	log.Error("bad", "id", "x")
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines", len(lines))
	}
	var rec map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &rec); err != nil {
		t.Fatal(err)
	}
	if rec["msg"] != "hello" || rec["n"] != float64(3) || rec["ok"] != true {
		t.Fatalf("got %v", rec)
	}
}

func TestLevelFilter(t *testing.T) {
	log, buf := capture(t, slog.LevelWarn)
	log.Info("hidden")
	log.Debug("hidden")
	log.Warn("shown")
	if strings.Contains(buf.String(), "hidden") || !strings.Contains(buf.String(), "shown") {
		t.Fatalf("got %s", buf.String())
	}
}

func TestTokenShapeRedactedWithoutKnownSecret(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, slog.LevelInfo, NewRedactor())
	shaped := strings.Repeat("A", 24) + "." + strings.Repeat("b", 6) + "." + strings.Repeat("c", 27)
	log.Info("x", "v", "header "+shaped)
	if strings.Contains(buf.String(), shaped) {
		t.Fatalf("token shaped value leaked: %s", buf.String())
	}
}
