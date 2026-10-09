package validate

import (
	"errors"
	"strings"
	"testing"
)

func TestDomainAccepts(t *testing.T) {
	for in, want := range map[string]string{
		"example.com":       "example.com",
		" YouTube.COM ":     "youtube.com",
		"a.b.c.d.e":         "a.b.c.d.e",
		"xn--bcher-kva.de":  "xn--bcher-kva.de",
		"my-site.co.uk":     "my-site.co.uk",
		"example.com.":      "example.com",
		"1.example":         "1.example",
		"github.io":         "github.io",
		"docs.example.com":  "docs.example.com",
		"x.y":               "x.y",
		"a-b-c.example.org": "a-b-c.example.org",
	} {
		got, err := Domain(in)
		if err != nil || got != want {
			t.Fatalf("%q: got %q %v", in, got, err)
		}
	}
}

func TestDomainRejects(t *testing.T) {
	for _, in := range []string{
		"", "localhost", ".com", "example..com", "-a.com", "a-.com", "exa mple.com",
		"http://example.com", "example.com/path", "ex@mple.com", "b\U000000FCcher.de",
		"*.example.com", "example.com:443", strings.Repeat("a", 64) + ".com",
		strings.Repeat("a.", 127) + "com", "a_b.com",
	} {
		if _, err := Domain(in); !errors.Is(err, ErrDomain) {
			t.Fatalf("%q: got %v", in, err)
		}
	}
}

func TestIntBounds(t *testing.T) {
	if Int(5, 1, 10) != nil || Int(1, 1, 10) != nil || Int(10, 1, 10) != nil {
		t.Fatal("in range rejected")
	}
	for _, v := range []int64{0, 11, -1 << 63, 1<<63 - 1} {
		if !errors.Is(Int(v, 1, 10), ErrRange) {
			t.Fatalf("%d accepted", v)
		}
	}
}

func FuzzDomain(f *testing.F) {
	for _, s := range []string{"example.com", "a..b", "x"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		got, err := Domain(s)
		if err != nil {
			return
		}
		again, err := Domain(got)
		if err != nil || again != got {
			t.Fatalf("not idempotent: %q gave %q then %q %v", s, got, again, err)
		}
	})
}
