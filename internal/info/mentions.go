package info

import (
	"strconv"
	"strings"

	"github.com/7kimchi/keroku/internal/validate"
)

// room for " and 99999999 more"
const moreRoom = 24

// mentionList renders ids as mentions like "<@&id>", within limit characters, ending with
// a count of what did not fit. Bad ids are skipped. Mentions in embeds never ping.
func mentionList(prefix string, ids []string, limit int) string {
	var b strings.Builder
	shown, valid, full := 0, 0, false
	for _, id := range ids {
		if _, err := validate.Snowflake(id); err != nil {
			continue
		}
		valid++
		m := "<" + prefix + id + ">"
		if full = full || b.Len()+2+len(m)+moreRoom > limit; full {
			continue
		}
		if shown > 0 {
			b.WriteString(", ")
		}
		b.WriteString(m)
		shown++
	}
	if valid == 0 {
		return "None"
	}
	if rest := valid - shown; rest > 0 {
		if shown > 0 {
			b.WriteString(" and ")
		}
		b.WriteString(strconv.Itoa(rest) + " more")
	}
	return b.String()
}
