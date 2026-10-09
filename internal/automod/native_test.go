package automod

import (
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/discord"
	"github.com/7kimchi/keroku/internal/store"
)

func TestSyncNativeLifecycle(t *testing.T) {
	f := discord.NewFake()
	f.SetIdentity("1", bs)
	ctx := t.Context()
	// A rule someone else made must survive every sync.
	f.SetIdentity("1", "100000000000000077")
	_ = f.CreateAutoModRule(ctx, gs, &discordgo.AutoModerationRule{Name: "Keroku mention spam"}, "")
	f.SetIdentity("1", bs)
	c := store.DefaultAutomod()
	c.MentionLimit, c.InvitesBlocked, c.Timeout = 5, true, time.Hour
	if err := SyncNative(ctx, f, gs, bs, c); err != nil {
		t.Fatal(err)
	}
	rules, _ := f.AutoModRules(ctx, gs)
	ours := map[string]*discordgo.AutoModerationRule{}
	for _, r := range rules {
		if r.CreatorID == bs {
			ours[r.Name] = r
		}
	}
	m, i := ours[mentionRuleName], ours[inviteRuleName]
	if len(rules) != 3 || m == nil || i == nil || m.TriggerMetadata.MentionTotalLimit != 5 || len(m.Actions) != 2 {
		t.Fatalf("rules %+v", rules)
	}
	if i.TriggerMetadata.RegexPatterns[0] != invitePattern || i.Actions[1].Metadata.Duration != 3600 {
		t.Fatalf("invite rule %+v", i)
	}
	c.MentionLimit, c.Timeout = 9, 0
	_ = SyncNative(ctx, f, gs, bs, c)
	rules, _ = f.AutoModRules(ctx, gs)
	if len(rules) != 3 || f.Calls("editAutoModRule") != 2 {
		t.Fatalf("edit created duplicates: %d rules", len(rules))
	}
	c.MentionLimit, c.InvitesBlocked = 0, false
	_ = SyncNative(ctx, f, gs, bs, c)
	if rules, _ = f.AutoModRules(ctx, gs); len(rules) != 1 || rules[0].CreatorID == bs {
		t.Fatalf("after off: %+v", rules)
	}
}

func TestSyncNativeError(t *testing.T) {
	f := discord.NewFake()
	f.FailNext("autoModRules", &discord.Error{Op: "autoModRules", Kind: discord.Forbidden}, 1)
	if err := SyncNative(t.Context(), f, gs, bs, store.DefaultAutomod()); !discord.Is(err, discord.Forbidden) {
		t.Fatalf("got %v", err)
	}
}
