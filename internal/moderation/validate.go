package moderation

import (
	"errors"
	"time"
	"unicode/utf8"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/validate"
)

var errBadAction = errors.New("moderation: invalid action")

// validateAction rejects anything a command handler should never have produced.
func validateAction(a Action) error {
	switch {
	case !a.Kind.Valid(), a.GuildID <= 0, a.TargetID <= 0, a.ModeratorID <= 0:
		return errBadAction
	case utf8.RuneCountInString(a.Reason) > validate.MaxReason:
		return errBadAction
	case a.InteractionID == 0 && a.IdempotencyKey == "":
		return errBadAction
	case a.DeleteSeconds < 0 || a.DeleteSeconds > MaxDeleteSeconds || (a.DeleteSeconds > 0 && a.Kind != cases.Ban):
		return errBadAction
	}
	switch a.Kind {
	case cases.Ban:
		if a.Duration != 0 && (a.Duration < time.Minute || a.Duration > MaxBan) {
			return errBadAction
		}
	case cases.Timeout:
		if a.Duration < time.Minute || a.Duration > MaxTimeout {
			return errBadAction
		}
	default:
		if a.Duration != 0 {
			return errBadAction
		}
	}
	return nil
}
