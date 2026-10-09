package gateway

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/websocket"

	"github.com/7kimchi/keroku/internal/safe"
)

// Config selects shards and intents.
type Config struct {
	Token          string
	ShardCount     int   // 0 uses the recommendation
	ShardIDs       []int // nil runs every shard
	MessageContent bool
	IdentifyWait   time.Duration     // pause between identify batches, 5s in production
	HTTPClient     *http.Client      // tests point these at a fake gateway
	Dialer         *websocket.Dialer // nil uses discordgo's default
}

// Gateway runs this instance's shard sessions.
type Gateway struct {
	cfg      Config
	router   Router
	guard    *safe.Guard
	log      *slog.Logger
	mu       sync.Mutex
	sessions []*discordgo.Session
}

// New prepares a gateway. Nothing connects until Open.
func New(cfg Config, r Router, g *safe.Guard, log *slog.Logger) *Gateway {
	if cfg.IdentifyWait <= 0 {
		cfg.IdentifyWait = 5 * time.Second
	}
	return &Gateway{cfg: cfg, router: r, guard: g, log: log}
}

// Open asks Discord for the shard recommendation and connects every planned shard,
// batch by batch. On failure every opened session is closed again.
func (g *Gateway) Open(ctx context.Context) error {
	probe, err := g.session(0, 1)
	if err != nil {
		return err
	}
	info, err := probe.GatewayBot(discordgo.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("gateway: fetch shard recommendation: %w", errors.New(errText(err)))
	}
	plan, err := PlanShards(info.Shards, g.cfg.ShardCount, g.cfg.ShardIDs,
		info.SessionStartLimit.MaxConcurrency, info.SessionStartLimit.Remaining)
	if err != nil {
		return err
	}
	for n, batch := range plan.Batches {
		if n > 0 {
			select {
			case <-ctx.Done():
				g.Close()
				return ctx.Err()
			case <-time.After(g.cfg.IdentifyWait):
			}
		}
		for _, id := range batch {
			if err := g.openShard(id, plan.Count); err != nil {
				g.Close()
				return err
			}
		}
	}
	g.log.Info("gateway connected", "shards", len(g.sessions), "shardCount", plan.Count)
	return nil
}
