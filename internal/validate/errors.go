// Package validate parses and checks untrusted input. It rejects bad input instead of fixing it.
package validate

import "errors"

// Sentinel errors so callers can map failures to short user messages.
var (
	ErrSnowflake = errors.New("invalid id")
	ErrText      = errors.New("invalid text")
	ErrTooLong   = errors.New("text too long")
	ErrDuration  = errors.New("invalid duration")
	ErrRange     = errors.New("value out of range")
	ErrDomain    = errors.New("invalid domain")
)
