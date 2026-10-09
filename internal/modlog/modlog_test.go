package modlog

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"sync"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"go.uber.org/goleak"

	"github.com/7kimchi/keroku/internal/discord"
	"github.com/7kimchi/keroku/internal/metrics"
	"github.com/7kimchi/keroku/internal/safe"
	"github.com/7kimchi/keroku/internal/store"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

type settings struct {
	mu   sync.Mutex
	m    map[int64]store.GuildSettings
	fail map[int64]bool
}

func (s *settings) Get(_ context.Context, g int64) (store.GuildSettings, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.fail[g] {
		return store.GuildSettings{}, errors.New("db down")
	}
	return s.m[g], nil
}

func setup(t *testing.T, queue int) (*Poster, *discord.Fake, *settings, *metrics.Metrics) {
	t.Helper()
	f := discord.NewFake()
	f.AddGuild("1", "o", 0)
	f.AddChannel("1", "10")
	f.AddChannel("1", "20")
	s := &settings{m: map[int64]store.GuildSettings{1: {ModlogChannelID: 10, LogChannelID: 20}}}
	m := metrics.New()
	p, err := New(f, s, 2, queue, m, slog.New(slog.DiscardHandler), safe.NewGuard(slog.New(slog.DiscardHandler), nil))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = p.Close(context.Background()) })
	return p, f, s, m
}

func embed(n int) *discordgo.MessageEmbed { return &discordgo.MessageEmbed{Title: strconv.Itoa(n)} }

func TestPostsToTheRightChannels(t *testing.T) {
	p, f, _, m := setup(t, 100)
	p.Case(1, embed(1))
	p.Log(1, embed(2))
	p.Case(2, embed(3)) // no settings for guild 2
	_ = p.Close(context.Background())
	if len(f.SentTo("10")) != 1 || len(f.SentTo("20")) != 1 || f.SentTo("10")[0].Title != "1" {
		t.Fatal("wrong channel routing")
	}
	if testutil.ToFloat64(m.Modlog.WithLabelValues("unset")) != 1 || testutil.ToFloat64(m.Modlog.WithLabelValues("ok")) != 2 {
		t.Fatal("metrics wrong")
	}
}

func TestDeletedChannelAndSettingsFailure(t *testing.T) {
	p, f, s, m := setup(t, 100)
	f.DeleteChannel("10")
	p.Case(1, embed(1))
	s.mu.Lock()
	s.fail = map[int64]bool{3: true}
	s.mu.Unlock()
	p.Log(3, embed(2))
	_ = p.Close(context.Background())
	if testutil.ToFloat64(m.Modlog.WithLabelValues("notFound")) != 1 || testutil.ToFloat64(m.Modlog.WithLabelValues("settingsError")) != 1 {
		t.Fatal("failures not counted")
	}
}
