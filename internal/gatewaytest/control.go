package gatewaytest

import (
	"fmt"
	"net/http"
	"net/url"
)

// Broadcast sends one dispatch event to every connected shard.
func (s *Server) Broadcast(event string) {
	s.mu.Lock()
	conns := append([]*conn(nil), s.conns...)
	s.mu.Unlock()
	for _, c := range conns {
		_ = c.write(s.frame(event))
	}
}

// Identified returns the identify payloads received so far.
func (s *Server) Identified() []map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]map[string]any(nil), s.identifies...)
}

// DropAll closes every socket, like Discord dropping connections.
func (s *Server) DropAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, c := range s.conns {
		_ = c.ws.Close()
	}
	s.conns = nil
}

// Client returns an HTTP client that sends every request to this server.
func (s *Server) Client() *http.Client {
	target, _ := url.Parse(s.srv.URL)
	return &http.Client{Transport: rewrite{target: target, next: s.srv.Client().Transport}}
}

type rewrite struct {
	target *url.URL
	next   http.RoundTripper
}

func (rw rewrite) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.URL.Scheme, r.URL.Host, r.Host = rw.target.Scheme, rw.target.Host, rw.target.Host
	return rw.next.RoundTrip(r)
}

func (s *Server) frame(event string) []byte {
	s.mu.Lock()
	s.seq++
	n := s.seq + 1
	s.mu.Unlock()
	return fmt.Appendf(nil, `{"op":0,"s":%d,%s}`, n, event)
}
