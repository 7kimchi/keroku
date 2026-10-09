package moderation

import (
	"strconv"
	"testing"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/perms"
)

func all(s *Service) []commands.Command {
	return []commands.Command{BanCommand{s}, UnbanCommand{s}, KickCommand{s}, TimeoutCommand{s},
		UntimeoutCommand{s}, WarnCommand{s}, NoteCommand{s}}
}

func TestDefinitionsRegister(t *testing.T) {
	reg, err := commands.NewRegistry(all(nil)...)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]int64{"ban": perms.BanMembers, "unban": perms.BanMembers, "kick": perms.KickMembers,
		"timeout": perms.ModerateMembers, "untimeout": perms.ModerateMembers, "warn": perms.ModerateMembers, "note": perms.ModerateMembers}
	for _, d := range reg.Definitions() {
		if *d.DefaultMemberPermissions != want[d.Name] {
			t.Fatalf("%s permission %d", d.Name, *d.DefaultMemberPermissions)
		}
		for _, o := range d.Options {
			if len(o.Description) > 100 || o.Description == "" {
				t.Fatalf("%s option %s description", d.Name, o.Name)
			}
		}
	}
}

var reqID = 1300000000000100000

func request(t *testing.T, name string, opts ...*discordgo.ApplicationCommandInteractionDataOption) *commands.Request {
	t.Helper()
	reqID++
	i := &discordgo.Interaction{ID: strconv.Itoa(reqID), Type: discordgo.InteractionApplicationCommand,
		GuildID: gs, ChannelID: "100000000000000050", AppPermissions: perms.BanMembers | perms.KickMembers | perms.ModerateMembers,
		Member: &discordgo.Member{User: &discordgo.User{ID: ms}, Roles: []string{"mod"},
			Permissions: perms.BanMembers | perms.KickMembers | perms.ModerateMembers},
		Data: discordgo.ApplicationCommandInteractionData{Name: name, CommandType: discordgo.ChatApplicationCommand, Options: opts}}
	r, err := commands.Parse(i, "ref")
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func o(name string, t discordgo.ApplicationCommandOptionType, v any) *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{Name: name, Type: t, Value: v}
}

const (
	tUser = discordgo.ApplicationCommandOptionUser
	tStr  = discordgo.ApplicationCommandOptionString
	tInt  = discordgo.ApplicationCommandOptionInteger
)

func TestBanCommand(t *testing.T) {
	e := setup(t)
	r := request(t, "ban", o("user", tUser, us), o("reason", tStr, " raid "), o("duration", tStr, "7d"), o("delete", tInt, 86400.0))
	reply, err := BanCommand{e.svc}.Handle(t.Context(), r)
	if err != nil {
		t.Fatal(err)
	}
	if reply.Title != "Member banned" || reply.Footer.Text != "Case 1" || !e.fake.Banned(gs, us) {
		t.Fatalf("reply %+v", reply)
	}
	got := map[string]string{}
	for _, f := range reply.Fields {
		got[f.Name] = f.Value
	}
	if got["Reason"] != "raid" || got["Duration"] != "7d" || got["DM"] != "Sent" {
		t.Fatalf("fields %v", got)
	}
}
