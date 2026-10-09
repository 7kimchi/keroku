package gateway

import (
	"net/http"
	"net/url"
)

type rewrite struct {
	target *url.URL
	next   http.RoundTripper
}

func (f *fakeGateway) client() *http.Client {
	target, _ := url.Parse(f.srv.URL)
	return &http.Client{Transport: rewrite{target: target, next: f.srv.Client().Transport}}
}

func (rw rewrite) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.URL.Scheme, r.URL.Host, r.Host = rw.target.Scheme, rw.target.Host, rw.target.Host
	return rw.next.RoundTrip(r)
}
