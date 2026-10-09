package embeds

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestLength(t *testing.T) {
	for s, want := range map[string]int{"": 0, "abc": 3, "\U000000E9": 1, "\U00004E16": 1, "\U0001F600": 2, "a\U0001F600b": 4} {
		if got := Length(s); got != want {
			t.Fatalf("%q: got %d want %d", s, got, want)
		}
	}
}

func TestTruncate(t *testing.T) {
	cases := []struct {
		in    string
		limit int
		want  string
	}{
		{"hello", 5, "hello"},
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"hello", 0, ""},
		{"hello", -1, ""},
		{"hello", 2, ".."},
		{"hello", 3, "..."},
		{"hello", 4, "h..."},
		{"\U0001F600\U0001F600\U0001F600", 5, "\U0001F600..."},
		{"\U0001F600\U0001F600\U0001F600", 4, "..."},
		{`abc\def`, 7, `abc\def`},
		{`ab\cdefgh`, 6, `ab...`},
		{`ab\\cdefgh`, 7, `ab\\...`},
		{`a\\\bcdefg`, 7, `a\\...`},
		{"", 5, ""},
	}
	for _, tc := range cases {
		got := Truncate(tc.in, tc.limit)
		if got != tc.want {
			t.Errorf("Truncate(%q, %d) = %q want %q", tc.in, tc.limit, got, tc.want)
		}
	}
}

func FuzzTruncate(f *testing.F) {
	f.Add("hello world", 5)
	f.Add(strings.Repeat("\U0001F600", 10), 7)
	f.Fuzz(func(t *testing.T, s string, limit int) {
		if limit > 10000 {
			return
		}
		got := Truncate(s, limit)
		if Length(got) > max(limit, 0) {
			t.Fatalf("length %d over %d", Length(got), limit)
		}
		if utf8.ValidString(s) && !utf8.ValidString(got) {
			t.Fatalf("split a rune: %q", got)
		}
		if Length(s) <= limit && got != s {
			t.Fatalf("changed a string that fit")
		}
	})
}
