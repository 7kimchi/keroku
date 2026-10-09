package discord

import (
	"errors"
	"strconv"
	"time"
)

// Kind groups API failures by what the caller should do about them.
type Kind int

// Failure kinds.
const (
	Unknown Kind = iota
	NotFound
	Forbidden
	RateLimited
	Unavailable
	BadRequest
	Expired
	Unauthorized
	Timeout
)

// String names the kind for logs and metrics.
func (k Kind) String() string {
	switch k {
	case NotFound:
		return "notFound"
	case Forbidden:
		return "forbidden"
	case RateLimited:
		return "rateLimited"
	case Unavailable:
		return "unavailable"
	case BadRequest:
		return "badRequest"
	case Expired:
		return "expired"
	case Unauthorized:
		return "unauthorized"
	case Timeout:
		return "timeout"
	}
	return "unknown"
}

// Error is a classified API failure. It never carries URLs or headers.
type Error struct {
	Op         string
	Kind       Kind
	Status     int
	Code       int
	RetryAfter time.Duration
}

func (e *Error) Error() string {
	s := "discord " + e.Op + ": " + e.Kind.String()
	if e.Status != 0 {
		s += " status " + strconv.Itoa(e.Status)
	}
	if e.Code != 0 {
		s += " code " + strconv.Itoa(e.Code)
	}
	return s
}

// KindOf returns the kind of a classified error, or Unknown.
func KindOf(err error) Kind {
	var e *Error
	if errors.As(err, &e) {
		return e.Kind
	}
	return Unknown
}

// Is reports whether err is a classified error of kind k.
func Is(err error, k Kind) bool {
	return err != nil && KindOf(err) == k
}
