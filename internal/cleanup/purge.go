package cleanup

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Purge limits.
const (
	MaxPurge  = 500
	pageLimit = 100
	maxPages  = 20
	// Discord refuses bulk deletes of messages 14 days old. A minute of margin keeps a
	// slow request from crossing the line.
	bulkAge = 14*24*time.Hour - time.Minute
)

// PurgeResult counts what happened.
type PurgeResult struct {
	Deleted int
	Stopped bool // reached messages too old to bulk delete
}

// Purge deletes up to count recent messages, optionally only from one author. Pinned
// messages are kept.
func (s *Service) Purge(ctx context.Context, channelID string, count int, authorID, reason string) (PurgeResult, error) {
	ids, stopped, err := s.collect(ctx, channelID, count, authorID)
	if err != nil {
		return PurgeResult{}, err
	}
	res := PurgeResult{Stopped: stopped}
	for len(ids) > 0 {
		n := min(len(ids), 100)
		chunk := ids[:n]
		ids = ids[n:]
		if len(chunk) == 1 {
			err = s.client.DeleteMessage(ctx, channelID, chunk[0], reason)
		} else {
			err = s.client.BulkDelete(ctx, channelID, chunk, reason)
		}
		if err != nil {
			return res, err
		}
		res.Deleted += len(chunk)
	}
	return res, nil
}

func (s *Service) collect(ctx context.Context, channelID string, count int, authorID string) ([]string, bool, error) {
	cutoff := s.now().Add(-bulkAge)
	var ids []string
	before := ""
	for page := 0; page < maxPages && len(ids) < count; page++ {
		msgs, err := s.client.Messages(ctx, channelID, pageLimit, before)
		if err != nil {
			return nil, false, err
		}
		for _, m := range msgs {
			if m.Timestamp.Before(cutoff) {
				return ids, true, nil
			}
			if keep(m, authorID) {
				continue
			}
			ids = append(ids, m.ID)
			if len(ids) == count {
				return ids, false, nil
			}
		}
		if len(msgs) < pageLimit {
			break
		}
		before = msgs[len(msgs)-1].ID
	}
	return ids, false, nil
}

func keep(m *discordgo.Message, authorID string) bool {
	return m.Pinned || (authorID != "" && (m.Author == nil || m.Author.ID != authorID))
}
