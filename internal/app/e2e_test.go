package app

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/gatewaytest"
	"github.com/7kimchi/keroku/internal/perms"
)

// A warn sent through a real gateway session ends as a stored case and a reply.
func TestWarnEndToEnd(t *testing.T) {
	gw := gatewaytest.New(t, 1)
	a, fake := newTestApp(t, testConfig())
	a.gatewayHTTP = gw.Client()
	g, bot, mod, user := "100000000000000001", "100000000000000999", "100000000000000004", "100000000000000005"
	fake.AddGuild(g, "100000000000000002", 0, &discordgo.Role{ID: "mod", Position: 2},
		&discordgo.Role{ID: "bot", Position: 9, Permissions: perms.ModerateMembers})
	fake.AddMember(g, bot, "bot")
	fake.AddMember(g, mod, "mod")
	fake.AddMember(g, user)
	cancel, done := runApp(t, a)
	defer cancel()
	waitUntil(t, func() bool {
		select {
		case err := <-done:
			t.Fatalf("Run returned early: %v", err)
		default:
		}
		return len(gw.Identified()) == 1
	})
	id := strconv.FormatInt((time.Now().UnixMilli()-1420070400000)<<22, 10)
	gw.Broadcast(fmt.Sprintf(`"t":"INTERACTION_CREATE","d":{"id":%q,"type":2,"guild_id":%q,"channel_id":"100000000000000010",`+
		`"app_permissions":"%d","member":{"user":{"id":%q},"roles":["mod"],"permissions":"%d"},`+
		`"data":{"name":"warn","type":1,"options":[{"name":"user","type":6,"value":%q},{"name":"reason","type":3,"value":"spam"}]}}`,
		id, g, perms.ModerateMembers, mod, perms.ModerateMembers, user))
	waitUntil(t, func() bool { return len(fake.Replies(id)) == 1 })
	reply := fake.Replies(id)[0]
	if reply.Title != "Member warned" || reply.Footer.Text != "Case 1" {
		t.Fatalf("reply %+v", reply)
	}
	if len(fake.SentTo("dm:"+user)) != 1 {
		t.Fatal("member not notified")
	}
	cancel()
	if err := waitDone(t, done); err != nil {
		t.Fatal(err)
	}
}

func waitUntil(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal("condition not met in time")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
