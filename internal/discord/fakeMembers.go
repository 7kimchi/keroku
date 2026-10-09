package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// Guild returns a copy of the guild and its roles.
func (f *Fake) Guild(ctx context.Context, guildID string) (*discordgo.Guild, error) {
	if err := f.enter(ctx, "guild"); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	g, ok := f.guilds[guildID]
	if !ok {
		return nil, notFound("guild", 10004)
	}
	cp := *g
	cp.Roles = make([]*discordgo.Role, len(g.Roles))
	for i, r := range g.Roles {
		rc := *r
		cp.Roles[i] = &rc
	}
	return &cp, nil
}

// Member returns a copy of the member.
func (f *Fake) Member(ctx context.Context, guildID, userID string) (*discordgo.Member, error) {
	if err := f.enter(ctx, "member"); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	m, ok := f.members[guildID][userID]
	if !ok {
		return nil, notFound("member", 10007)
	}
	cp := *m
	cp.Roles = append([]string(nil), m.Roles...)
	user := *m.User
	cp.User = &user
	return &cp, nil
}

// Ban bans and removes the member. Banning twice succeeds, like Discord.
func (f *Fake) Ban(ctx context.Context, guildID, userID string, deleteSeconds int, _ string) error {
	if err := f.enter(ctx, "ban"); err != nil {
		return err
	}
	if deleteSeconds < 0 || deleteSeconds > 604800 {
		return &Error{Op: "ban", Kind: BadRequest, Status: 400}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.guilds[guildID]; !ok {
		return notFound("ban", 10004)
	}
	f.bans[guildID][userID] = true
	delete(f.members[guildID], userID)
	return nil
}

// Unban lifts a ban. Unbanning someone not banned is NotFound.
func (f *Fake) Unban(ctx context.Context, guildID, userID, _ string) error {
	if err := f.enter(ctx, "unban"); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.bans[guildID][userID] {
		return notFound("unban", 10026)
	}
	delete(f.bans[guildID], userID)
	return nil
}

// IsBanned reports the ban state.
func (f *Fake) IsBanned(ctx context.Context, guildID, userID string) (bool, error) {
	if err := f.enter(ctx, "getBan"); err != nil {
		return false, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.bans[guildID][userID], nil
}
