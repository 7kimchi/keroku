package metrics

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/safe"
)

func guard() *safe.Guard { return safe.NewGuard(slog.New(slog.DiscardHandler), nil) }

func TestServeRefusesNonLoopback(t *testing.T) {
	for _, addr := range []string{"0.0.0.0:0", ":0", "[::]:0", "localhost:0", "192.168.1.1:0", "bad"} {
		if err := New().Serve(t.Context(), addr, guard()); err == nil {
			t.Fatalf("%s accepted", addr)
		}
	}
}

func TestServeAndShutdown(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	m := New()
	m.Raids.Inc()
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- m.serve(ctx, ln, guard()) }()
	base := "http://" + ln.Addr().String()

	body := get(t, base+"/metrics", http.StatusOK)
	if !strings.Contains(body, "keroku_raids_total 1") {
		t.Fatalf("metric missing: %.300s", body)
	}
	get(t, base+"/", http.StatusNotFound)
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodPost, base+"/metrics", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("POST got %d", resp.StatusCode)
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("shutdown error %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("server did not stop")
	}
	http.DefaultClient.CloseIdleConnections()
}

func TestServePortInUse(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	if err := New().Serve(t.Context(), ln.Addr().String(), guard()); err == nil {
		t.Fatal("bound a port already in use")
	}
}

func get(t *testing.T, url string, want int) string {
	t.Helper()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != want {
		t.Fatalf("%s: got %d", url, resp.StatusCode)
	}
	return string(b)
}
