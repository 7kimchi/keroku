package records

import (
	"context"
	"strconv"
	"sync"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
	"go.uber.org/goleak"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/dbtest"
	"github.com/7kimchi/keroku/internal/store"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreTopFunction("github.com/jackc/pgx/v5/pgxpool.(*Pool).backgroundHealthCheck"))
}

const (
	gs     = "100000000000000001"
	guild  = int64(100000000000000001)
	target = int64(100000000000000005)
	ts     = "100000000000000005"
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

func setup(t *testing.T) (Deps, *modlog) {
	t.Helper()
	ml := &modlog{}
	return Deps{Store: store.New(dbtest.New(t)), Modlog: ml, BotID: "100000000000000003"}, ml
}

func seed(t *testing.T, d Deps, g int64, kind cases.Kind, reason string) cases.Case {
	t.Helper()
	var c cases.Case
	err := d.Store.InTx(context.Background(), func(tx pgx.Tx) error {
		var err error
		c, err = cases.Insert(context.Background(), tx, cases.New{GuildID: g, Kind: kind, TargetID: target, ModeratorID: 4, Reason: reason})
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

var reqID = 1300000000000200000

func request(t *testing.T, name, sub string, opts ...*discordgo.ApplicationCommandInteractionDataOption) *commands.Request {
	t.Helper()
	reqID++
	if sub != "" {
		opts = []*discordgo.ApplicationCommandInteractionDataOption{{Name: sub, Type: discordgo.ApplicationCommandOptionSubCommand, Options: opts}}
	}
	i := &discordgo.Interaction{ID: strconv.Itoa(reqID), Type: discordgo.InteractionApplicationCommand, GuildID: gs,
		ChannelID: "100000000000000050", Member: &discordgo.Member{User: &discordgo.User{ID: "100000000000000004"}},
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
