package commands

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/embeds"
)

// NewRef returns a short random correlation id.
func NewRef() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func errorEmbed(title, detail string) *discordgo.MessageEmbed {
	return embeds.New(title).Description(detail).Build()
}

func internalError(ref string) *discordgo.MessageEmbed {
	return errorEmbed("Command failed", "Unexpected error. Ref "+ref+".")
}

func deferred() *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral},
	}
}

func immediate(e *discordgo.MessageEmbed) *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral, Embeds: []*discordgo.MessageEmbed{e}},
	}
}
