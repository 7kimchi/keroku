package validate

import (
	"errors"
	"testing"
	"time"
)

const day = 24 * time.Hour

func TestDurationAccepts(t *testing.T) {
	for in, want := range map[string]time.Duration{
		"30s":          30 * time.Second,
		"1m":           time.Minute,
		"1d4h":         day + 4*time.Hour,
		"1W2D3H4M5S":   7*day + 2*day + 3*time.Hour + 4*time.Minute + 5*time.Second,
		" 2h ":         2 * time.Hour,
		"90m":          90 * time.Minute,
		"28d":          28 * day,
		"0001h":        time.Hour,
		"4w":           28 * day,
		"1w0d0h0m1s":   7*day + time.Second,
		"2419200s":     28 * day,
		"40320m":       28 * day,
		"672h":         28 * day,
		"1h1s":         time.Hour + time.Second,
		"10d":          10 * day,
		"3m3s":         3*time.Minute + 3*time.Second,
		"5d23h59m59s":  6*day - time.Second,
		"0w0d0h0m60s":  time.Minute,
		"1w6d23h59m":   14*day - time.Minute,
		"0h1m":         time.Minute,
		"1d0s":         day,
		"0d1s":         time.Second,
		"00000000001s": time.Second,
	} {
		got, err := Duration(in, time.Second, 28*day)
		if err != nil || got != want {
			t.Fatalf("%q: got %v %v", in, got, err)
		}
	}
}

func TestDurationRejectsSyntax(t *testing.T) {
	for _, in := range []string{
		"", " ", "h", "1", "1x", "-1h", "+1h", "1.5h", "1h1h", "1m1h", "1s1d", "1 h", "1h 2m",
		"1hh", "one hour", "\U0000FF11h", "1h\x00", "999999999999999999999999999999s", "1y", "1ms",
	} {
		_, err := Duration(in, time.Second, 28*day)
		if !errors.Is(err, ErrDuration) && !errors.Is(err, ErrRange) {
			t.Fatalf("%q: got %v", in, err)
		}
	}
}

func TestDurationBounds(t *testing.T) {
	if _, err := Duration("28d1s", time.Second, 28*day); !errors.Is(err, ErrRange) {
		t.Fatalf("over max: %v", err)
	}
	if _, err := Duration("0s", time.Second, 28*day); !errors.Is(err, ErrRange) {
		t.Fatalf("under min: %v", err)
	}
	if _, err := Duration("9999999w", time.Second, 365*day); !errors.Is(err, ErrRange) {
		t.Fatalf("overflow candidate: %v", err)
	}
	if _, err := Duration("1s", 0, -1); !errors.Is(err, ErrDuration) {
		t.Fatalf("negative max: %v", err)
	}
}

func FuzzDuration(f *testing.F) {
	for _, s := range []string{"1d", "1w2d3h4m5s", "99999w", "-1s", "1h1h"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		d, err := Duration(s, time.Second, 365*day)
		if err == nil && (d < time.Second || d > 365*day) {
			t.Fatalf("%q gave out of range %v", s, d)
		}
	})
}
