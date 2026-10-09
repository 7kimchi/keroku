package discord

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

// Calls returns how many times op was called.
func (f *Fake) Calls(op string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls[op]
}

// SentTo returns embeds posted to a channel, or to "dm:<userID>" for DMs.
func (f *Fake) SentTo(to string) []*discordgo.MessageEmbed {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []*discordgo.MessageEmbed
	for _, s := range f.sent {
		if s.To == to {
			out = append(out, s.Embed)
		}
	}
	return out
}

// Replies returns every embed the bot used to answer an interaction, in order.
func (f *Fake) Replies(interactionID string) []*discordgo.MessageEmbed {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []*discordgo.MessageEmbed
	for _, s := range f.edits {
		if s.To == interactionID {
			out = append(out, s.Embed)
		}
	}
	return out
}

// Responded reports whether the interaction got its initial response.
func (f *Fake) Responded(interactionID string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.responded[interactionID]
}

// Banned reports the current ban state without counting a call.
func (f *Fake) Banned(guildID, userID string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.bans[guildID][userID]
}

// IsMember reports whether the user is in the guild without counting a call.
func (f *Fake) IsMember(guildID, userID string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, ok := f.members[guildID][userID]
	return ok
}

// TimeoutUntil returns the member's timeout end, or nil.
func (f *Fake) TimeoutUntil(guildID, userID string) *time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	if m := f.members[guildID][userID]; m != nil {
		return m.CommunicationDisabledUntil
	}
	return nil
}
