package automod

import (
	"context"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/discord"
	"github.com/7kimchi/keroku/internal/store"
)

// Names of the native rules Keroku owns. Rules with other names are never touched.
const (
	mentionRuleName = "Keroku mention spam"
	inviteRuleName  = "Keroku invite links"
)

// discordgo has no constant for the mention spam trigger.
const mentionSpamTrigger discordgo.AutoModerationRuleTriggerType = 5

// invitePattern is Discord's Rust regex flavor. It matches invite links with or without a scheme.
const invitePattern = `(?i)(discord\.gg|discord(app)?\.com/invite|dsc\.gg)/[a-z0-9-]+`

// SyncNative creates, updates or removes Keroku's native rules to match cfg.
func SyncNative(ctx context.Context, client discord.AutoMod, guildID, botID string, cfg store.AutomodSettings) error {
	rules, err := client.AutoModRules(ctx, guildID)
	if err != nil {
		return err
	}
	owned := map[string]string{}
	for _, r := range rules {
		if r.CreatorID == botID && strings.HasPrefix(r.Name, "Keroku ") {
			owned[r.Name] = r.ID
		}
	}
	var mention, invite *discordgo.AutoModerationRule
	if cfg.MentionLimit > 0 {
		mention = rule(mentionRuleName, mentionSpamTrigger, &discordgo.AutoModerationTriggerMetadata{MentionTotalLimit: cfg.MentionLimit}, cfg)
	}
	if cfg.InvitesBlocked {
		invite = rule(inviteRuleName, discordgo.AutoModerationEventTriggerKeyword,
			&discordgo.AutoModerationTriggerMetadata{RegexPatterns: []string{invitePattern}}, cfg)
	}
	for name, want := range map[string]*discordgo.AutoModerationRule{mentionRuleName: mention, inviteRuleName: invite} {
		id, exists := owned[name]
		switch {
		case want == nil && exists:
			err = client.DeleteAutoModRule(ctx, guildID, id, "Automod setting turned off.")
		case want != nil && exists:
			err = client.EditAutoModRule(ctx, guildID, id, want, "Automod setting changed.")
		case want != nil:
			err = client.CreateAutoModRule(ctx, guildID, want, "Automod setting turned on.")
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func rule(name string, trigger discordgo.AutoModerationRuleTriggerType, meta *discordgo.AutoModerationTriggerMetadata, cfg store.AutomodSettings) *discordgo.AutoModerationRule {
	enabled := true
	actions := []discordgo.AutoModerationAction{{Type: discordgo.AutoModerationRuleActionBlockMessage}}
	if cfg.Timeout > 0 {
		actions = append(actions, discordgo.AutoModerationAction{Type: discordgo.AutoModerationRuleActionTimeout,
			Metadata: &discordgo.AutoModerationActionMetadata{Duration: int(cfg.Timeout.Seconds())}})
	}
	return &discordgo.AutoModerationRule{Name: name, EventType: discordgo.AutoModerationEventMessageSend,
		TriggerType: trigger, TriggerMetadata: meta, Actions: actions, Enabled: &enabled}
}
