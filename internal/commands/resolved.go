package commands

import "github.com/bwmarrin/discordgo"

// Role returns a role option id.
func (r *Request) Role(name string) (string, bool, error) {
	return r.Snowflake(name, discordgo.ApplicationCommandOptionRole)
}

// ResolvedUser returns the user Discord sent with the interaction for id.
func (r *Request) ResolvedUser(id string) (*discordgo.User, bool) {
	if r.resolved == nil {
		return nil, false
	}
	u := r.resolved.Users[id]
	if u == nil || u.ID != id {
		return nil, false
	}
	return u, true
}

// ResolvedRole returns the role Discord sent with the interaction for id.
func (r *Request) ResolvedRole(id string) (*discordgo.Role, bool) {
	if r.resolved == nil {
		return nil, false
	}
	role := r.resolved.Roles[id]
	if role == nil || role.ID != id {
		return nil, false
	}
	return role, true
}
