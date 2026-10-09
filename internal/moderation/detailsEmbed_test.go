package moderation

import (
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/cases"
)

func embedField(c cases.Case, name string) string {
	for _, f := range CaseEmbed(c, bs).Fields {
		if f.Name == name {
			return f.Value
		}
	}
	return ""
}

func TestCaseEmbedDetails(t *testing.T) {
	c := cases.Case{Number: 4, Kind: cases.Ban, TargetID: user, ModeratorID: bot, CreatedAt: time.Now(),
		Details: cases.Details{Source: cases.FromAutomod, Rule: "mention **spam**", DeleteSeconds: 3600}}
	if embedField(c, "Source") != "Automod" || embedField(c, "Rule") != "mention \\*\\*spam\\*\\*" ||
		embedField(c, "Messages deleted") != "Last 1h" {
		t.Fatalf("%+v", CaseEmbed(c, bs).Fields)
	}
	c.Details = cases.Details{Source: cases.FromEscalation, WarnCase: 12}
	if embedField(c, "Source") != "Warn escalation" || embedField(c, "Triggered by") != "Case 12" {
		t.Fatalf("%+v", CaseEmbed(c, bs).Fields)
	}
}

// Command cases show no source field. Old cases with empty details render as before.
func TestCaseEmbedNoDetails(t *testing.T) {
	for _, d := range []cases.Details{{}, {Source: cases.FromCommand}} {
		c := cases.Case{Number: 1, Kind: cases.Warn, TargetID: user, ModeratorID: mod, Details: d}
		for _, f := range []string{"Source", "Rule", "Triggered by", "Messages deleted"} {
			if embedField(c, f) != "" {
				t.Fatalf("%+v shows %s", d, f)
			}
		}
	}
}

func TestSourceNames(t *testing.T) {
	want := map[cases.Source]string{cases.FromRaid: "Raid protection", cases.FromTimer: "Scheduled",
		cases.FromAutomod: "Automod", cases.FromEscalation: "Warn escalation", cases.FromCommand: "", "x": ""}
	for s, w := range want {
		if got := sourceName(s); got != w {
			t.Fatalf("%s: %q", s, got)
		}
	}
}
