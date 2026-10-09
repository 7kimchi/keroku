package discord

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
)

func TestRateLimitHonorsRetryAfter(t *testing.T) {
	a, rest := newServer(t, func(w http.ResponseWriter, r *http.Request, hit int) {
		if hit == 1 {
			reply(w, 429, `{"message":"slow down","retry_after":0.3,"global":false}`)
			return
		}
		noContent(w, r, hit)
	})
	start := time.Now()
	if err := rest.Kick(t.Context(), "1", "2", ""); err != nil {
		t.Fatal(err)
	}
	if took := time.Since(start); took < 300*time.Millisecond {
		t.Fatalf("retried after %v, before retry_after", took)
	}
	if a.count() != 2 {
		t.Fatalf("%d hits", a.count())
	}
}

func TestRateLimitLongerThanDeadlineFailsFast(t *testing.T) {
	a, rest := newServer(t, func(w http.ResponseWriter, _ *http.Request, _ int) {
		reply(w, 429, `{"message":"slow down","retry_after":30,"global":false}`)
	})
	start := time.Now()
	err := rest.Send(t.Context(), "1", testEmbed())
	if !Is(err, RateLimited) || time.Since(start) > time.Second || a.count() != 1 {
		t.Fatalf("err %v after %v and %d hits", err, time.Since(start), a.count())
	}
	if e := err.(*Error); e.RetryAfter != 30*time.Second {
		t.Fatalf("retry after %v", e.RetryAfter)
	}
}

func TestRateLimitRetriesAreBounded(t *testing.T) {
	a, rest := newServer(t, func(w http.ResponseWriter, _ *http.Request, _ int) {
		reply(w, 429, `{"message":"slow down","retry_after":0.01,"global":false}`)
	})
	if err := rest.Kick(t.Context(), "1", "2", ""); !Is(err, RateLimited) {
		t.Fatalf("got %v", err)
	}
	if a.count() != rest.retries+1 {
		t.Fatalf("%d hits", a.count())
	}
}

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

func TestCallerCancelStopsRetries(t *testing.T) {
	_, rest := newServer(t, func(w http.ResponseWriter, _ *http.Request, _ int) {
		reply(w, 429, `{"message":"x","retry_after":0.5,"global":false}`)
	})
	ctx, cancel := context.WithTimeout(t.Context(), 700*time.Millisecond)
	defer cancel()
	start := time.Now()
	if err := rest.Kick(ctx, "1", "2", ""); err == nil || time.Since(start) > 1500*time.Millisecond {
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

func TestErrorsNeverCarrySecrets(t *testing.T) {
	_, rest := newServer(t, func(w http.ResponseWriter, _ *http.Request, _ int) {
		reply(w, 404, `{"code":10015,"message":"Unknown Webhook"}`)
	})
	i := &discordgo.Interaction{AppID: "1", ID: "2", Token: "interactionsecrettoken"}
	err := rest.EditResponse(t.Context(), i, testEmbed())
	if !Is(err, Expired) {
		t.Fatalf("got %v", err)
	}
	if s := err.Error(); strings.Contains(s, "interactionsecrettoken") || strings.Contains(s, testToken) || strings.Contains(s, "http") {
		t.Fatalf("leak: %s", s)
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
