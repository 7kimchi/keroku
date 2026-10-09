package sweeper

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/store"
)

// maxAttempts is how often a failing timer is retried before it is marked failed.
const maxAttempts = 10

// outcome is what to do with a timer after one attempt.
type outcome struct {
	status   string    // done or failed; empty keeps it pending
	next     time.Time // next run when pending
	attempts int
	err      error
}

// Once claims and processes one due timer. It reports whether there was one.
func (s *Sweeper) Once(ctx context.Context) (bool, error) {
	worked := false
	err := s.d.Store.InTx(ctx, func(tx pgx.Tx) error {
		job, ok, err := store.ClaimDue(ctx, tx, s.d.Now())
		if err != nil || !ok {
			return err
		}
		worked = true
		jctx, cancel := context.WithTimeout(ctx, s.d.JobTimeout)
		o := s.process(jctx, job)
		cancel()
		s.d.Metrics.Sweeper.WithLabelValues(job.Kind, resultLabel(o)).Inc()
		msg := ""
		if o.err != nil {
			msg = o.err.Error()
			s.d.Log.Warn("timer attempt failed", "id", job.ID, "kind", job.Kind, "guildId", job.GuildID, "err", o.err)
		}
		if o.status == "failed" {
			s.notifyFailure(job, o.err)
		}
		if o.status != "" {
			return store.FinishJob(ctx, tx, job.ID, o.status, msg)
		}
		return store.RescheduleJob(ctx, tx, job.ID, o.next, o.attempts, msg)
	})
	return worked, err
}

// retry backs off on transient errors and gives up on refusals or after maxAttempts.
func (s *Sweeper) retry(job store.Job, err error) outcome {
	var refused *commands.UserError
	attempts := job.Attempts + 1
	if (errors.As(err, &refused) && refused.Cause == nil) || attempts >= maxAttempts {
		return outcome{status: "failed", err: err}
	}
	backoff := min(time.Duration(1<<attempts)*time.Minute, time.Hour)
	return outcome{next: s.d.Now().Add(backoff), attempts: attempts, err: err}
}

func resultLabel(o outcome) string {
	switch {
	case o.status != "":
		return o.status
	case o.err != nil:
		return "retry"
	}
	return "rescheduled"
}
