package gatewaytest

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestHandshakeBroadcastAndDrop(t *testing.T) {
	s := New(t, 2, `"t":"X","d":{}`)
	ws, resp, err := websocket.DefaultDialer.Dial(s.wsURL(), nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	defer func() { _ = ws.Close() }()
	read := func() string {
		_ = ws.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, b, err := ws.ReadMessage()
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	if !strings.Contains(read(), `"op":10`) {
		t.Fatal("no hello")
	}
	_ = ws.WriteMessage(websocket.TextMessage, []byte(`{"op":2,"d":{"token":"t","shard":[0,2]}}`))
	if !strings.Contains(read(), "READY") || !strings.Contains(read(), `"t":"X"`) {
		t.Fatal("no ready or startup event")
	}
	for len(s.Identified()) == 0 {
		time.Sleep(time.Millisecond)
	}
	if s.Identified()[0]["token"] != "t" {
		t.Fatal("identify not recorded")
	}
	s.Broadcast(`"t":"Y","d":{}`)
	if !strings.Contains(read(), `"t":"Y"`) {
		t.Fatal("broadcast missing")
	}
	s.DropAll()
	_ = ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, _, err := ws.ReadMessage(); err == nil {
		t.Fatal("socket still open after drop")
	}
}

func TestRESTRoutesAndReject(t *testing.T) {
	s := New(t, 3)
	s.REST = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) })
	get := func(path string) int {
		req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://discord.example"+path, nil)
		resp, err := s.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		return resp.StatusCode
	}
	if get("/api/v9/gateway/bot") != 200 || get("/api/v9/guilds/1") != http.StatusTeapot {
		t.Fatal("routing wrong")
	}
	s.Reject = true
	if get("/api/v9/gateway/bot") != http.StatusUnauthorized {
		t.Fatal("reject ignored")
	}
}
