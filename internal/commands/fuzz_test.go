package commands

import (
	"encoding/json"
	"testing"

	"github.com/bwmarrin/discordgo"
)

// FuzzParse feeds raw gateway payloads through discordgo's decoder and our parser.
// Nothing may panic, and anything accepted must carry valid ids.
func FuzzParse(f *testing.F) {
	f.Add(`{"id":"1100000000000000001","type":2,"guild_id":"1","channel_id":"2","member":{"user":{"id":"3"},"permissions":"4"},"data":{"name":"ban","type":1,"options":[{"name":"user","type":6,"value":"5"}]}}`)
	f.Add(`{"id":"1","type":2,"member":{"user":{}},"data":{"name":"x","type":1,"options":[{"type":1,"options":[null]}]}}`)
	f.Add(`{"type":2,"data":null}`)
	f.Fuzz(func(t *testing.T, raw string) {
		var i discordgo.Interaction
		if err := json.Unmarshal([]byte(raw), &i); err != nil {
			return
		}
		r, err := Parse(&i, "ref")
		if err != nil {
			return
		}
		if r.GuildID == "" || r.UserID == "" || r.InteractionID() <= 0 {
			t.Fatalf("accepted bad ids: %+v", r)
		}
		for name := range r.opts {
			_, _, _ = r.Text(name, 512)
			_, _, _ = r.Int(name, -10, 10)
			_, _, _ = r.User(name)
			_, _, _ = r.Bool(name)
		}
	})
}
