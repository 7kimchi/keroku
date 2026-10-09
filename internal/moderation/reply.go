package moderation

import (
	"strconv"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/embeds"
)

// Reply renders a result for the moderator who ran the command.
func Reply(r Result, botID string) *discordgo.MessageEmbed {
	if r.Duplicate {
		detail := "This action was already handled."
		if r.Case.Number > 0 {
			detail = "Already handled in case " + strconv.FormatInt(r.Case.Number, 10) + "."
		}
		return embeds.New("No change").Description(detail).Build()
	}
	e := CaseEmbed(r.Case, botID)
	switch r.DM {
	case dmSent:
		e.Fields = append(e.Fields, &discordgo.MessageEmbedField{Name: "DM", Value: "Sent", Inline: true})
	case dmFailed:
		e.Fields = append(e.Fields, &discordgo.MessageEmbedField{Name: "DM", Value: "Not delivered", Inline: true})
	}
	if esc := r.Escalation; esc != nil && !esc.Duplicate {
		e.Fields = append(e.Fields, &discordgo.MessageEmbedField{
			Name: "Escalation", Value: Title(esc.Case.Kind) + ", case " + strconv.FormatInt(esc.Case.Number, 10) + ".",
		})
	}
	return e
}
