package logging

import (
	"strings"
	"testing"
)

func TestRedactorIgnoresShortSecrets(t *testing.T) {
	r := NewRedactor("ab", "")
	if got := r.String("abc"); got != "abc" {
		t.Fatalf("short secret redacted everything: %q", got)
	}
}

func TestSensitiveKey(t *testing.T) {
	for k, want := range map[string]bool{
		"token": true, "botToken": true, "PASSWORD": true, "clientSecret": true,
		"authorization": true, "databaseUrl": true, "guildId": false, "reason": false,
	} {
		if sensitiveKey(k) != want {
			t.Fatalf("%s: want %v", k, want)
		}
	}
}

func FuzzRedactor(f *testing.F) {
	f.Add("prefix " + secret + " suffix")
	f.Fuzz(func(t *testing.T, s string) {
		out := NewRedactor(secret).String(s + secret)
		if strings.Contains(out, secret) {
			t.Fatalf("secret survived in %q", out)
		}
	})
}
