package gateway

import "github.com/bwmarrin/discordgo"

// Router receives gateway events. Every method runs on the websocket reader and must not
// block: hand work to a bounded queue and return.
type Router interface {
	Interaction(i *discordgo.Interaction)
	MessageCreate(m *discordgo.Message)
	MessageUpdate(m *discordgo.MessageUpdate)
	MessageDelete(m *discordgo.MessageDelete)
	MemberAdd(m *discordgo.GuildMemberAdd)
	MemberRemove(m *discordgo.GuildMemberRemove)
}
