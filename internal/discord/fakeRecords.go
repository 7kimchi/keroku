package discord

import (
	"context"
	"strings"
)

// Commands returns the registered command names.
func (f *Fake) Commands() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	names := make([]string, 0, len(f.commands))
	for _, c := range f.commands {
		names = append(names, c.Name)
	}
	return names
}

// DMCount counts DMs delivered to anyone.
func (f *Fake) DMCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, s := range f.sent {
		if strings.HasPrefix(s.To, "dm:") {
			n++
		}
	}
	return n
}

// MessageCount returns how many messages remain in a channel.
func (f *Fake) MessageCount(channelID string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.messages[channelID])
}

// SetIdentity sets what Identity returns.
func (f *Fake) SetIdentity(appID, botUserID string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.appID, f.botID = appID, botUserID
}

// Identity returns the configured application and bot ids.
func (f *Fake) Identity(ctx context.Context) (string, string, error) {
	if err := f.enter(ctx, "identity"); err != nil {
		return "", "", err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.appID, f.botID, nil
}
