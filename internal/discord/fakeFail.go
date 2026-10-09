package discord

import (
	"context"
	"time"
)

// FailNext makes the next n calls to op return err. op "*" matches every operation.
func (f *Fake) FailNext(op string, err error, n int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for range n {
		f.failures[op] = append(f.failures[op], err)
	}
}

// SetDelay makes every call to op wait d first, honoring the caller's context.
func (f *Fake) SetDelay(op string, d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.delay[op] = d
}

// enter counts the call, applies delay and returns an injected failure, if any.
// The lock is not held on return.
func (f *Fake) enter(ctx context.Context, op string) error {
	f.mu.Lock()
	f.calls[op]++
	d := f.delay[op] + f.delay["*"]
	var err error
	for _, key := range []string{op, "*"} {
		if q := f.failures[key]; len(q) > 0 && err == nil {
			err, f.failures[key] = q[0], q[1:]
		}
	}
	hook := f.hooks[op]
	f.mu.Unlock()
	if hook != nil {
		hook()
	}
	if d > 0 {
		t := time.NewTimer(d)
		defer t.Stop()
		select {
		case <-ctx.Done():
			return &Error{Op: op, Kind: Timeout}
		case <-t.C:
		}
	}
	if err == nil && ctx.Err() != nil {
		return &Error{Op: op, Kind: Timeout}
	}
	return err
}

func notFound(op string, code int) *Error {
	return &Error{Op: op, Kind: NotFound, Status: 404, Code: code}
}

// OnCall runs fn at the start of every call to op, before any delay or failure.
func (f *Fake) OnCall(op string, fn func()) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.hooks[op] = fn
}

// SetNow replaces the clock the fake uses for Discord's time limits.
func (f *Fake) SetNow(now func() time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.now = now
}
