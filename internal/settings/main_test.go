package settings

import (
	"strconv"
	"testing"
	"time"

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

const gs, bs = "100000000000000001", "100000000000000003"

func setup(t *testing.T) (ConfigCommand, *discord.Fake) {
	t.Helper()
	f := discord.NewFake()
	f.AddGuild(gs, "100000000000000002", perms.ViewChannel|perms.SendMessages|perms.EmbedLinks,
		&discordgo.Role{ID: "botrole", Position: 3})
	f.AddMember(gs, bs, "botrole")
	f.AddChannel(gs, "100000000000000010")
	f.AddChannel("100000000000000099", "100000000000000011")
	f.AddChannel(gs, "100000000000000012")
	_ = f.SetRoleOverwrite(t.Context(), "100000000000000012", gs, 0, perms.SendMessages, "")
	st := store.New(dbtest.New(t))
	cache, _ := store.NewSettingsCache(st, 100, time.Minute)
	am, _ := store.NewCached(st.Automod, 100, time.Minute)
	rd, _ := store.NewCached(st.Raid, 100, time.Minute)
	f.SetIdentity("1", bs)
	return ConfigCommand{Deps{Store: st, Settings: cache, Automod: am, Raid: rd, Client: f, BotID: bs}}, f
}

var reqID = 1300000000000300000

func request(t *testing.T, path []string, opts ...*discordgo.ApplicationCommandInteractionDataOption) *commands.Request {
	t.Helper()
	reqID++
	for i := len(path) - 1; i >= 0; i-- {
		kind := discordgo.ApplicationCommandOptionSubCommand
		if i < len(path)-1 {
			kind = discordgo.ApplicationCommandOptionSubCommandGroup
		}
		opts = []*discordgo.ApplicationCommandInteractionDataOption{{Name: path[i], Type: kind, Options: opts}}
	}
	i := &discordgo.Interaction{ID: strconv.Itoa(reqID), Type: discordgo.InteractionApplicationCommand, GuildID: gs,
		ChannelID: "100000000000000050", Member: &discordgo.Member{User: &discordgo.User{ID: "100000000000000004"}},
		Data: discordgo.ApplicationCommandInteractionData{Name: "config", CommandType: discordgo.ChatApplicationCommand, Options: opts}}
	r, err := commands.Parse(i, "ref")
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func opt(name string, t discordgo.ApplicationCommandOptionType, v any) *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{Name: name, Type: t, Value: v}
}

const tChan, tInt, tStr = discordgo.ApplicationCommandOptionChannel, discordgo.ApplicationCommandOptionInteger, discordgo.ApplicationCommandOptionString

const tBool = discordgo.ApplicationCommandOptionBoolean
