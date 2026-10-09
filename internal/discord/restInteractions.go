package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// Respond sends the initial interaction response. Mentions are always disabled.
func (r *REST) Respond(ctx context.Context, i *discordgo.Interaction, resp *discordgo.InteractionResponse) error {
	if resp.Data != nil {
		resp.Data.AllowedMentions = NoMentions()
	}
	return r.call(ctx, "respond", false, func(o ...discordgo.RequestOption) error {
		return r.s.InteractionRespond(i, resp, o...)
	})
}

// EditResponse replaces the original response with one embed.
func (r *REST) EditResponse(ctx context.Context, i *discordgo.Interaction, embed *discordgo.MessageEmbed) error {
	if embed == nil {
		return &Error{Op: "editResponse", Kind: BadRequest}
	}
	edit := &discordgo.WebhookEdit{Embeds: &[]*discordgo.MessageEmbed{embed}, AllowedMentions: NoMentions()}
	return r.call(ctx, "editResponse", true, func(o ...discordgo.RequestOption) error {
		_, e := r.s.InteractionResponseEdit(i, edit, o...)
		return e
	})
}

// AutoModRules lists the guild's native AutoMod rules.
func (r *REST) AutoModRules(ctx context.Context, guildID string) (rules []*discordgo.AutoModerationRule, err error) {
	err = r.call(ctx, "autoModRules", true, func(o ...discordgo.RequestOption) (e error) {
		rules, e = r.s.AutoModerationRules(guildID, o...)
		return
	})
	return rules, err
}

// CreateAutoModRule adds a native rule.
func (r *REST) CreateAutoModRule(ctx context.Context, guildID string, rule *discordgo.AutoModerationRule, reason string) error {
	return r.call(ctx, "createAutoModRule", false, func(o ...discordgo.RequestOption) error {
		_, e := r.s.AutoModerationRuleCreate(guildID, rule, with(o, reason)...)
		return e
	})
}

// EditAutoModRule updates a native rule.
func (r *REST) EditAutoModRule(ctx context.Context, guildID, ruleID string, rule *discordgo.AutoModerationRule, reason string) error {
	return r.call(ctx, "editAutoModRule", true, func(o ...discordgo.RequestOption) error {
		_, e := r.s.AutoModerationRuleEdit(guildID, ruleID, rule, with(o, reason)...)
		return e
	})
}

// DeleteAutoModRule removes a native rule.
func (r *REST) DeleteAutoModRule(ctx context.Context, guildID, ruleID, reason string) error {
	return r.call(ctx, "deleteAutoModRule", true, func(o ...discordgo.RequestOption) error {
		return r.s.AutoModerationRuleDelete(guildID, ruleID, with(o, reason)...)
	})
}

// OverwriteCommands replaces every global command in one call.
func (r *REST) OverwriteCommands(ctx context.Context, appID string, cmds []*discordgo.ApplicationCommand) error {
	return r.call(ctx, "overwriteCommands", true, func(o ...discordgo.RequestOption) error {
		_, e := r.s.ApplicationCommandBulkOverwrite(appID, "", cmds, o...)
		return e
	})
}
