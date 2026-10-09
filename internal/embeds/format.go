package embeds

import (
	"strconv"
	"time"
)

// Duration renders d by its largest unit plus the next unit down when that is non zero,
// like "1d 4h" or "30m". Smaller remainders are dropped.
func Duration(d time.Duration) string {
	if d < time.Second {
		return "0s"
	}
	units := []struct {
		suffix string
		size   time.Duration
	}{{"d", 24 * time.Hour}, {"h", time.Hour}, {"m", time.Minute}, {"s", time.Second}}
	out := ""
	parts := 0
	for _, u := range units {
		n := d / u.size
		if n == 0 {
			if parts > 0 {
				break
			}
			continue
		}
		if out != "" {
			out += " "
		}
		out += strconv.FormatInt(int64(n), 10) + u.suffix
		d -= n * u.size
		if parts++; parts == 2 {
			break
		}
	}
	return out
}

// Relative renders t as a Discord relative timestamp.
func Relative(t time.Time) string {
	return "<t:" + strconv.FormatInt(t.Unix(), 10) + ":R>"
}

// User renders a user mention followed by the raw id. Mentions in embeds never ping.
func User(id string) string {
	return "<@" + id + "> (" + id + ")"
}

// Channel renders a channel mention.
func Channel(id string) string {
	return "<#" + id + ">"
}
