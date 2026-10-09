package store

import (
	"context"
	"time"

	"github.com/7kimchi/keroku/internal/cache"
)

// SettingsCache keeps recently used guild settings in memory. Writes through this cache
// take effect at once on this instance and within the TTL on others.
type SettingsCache struct {
	s     *Store
	cache *cache.Cache[int64, GuildSettings]
}

// NewSettingsCache bounds the cache to maxGuilds entries, each kept for ttl.
func NewSettingsCache(s *Store, maxGuilds int, ttl time.Duration) (*SettingsCache, error) {
	c, err := cache.New[int64, GuildSettings](maxGuilds, ttl, nil)
	if err != nil {
		return nil, err
	}
	return &SettingsCache{s: s, cache: c}, nil
}

// Get returns cached settings or loads them.
func (c *SettingsCache) Get(ctx context.Context, guildID int64) (GuildSettings, error) {
	if g, ok := c.cache.Get(guildID); ok {
		return g, nil
	}
	g, err := c.s.GuildSettings(ctx, guildID)
	if err != nil {
		return GuildSettings{}, err
	}
	c.cache.Set(guildID, g)
	return g, nil
}

// SetModlog stores the modlog channel and drops the cached copy.
func (c *SettingsCache) SetModlog(ctx context.Context, guildID, channelID int64) error {
	defer c.cache.Delete(guildID)
	return c.s.SetModlogChannel(ctx, guildID, channelID)
}

// SetLog stores the log channel and drops the cached copy.
func (c *SettingsCache) SetLog(ctx context.Context, guildID, channelID int64) error {
	defer c.cache.Delete(guildID)
	return c.s.SetLogChannel(ctx, guildID, channelID)
}

// Sweep drops expired entries.
func (c *SettingsCache) Sweep() { c.cache.Sweep(1000) }
