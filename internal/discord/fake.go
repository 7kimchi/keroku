package discord

import (
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Fake is an in memory Client for tests. It mimics Discord's responses for the calls the bot
// makes, records what was sent, and can inject failures and latency per operation.
type Fake struct {
	mu        sync.Mutex
	guilds    map[string]*discordgo.Guild
	members   map[string]map[string]*discordgo.Member
	bans      map[string]map[string]bool
	channels  map[string]*discordgo.Channel
	messages  map[string][]*discordgo.Message
	blocked   map[string]bool
	rules     map[string][]*discordgo.AutoModerationRule
	responded map[string]bool
	expired   map[string]bool
	sent      []Sent
	edits     []Sent
	commands  []*discordgo.ApplicationCommand
	calls     map[string]int
	failures  map[string][]error
	delay     map[string]time.Duration
	hooks     map[string]func()
	now       func() time.Time
	nextID    int
	appID     string
	botID     string
}

// Sent records one embed the bot delivered. To is a channel, user or interaction id.
type Sent struct {
	To    string
	Embed *discordgo.MessageEmbed
}

// NewFake returns an empty fake.
func NewFake() *Fake {
	return &Fake{
		guilds: map[string]*discordgo.Guild{}, members: map[string]map[string]*discordgo.Member{},
		bans: map[string]map[string]bool{}, channels: map[string]*discordgo.Channel{},
		messages: map[string][]*discordgo.Message{}, blocked: map[string]bool{},
		rules: map[string][]*discordgo.AutoModerationRule{}, responded: map[string]bool{},
		expired: map[string]bool{}, calls: map[string]int{}, failures: map[string][]error{},
		delay: map[string]time.Duration{}, hooks: map[string]func(){}, now: time.Now, nextID: 1000,
	}
}

func (f *Fake) id() string {
	f.nextID++
	return itoa(f.nextID)
}
