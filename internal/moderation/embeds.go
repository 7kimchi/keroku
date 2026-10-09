package moderation

import (
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/embeds"
	"github.com/7kimchi/keroku/internal/validate"
)

// Title returns the modlog title for a case kind.
func Title(k cases.Kind) string {
	switch k {
	case cases.Ban:
		return "Member banned"
	case cases.Unban:
		return "Member unbanned"
	case cases.Kick:
		return "Member kicked"
	case cases.Timeout:
		return "Member timed out"
	case cases.Untimeout:
		return "Timeout removed"
	case cases.Warn:
		return "Member warned"
	}
	return "Note added"
}

// CaseEmbed renders a case for the modlog and for command replies.
func CaseEmbed(c cases.Case, botID string) *discordgo.MessageEmbed {
	mod := embeds.User(validate.FormatSnowflake(c.ModeratorID))
	if validate.FormatSnowflake(c.ModeratorID) == botID {
		mod += ", automatic"
	}
	b := embeds.New(Title(c.Kind)).
		Field("Member", embeds.User(validate.FormatSnowflake(c.TargetID)), true).
		Field("Moderator", mod, true).
		Field("Reason", reasonText(c.Reason), false)
	if c.Duration > 0 {
		b.Field("Duration", embeds.Duration(c.Duration), true).
			Field("Ends", embeds.Relative(c.CreatedAt.Add(c.Duration)), true)
	}
	return b.Footer("Case " + strconv.FormatInt(c.Number, 10)).Timestamp(c.CreatedAt).Build()
}

// dmEmbed is the notice sent to the member.
func dmEmbed(a Action, guildName string, now time.Time) *discordgo.MessageEmbed {
	b := embeds.New(dmTitle(a.Kind)+" "+embeds.Escape(guildName)).Field("Reason", reasonText(a.Reason), false)
	if a.Duration > 0 {
		b.Field("Duration", embeds.Duration(a.Duration), true).Field("Ends", embeds.Relative(now.Add(a.Duration)), true)
	}
	return b.Timestamp(now).Build()
}

func dmTitle(k cases.Kind) string {
	switch k {
	case cases.Ban:
		return "Banned from"
	case cases.Unban:
		return "Unbanned from"
	case cases.Kick:
		return "Kicked from"
	case cases.Timeout:
		return "Timed out in"
	case cases.Untimeout:
		return "Timeout removed in"
	}
	return "Warning in"
}

func reasonText(r string) string {
	if r == "" {
		return "No reason given."
	}
	return embeds.Escape(r)
}
