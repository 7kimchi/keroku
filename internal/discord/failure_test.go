package discord

import (
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestServerErrorsRetryOnlyIdempotent(t *testing.T) {
	a, rest := newServer(t, func(w http.ResponseWriter, _ *http.Request, _ int) { reply(w, 503, `{}`) })
	if err := rest.Kick(t.Context(), "1", "2", ""); !Is(err, Unavailable) || a.count() != 4 {
		t.Fatalf("kick: %v, %d hits", err, a.count())
	}
	b, rest2 := newServer(t, func(w http.ResponseWriter, _ *http.Request, _ int) { reply(w, 502, `{}`) })
	if err := rest2.Send(t.Context(), "1", testEmbed()); !Is(err, Unavailable) || b.count() != 1 {
		t.Fatalf("send: %v, %d hits", err, b.count())
	}
}

func TestServerErrorThenRecovery(t *testing.T) {
	_, rest := newServer(t, func(w http.ResponseWriter, r *http.Request, hit int) {
		if hit < 3 {
			reply(w, 500, `{}`)
			return
		}
		reply(w, 200, `{"id":"9","name":"g"}`)
	})
	g, err := rest.Guild(t.Context(), "9")
	if err != nil || g.ID != "9" {
		t.Fatalf("got %v %v", g, err)
	}
}

func TestSlowResponseTimesOut(t *testing.T) {
	_, rest := newServer(t, func(w http.ResponseWriter, r *http.Request, _ int) {
		select {
		case <-time.After(5 * time.Second):
		case <-r.Context().Done():
		}
	})
	rest.timeout = 300 * time.Millisecond
	start := time.Now()
	err := rest.Send(t.Context(), "1", testEmbed())
	if !Is(err, Timeout) || time.Since(start) > 2*time.Second {
		t.Fatalf("got %v after %v", err, time.Since(start))
	}
}

func TestDroppedConnection(t *testing.T) {
	a, rest := newServer(t, func(w http.ResponseWriter, _ *http.Request, _ int) {
		conn, _, err := w.(http.Hijacker).Hijack()
		if err == nil {
			_ = conn.Close()
		}
	})
	if err := rest.Kick(t.Context(), "1", "2", ""); !Is(err, Unavailable) || a.count() != 4 {
		t.Fatalf("got %v after %d hits", err, a.count())
	}
}

func TestConcurrentCalls(t *testing.T) {
	a, rest := newServer(t, func(w http.ResponseWriter, _ *http.Request, _ int) {
		reply(w, 200, `{"user":{"id":"2"},"roles":[]}`)
	})
	var wg sync.WaitGroup
	errs := make(chan error, 300)
	for range 300 {
		wg.Go(func() {
			_, err := rest.Member(t.Context(), "1", "2")
			errs <- err
		})
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if a.count() != 300 {
		t.Fatalf("%d hits", a.count())
	}
}
