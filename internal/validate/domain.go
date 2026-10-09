package validate

import "strings"

// Domain checks a hostname for the link allowlist and returns it lowercased.
// Only ASCII is accepted, so internationalized names must be given in punycode.
func Domain(s string) (string, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.TrimSuffix(s, ".")
	if len(s) == 0 || len(s) > 253 || !strings.Contains(s, ".") {
		return "", ErrDomain
	}
	for _, label := range strings.Split(s, ".") {
		if !validLabel(label) {
			return "", ErrDomain
		}
	}
	return s, nil
}

func validLabel(l string) bool {
	if len(l) == 0 || len(l) > 63 || l[0] == '-' || l[len(l)-1] == '-' {
		return false
	}
	for i := 0; i < len(l); i++ {
		c := l[i]
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '-' {
			return false
		}
	}
	return true
}
