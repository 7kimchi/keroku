// Package cleanup serves channel and server housekeeping: purge, slowmode, channel locks
// and server lockdown. Locks and lockdowns are also used by raid protection and the sweeper.
package cleanup

import (
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/discord"
	"github.com/7kimchi/keroku/internal/store"
)

// Modlog receives event embeds.
type Modlog interface {
	Case(guildID int64, e *discordgo.MessageEmbed) bool
}

// Service runs cleanup actions.
type Service struct {
	store  *store.Store
	client discord.Client
	modlog Modlog
	now    func() time.Time
}

// New builds a service. now may be nil.
func New(s *store.Store, c discord.Client, m Modlog, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{store: s, client: c, modlog: m, now: now}
}
