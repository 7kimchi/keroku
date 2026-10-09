package workers

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"go.uber.org/goleak"

	"github.com/7kimchi/keroku/internal/safe"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func newPool(t *testing.T, lanes, queue int) *Pool {
	t.Helper()
	p, err := New("test", lanes, queue, safe.NewGuard(slog.New(slog.DiscardHandler), nil))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = p.Close(context.Background()) })
	return p
}

func TestNewRejectsBadSizes(t *testing.T) {
	g := safe.NewGuard(slog.New(slog.DiscardHandler), nil)
	for _, tc := range [][2]int{{0, 1}, {1, 0}, {-1, 5}} {
		if _, err := New("x", tc[0], tc[1], g); err == nil {
			t.Fatalf("%v accepted", tc)
		}
	}
}

func TestSameKeyRunsInOrder(t *testing.T) {
	p := newPool(t, 8, 1000)
	var mu sync.Mutex
	var got []int
	for i := range 500 {
		if !p.Submit("guild", func(context.Context) {
			mu.Lock()
			got = append(got, i)
			mu.Unlock()
		}) {
			t.Fatal("submit refused")
		}
	}
	if err := p.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	for i, v := range got {
		if v != i {
			t.Fatalf("position %d ran job %d", i, v)
		}
	}
	if len(got) != 500 {
		t.Fatalf("ran %d", len(got))
	}
}

func TestFullLaneRefusesInsteadOfBlocking(t *testing.T) {
	p := newPool(t, 1, 2)
	release := make(chan struct{})
	started := make(chan struct{})
	p.Submit("k", func(context.Context) { close(started); <-release })
	<-started
	for range 2 {
		if !p.Submit("k", func(context.Context) {}) {
			t.Fatal("queue slot refused")
		}
	}
	start := time.Now()
	if p.Submit("k", func(context.Context) {}) {
		t.Fatal("over capacity accepted")
	}
	if time.Since(start) > 50*time.Millisecond {
		t.Fatal("submit blocked")
	}
	if p.Depth() != 2 {
		t.Fatalf("depth %d", p.Depth())
	}
	close(release)
}
