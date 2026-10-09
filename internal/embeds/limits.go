// Package embeds builds Discord embeds that always fit Discord's limits.
package embeds

import (
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// Discord embed limits. Lengths are measured in UTF-16 units, which never undercounts.
const (
	MaxTitle       = 256
	MaxDescription = 4096
	MaxFields      = 25
	MaxFieldName   = 256
	MaxFieldValue  = 1024
	MaxFooter      = 2048
	MaxAuthor      = 256
	MaxTotal       = 6000
)

const ellipsis = "..."

// Length returns s's length as Discord counts it.
func Length(s string) int {
	n := 0
	for _, r := range s {
		n += len(utf16.Encode([]rune{r}))
	}
	return n
}

// Truncate shortens s to at most limit units, marking the cut with "...".
func Truncate(s string, limit int) string {
	if limit <= 0 {
		return ""
	}
	if Length(s) <= limit {
		return s
	}
	if limit <= len(ellipsis) {
		return ellipsis[:limit]
	}
	budget := limit - len(ellipsis)
	used, cut := 0, 0
	for cut < len(s) {
		r, size := utf8.DecodeRuneInString(s[cut:])
		w := 1
		if r > 0xFFFF {
			w = 2
		}
		if used+w > budget {
			break
		}
		used += w
		cut += size
	}
	out := s[:cut]
	// An odd run of trailing backslashes would escape the marker, so drop one.
	if trailing := len(out) - len(strings.TrimRight(out, `\`)); trailing%2 == 1 {
		out = out[:len(out)-1]
	}
	return out + ellipsis
}
