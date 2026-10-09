package app

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/gatewaytest"
	"github.com/7kimchi/keroku/internal/perms"
	"github.com/7kimchi/keroku/internal/store"
)

// Spam and a join burst sent through a real gateway session reach automod and raid.
func TestAutomodAndRaidEndToEnd(t *testing.T) {
	gw := gatewaytest.New(t, 1)
	a, fake := newTestApp(t, testConfig())
	a.gatewayHTTP = gw.Client()
	const g, ch, spammer = "100000000000000001", "100000000000000010", "100000000000000007"
	fake.AddGuild(g, "100000000000000002", perms.SendMessages,
		&discordgo.Role{ID: "bot", Position: 9, Permissions: perms.KickMembers | perms.ManageMessages})
	fake.AddMember(g, "100000000000000999", "bot")
	fake.AddChannel(g, ch)
	am := store.DefaultAutomod()
	am.SpamEnabled, am.SpamMessages = true, 3
	_ = a.store.SetAutomod(t.Context(), 100000000000000001, am)
	_ = a.store.SetRaid(t.Context(), 100000000000000001, store.RaidSettings{Enabled: true, JoinLimit: 3,
		Window: time.Minute, Action: "kick", LockdownFor: time.Hour})
	cancel, done := runApp(t, a)
	defer cancel()
	waitUntil(t, func() bool { return len(gw.Identified()) == 1 })
	for i := range 5 {
		id := strconv.Itoa(1300000000000700000 + i)
		fake.AddMessage(ch, id, spammer, time.Now())
		gw.Broadcast(fmt.Sprintf(`"t":"MESSAGE_CREATE","d":{"id":%q,"channel_id":%q,"guild_id":%q,"content":"m%d",`+
			`"author":{"id":%q},"member":{"roles":[]}}`, id, ch, g, i, spammer))
	}
	for i := range 3 {
		id := strconv.Itoa(1300000000000800000 + i)
		fake.AddMember(g, id)
		gw.Broadcast(fmt.Sprintf(`"t":"GUILD_MEMBER_ADD","d":{"guild_id":%q,"user":{"id":%q},"roles":[],"joined_at":"%s"}`,
			g, id, time.Now().UTC().Format(time.RFC3339)))
	}
	waitUntil(t, func() bool { return fake.Calls("deleteMessage") == 3 && fake.Calls("kick") == 3 })
	cancel()
	if err := waitDone(t, done); err != nil {
		t.Fatal(err)
	}
}
