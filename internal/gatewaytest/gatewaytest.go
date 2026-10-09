// Package gatewaytest is a fake Discord gateway and REST host for tests that open real sessions.
package gatewaytest

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

// Server answers /gateway, /gateway/bot and the websocket. Other REST paths go to REST.
type Server struct {
	srv        *httptest.Server
	shards     int
	events     []string
	REST       http.Handler
	Reject     bool // answer every gateway lookup with 401, like a revoked token
	mu         sync.Mutex
	identifies []map[string]any
	conns      []*conn
	seq        int
}

type conn struct {
	mu sync.Mutex
	ws *websocket.Conn
}

func (c *conn) write(b []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ws.WriteMessage(websocket.TextMessage, b)
}

// New starts a server recommending shards and sending events to each shard after READY.
// Events are the inner part of a dispatch, like `"t":"MESSAGE_CREATE","d":{...}`.
func New(t testing.TB, shards int, events ...string) *Server {
	t.Helper()
	s := &Server{shards: shards, events: events, REST: http.NotFoundHandler()}
	s.srv = httptest.NewServer(http.HandlerFunc(s.serve))
	t.Cleanup(s.srv.Close)
	return s
}

func (s *Server) wsURL() string { return "ws" + strings.TrimPrefix(s.srv.URL, "http") + "/ws" }

func (s *Server) serve(w http.ResponseWriter, r *http.Request) {
	switch {
	case s.Reject:
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"401: Unauthorized","code":0}`))
	case strings.HasSuffix(r.URL.Path, "/gateway/bot"):
		fmt.Fprintf(w, `{"url":%q,"shards":%d,"session_start_limit":{"total":1000,"remaining":1000,"max_concurrency":16}}`, s.wsURL(), s.shards)
	case strings.HasSuffix(r.URL.Path, "/gateway"):
		fmt.Fprintf(w, `{"url":%q}`, s.wsURL())
	case strings.HasPrefix(r.URL.Path, "/ws"):
		s.socket(w, r)
	default:
		s.REST.ServeHTTP(w, r)
	}
}

func (s *Server) socket(w http.ResponseWriter, r *http.Request) {
	ws, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer func() { _ = ws.Close() }()
	c := &conn{ws: ws}
	_ = c.write([]byte(`{"op":10,"d":{"heartbeat_interval":45000}}`))
	_, raw, err := ws.ReadMessage()
	if err != nil {
		return
	}
	var id struct {
		D map[string]any `json:"d"`
	}
	_ = json.Unmarshal(raw, &id)
	s.mu.Lock()
	s.identifies = append(s.identifies, id.D)
	s.conns = append(s.conns, c)
	s.mu.Unlock()
	_ = c.write(fmt.Appendf(nil, `{"op":0,"t":"READY","s":1,"d":{"v":10,"session_id":"s","resume_gateway_url":%q,"user":{"id":"999"},"guilds":[]}}`, s.wsURL()))
	for _, e := range s.events {
		_ = c.write(s.frame(e))
	}
	for {
		if _, _, err := ws.ReadMessage(); err != nil {
			return
		}
	}
}
