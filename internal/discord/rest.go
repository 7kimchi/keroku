package discord

import (
	"context"
	"net/url"
	"time"

	"github.com/bwmarrin/discordgo"
)

// MaxAuditReason is Discord's limit for the audit log reason header.
const MaxAuditReason = 512

// REST implements Client on top of a discordgo session.
type REST struct {
	s       *discordgo.Session
	timeout time.Duration
	retries int
	backoff time.Duration
}

// NewREST wraps a session. Each call gets timeout unless the caller's context ends sooner.
func NewREST(s *discordgo.Session, timeout time.Duration) *REST {
	return &REST{s: s, timeout: timeout, retries: 3, backoff: 250 * time.Millisecond}
}

// call runs fn with a bounded context. 429s are retried after the delay Discord asks for.
// Server errors and dropped connections are retried only when the request is idempotent.
func (r *REST) call(ctx context.Context, op string, idempotent bool, fn func(opts ...discordgo.RequestOption) error) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	for attempt := 0; ; attempt++ {
		err := fn(discordgo.WithContext(ctx), discordgo.WithRetryOnRatelimit(false), discordgo.WithRestRetries(0))
		if err == nil {
			return nil
		}
		e := classify(ctx, op, err)
		var wait time.Duration
		switch {
		case attempt >= r.retries:
			return e
		case e.Kind == RateLimited:
			wait = e.RetryAfter
		case e.Kind == Unavailable && idempotent:
			wait = r.backoff << attempt
		default:
			return e
		}
		if deadline, ok := ctx.Deadline(); ok && time.Until(deadline) < wait {
			return e
		}
		t := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			t.Stop()
			return &Error{Op: op, Kind: Timeout}
		case <-t.C:
		}
	}
}

// reasonOption encodes the audit log reason the way Discord expects it.
func reasonOption(reason string) discordgo.RequestOption {
	runes := []rune(reason)
	if len(runes) > MaxAuditReason {
		runes = runes[:MaxAuditReason]
	}
	return discordgo.WithAuditLogReason(url.PathEscape(string(runes)))
}

func with(opts []discordgo.RequestOption, reason string) []discordgo.RequestOption {
	if reason == "" {
		return opts
	}
	return append(opts, reasonOption(reason))
}
