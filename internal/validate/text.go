package validate

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// MaxReason is the longest reason accepted, matching the audit log limit.
const MaxReason = 512

// Text trims s, strips control and invisible formatting characters, then rejects it if
// it is still longer than maxRunes.
func Text(s string, maxRunes int) (string, error) {
	if !utf8.ValidString(s) {
		return "", ErrText
	}
	// Bound the work on huge input before allocating.
	if len(s) > maxRunes*4+1024 {
		return "", ErrTooLong
	}
	cleaned := strings.TrimSpace(strings.Map(keepRune, s))
	if utf8.RuneCountInString(cleaned) > maxRunes {
		return "", ErrTooLong
	}
	return cleaned, nil
}

// Reason validates a moderator supplied reason.
func Reason(s string) (string, error) {
	return Text(s, MaxReason)
}

func keepRune(r rune) rune {
	if r == '\n' || r == '\t' {
		return ' '
	}
	if unicode.IsControl(r) || isInvisible(r) {
		return -1
	}
	return r
}

// isInvisible covers zero width and bidi override characters used to spoof text.
func isInvisible(r rune) bool {
	switch {
	case r >= 0x200B && r <= 0x200F:
		return true
	case r >= 0x202A && r <= 0x202E:
		return true
	case r >= 0x2060 && r <= 0x206F:
		return true
	case r == 0xFEFF, r == 0x00AD, r == 0x180E:
		return true
	}
	return false
}
