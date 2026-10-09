package discord

import (
	"context"
	"encoding/json"

	"github.com/bwmarrin/discordgo"
)

// Identity returns the application id and the bot's user id.
func (r *REST) Identity(ctx context.Context) (string, string, error) {
	var app discordgo.Application
	err := r.call(ctx, "application", true, func(o ...discordgo.RequestOption) error {
		ep := discordgo.EndpointOAuth2Application("@me")
		body, e := r.s.RequestWithBucketID("GET", ep, nil, ep, o...)
		if e != nil {
			return e
		}
		return json.Unmarshal(body, &app)
	})
	if err != nil {
		return "", "", err
	}
	var me *discordgo.User
	err = r.call(ctx, "currentUser", true, func(o ...discordgo.RequestOption) (e error) {
		me, e = r.s.User("@me", o...)
		return
	})
	if err != nil {
		return "", "", err
	}
	return app.ID, me.ID, nil
}
