package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"
)

func TestSecretFromFile(t *testing.T) {
	m := base()
	delete(m, "DISCORD_TOKEN")
	m["DISCORD_TOKEN_FILE"] = "/run/secrets/token"
	c, err := Load(env(m, map[string]string{"/run/secrets/token": fakeToken + "\r\n"}))
	if err != nil || c.Token.Reveal() != fakeToken {
		t.Fatalf("got %v", err)
	}
}

func TestSecretRules(t *testing.T) {
	m := base()
	m["DISCORD_TOKEN_FILE"] = "/x"
	mustFail(t, m, map[string]string{"/x": fakeToken})

	for _, bad := range []string{"", "Bot " + fakeToken, "a.b", "a..b", ".a.b", "a.b.", "a.b.c.d", "a.b.c\t", "a.b.\x00c", "a.b.c\U000000E9"} {
		mustFail(t, with("DISCORD_TOKEN", bad), nil)
	}
	mustFail(t, with("DISCORD_TOKEN", strings.Repeat("a", 5000)+".b.c"), nil)

	m = base()
	delete(m, "DATABASE_URL")
	mustFail(t, m, nil)

	m = base()
	delete(m, "DISCORD_TOKEN")
	m["DISCORD_TOKEN_FILE"] = "/big"
	mustFail(t, m, map[string]string{"/big": strings.Repeat("a", 4097)})
}

func TestSecretNeverPrints(t *testing.T) {
	s := NewSecret(fakeToken)
	type holder struct{ Token Secret }
	h := holder{s}
	outputs := []string{
		fmt.Sprint(s), fmt.Sprintf("%v %s %q %+v %#v %x", s, s, s, s, s, s),
		fmt.Sprintf("%v %+v %#v", h, h, h),
	}
	j, _ := json.Marshal(h)
	outputs = append(outputs, string(j))
	var b strings.Builder
	slog.New(slog.NewJSONHandler(&b, nil)).Info("x", "token", s, "holder", h)
	slog.New(slog.NewTextHandler(&b, nil)).Info("x", "token", s)
	outputs = append(outputs, b.String())
	for _, out := range outputs {
		if strings.Contains(out, "fake") || strings.Contains(out, "token.") {
			t.Fatalf("secret leaked: %s", out)
		}
	}
	if s.Empty() || !NewSecret("").Empty() {
		t.Fatal("Empty is wrong")
	}
}

func TestErrorsDoNotContainSecret(t *testing.T) {
	m := base()
	m["DISCORD_TOKEN"] = "not a token secretvalue"
	err := loadErr(t, m, nil)
	if strings.Contains(err.Error(), "secretvalue") {
		t.Fatalf("error leaks value: %v", err)
	}
}

func FuzzLoad(f *testing.F) {
	f.Add(fakeToken, "4", "0-3", "32", "127.0.0.1:1")
	f.Fuzz(func(t *testing.T, token, count, ids, workers, addr string) {
		// The marker keeps the leak check from matching ordinary words in messages.
		token = "zqk" + token
		m := map[string]string{"DISCORD_TOKEN": token, "DATABASE_URL": "postgres://x",
			"SHARD_COUNT": count, "SHARD_IDS": ids, "WORKERS": workers, "METRICS_ADDR": addr}
		c, err := Load(env(m, nil))
		if err != nil {
			if strings.Contains(err.Error(), token) {
				t.Fatalf("error echoes token: %v", err)
			}
			return
		}
		for _, id := range c.ShardIDs {
			if id < 0 || id >= c.ShardCount {
				t.Fatalf("shard %d outside count %d", id, c.ShardCount)
			}
		}
	})
}
