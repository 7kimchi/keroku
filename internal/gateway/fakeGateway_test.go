package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
)

// fakeGateway speaks the handful of gateway and REST messages a session needs to connect.
type fakeGateway struct {
	srv        *httptest.Server
	shards     int
	mu         sync.Mutex
	identifies []map[string]any
	conns      []*websocket.Conn
	events     []string // raw dispatch payloads sent after READY
}

func newFakeGateway(t *testing.T, shards int, events ...string) *fakeGateway {
	t.Helper()
	f := &fakeGateway{shards: shards, events: events}
	f.srv = httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeGateway) wsURL() string { return "ws" + strings.TrimPrefix(f.srv.URL, "http") + "/ws" }

func (f *fakeGateway) serve(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasSuffix(r.URL.Path, "/gateway/bot"):
		fmt.Fprintf(w, `{"url":%q,"shards":%d,"session_start_limit":{"total":1000,"remaining":1000,"max_concurrency":2}}`, f.wsURL(), f.shards)
	case strings.HasSuffix(r.URL.Path, "/gateway"):
		fmt.Fprintf(w, `{"url":%q}`, f.wsURL())
	case strings.HasPrefix(r.URL.Path, "/ws"):
		f.socket(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (f *fakeGateway) socket(w http.ResponseWriter, r *http.Request) {
	up := websocket.Upgrader{}
	c, err := up.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	f.mu.Lock()
	f.conns = append(f.conns, c)
	f.mu.Unlock()
	defer func() { _ = c.Close() }()
	_ = c.WriteMessage(websocket.TextMessage, []byte(`{"op":10,"d":{"heartbeat_interval":45000}}`))
	_, raw, err := c.ReadMessage()
	if err != nil {
		return
	}
	var id struct {
		D map[string]any `json:"d"`
	}
	_ = json.Unmarshal(raw, &id)
	f.mu.Lock()
	f.identifies = append(f.identifies, id.D)
	f.mu.Unlock()
	ready := fmt.Sprintf(`{"op":0,"t":"READY","s":1,"d":{"v":10,"session_id":"s","resume_gateway_url":%q,"user":{"id":"999"},"guilds":[]}}`, f.wsURL())
	_ = c.WriteMessage(websocket.TextMessage, []byte(ready))
	for n, e := range f.events {
		_ = c.WriteMessage(websocket.TextMessage, fmt.Appendf(nil, `{"op":0,"s":%d,%s}`, n+2, e))
	}
	for {
		if _, _, err := c.ReadMessage(); err != nil {
			return
		}
	}
}

func (f *fakeGateway) identified() []map[string]any {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]map[string]any(nil), f.identifies...)
}

// dropAll closes every socket, like Discord dropping the connection.
func (f *fakeGateway) dropAll() {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, c := range f.conns {
		_ = c.Close()
	}
}
