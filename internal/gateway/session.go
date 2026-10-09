package gateway

import (
	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/safe"
)

// newSession builds one shard's session. The state cache is off so memory stays bounded,
// and events are delivered synchronously so nothing spawns a goroutine per event.
func newSession(token string, shard, count int, intents discordgo.Intent, r Router, g *safe.Guard) (*discordgo.Session, error) {
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}
	s.ShardID, s.ShardCount = shard, count
	s.Identify.Intents = intents
	s.StateEnabled = false
	s.State.MaxMessageCount = 0
	s.SyncEvents = true
	s.LogLevel = discordgo.LogWarning
	s.ShouldReconnectOnError = true
	register(s, r, g)
	return s, nil
}

// register adds one guarded handler per event type. A panic drops that event only.
func register(s *discordgo.Session, r Router, g *safe.Guard) {
	s.AddHandler(func(_ *discordgo.Session, e *discordgo.InteractionCreate) {
		g.Run("interactionCreate", func() { r.Interaction(e.Interaction) })
	})
	s.AddHandler(func(_ *discordgo.Session, e *discordgo.MessageCreate) {
		g.Run("messageCreate", func() { r.MessageCreate(e.Message) })
	})
	s.AddHandler(func(_ *discordgo.Session, e *discordgo.MessageUpdate) {
		g.Run("messageUpdate", func() { r.MessageUpdate(e) })
	})
	s.AddHandler(func(_ *discordgo.Session, e *discordgo.MessageDelete) {
		g.Run("messageDelete", func() { r.MessageDelete(e) })
	})
	s.AddHandler(func(_ *discordgo.Session, e *discordgo.GuildMemberAdd) {
		g.Run("memberAdd", func() { r.MemberAdd(e) })
	})
	s.AddHandler(func(_ *discordgo.Session, e *discordgo.GuildMemberRemove) {
		g.Run("memberRemove", func() { r.MemberRemove(e) })
	})
}
