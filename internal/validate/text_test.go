package validate

import (
	"errors"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

func TestTextCleans(t *testing.T) {
	for in, want := range map[string]string{
		"  spam  ":                   "spam",
		"":                           "",
		"line\nbreak\ttab":           "line break tab",
		"a\x00b\x07c\x1bd\x7f":       "abcd",
		"zero\U0000200Bwidth":            "zerowidth",
		"bidi\U0000202Eesrever":          "bidiesrever",
		"bom\U0000FEFF":                  "bom",
		"caf\U000000E9 \U00004E16\U0001F600": "caf\U000000E9 \U00004E16\U0001F600",
		"@everyone":                  "@everyone",
	} {
		got, err := Text(in, 100)
		if err != nil || got != want {
			t.Fatalf("%q: got %q %v", in, got, err)
		}
	}
}

func TestTextRejectsInvalidUTF8(t *testing.T) {
	if _, err := Text("ok\xffbad", 100); !errors.Is(err, ErrText) {
		t.Fatalf("got %v", err)
	}
}

func TestTextLengthBoundary(t *testing.T) {
	exact := strings.Repeat("\U00004E16", 512)
	if got, err := Reason(exact); err != nil || got != exact {
		t.Fatalf("512 runes rejected: %v", err)
	}
	if _, err := Reason(exact + "x"); !errors.Is(err, ErrTooLong) {
		t.Fatalf("513 runes: got %v", err)
	}
	// Stripped characters do not count toward the limit.
	padded := strings.Repeat("\U0000200B", 600) + "short"
	if got, err := Reason(padded); err != nil || got != "short" {
		t.Fatalf("got %q %v", got, err)
	}
}

func TestTextHugeInputIsCheap(t *testing.T) {
	huge := strings.Repeat("a", 10<<20)
	if _, err := Reason(huge); !errors.Is(err, ErrTooLong) {
		t.Fatalf("got %v", err)
	}
}

func TestTextOnlyWhitespace(t *testing.T) {
	got, err := Text(" \n\t\U0000200B ", 10)
	if err != nil || got != "" {
		t.Fatalf("got %q %v", got, err)
	}
}

func FuzzText(f *testing.F) {
	for _, s := range []string{"", "hi", "a\x00b", "\U0000202E", "\xff", strings.Repeat("x", 600)} {
		f.Add(s, 50)
	}
	f.Fuzz(func(t *testing.T, s string, limit int) {
		if limit < 0 || limit > 5000 {
			return
		}
		got, err := Text(s, limit)
		if err != nil {
			return
		}
		if !utf8.ValidString(got) || utf8.RuneCountInString(got) > limit {
			t.Fatalf("bad output %q for limit %d", got, limit)
		}
		if got != strings.TrimSpace(got) {
			t.Fatalf("untrimmed %q", got)
		}
		for _, r := range got {
			if unicode.IsControl(r) || isInvisible(r) {
				t.Fatalf("control rune %U survived", r)
			}
		}
	})
}
