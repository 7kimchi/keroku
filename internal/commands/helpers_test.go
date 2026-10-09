package commands

import (
	"context"
	"log/slog"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/goleak"

	"github.com/7kimchi/keroku/internal/cache"
	"github.com/7kimchi/keroku/internal/discord"
	"github.com/7kimchi/keroku/internal/metrics"
	"github.com/7kimchi/keroku/internal/ratelimit"
	"github.com/7kimchi/keroku/internal/safe"
	"github.com/7kimchi/keroku/internal/workers"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

var idCounter atomic.Int64

// snowflakeAt builds a unique id whose timestamp is t.
func snowflakeAt(t time.Time) string {
	ms := t.UnixMilli() - 1420070400000
	return strconv.FormatInt(ms<<22|(idCounter.Add(1)&0x3FFFFF), 10)
}

func opt(name string, t discordgo.ApplicationCommandOptionType, v any) *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{Name: name, Type: t, Value: v}
}

func interaction(name string, opts ...*discordgo.ApplicationCommandInteractionDataOption) *discordgo.Interaction {
	return &discordgo.Interaction{
		ID: snowflakeAt(time.Now()), Type: discordgo.InteractionApplicationCommand,
		GuildID: "100000000000000001", ChannelID: "100000000000000002", AppID: "1",
		Member: &discordgo.Member{User: &discordgo.User{ID: "100000000000000003"}, Permissions: 4},
		Data:   discordgo.ApplicationCommandInteractionData{Name: name, CommandType: discordgo.ChatApplicationCommand, Options: opts},
	}
}

// stub is a command whose handler is a function.
type stub struct {
	name string
	fn   func(context.Context, *Request) (*discordgo.MessageEmbed, error)
}

func (s stub) Definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{Name: s.name, Description: "test", DefaultMemberPermissions: Perm(4), Contexts: GuildOnly()}
}

func (s stub) Handle(ctx context.Context, r *Request) (*discordgo.MessageEmbed, error) {
	return s.fn(ctx, r)
}

type harness struct {
	d    *Dispatcher
	fake *discord.Fake
	m    *metrics.Metrics
	ack  *workers.Pool
	lane *workers.Pool
}

func newHarness(t *testing.T, burst int, cmds ...Command) *harness {
	t.Helper()
	g := safe.NewGuard(slog.New(slog.DiscardHandler), nil)
	reg, err := NewRegistry(cmds...)
	if err != nil {
		t.Fatal(err)
	}
	ack, _ := workers.New("ack", 8, 4096, g)
	lane, _ := workers.New("lanes", 8, 4096, g)
	user, _ := ratelimit.New(time.Second, burst, 10000, nil)
	guild, _ := ratelimit.New(time.Millisecond, burst*1000, 10000, nil)
	seen, _ := cache.New[string, struct{}](10000, 15*time.Minute, nil)
	h := &harness{fake: discord.NewFake(), m: metrics.New(), ack: ack, lane: lane}
	h.d = NewDispatcher(Deps{Client: h.fake, Registry: reg, Ack: ack, Lanes: lane, UserLimit: user, GuildLimit: guild,
		Seen: seen, Metrics: h.m, Log: slog.New(slog.DiscardHandler), Guard: g, HandlerTimeout: 2 * time.Second})
	t.Cleanup(h.drain)
	return h
}

// drain waits for every queued job. Safe to call twice.
func (h *harness) drain() {
	_ = h.ack.Close(context.Background())
	_ = h.lane.Close(context.Background())
}
