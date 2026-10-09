package discord

import (
	"context"
	"time"
)

// Kick removes the member.
func (f *Fake) Kick(ctx context.Context, guildID, userID, _ string) error {
	if err := f.enter(ctx, "kick"); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.members[guildID][userID]; !ok {
		return notFound("kick", 10007)
	}
	delete(f.members[guildID], userID)
	return nil
}

// Timeout sets or clears the member's timeout. Discord refuses more than 28 days.
func (f *Fake) Timeout(ctx context.Context, guildID, userID string, until *time.Time, _ string) error {
	if err := f.enter(ctx, "timeout"); err != nil {
		return err
	}
	if until != nil && time.Until(*until) > 28*24*time.Hour+time.Minute {
		return &Error{Op: "timeout", Kind: BadRequest, Status: 400}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	m, ok := f.members[guildID][userID]
	if !ok {
		return notFound("timeout", 10007)
	}
	m.CommunicationDisabledUntil = until
	return nil
}
