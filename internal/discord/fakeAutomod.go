package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// AutoModRules returns copies of the guild's native rules.
func (f *Fake) AutoModRules(ctx context.Context, guildID string) ([]*discordgo.AutoModerationRule, error) {
	if err := f.enter(ctx, "autoModRules"); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]*discordgo.AutoModerationRule, 0, len(f.rules[guildID]))
	for _, r := range f.rules[guildID] {
		cp := *r
		out = append(out, &cp)
	}
	return out, nil
}

// CreateAutoModRule stores a rule with a fresh id.
func (f *Fake) CreateAutoModRule(ctx context.Context, guildID string, rule *discordgo.AutoModerationRule, _ string) error {
	if err := f.enter(ctx, "createAutoModRule"); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := *rule
	cp.ID, cp.GuildID, cp.CreatorID = f.id(), guildID, f.botID
	f.rules[guildID] = append(f.rules[guildID], &cp)
	return nil
}

// EditAutoModRule replaces a stored rule.
func (f *Fake) EditAutoModRule(ctx context.Context, guildID, ruleID string, rule *discordgo.AutoModerationRule, _ string) error {
	if err := f.enter(ctx, "editAutoModRule"); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for i, r := range f.rules[guildID] {
		if r.ID == ruleID {
			cp := *rule
			cp.ID, cp.GuildID, cp.CreatorID = ruleID, guildID, r.CreatorID
			f.rules[guildID][i] = &cp
			return nil
		}
	}
	return notFound("editAutoModRule", 10066)
}

// DeleteAutoModRule removes a stored rule.
func (f *Fake) DeleteAutoModRule(ctx context.Context, guildID, ruleID, _ string) error {
	if err := f.enter(ctx, "deleteAutoModRule"); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	rules := f.rules[guildID]
	for i, r := range rules {
		if r.ID == ruleID {
			f.rules[guildID] = append(rules[:i:i], rules[i+1:]...)
			return nil
		}
	}
	return notFound("deleteAutoModRule", 10066)
}
