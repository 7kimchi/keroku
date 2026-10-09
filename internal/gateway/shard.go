package gateway

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func (g *Gateway) openShard(id, count int) error {
	s, err := g.session(id, count)
	if err != nil {
		return err
	}
	if err := s.Open(); err != nil {
		return fmt.Errorf("gateway: open shard %d: %s", id, errText(err))
	}
	g.mu.Lock()
	g.sessions = append(g.sessions, s)
	g.mu.Unlock()
	return nil
}

func (g *Gateway) session(id, count int) (*discordgo.Session, error) {
	s, err := newSession(g.cfg.Token, id, count, Intents(g.cfg.MessageContent), g.router, g.guard)
	if err != nil {
		return nil, err
	}
	if g.cfg.HTTPClient != nil {
		s.Client = g.cfg.HTTPClient
	}
	if g.cfg.Dialer != nil {
		s.Dialer = g.cfg.Dialer
	}
	return s, nil
}

// Close disconnects every shard. Safe to call more than once.
func (g *Gateway) Close() {
	g.mu.Lock()
	sessions := g.sessions
	g.sessions = nil
	g.mu.Unlock()
	for _, s := range sessions {
		if err := s.Close(); err != nil {
			g.log.Warn("gateway close", "shard", s.ShardID, "err", errText(err))
		}
	}
}
