package discord

import (
	"context"
	"encoding/json"

	"github.com/bwmarrin/discordgo"
)

// GuildCounts fetches the guild with approximate member and online counts.
func (r *REST) GuildCounts(ctx context.Context, guildID string) (g *discordgo.Guild, err error) {
	err = r.call(ctx, "guildCounts", true, func(o ...discordgo.RequestOption) (e error) {
		g, e = r.s.GuildWithCounts(guildID, o...)
		return
	})
	return g, err
}

// ServerCount returns Discord's approximate count of servers the app is in.
func (r *REST) ServerCount(ctx context.Context) (int, error) {
	// discordgo's Application type has no field for this count.
	var app struct {
		Count int `json:"approximate_guild_count"`
	}
	err := r.call(ctx, "serverCount", true, func(o ...discordgo.RequestOption) error {
		ep := discordgo.EndpointApplication("@me")
		body, e := r.s.RequestWithBucketID("GET", ep, nil, ep, o...)
		if e != nil {
			return e
		}
		return json.Unmarshal(body, &app)
	})
	if err != nil {
		return 0, err
	}
	return max(app.Count, 0), nil
}
