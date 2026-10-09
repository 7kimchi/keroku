package eventlog

import (
	"context"
	"log/slog"
	"strconv"
	"sync"
	"testing"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/goleak"

	"github.com/7kimchi/keroku/internal/metrics"
	"github.com/7kimchi/keroku/internal/safe"
	"github.com/7kimchi/keroku/internal/store"
	"github.com/7kimchi/keroku/internal/workers"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

const gs, cs, us = "100000000000000001", "100000000000000010", "100000000000000005"

type poster struct {
	mu   sync.Mutex
	sent []*discordgo.MessageEmbed
}

func (p *poster) Log(_ int64, e *discordgo.MessageEmbed) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sent = append(p.sent, e)
	return true
}

type settings struct{ logChannel int64 }

func (s settings) Get(context.Context, int64) (store.GuildSettings, error) {
	return store.GuildSettings{LogChannelID: s.logChannel}, nil
}

func setup(t *testing.T, logChannel int64) (*Logger, *poster, *workers.Pool) {
	t.Helper()
	pool, _ := workers.New("events", 2, 100_000, safe.NewGuard(slog.New(slog.DiscardHandler), nil))
	p := &poster{}
	l, err := New(p, settings{logChannel}, pool, metrics.New(), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close(context.Background()) })
	return l, p, pool
}

func msg(id, content string) *discordgo.Message {
	return &discordgo.Message{ID: id, GuildID: gs, ChannelID: cs, Content: content, Author: &discordgo.User{ID: us}}
}

func fields(e *discordgo.MessageEmbed) map[string]string {
	out := map[string]string{}
	for _, f := range e.Fields {
		out[f.Name] = f.Value
	}
	return out
}

func itoa(n int) string { return strconv.Itoa(n) }
