package commands

// Laned is implemented by commands that pick their own ordering key.
type Laned interface {
	Lane(r *Request) string
}

// laneKey is what a command runs in order with. Actions on one member are ordered, so are
// commands that declare a narrower key. The rest is ordered per guild. Unrelated work in
// one guild no longer waits behind a slow purge or another moderator's ban.
func laneKey(cmd Command, r *Request) string {
	if l, ok := cmd.(Laned); ok {
		return r.GuildID + "/" + l.Lane(r)
	}
	if id, ok, err := r.User("user"); err == nil && ok {
		return r.GuildID + "/u" + id
	}
	return r.GuildID
}
