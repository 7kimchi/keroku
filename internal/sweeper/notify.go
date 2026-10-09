package sweeper

import (
	"errors"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/embeds"
	"github.com/7kimchi/keroku/internal/store"
	"github.com/7kimchi/keroku/internal/validate"
)

func asUserErr(err error, target **commands.UserError) bool { return errors.As(err, target) }

// notifyFailure tells the guild's moderators that a timer could not finish.
func (s *Sweeper) notifyFailure(job store.Job, err error) {
	detail := "Keroku could not complete it. Check its role position and permissions."
	var refused *commands.UserError
	if asUserErr(err, &refused) && refused.Cause == nil {
		detail = refused.Detail
	}
	s.d.Modlog.Case(job.GuildID, embeds.New("Timer failed").
		Field("Timer", timerName(job.Kind), true).
		Field("Target", target(job), true).
		Description(detail).Build())
}

func timerName(kind string) string {
	switch kind {
	case store.TimerUnban:
		return "Temporary ban end"
	case store.TimerTimeoutRenew:
		return "Timeout renewal"
	case store.TimerChannelUnlock:
		return "Channel unlock"
	}
	return "Lockdown end"
}

func target(job store.Job) string {
	id := validate.FormatSnowflake(job.TargetID)
	switch job.Kind {
	case store.TimerChannelUnlock:
		return embeds.Channel(id)
	case store.TimerLockdownEnd:
		return "Server"
	}
	return embeds.User(id)
}
