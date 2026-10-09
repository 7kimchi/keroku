package moderation

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/commands"
)

// Execute runs one action. Everything that decides whether the action happens runs inside
// one transaction holding a lock on the target, so double submits, replays and two
// moderators acting at once resolve to one action and one case.
func (s *Service) Execute(ctx context.Context, a Action) (Result, error) {
	if err := validateAction(a); err != nil {
		return Result{}, err
	}
	var res Result
	applied := false
	err := s.store.InTx(ctx, func(tx pgx.Tx) error {
		dup, err := s.duplicate(ctx, tx, a)
		if err != nil || dup != nil {
			if dup != nil {
				res = *dup
			}
			return err
		}
		st, err := s.inspect(ctx, a)
		if err != nil {
			return err
		}
		res.GuildName = st.guild.Name
		if a.Kind == cases.Ban || a.Kind == cases.Kick {
			// After a ban or kick there is no shared server left to DM through.
			res.DM = s.notify(ctx, a, st.guild.Name)
		}
		if err := s.apply(ctx, a); err != nil {
			return err
		}
		applied = true
		c, err := cases.Insert(ctx, tx, cases.New{GuildID: a.GuildID, Kind: a.Kind, TargetID: a.TargetID,
			ModeratorID: a.ModeratorID, Reason: a.Reason, Duration: a.Duration,
			InteractionID: a.InteractionID, IdempotencyKey: a.IdempotencyKey})
		if err != nil {
			return err
		}
		res.Case = c
		return s.timers(ctx, tx, a)
	})
	if err != nil {
		return res, s.failure(ctx, a, applied, err)
	}
	if res.Duplicate {
		return res, nil
	}
	s.after(ctx, a, &res)
	return res, nil
}

// duplicate claims the interaction, takes the target lock and looks for an earlier case
// that already covers this action.
func (s *Service) duplicate(ctx context.Context, tx pgx.Tx, a Action) (*Result, error) {
	if a.InteractionID != 0 {
		ok, err := cases.Claim(ctx, tx, a.InteractionID, a.GuildID)
		if err != nil || !ok {
			return &Result{Duplicate: true}, err
		}
	}
	if a.IdempotencyKey != "" {
		if c, ok, err := cases.ByKey(ctx, tx, a.GuildID, a.IdempotencyKey); err != nil || ok {
			return &Result{Duplicate: true, Case: c}, err
		}
	}
	if err := cases.LockTarget(ctx, tx, a.GuildID, a.TargetID); err != nil {
		return nil, err
	}
	if a.Automated {
		return nil, nil
	}
	c, ok, err := cases.Recent(ctx, tx, a.GuildID, a.TargetID, a.Kind, s.now().Add(-duplicateWindow))
	if err != nil || ok {
		return &Result{Duplicate: true, Case: c}, err
	}
	return nil, nil
}

// failure turns an error into what the moderator sees. If Discord already applied the
// action but the case was not saved, the action is reverted when Discord allows it.
func (s *Service) failure(ctx context.Context, a Action, applied bool, err error) error {
	var userErr *commands.UserError
	if errors.As(err, &userErr) {
		return err
	}
	if !applied {
		return discordFailure(a, err)
	}
	return s.compensate(ctx, a, err)
}
