package perms

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

// guild has roles: everyone(0), member(1), mod(5), admin(10). Owner is "owner".
func guild() *discordgo.Guild {
	return &discordgo.Guild{
		ID:      "g",
		OwnerID: "owner",
		Roles: []*discordgo.Role{
			{ID: "g", Position: 0, Permissions: ViewChannel | SendMessages},
			{ID: "member", Position: 1, Permissions: AddReactions},
			{ID: "mod", Position: 5, Permissions: BanMembers | KickMembers},
			{ID: "admin", Position: 10, Permissions: Administrator},
			{ID: "mod2", Position: 5, Permissions: BanMembers},
		},
	}
}

func baseRequest() Request {
	return Request{
		Guild: guild(), Need: BanMembers,
		InvokerID: "inv", InvokerRoles: []string{"mod"}, InvokerPerms: BanMembers,
		BotID: "bot", BotRoles: []string{"admin"}, BotPerms: BanMembers,
		TargetID: "tgt", TargetRoles: []string{"member"}, TargetMember: true,
	}
}
