package eventlog

import (
	"github.com/7kimchi/keroku/internal/embeds"
	"github.com/7kimchi/keroku/internal/validate"
)

// text cleans and caps message text before it is kept or shown.
func text(s string) string {
	clean, err := validate.Text(s, 4000)
	if err != nil {
		return ""
	}
	return embeds.Truncate(clean, maxText)
}

func show(s string) string {
	if s == "" {
		return "No text."
	}
	return embeds.Escape(s)
}

func beforeText(c cached, known bool) string {
	if !known {
		return "Not cached."
	}
	return show(c.text)
}

func userOrUnknown(id string) string {
	if id == "" {
		return "Unknown"
	}
	return embeds.User(id)
}

func jump(guildID, channelID, messageID string) string {
	return "https://discord.com/channels/" + guildID + "/" + channelID + "/" + messageID
}
