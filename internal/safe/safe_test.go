package safe

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func newGuard() (*Guard, *bytes.Buffer, *atomic.Int64) {
	var buf bytes.Buffer
	var n atomic.Int64
	log := slog.New(slog.NewJSONHandler(&lockedWriter{w: &buf}, nil))
	return NewGuard(log, func(string) { n.Add(1) }), &buf, &n
}

type lockedWriter struct {
	mu sync.Mutex
	w  *bytes.Buffer
}

func (l *lockedWriter) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w.Write(p)
}

func TestRunNoPanic(t *testing.T) {
	g, buf, n := newGuard()
	ran := false
	if g.Run("x", func() { ran = true }) || !ran || n.Load() != 0 || buf.Len() != 0 {
		t.Fatal("clean run misreported")
	}
}

func TestRunRecoversEveryPanicKind(t *testing.T) {
	g, buf, n := newGuard()
	values := []any{"boom", errors.New("err"), 42, nil}
	for _, v := range values[:3] {
		if !g.Run("handler", func() { panic(v) }) {
			t.Fatalf("panic(%v) not reported", v)
		}
	}
	var nilMap map[string]int
	if !g.Run("nilmap", func() { nilMap["x"] = 1 }) {
		t.Fatal("runtime panic not recovered")
	}
	if n.Load() != 4 {
		t.Fatalf("counted %d", n.Load())
	}
	out := buf.String()
	if !strings.Contains(out, "boom") || !strings.Contains(out, "stack") || !strings.Contains(out, "nilmap") {
		t.Fatalf("log missing details: %s", out)
	}
}

func TestNilCounterIsAllowed(t *testing.T) {
	g := NewGuard(slog.New(slog.DiscardHandler), nil)
	if !g.Run("x", func() { panic("y") }) {
		t.Fatal("not recovered")
	}
}

func TestGoUnderLoad(t *testing.T) {
	g, _, n := newGuard()
	var wg sync.WaitGroup
	var ok atomic.Int64
	for i := range 5000 {
		g.Go(&wg, "worker", func() {
			if i%2 == 0 {
				panic("half of them fail")
			}
			ok.Add(1)
		})
	}
	wg.Wait()
	if n.Load() != 2500 || ok.Load() != 2500 {
		t.Fatalf("panics %d ok %d", n.Load(), ok.Load())
	}
}
