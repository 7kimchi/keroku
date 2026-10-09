// Package cases stores moderation cases: numbering, lookups, history and reason edits.
package cases

import "time"

// Kind is the action a case records.
type Kind string

// Case kinds. They match the database check constraint.
const (
	Ban       Kind = "ban"
	Unban     Kind = "unban"
	Kick      Kind = "kick"
	Timeout   Kind = "timeout"
	Untimeout Kind = "untimeout"
	Warn      Kind = "warn"
	Note      Kind = "note"
)

// Valid reports whether k is a known kind.
func (k Kind) Valid() bool {
	switch k {
	case Ban, Unban, Kick, Timeout, Untimeout, Warn, Note:
		return true
	}
	return false
}

// Case is one stored record. Ids are Discord snowflakes as numbers.
type Case struct {
	ID             int64
	GuildID        int64
	Number         int64
	Kind           Kind
	TargetID       int64
	ModeratorID    int64
	Reason         string
	Duration       time.Duration // zero when the action is permanent
	InteractionID  int64         // zero for automatic actions
	IdempotencyKey string        // set for automatic actions
	CreatedAt      time.Time
}

// New describes a case to insert. The number is assigned on insert.
type New struct {
	GuildID        int64
	Kind           Kind
	TargetID       int64
	ModeratorID    int64
	Reason         string
	Duration       time.Duration
	InteractionID  int64
	IdempotencyKey string
}
