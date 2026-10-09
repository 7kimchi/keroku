package discord

import (
	"net/http"
	"testing"
)

func TestGuildCountsRoute(t *testing.T) {
	var query string
	a, rest := newServer(t, func(w http.ResponseWriter, r *http.Request, _ int) {
		query = r.URL.RawQuery
		reply(w, 200, `{"id":"1","name":"g","approximate_member_count":42,"approximate_presence_count":7}`)
	})
	g, err := rest.GuildCounts(t.Context(), "1")
	if err != nil || g.ApproximateMemberCount != 42 || g.ApproximatePresenceCount != 7 {
		t.Fatalf("g %+v err %v", g, err)
	}
	if l := a.last(); l.method != "GET" || l.path != "/api/v9/guilds/1" || query != "with_counts=true" {
		t.Fatalf("%s %s ? %s", l.method, l.path, query)
	}
}

func TestServerCount(t *testing.T) {
	cases := map[string]int{
		`{"id":"1","approximate_guild_count":1234}`: 1234,
		`{"id":"1"}`: 0,
		`{"id":"1","approximate_guild_count":-5}`: 0,
	}
	for body, want := range cases {
		a, rest := newServer(t, func(w http.ResponseWriter, _ *http.Request, _ int) { reply(w, 200, body) })
		n, err := rest.ServerCount(t.Context())
		if err != nil || n != want {
			t.Fatalf("%s: n %d err %v", body, n, err)
		}
		if l := a.last(); l.path != "/api/v9/applications/@me" {
			t.Fatalf("path %s", l.path)
		}
	}
}

func TestServerCountMalformed(t *testing.T) {
	for _, body := range []string{`not json`, `{"approximate_guild_count":"many"}`, `[]`} {
		_, rest := newServer(t, func(w http.ResponseWriter, _ *http.Request, _ int) { reply(w, 200, body) })
		if _, err := rest.ServerCount(t.Context()); err == nil {
			t.Fatalf("%s accepted", body)
		}
	}
}

// Both calls are reads, so 5xx and 429 are retried and a 404 is reported as such.
func TestInfoCallsRetryAndClassify(t *testing.T) {
	a, rest := newServer(t, func(w http.ResponseWriter, _ *http.Request, hit int) {
		switch hit {
		case 1:
			reply(w, 503, `{}`)
		case 2:
			reply(w, 429, `{"message":"slow down","retry_after":0.05,"global":false}`)
		default:
			reply(w, 200, `{"id":"1","approximate_member_count":3}`)
		}
	})
	if g, err := rest.GuildCounts(t.Context(), "1"); err != nil || g.ApproximateMemberCount != 3 || a.count() != 3 {
		t.Fatalf("g %+v err %v hits %d", g, err, a.count())
	}
	_, rest = newServer(t, func(w http.ResponseWriter, _ *http.Request, _ int) {
		reply(w, 404, `{"message":"Unknown Guild","code":10004}`)
	})
	if _, err := rest.GuildCounts(t.Context(), "1"); !Is(err, NotFound) {
		t.Fatal(err)
	}
	_, rest = newServer(t, func(w http.ResponseWriter, _ *http.Request, _ int) { reply(w, 500, `{}`) })
	if _, err := rest.ServerCount(t.Context()); !Is(err, Unavailable) {
		t.Fatal(err)
	}
}
