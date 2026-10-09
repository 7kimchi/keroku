package discord

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
)

const testToken = "fake.test.token"

// apiServer is an httptest server that speaks enough of the REST API for the client tests.
type apiServer struct {
	srv  *httptest.Server
	mu   sync.Mutex
	hits int
	reqs []recorded
	fn   func(w http.ResponseWriter, r *http.Request, hit int)
}

type recorded struct {
	method, path, reason, auth string
	body                       []byte
}

func newServer(t *testing.T, fn func(w http.ResponseWriter, r *http.Request, hit int)) (*apiServer, *REST) {
	t.Helper()
	a := &apiServer{fn: fn}
	a.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		a.mu.Lock()
		a.hits++
		hit := a.hits
		a.reqs = append(a.reqs, recorded{r.Method, r.URL.Path, r.Header.Get("X-Audit-Log-Reason"), r.Header.Get("Authorization"), body})
		a.mu.Unlock()
		a.fn(w, r, hit)
	}))
	t.Cleanup(a.srv.Close)
	s, err := discordgo.New("Bot " + testToken)
	if err != nil {
		t.Fatal(err)
	}
	target, _ := url.Parse(a.srv.URL)
	s.Client = &http.Client{Transport: rewrite{target: target, next: a.srv.Client().Transport}}
	rest := NewREST(s, 2*time.Second)
	rest.backoff = 10 * time.Millisecond
	return a, rest
}

func (a *apiServer) count() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.hits
}

func (a *apiServer) last() recorded {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.reqs[len(a.reqs)-1]
}

// rewrite sends every request to the test server, keeping the path.
type rewrite struct {
	target *url.URL
	next   http.RoundTripper
}

func (rw rewrite) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.URL.Scheme, r.URL.Host, r.Host = rw.target.Scheme, rw.target.Host, rw.target.Host
	return rw.next.RoundTrip(r)
}

func reply(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = io.WriteString(w, body)
}

func ok(w http.ResponseWriter, _ *http.Request, _ int) { reply(w, http.StatusOK, `{"id":"1"}`) }

func noContent(w http.ResponseWriter, _ *http.Request, _ int) { w.WriteHeader(http.StatusNoContent) }

func testEmbed() *discordgo.MessageEmbed { return &discordgo.MessageEmbed{Title: "x"} }
