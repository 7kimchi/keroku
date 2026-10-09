package commands

import (
	"context"
	"fmt"
	"regexp"

	"github.com/bwmarrin/discordgo"
)

// Command is one slash command.
type Command interface {
	Definition() *discordgo.ApplicationCommand
	Handle(ctx context.Context, r *Request) (*discordgo.MessageEmbed, error)
}

// Registry holds commands by name.
type Registry struct {
	byName map[string]Command
	order  []string
}

// NewRegistry checks every definition: lowercase unique names, guild only, and a default
// member permission so members without it never see the command.
func NewRegistry(cmds ...Command) (*Registry, error) {
	name := regexp.MustCompile(`^[a-z0-9_-]{1,32}$`)
	r := &Registry{byName: map[string]Command{}}
	for _, c := range cmds {
		d := c.Definition()
		switch {
		case d == nil || !name.MatchString(d.Name):
			return nil, fmt.Errorf("commands: bad name %v", d)
		case r.byName[d.Name] != nil:
			return nil, fmt.Errorf("commands: duplicate %s", d.Name)
		case d.Description == "" || len([]rune(d.Description)) > 100:
			return nil, fmt.Errorf("commands: %s needs a description up to 100 characters", d.Name)
		case d.DefaultMemberPermissions == nil:
			return nil, fmt.Errorf("commands: %s has no default member permissions", d.Name)
		case d.Contexts == nil || len(*d.Contexts) != 1 || (*d.Contexts)[0] != discordgo.InteractionContextGuild:
			return nil, fmt.Errorf("commands: %s must be guild only", d.Name)
		}
		r.byName[d.Name] = c
		r.order = append(r.order, d.Name)
	}
	return r, nil
}

// Get finds a command.
func (r *Registry) Get(name string) (Command, bool) {
	c, ok := r.byName[name]
	return c, ok
}

// Definitions lists every definition in registration order.
func (r *Registry) Definitions() []*discordgo.ApplicationCommand {
	out := make([]*discordgo.ApplicationCommand, 0, len(r.order))
	for _, n := range r.order {
		out = append(out, r.byName[n].Definition())
	}
	return out
}

// GuildOnly returns the contexts value that limits a command to servers.
func GuildOnly() *[]discordgo.InteractionContextType {
	return &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild}
}

// Perm wraps a permission bit for DefaultMemberPermissions.
func Perm(bit int64) *int64 { return &bit }
