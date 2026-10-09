package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// GuildCounts returns the guild with its member count filled from the fake's members.
func (f *Fake) GuildCounts(ctx context.Context, guildID string) (*discordgo.Guild, error) {
	if err := f.enter(ctx, "guildCounts"); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	g, ok := f.guilds[guildID]
	if !ok {
		return nil, notFound("guildCounts", 10004)
	}
	cp := *g
	cp.Roles = append([]*discordgo.Role(nil), g.Roles...)
	cp.ApproximateMemberCount = len(f.members[guildID])
	return &cp, nil
}

// ServerCount returns how many guilds the fake knows.
func (f *Fake) ServerCount(ctx context.Context) (int, error) {
	if err := f.enter(ctx, "serverCount"); err != nil {
		return 0, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.guilds), nil
}
