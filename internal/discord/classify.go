package discord

import (
	"context"
	"errors"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

// classify turns any error from discordgo into an *Error, dropping request details.
func classify(ctx context.Context, op string, err error) *Error {
	var rl *discordgo.RateLimitError
	if errors.As(err, &rl) && rl.RateLimit != nil && rl.TooManyRequests != nil {
		return &Error{Op: op, Kind: RateLimited, Status: http.StatusTooManyRequests, RetryAfter: rl.RetryAfter}
	}
	var rest *discordgo.RESTError
	if errors.As(err, &rest) {
		e := &Error{Op: op}
		if rest.Response != nil {
			e.Status = rest.Response.StatusCode
		}
		if rest.Message != nil {
			e.Code = rest.Message.Code
		}
		e.Kind = kindFor(e.Status, e.Code)
		return e
	}
	if errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
		return &Error{Op: op, Kind: Timeout}
	}
	if errors.Is(err, discordgo.ErrUnauthorized) {
		return &Error{Op: op, Kind: Unauthorized, Status: http.StatusUnauthorized}
	}
	// Network failures, broken connections and bad bodies.
	return &Error{Op: op, Kind: Unavailable}
}

func kindFor(status, code int) Kind {
	switch code {
	case 10003, 10004, 10007, 10008, 10011, 10013, 10026, 10066:
		return NotFound
	case 10015, 10062, 40060:
		return Expired
	case 50001, 50007, 50013:
		return Forbidden
	}
	switch {
	case status == http.StatusNotFound:
		return NotFound
	case status == http.StatusUnauthorized:
		return Unauthorized
	case status == http.StatusForbidden:
		return Forbidden
	case status == http.StatusTooManyRequests:
		return RateLimited
	case status >= 500:
		return Unavailable
	case status >= 400:
		return BadRequest
	}
	return Unknown
}
