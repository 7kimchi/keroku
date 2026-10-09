package settings

import (
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/discord"
)

const gid = int64(100000000000000001)

func TestAutomodCommands(t *testing.T) {
	c, f := setup(t)
	ctx := t.Context()
	steps := [][]any{
		{[]string{"automod", "spam"}, opt("enabled", tBool, true), opt("messages", tInt, 4.0), opt("seconds", tInt, 3.0)},
		{[]string{"automod", "duplicates"}, opt("enabled", tBool, true), opt("count", tInt, 5.0)},
		{[]string{"automod", "links"}, opt("enabled", tBool, true), opt("allow", tStr, " YouTube.com, ,github.com ")},
		{[]string{"automod", "mentions"}, opt("limit", tInt, 6.0)},
		{[]string{"automod", "invites"}, opt("enabled", tBool, true)},
		{[]string{"automod", "timeout"}, opt("duration", tStr, "10m")},
	}
	for _, s := range steps {
		r := request(t, s[0].([]string), toOpts(s[1:])...)
		if _, err := c.Handle(ctx, r); err != nil {
			t.Fatalf("%v: %v", s[0], err)
		}
	}
	cfg, _ := c.D.Store.Automod(ctx, gid)
	if !cfg.SpamEnabled || cfg.SpamMessages != 4 || cfg.SpamWindow != 3*time.Second || cfg.DuplicateCount != 5 ||
		strings.Join(cfg.AllowedDomains, ",") != "youtube.com,github.com" || cfg.MentionLimit != 6 || !cfg.InvitesBlocked ||
		cfg.Timeout != 10*time.Minute {
		t.Fatalf("stored %+v", cfg)
	}
	if rules, _ := f.AutoModRules(ctx, gs); len(rules) != 2 {
		t.Fatalf("%d native rules", len(rules))
	}
	if cached, _ := c.D.Automod.Get(ctx, gid); cached.Timeout != 10*time.Minute {
		t.Fatal("cache not invalidated")
	}
}

func TestAutomodRejects(t *testing.T) {
	c, f := setup(t)
	for _, r := range [][]any{
		{[]string{"automod", "links"}, opt("enabled", tBool, true), opt("allow", tStr, "good.com, bad domain")},
		{[]string{"automod", "links"}, opt("enabled", tBool, true), opt("allow", tStr, strings.Repeat("a.com,", 51))},
		{[]string{"automod", "timeout"}, opt("duration", tStr, "29d")},
		{[]string{"automod", "spam"}, opt("enabled", tBool, true), opt("messages", tInt, 1.0)},
	} {
		if _, err := c.Handle(t.Context(), request(t, r[0].([]string), toOpts(r[1:])...)); err == nil {
			t.Fatalf("%v accepted", r)
		}
	}
	f.FailNext("autoModRules", &discord.Error{Op: "autoModRules", Kind: discord.Forbidden}, 1)
	_, err := c.Handle(t.Context(), request(t, []string{"automod", "invites"}, opt("enabled", tBool, true)))
	if err == nil || !strings.Contains(err.Error(), "Manage Server") {
		t.Fatalf("got %v", err)
	}
}

func TestRaidCommand(t *testing.T) {
	c, _ := setup(t)
	r := request(t, []string{"raid", "set"}, opt("enabled", tBool, true), opt("joins", tInt, 20.0), opt("seconds", tInt, 30.0),
		opt("minage", tStr, "7d"), opt("action", tStr, "lockdown"), opt("lockdown", tStr, "1h"))
	e, err := c.Handle(t.Context(), r)
	if err != nil || !strings.Contains(e.Description, "20 joins in 30s: lockdown for 1h") {
		t.Fatalf("%+v %v", e, err)
	}
	cfg, _ := c.D.Store.Raid(t.Context(), gid)
	if cfg.MinAccountAge != 7*24*time.Hour {
		t.Fatalf("stored %+v", cfg)
	}
	if _, err := c.Handle(t.Context(), request(t, []string{"raid", "set"}, opt("minage", tStr, "off"))); err != nil {
		t.Fatal(err)
	}
	if cfg, _ = c.D.Store.Raid(t.Context(), gid); cfg.MinAccountAge != 0 || !cfg.Enabled || cfg.JoinLimit != 20 {
		t.Fatalf("partial update lost values: %+v", cfg)
	}
	for _, bad := range []string{"soon", "400d"} {
		if _, err := c.Handle(t.Context(), request(t, []string{"raid", "set"}, opt("minage", tStr, bad))); err == nil {
			t.Fatalf("minage %s accepted", bad)
		}
	}
	view, _ := c.Handle(t.Context(), request(t, []string{"view"}))
	if len(view.Fields) != 5 || view.Fields[3].Name != "Automod" || view.Fields[4].Value == "" {
		t.Fatalf("view %+v", view.Fields)
	}
}

func toOpts(in []any) []*discordgo.ApplicationCommandInteractionDataOption {
	out := make([]*discordgo.ApplicationCommandInteractionDataOption, 0, len(in))
	for _, o := range in {
		out = append(out, o.(*discordgo.ApplicationCommandInteractionDataOption))
	}
	return out
}
