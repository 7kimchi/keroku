package logging

import (
	"regexp"
	"strings"
)

const redacted = "[redacted]"

// Redactor scrubs known secrets and anything shaped like a Discord token.
type Redactor struct {
	secrets []string
	token   *regexp.Regexp
}

// NewRedactor builds a redactor for the given secret values. Empty values are ignored.
func NewRedactor(secrets ...string) *Redactor {
	r := &Redactor{token: regexp.MustCompile(`[A-Za-z0-9_-]{20,}\.[A-Za-z0-9_-]{5,}\.[A-Za-z0-9_-]{20,}`)}
	for _, s := range secrets {
		if len(s) >= 4 {
			r.secrets = append(r.secrets, s)
		}
	}
	return r
}

// String returns s with every secret replaced.
func (r *Redactor) String(s string) string {
	for _, secret := range r.secrets {
		s = strings.ReplaceAll(s, secret, redacted)
	}
	return r.token.ReplaceAllLiteralString(s, redacted)
}

// sensitiveKey reports keys whose values are never logged.
func sensitiveKey(k string) bool {
	k = strings.ToLower(k)
	for _, word := range []string{"token", "password", "secret", "authorization", "databaseurl"} {
		if strings.Contains(k, word) {
			return true
		}
	}
	return false
}
