package settings

import (
	"strconv"

	"github.com/7kimchi/keroku/internal/embeds"
	"github.com/7kimchi/keroku/internal/store"
)

func raidSummary(r store.RaidSettings) string {
	if !r.Enabled {
		return "Off"
	}
	out := strconv.Itoa(r.JoinLimit) + " joins in " + embeds.Duration(r.Window) + ": " + r.Action
	if r.Action == "lockdown" {
		out += " for " + embeds.Duration(r.LockdownFor)
	}
	if r.MinAccountAge > 0 {
		out += "\nAccounts younger than " + embeds.Duration(r.MinAccountAge) + " are removed"
	}
	return out
}
