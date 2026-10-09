package perms

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestHas(t *testing.T) {
	if !Has(BanMembers|KickMembers, BanMembers) || Has(KickMembers, BanMembers) {
		t.Fatal("single bit")
	}
	if !Has(Administrator, ManageGuild|BanMembers) || Has(BanMembers, BanMembers|KickMembers) {
		t.Fatal("admin or multi bit")
	}
	if !Has(0, 0) {
		t.Fatal("empty need")
	}
}

func TestNames(t *testing.T) {
	for _, bit := range []int64{BanMembers, KickMembers, ModerateMembers, ManageMessages, ManageChannels,
		ManageRoles, ManageGuild, ReadHistory, SendMessages, EmbedLinks, ViewChannel, Administrator} {
		if Name(bit) == "Unknown" {
			t.Fatalf("bit %d has no name", bit)
		}
	}
	if Name(BanMembers|KickMembers) != "Unknown" {
		t.Fatal("combined bits should not have a name")
	}
}

func TestBase(t *testing.T) {
	g := guild()
	if Base(g, "owner", nil) != allBits {
		t.Fatal("owner is not all powerful")
	}
	if got := Base(g, "u", nil); got != ViewChannel|SendMessages {
		t.Fatalf("everyone only: %b", got)
	}
	if got := Base(g, "u", []string{"mod", "member"}); got != ViewChannel|SendMessages|AddReactions|BanMembers|KickMembers {
		t.Fatalf("roles: %b", got)
	}
	if Base(g, "u", []string{"admin"}) != allBits {
		t.Fatal("admin role")
	}
}

func overwrite(id string, kind discordgo.PermissionOverwriteType, allow, deny int64) *discordgo.PermissionOverwrite {
	return &discordgo.PermissionOverwrite{ID: id, Type: kind, Allow: allow, Deny: deny}
}

func TestInChannelOrder(t *testing.T) {
	g := guild()
	role, member := discordgo.PermissionOverwriteTypeRole, discordgo.PermissionOverwriteTypeMember
	ch := &discordgo.Channel{PermissionOverwrites: []*discordgo.PermissionOverwrite{
		overwrite("member", role, 0, SendMessages),
		overwrite("u", member, SendMessages, 0),
		overwrite("g", role, ManageMessages, ViewChannel),
		overwrite("mod", role, ViewChannel, 0),
	}}
	got := InChannel(g, ch, "u", []string{"member"})
	if got&SendMessages == 0 || got&ViewChannel != 0 || got&ManageMessages == 0 {
		t.Fatalf("member allow should win, everyone deny should hold: %b", got)
	}
	got = InChannel(g, ch, "other", []string{"member", "mod"})
	if got&SendMessages != 0 || got&ViewChannel == 0 {
		t.Fatalf("role deny and role allow: %b", got)
	}
	if InChannel(g, ch, "x", []string{"admin"}) != allBits {
		t.Fatal("admin ignores overwrites")
	}
	if InChannel(g, ch, "owner", nil) != allBits {
		t.Fatal("owner ignores overwrites")
	}
}

func TestTopPosition(t *testing.T) {
	g := guild()
	if TopPosition(g, nil) != 0 || TopPosition(g, []string{"g"}) != 0 {
		t.Fatal("no roles is zero")
	}
	if TopPosition(g, []string{"member", "admin", "mod"}) != 10 {
		t.Fatal("max not found")
	}
	if TopPosition(g, []string{"ghost"}) != 0 {
		t.Fatal("unknown role counted")
	}
}
