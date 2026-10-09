package raid

import (
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/perms"
)

const old = 365 * 24 * time.Hour

func TestJoinBurstKicksEveryJoiner(t *testing.T) {
	v := setup(t, raidCfg())
	var ids []string
	for i := range 5 {
		ids = append(ids, v.join(i, old))
	}
	v.drain()
	for _, id := range ids {
		if v.fake.IsMember(gs, id) {
			t.Fatalf("%s still in the server", id)
		}
	}
	if v.log.count("Raid detected") != 1 || v.log.count("Member kicked") != 5 {
		t.Fatalf("alerts %d kicks %d", v.log.count("Raid detected"), v.log.count("Member kicked"))
	}
}

func TestBanAction(t *testing.T) {
	c := raidCfg()
	c.Action = "ban"
	v := setup(t, c)
	var ids []string
	for i := range 3 {
		ids = append(ids, v.join(i, old))
	}
	v.drain()
	for _, id := range ids {
		if !v.fake.Banned(gs, id) {
			t.Fatalf("%s not banned", id)
		}
	}
}

func TestLockdownActionRunsOnce(t *testing.T) {
	c := raidCfg()
	c.Action = "lockdown"
	v := setup(t, c)
	for i := range 20 {
		v.join(i, old)
	}
	v.drain()
	g, _ := v.fake.Guild(t.Context(), gs)
	if g.Roles[0].Permissions&perms.SendMessages != 0 || v.fake.Calls("rolePermissions") != 1 || v.fake.Calls("kick") != 0 {
		t.Fatalf("lockdown wrong: perms %b edits %d kicks %d", g.Roles[0].Permissions, v.fake.Calls("rolePermissions"), v.fake.Calls("kick"))
	}
}

func TestYoungAccountsRemoved(t *testing.T) {
	c := raidCfg()
	c.JoinLimit, c.MinAccountAge = 500, 7*24*time.Hour
	v := setup(t, c)
	young := v.join(1, time.Hour)
	grown := v.join(2, 30*24*time.Hour)
	v.drain()
	if v.fake.IsMember(gs, young) || !v.fake.IsMember(gs, grown) {
		t.Fatal("age check wrong")
	}
}

func TestDisabledDoesNothing(t *testing.T) {
	c := raidCfg()
	c.Enabled = false
	v := setup(t, c)
	for i := range 50 {
		v.join(i, time.Minute)
	}
	v.drain()
	if v.fake.Calls("kick") != 0 || v.log.count("Raid detected") != 0 {
		t.Fatal("disabled raid protection acted")
	}
}

func TestThousandJoinFlood(t *testing.T) {
	c := raidCfg()
	c.JoinLimit = 50
	v := setup(t, c)
	for i := range 1000 {
		v.join(i, old)
	}
	v.drain()
	if v.fake.Calls("kick") != 1000 || v.log.count("Raid detected") != 1 {
		t.Fatalf("kicks %d alerts %d", v.fake.Calls("kick"), v.log.count("Raid detected"))
	}
	v.x.Sweep()
}
