package app

import (
	"fmt"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/gatewaytest"
	"github.com/7kimchi/keroku/internal/perms"
)

// Info commands go out with the rest at startup and answer through a real gateway session.
func TestInfoEndToEnd(t *testing.T) {
	gw := gatewaytest.New(t, 1)
	a, fake := newTestApp(t, testConfig())
	a.gatewayHTTP = gw.Client()
	g, member, other := "100000000000000001", "100000000000000004", "100000000000000005"
	fake.AddGuild(g, "100000000000000002", 0)
	fake.AddMember(g, member)
	cancel, done := runApp(t, a)
	defer cancel()
	waitUntil(t, func() bool { return len(gw.Identified()) == 1 })
	names := fake.Commands()
	for _, n := range []string{"serverinfo", "botinfo", "userinfo", "roleinfo"} {
		if !slices.Contains(names, n) {
			t.Fatalf("%s not registered: %v", n, names)
		}
	}
	send := func(n int64, data string) string {
		id := strconv.FormatInt(((time.Now().UnixMilli()-1420070400000)<<22)+n, 10)
		gw.Broadcast(fmt.Sprintf(`"t":"INTERACTION_CREATE","d":{"id":%q,"type":2,"guild_id":%q,"channel_id":"100000000000000010",`+
			`"member":{"user":{"id":%q,"username":"me"},"roles":[],"permissions":"%d"},"data":%s}`,
			id, g, member, perms.ViewChannel, data))
		return id
	}
	server := send(1, `{"name":"serverinfo","type":1}`)
	user := send(2, fmt.Sprintf(`{"name":"userinfo","type":1,"options":[{"name":"user","type":6,"value":%q}],`+
		`"resolved":{"users":{%q:{"id":%q,"username":"other"}}}}`, other, other, other))
	waitUntil(t, func() bool { return len(fake.Replies(server)) == 1 && len(fake.Replies(user)) == 1 })
	if r := fake.Replies(server)[0]; r.Footer == nil || r.Footer.Text != "ID "+g {
		t.Fatalf("serverinfo %+v", r)
	}
	if r := fake.Replies(user)[0]; r.Title != "other" || r.Fields[len(r.Fields)-1].Value != "Not in this server" {
		t.Fatalf("userinfo %+v", r)
	}
	cancel()
	if err := waitDone(t, done); err != nil {
		t.Fatal(err)
	}
}
