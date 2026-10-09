package info

import (
	"runtime/debug"
	"testing"
)

func TestVersionFrom(t *testing.T) {
	vcs := func(kv ...string) *debug.BuildInfo {
		bi := &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}
		for i := 0; i < len(kv); i += 2 {
			bi.Settings = append(bi.Settings, debug.BuildSetting{Key: kv[i], Value: kv[i+1]})
		}
		return bi
	}
	cases := []struct {
		bi   *debug.BuildInfo
		ok   bool
		want string
	}{
		{nil, false, "dev"},
		{nil, true, "dev"},
		{&debug.BuildInfo{Main: debug.Module{Version: "v1.2.3"}}, true, "v1.2.3"},
		{vcs(), true, "dev"},
		{vcs("vcs.revision", "0123456789abcdef0123"), true, "0123456789ab"},
		{vcs("vcs.revision", "abc", "vcs.modified", "true"), true, "abc (modified)"},
		{vcs("vcs.revision", "abc", "vcs.modified", "false"), true, "abc"},
	}
	for _, c := range cases {
		if got := versionFrom(c.bi, c.ok); got != c.want {
			t.Fatalf("got %q want %q", got, c.want)
		}
	}
	if Version() == "" {
		t.Fatal("empty version")
	}
}
