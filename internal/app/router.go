package app

import "github.com/bwmarrin/discordgo"

// interactionSink is the dispatcher's entry point.
type interactionSink interface {
	Interaction(i *discordgo.Interaction)
}

// router fans gateway events out to the features that use them. Every hook is non blocking.
type router struct {
	interactions interactionSink
	messages     []func(*discordgo.Message)
	updates      []func(*discordgo.MessageUpdate)
	deletes      []func(*discordgo.MessageDelete)
	joins        []func(*discordgo.GuildMemberAdd)
	leaves       []func(*discordgo.GuildMemberRemove)
}

func (r *router) Interaction(i *discordgo.Interaction) { r.interactions.Interaction(i) }

func (r *router) MessageCreate(m *discordgo.Message) {
	if m == nil || m.GuildID == "" || m.Author == nil || m.Author.Bot {
		return
	}
	for _, fn := range r.messages {
		fn(m)
	}
}

func (r *router) MessageUpdate(m *discordgo.MessageUpdate) {
	if m == nil || m.Message == nil || m.GuildID == "" {
		return
	}
	for _, fn := range r.updates {
		fn(m)
	}
}

func (r *router) MessageDelete(m *discordgo.MessageDelete) {
	if m == nil || m.Message == nil || m.GuildID == "" {
		return
	}
	for _, fn := range r.deletes {
		fn(m)
	}
}

func (r *router) MemberAdd(m *discordgo.GuildMemberAdd) {
	if m == nil || m.Member == nil || m.User == nil {
		return
	}
	for _, fn := range r.joins {
		fn(m)
	}
}

func (r *router) MemberRemove(m *discordgo.GuildMemberRemove) {
	if m == nil || m.Member == nil || m.User == nil {
		return
	}
	for _, fn := range r.leaves {
		fn(m)
	}
}
