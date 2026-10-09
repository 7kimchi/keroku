package cases

import "unicode/utf8"

// Source says what started an action.
type Source string

// Action sources.
const (
	FromCommand    Source = "command"
	FromAutomod    Source = "automod"
	FromRaid       Source = "raid"
	FromEscalation Source = "escalation"
	FromTimer      Source = "timer"
)

// Limits on details.
const (
	MaxRule          = 64
	maxDeleteSeconds = 604800
)

// Details are per action extras stored as a jsonb object. Every field is optional.
type Details struct {
	Source        Source `json:"source,omitempty"`
	DeleteSeconds int    `json:"deleteSeconds,omitempty"` // ban message deletion window
	Rule          string `json:"rule,omitempty"`          // automod rule that fired
	WarnCase      int64  `json:"warnCase,omitempty"`      // warning case that set off an escalation
	TimerID       int64  `json:"timerId,omitempty"`       // timer that ran the action
}

// Valid reports whether d is safe to store.
func (d Details) Valid() bool {
	switch d.Source {
	case "", FromCommand, FromAutomod, FromRaid, FromEscalation, FromTimer:
	default:
		return false
	}
	return d.DeleteSeconds >= 0 && d.DeleteSeconds <= maxDeleteSeconds && d.WarnCase >= 0 && d.TimerID >= 0 &&
		utf8.ValidString(d.Rule) && utf8.RuneCountInString(d.Rule) <= MaxRule && !hasControl(d.Rule)
}

func hasControl(s string) bool {
	for _, r := range s {
		if r < 0x20 || r == 0x7f {
			return true
		}
	}
	return false
}
