package moderation

import (
	"strconv"
	"time"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/embeds"
)

// detailFields adds what the case details record. Commands add nothing beyond the ban window.
func detailFields(b *embeds.Builder, d cases.Details) {
	if s := sourceName(d.Source); s != "" {
		b.Field("Source", s, true)
	}
	if d.Rule != "" {
		b.Field("Rule", embeds.Escape(d.Rule), true)
	}
	if d.WarnCase > 0 {
		b.Field("Triggered by", "Case "+strconv.FormatInt(d.WarnCase, 10), true)
	}
	if d.DeleteSeconds > 0 {
		b.Field("Messages deleted", "Last "+embeds.Duration(time.Duration(d.DeleteSeconds)*time.Second), true)
	}
}

func sourceName(s cases.Source) string {
	switch s {
	case cases.FromAutomod:
		return "Automod"
	case cases.FromRaid:
		return "Raid protection"
	case cases.FromEscalation:
		return "Warn escalation"
	case cases.FromTimer:
		return "Scheduled"
	}
	return ""
}
