package config

import "testing"

func TestShardIDs(t *testing.T) {
	mustFail(t, with("SHARD_IDS", "0"), nil)
	for _, bad := range []string{"4", "-1", "0,0", "0-1,1", "2-1", "a", "0-", "-", "", "1-4", "01", " , "} {
		m := with("SHARD_COUNT", "4")
		m["SHARD_IDS"] = bad
		if bad == "" {
			continue
		}
		mustFail(t, m, nil)
	}
	m := with("SHARD_COUNT", "4")
	m["SHARD_IDS"] = "3,0-1"
	c, err := Load(env(m, nil))
	if err != nil || len(c.ShardIDs) != 3 || c.ShardIDs[0] != 0 || c.ShardIDs[2] != 3 {
		t.Fatalf("got %v %v", c.ShardIDs, err)
	}
}

func TestShardIDsFullRange(t *testing.T) {
	m := with("SHARD_COUNT", "4096")
	m["SHARD_IDS"] = "0-4095"
	c, err := Load(env(m, nil))
	if err != nil || len(c.ShardIDs) != 4096 {
		t.Fatalf("got %d %v", len(c.ShardIDs), err)
	}
}
