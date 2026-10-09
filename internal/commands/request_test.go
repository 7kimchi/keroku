package commands

import (
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
)

func TestParseHappyPath(t *testing.T) {
	i := interaction("ban", opt("user", discordgo.ApplicationCommandOptionUser, "100000000000000009"))
	i.AppPermissions = 8
	r, err := Parse(i, "ref1")
	if err != nil {
		t.Fatal(err)
	}
	if r.Name != "ban" || r.Sub != "" || r.UserID != "100000000000000003" || r.Permissions != 4 || r.AppPermissions != 8 {
		t.Fatalf("got %+v", r)
	}
	if time.Since(r.At) > time.Minute || r.Ref != "ref1" || r.InteractionID() <= 0 {
		t.Fatalf("time %v ref %s", r.At, r.Ref)
	}
}

func TestParseSubcommands(t *testing.T) {
	sub := &discordgo.ApplicationCommandInteractionDataOption{Name: "reason", Type: discordgo.ApplicationCommandOptionSubCommand,
		Options: []*discordgo.ApplicationCommandInteractionDataOption{opt("case", discordgo.ApplicationCommandOptionInteger, 5.0)}}
	r, err := Parse(interaction("case", sub), "x")
	if err != nil || r.Sub != "reason" {
		t.Fatalf("got %v %v", r, err)
	}
	if n, ok, err := r.Int("case", 1, 10); !ok || err != nil || n != 5 {
		t.Fatalf("option lost: %d %v %v", n, ok, err)
	}
	group := &discordgo.ApplicationCommandInteractionDataOption{Name: "automod", Type: discordgo.ApplicationCommandOptionSubCommandGroup,
		Options: []*discordgo.ApplicationCommandInteractionDataOption{sub}}
	if r, err := Parse(interaction("config", group), "x"); err != nil || r.Sub != "automod reason" {
		t.Fatalf("group: %v %v", r, err)
	}
	deep := &discordgo.ApplicationCommandInteractionDataOption{Name: "x", Type: discordgo.ApplicationCommandOptionSubCommandGroup,
		Options: []*discordgo.ApplicationCommandInteractionDataOption{group}}
	if _, err := Parse(interaction("config", deep), "x"); err == nil {
		t.Fatal("three levels accepted")
	}
}

func TestParseRejectsMalformed(t *testing.T) {
	mutations := map[string]func(*discordgo.Interaction){
		"nil member":    func(i *discordgo.Interaction) { i.Member = nil },
		"nil user":      func(i *discordgo.Interaction) { i.Member.User = nil },
		"bad guild":     func(i *discordgo.Interaction) { i.GuildID = "0x1" },
		"empty channel": func(i *discordgo.Interaction) { i.ChannelID = "" },
		"bad user":      func(i *discordgo.Interaction) { i.Member.User.ID = "-1" },
		"bad id":        func(i *discordgo.Interaction) { i.ID = "abc" },
		"wrong type":    func(i *discordgo.Interaction) { i.Type = discordgo.InteractionMessageComponent },
		"wrong data":    func(i *discordgo.Interaction) { i.Data = discordgo.MessageComponentInteractionData{} },
		"user command": func(i *discordgo.Interaction) {
			d := i.Data.(discordgo.ApplicationCommandInteractionData)
			d.CommandType = 2
			i.Data = d
		},
		"nil option": func(i *discordgo.Interaction) { setOpts(i, nil) },
		"dup option": func(i *discordgo.Interaction) { o := opt("a", 3, "x"); setOpts(i, o, o) },
		"stray sub": func(i *discordgo.Interaction) {
			setOpts(i, opt("a", 3, "x"), &discordgo.ApplicationCommandInteractionDataOption{Name: "s", Type: 1})
		},
		"too many opts": func(i *discordgo.Interaction) { setOpts(i, many(26)...) },
	}
	for name, mutate := range mutations {
		i := interaction("ban")
		mutate(i)
		if _, err := Parse(i, "x"); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
	if _, err := Parse(nil, "x"); err == nil {
		t.Fatal("nil accepted")
	}
}

func setOpts(i *discordgo.Interaction, o ...*discordgo.ApplicationCommandInteractionDataOption) {
	d := i.Data.(discordgo.ApplicationCommandInteractionData)
	d.Options = o
	i.Data = d
}

func many(n int) []*discordgo.ApplicationCommandInteractionDataOption {
	out := make([]*discordgo.ApplicationCommandInteractionDataOption, n)
	for k := range out {
		out[k] = opt(string(rune('a'+k%26))+string(rune('a'+k/26)), 3, "x")
	}
	return out
}
