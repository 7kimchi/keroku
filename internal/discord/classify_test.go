package discord

import (
	"net/http"
	"strings"
	"testing"
)

func TestStatusAndCodeClassification(t *testing.T) {
	cases := []struct {
		status int
		body   string
		want   Kind
	}{
		{404, `{"code":10007,"message":"Unknown Member"}`, NotFound},
		{404, `{}`, NotFound},
		{403, `{"code":50013,"message":"Missing Permissions"}`, Forbidden},
		{403, `{"code":50007,"message":"Cannot send messages to this user"}`, Forbidden},
		{404, `{"code":10062,"message":"Unknown interaction"}`, Expired},
		{400, `{"code":40060,"message":"Interaction has already been acknowledged."}`, Expired},
		{400, `{"code":50034,"message":"too old"}`, BadRequest},
		{401, `{"code":0,"message":"401: Unauthorized"}`, Unauthorized},
		{400, `not json`, BadRequest},
	}
	for _, tc := range cases {
		_, rest := newServer(t, func(w http.ResponseWriter, _ *http.Request, _ int) { reply(w, tc.status, tc.body) })
		err := rest.Send(t.Context(), "1", testEmbed())
		if KindOf(err) != tc.want {
			t.Errorf("%d %s: got %v", tc.status, tc.body, err)
		}
	}
}

func TestIsBanned(t *testing.T) {
	_, rest := newServer(t, func(w http.ResponseWriter, r *http.Request, _ int) {
		if strings.HasSuffix(r.URL.Path, "/2") {
			reply(w, 200, `{"user":{"id":"2"}}`)
			return
		}
		reply(w, 404, `{"code":10026,"message":"Unknown Ban"}`)
	})
	if b, err := rest.IsBanned(t.Context(), "1", "2"); !b || err != nil {
		t.Fatalf("banned: %v %v", b, err)
	}
	if b, err := rest.IsBanned(t.Context(), "1", "3"); b || err != nil {
		t.Fatalf("not banned: %v %v", b, err)
	}
}
