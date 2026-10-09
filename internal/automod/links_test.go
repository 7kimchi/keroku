package automod

import "testing"

func TestLinksAllowlist(t *testing.T) {
	c, p := cfg(), linkPattern()
	for content, want := range map[string]bool{
		"see https://example.com/x":                  false,
		"see https://docs.example.com/x":             false,
		"HTTPS://WWW.YOUTUBE.COM/watch?v=1":          false,
		"https://example.com.":                       false,
		"no links here, just example.com text":       false,
		"https://evil.com":                           true,
		"https://evilexample.com":                    true,
		"https://example.com.evil.com":               true,
		"https://example.com@evil.com":               true,
		"https://user:pass@evil.com/example.com":     true,
		"[example.com](https://evil.com)":            true,
		"<https://evil.com>":                         true,
		"ok https://example.com then https://bad.io": true,
		"https://xn--exmple-cua.com":                 true,
		"http://127.0.0.1/":                          true,
		"https:// is not a link":                     false,
	} {
		if got := links(c, content, p); got != want {
			t.Errorf("%q: got %v want %v", content, got, want)
		}
	}
	c.LinksEnabled = false
	if links(c, "https://evil.com", p) {
		t.Fatal("disabled rule fired")
	}
}

func FuzzLinks(f *testing.F) {
	f.Add("https://example.com https://evil.com")
	p, c := linkPattern(), cfg()
	f.Fuzz(func(t *testing.T, s string) {
		_ = links(c, s, p)
	})
}
