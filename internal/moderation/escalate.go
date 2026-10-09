package moderation

import (
	"context"
	"strconv"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/store"
	"github.com/7kimchi/keroku/internal/validate"
)

// escalate fires the configured step when a warning brings the member to its threshold.
// The warning's case id is the idempotency key, so a step fires at most once per warning.
func (s *Service) escalate(ctx context.Context, warn Action, w cases.Case) *Result {
	steps, err := s.store.EscalationSteps(ctx, warn.GuildID)
	if err != nil || len(steps) == 0 {
		if err != nil {
			s.log.Warn("escalation lookup failed", "guildId", warn.GuildID, "err", err)
		}
		return nil
	}
	count, err := cases.CountUpTo(ctx, s.store.Pool(), warn.GuildID, warn.TargetID, cases.Warn, w.Number)
	if err != nil {
		s.log.Warn("warning count failed", "guildId", warn.GuildID, "err", err)
		return nil
	}
	step, ok := stepFor(steps, count)
	if !ok {
		return nil
	}
	botID, _ := validate.Snowflake(s.botID)
	a := Action{
		Kind: cases.Kind(step.Action), GuildID: warn.GuildID, TargetID: warn.TargetID, ModeratorID: botID,
		Reason:         "Reached " + strconv.Itoa(count) + " warnings.",
		IdempotencyKey: "escalation:" + strconv.FormatInt(w.ID, 10), Automated: true,
	}
	if a.Kind == cases.Timeout || a.Kind == cases.Ban {
		a.Duration = step.Duration
	}
	res, err := s.Execute(ctx, a)
	if err != nil {
		s.log.Warn("escalation failed", "guildId", warn.GuildID, "targetId", warn.TargetID, "err", err)
		return nil
	}
	return &res
}

func stepFor(steps []store.EscalationStep, count int) (store.EscalationStep, bool) {
	for _, st := range steps {
		if st.WarnCount == count {
			return st, true
		}
	}
	return store.EscalationStep{}, false
}
