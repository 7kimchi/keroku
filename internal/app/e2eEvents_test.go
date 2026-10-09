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

// A message deleted after it was seen shows up in the log channel with its text.
func TestDeleteLogEndToEnd(t *testing.T) {
	gw := gatewaytest.New(t, 1)
	a, fake := newTestApp(t, testConfig())
	a.gatewayHTTP = gw.Client()
	const g, ch, logs = "100000000000000001", "100000000000000010", "100000000000000011"
	fake.AddGuild(g, "100000000000000002", 0)
	fake.AddChannel(g, ch)
	fake.AddChannel(g, logs)
	_ = a.store.SetLogChannel(t.Context(), 100000000000000001, 100000000000000011)
	cancel, done := runApp(t, a)
	defer cancel()
	waitUntil(t, func() bool { return len(gw.Identified()) == 1 })
	gw.Broadcast(fmt.Sprintf(`"t":"MESSAGE_CREATE","d":{"id":"5","channel_id":%q,"guild_id":%q,"content":"secret plan",`+
		`"author":{"id":"100000000000000007"},"member":{"roles":[]}}`, ch, g))
	gw.Broadcast(fmt.Sprintf(`"t":"MESSAGE_DELETE","d":{"id":"5","channel_id":%q,"guild_id":%q}`, ch, g))
	waitUntil(t, func() bool { return len(fake.SentTo(logs)) == 1 })
	if got := fake.SentTo(logs)[0]; got.Title != "Message deleted" || got.Fields[2].Value != "secret plan" {
		t.Fatalf("log %+v", got)
	}
	cancel()
	if err := waitDone(t, done); err != nil {
		t.Fatal(err)
	}
}
