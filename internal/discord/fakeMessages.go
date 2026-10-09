package discord

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Messages lists messages before beforeID, newest first, like Discord.
func (f *Fake) Messages(ctx context.Context, channelID string, limit int, beforeID string) ([]*discordgo.Message, error) {
	if err := f.enter(ctx, "messages"); err != nil {
		return nil, err
	}
	if limit < 1 || limit > 100 {
		return nil, &Error{Op: "messages", Kind: BadRequest, Status: 400}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.channels[channelID]; !ok {
		return nil, notFound("messages", 10003)
	}
	all := f.messages[channelID]
	end := len(all)
	if beforeID != "" {
		end = 0
		for i, m := range all {
			if m.ID == beforeID {
				end = i
				break
			}
		}
	}
	var out []*discordgo.Message
	for i := end - 1; i >= 0 && len(out) < limit; i-- {
		cp := *all[i]
		out = append(out, &cp)
	}
	return out, nil
}

// BulkDelete enforces Discord's rules: 2 to 100 ids, none older than 14 days.
func (f *Fake) BulkDelete(ctx context.Context, channelID string, ids []string, _ string) error {
	if err := f.enter(ctx, "bulkDelete"); err != nil {
		return err
	}
	if len(ids) < 2 || len(ids) > 100 {
		return &Error{Op: "bulkDelete", Kind: BadRequest, Status: 400, Code: 50016}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	drop := map[string]bool{}
	for _, id := range ids {
		drop[id] = true
	}
	cutoff := time.Now().Add(-14 * 24 * time.Hour)
	for _, m := range f.messages[channelID] {
		if drop[m.ID] && m.Timestamp.Before(cutoff) {
			return &Error{Op: "bulkDelete", Kind: BadRequest, Status: 400, Code: 50034}
		}
	}
	f.removeMessages(channelID, drop)
	return nil
}

// DeleteMessage removes a single message.
func (f *Fake) DeleteMessage(ctx context.Context, channelID, messageID, _ string) error {
	if err := f.enter(ctx, "deleteMessage"); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	before := len(f.messages[channelID])
	f.removeMessages(channelID, map[string]bool{messageID: true})
	if len(f.messages[channelID]) == before {
		return notFound("deleteMessage", 10008)
	}
	return nil
}

func (f *Fake) removeMessages(channelID string, drop map[string]bool) {
	kept := f.messages[channelID][:0]
	for _, m := range f.messages[channelID] {
		if !drop[m.ID] {
			kept = append(kept, m)
		}
	}
	f.messages[channelID] = kept
}
