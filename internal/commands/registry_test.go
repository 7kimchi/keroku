package commands

import (
	"context"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

type defCmd struct{ def *discordgo.ApplicationCommand }

func (d defCmd) Definition() *discordgo.ApplicationCommand { return d.def }
func (defCmd) Handle(context.Context, *Request) (*discordgo.MessageEmbed, error) {
	return nil, nil
}

func good(name string) *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{Name: name, Description: "d", DefaultMemberPermissions: Perm(4), Contexts: GuildOnly()}
}

func TestRegistryAccepts(t *testing.T) {
	r, err := NewRegistry(defCmd{good("ban")}, defCmd{good("kick")})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := r.Get("ban"); !ok {
		t.Fatal("ban missing")
	}
	if _, ok := r.Get("nope"); ok {
		t.Fatal("unknown found")
	}
	defs := r.Definitions()
	if len(defs) != 2 || defs[0].Name != "ban" || defs[1].Name != "kick" {
		t.Fatal("order lost")
	}
}

func TestRegistryRejects(t *testing.T) {
	mutate := map[string]func(*discordgo.ApplicationCommand){
		"upper":       func(d *discordgo.ApplicationCommand) { d.Name = "Ban" },
		"space":       func(d *discordgo.ApplicationCommand) { d.Name = "b an" },
		"long":        func(d *discordgo.ApplicationCommand) { d.Name = strings.Repeat("a", 33) },
		"no desc":     func(d *discordgo.ApplicationCommand) { d.Description = "" },
		"long desc":   func(d *discordgo.ApplicationCommand) { d.Description = strings.Repeat("a", 101) },
		"no perms":    func(d *discordgo.ApplicationCommand) { d.DefaultMemberPermissions = nil },
		"no contexts": func(d *discordgo.ApplicationCommand) { d.Contexts = nil },
		"dm allowed": func(d *discordgo.ApplicationCommand) {
			d.Contexts = &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild, discordgo.InteractionContextBotDM}
		},
	}
	for name, m := range mutate {
		d := good("ban")
		m(d)
		if _, err := NewRegistry(defCmd{d}); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
	if _, err := NewRegistry(defCmd{good("ban")}, defCmd{good("ban")}); err == nil {
		t.Fatal("duplicate accepted")
	}
	if _, err := NewRegistry(defCmd{nil}); err == nil {
		t.Fatal("nil definition accepted")
	}
}
