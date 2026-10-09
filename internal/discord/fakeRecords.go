package discord

import "strings"

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
