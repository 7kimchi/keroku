package moderation

import (
	"context"
	"time"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/validate"
)

const dmTimeout = 3 * time.Second

// DM outcomes.
const (
	dmSent    = "sent"
	dmFailed  = "failed"
	dmSkipped = "skipped"
)

// notify DMs the target. It never fails the action: closed DMs and timeouts are reported.
func (s *Service) notify(ctx context.Context, a Action, guildName string) string {
	if a.Kind == cases.Note {
		return dmSkipped
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), dmTimeout)
	defer cancel()
	if err := s.client.DM(ctx, validate.FormatSnowflake(a.TargetID), dmEmbed(a, guildName, s.now())); err != nil {
		return dmFailed
	}
	return dmSent
}

// after runs once the case is committed: DM for actions that keep the member, the modlog
// post, and warn escalation.
func (s *Service) after(ctx context.Context, a Action, res *Result) {
	if res.DM == "" {
		res.DM = s.notify(ctx, a, res.GuildName)
	}
	if !s.modlog.Case(a.GuildID, CaseEmbed(res.Case, s.botID)) {
		s.log.Warn("modlog queue full", "guildId", a.GuildID, "case", res.Case.Number)
	}
	if a.Kind == cases.Warn {
		res.Escalation = s.escalate(ctx, a, res.Case)
	}
}

func failWith(a Action, detail string, cause error) error {
	return commands.FailWith(verb(a.Kind)+" failed", detail, cause)
}

func incomplete(a Action, detail string, cause error) error {
	return commands.FailWith(verb(a.Kind)+" incomplete", detail, cause)
}
