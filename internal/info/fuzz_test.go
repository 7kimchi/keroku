package info

import (
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
)

func FuzzUserEmbed(f *testing.F) {
	f.Add("name", "global", "nick", rs, int64(0))
	f.Add("@everyone", "\x00‮", "[a](b)", "<@&1>", int64(-1))
	f.Fuzz(func(t *testing.T, user, global, nick, role string, joined int64) {
		u := &discordgo.User{ID: us, Username: user, GlobalName: global}
		m := &discordgo.Member{Nick: nick, Roles: []string{role, rs}, JoinedAt: time.Unix(joined%1e10, 0)}
		fits(t, userEmbed(u, m, us, time.Now()))
	})
}

func FuzzRoleEmbed(f *testing.F) {
	f.Add("role", 0, 0, int64(0))
	f.Add("**@here**", -1, 1<<30, int64(-1))
	f.Fuzz(func(t *testing.T, name string, color, pos int, p int64) {
		fits(t, roleEmbed(&discordgo.Role{ID: rs, Name: name, Color: color, Position: pos, Permissions: p}, gs))
	})
}

func FuzzServerEmbed(f *testing.F) {
	f.Add("guild", "100000000000000002", 5, 2)
	f.Fuzz(func(t *testing.T, name, owner string, members, tier int) {
		g := &discordgo.Guild{Name: name, OwnerID: owner, ApproximateMemberCount: members,
			PremiumTier: discordgo.PremiumTier(tier)}
		fits(t, serverEmbed(g, gs))
	})
}
