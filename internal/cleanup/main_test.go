package cleanup

import (
	"strconv"
	"sync"
	"testing"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/goleak"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/dbtest"
	"github.com/7kimchi/keroku/internal/discord"
	"github.com/7kimchi/keroku/internal/perms"
	"github.com/7kimchi/keroku/internal/store"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreTopFunction("github.com/jackc/pgx/v5/pgxpool.(*Pool).backgroundHealthCheck"))
}

const (
	gs    = "100000000000000001"
	guild = int64(100000000000000001)
	cs    = "100000000000000010"
	chan_ = int64(100000000000000010)
	all   = perms.ManageMessages | perms.ReadHistory | perms.ManageChannels | perms.ManageRoles | perms.ManageGuild
)

type modlog struct {
	mu   sync.Mutex
	sent []*discordgo.MessageEmbed
}

func (m *modlog) Case(_ int64, e *discordgo.MessageEmbed) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sent = append(m.sent, e)
	return true
}

func setup(t *testing.T) (*Service, *discord.Fake, *store.Store) {
	t.Helper()
	f := discord.NewFake()
	f.AddGuild(gs, "100000000000000002", perms.ViewChannel|perms.SendMessages|perms.AddReactions|perms.EmbedLinks)
	f.AddChannel(gs, cs)
	st := store.New(dbtest.New(t))
	return New(st, f, &modlog{}, nil), f, st
}

var reqID = 1300000000000400000

func request(t *testing.T, name, sub string, opts ...*discordgo.ApplicationCommandInteractionDataOption) *commands.Request {
	t.Helper()
	reqID++
	if sub != "" {
		opts = []*discordgo.ApplicationCommandInteractionDataOption{{Name: sub, Type: discordgo.ApplicationCommandOptionSubCommand, Options: opts}}
	}
	i := &discordgo.Interaction{ID: strconv.Itoa(reqID), Type: discordgo.InteractionApplicationCommand, GuildID: gs,
		ChannelID: cs, AppPermissions: all, Member: &discordgo.Member{User: &discordgo.User{ID: "100000000000000004"}, Permissions: all},
		Data: discordgo.ApplicationCommandInteractionData{Name: name, CommandType: discordgo.ChatApplicationCommand, Options: opts}}
	r, err := commands.Parse(i, "ref")
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func opt(name string, t discordgo.ApplicationCommandOptionType, v any) *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{Name: name, Type: t, Value: v}
}

const (
	tInt  = discordgo.ApplicationCommandOptionInteger
	tStr  = discordgo.ApplicationCommandOptionString
	tUser = discordgo.ApplicationCommandOptionUser
)

func TestDefinitionsRegister(t *testing.T) {
	if _, err := commands.NewRegistry(PurgeCommand{}, SlowmodeCommand{}, LockCommand{}, UnlockCommand{}, LockdownCommand{}); err != nil {
		t.Fatal(err)
	}
}
