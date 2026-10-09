package embeds

import "strings"

// Escape neutralizes markdown, masked links and mention syntax in user supplied text.
func Escape(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 8)
	for _, r := range s {
		switch r {
		case '\\', '*', '_', '~', '`', '|', '>', '<', '[', ']', '(', ')', '#', '-':
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}
