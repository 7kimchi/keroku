package gateway

import (
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/safe"
)

type recorder struct {
	mu           sync.Mutex
	interactions []string
	messages     int
	panicOn      string
}

func (r *recorder) Interaction(i *discordgo.Interaction) {
	if i.ID == r.panicOn {
		panic("handler bug")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.interactions = append(r.interactions, i.ID)
}
func (r *recorder) MessageCreate(*discordgo.Message)          { r.mu.Lock(); r.messages++; r.mu.Unlock() }
func (r *recorder) MessageUpdate(*discordgo.MessageUpdate)    {}
func (r *recorder) MessageDelete(*discordgo.MessageDelete)    {}
func (r *recorder) MemberAdd(*discordgo.GuildMemberAdd)       {}
func (r *recorder) MemberRemove(*discordgo.GuildMemberRemove) {}

func (r *recorder) seen() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.interactions...)
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestOpenDeliversEventsAndSurvivesPanics(t *testing.T) {
	events := []string{
		`"t":"INTERACTION_CREATE","d":{"id":"bad","type":2,"data":{"name":"x"}}`,
		`"t":"INTERACTION_CREATE","d":{"id":"good","type":2,"data":{"name":"x"}}`,
		`"t":"MESSAGE_CREATE","d":{"id":"m","channel_id":"c","content":"hi","author":{"id":"u"}}`,
	}
	f := newFakeGateway(t, 3, events...)
	rec := &recorder{panicOn: "bad"}
	var panics int
	var mu sync.Mutex
	g := safe.NewGuard(slog.New(slog.DiscardHandler), func(string) { mu.Lock(); panics++; mu.Unlock() })
	gw := New(Config{Token: "fake.test.token", IdentifyWait: 10 * time.Millisecond, HTTPClient: f.client()}, rec, g, slog.New(slog.DiscardHandler))
	if err := gw.Open(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer gw.Close()
	waitFor(t, "identify of 3 shards", func() bool { return len(f.identified()) == 3 })
	waitFor(t, "good interaction on every shard", func() bool { return len(rec.seen()) == 3 })
	for _, id := range f.identified() {
		if id["token"] != "Bot fake.test.token" || int64(id["intents"].(float64)) != int64(Intents(false)) {
			t.Fatalf("identify %v", id)
		}
	}
	mu.Lock()
	defer mu.Unlock()
	if panics != 3 {
		t.Fatalf("recovered %d panics, want 3", panics)
	}
}
