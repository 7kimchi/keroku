// Package moderation runs moderation actions end to end: checks, the Discord call, the case
// record, timers, DMs, the modlog post and warn escalation.
package moderation

import (
	"time"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/perms"
)

// Limits on action input.
const (
	MaxTimeout       = 365 * 24 * time.Hour // longer than Discord allows; renewed by the sweeper
	DiscordTimeout   = 28 * 24 * time.Hour
	MaxBan           = 365 * 24 * time.Hour
	MaxDeleteSeconds = 604800
	duplicateWindow  = 10 * time.Second
)

// Action is one requested moderation action.
type Action struct {
	Kind          cases.Kind
	GuildID       int64
	TargetID      int64
	ModeratorID   int64 // the bot's id for automatic actions
	Reason        string
	Duration      time.Duration // ban length or timeout length; 0 means permanent ban
	DeleteSeconds int           // ban message deletion window

	InteractionID  int64  // set for commands
	IdempotencyKey string // set for automatic actions
	Automated      bool

	InvokerRoles []string
	InvokerPerms int64
	BotPerms     int64 // channel permissions from the interaction; 0 for automatic actions
}

// Result describes what happened.
type Result struct {
	Case       cases.Case
	Duplicate  bool   // the action was already handled; Case is the earlier case if known
	DM         string // "sent", "failed" or "skipped"
	GuildName  string
	Escalation *Result
}

// need returns the permission an action requires.
func need(k cases.Kind) int64 {
	switch k {
	case cases.Ban, cases.Unban:
		return perms.BanMembers
	case cases.Kick:
		return perms.KickMembers
	}
	return perms.ModerateMembers
}

// verb is the noun used in failure titles, like "Ban failed".
func verb(k cases.Kind) string {
	switch k {
	case cases.Ban:
		return "Ban"
	case cases.Unban:
		return "Unban"
	case cases.Kick:
		return "Kick"
	case cases.Timeout:
		return "Timeout"
	case cases.Untimeout:
		return "Timeout removal"
	case cases.Warn:
		return "Warning"
	}
	return "Note"
}
