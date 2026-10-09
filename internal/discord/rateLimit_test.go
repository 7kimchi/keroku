package discord

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
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
	var e *Error
	if !errors.As(err, &e) || e.RetryAfter != 30*time.Second {
		t.Fatalf("retry after %v", err)
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
