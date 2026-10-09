package records

import (
	"strconv"
	"strings"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/embeds"
	"github.com/7kimchi/keroku/internal/moderation"
)

const pageSize = 10

// line renders one case as a list row, like "Case 12, Member banned, 2 hours ago: spam".
func line(c cases.Case) string {
	reason := c.Reason
	if reason == "" {
		reason = "No reason given."
	}
	return "Case " + strconv.FormatInt(c.Number, 10) + ", " + moderation.Title(c.Kind) + ", " +
		embeds.Relative(c.CreatedAt) + ": " + embeds.Truncate(embeds.Escape(reason), 200)
}

func lines(list []cases.Case) string {
	out := make([]string, 0, len(list))
	for _, c := range list {
		out = append(out, line(c))
	}
	return strings.Join(out, "\n")
}

// count renders "1 case" or "3 cases".
func count(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return strconv.Itoa(n) + " " + noun + "s"
}
