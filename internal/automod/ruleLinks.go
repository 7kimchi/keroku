package automod

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/7kimchi/keroku/internal/store"
)

// maxLinks bounds the work per message.
const maxLinks = 20

// links fires when a message has a link whose host is not on the allowlist. A host is
// allowed when it equals an allowed domain or is a subdomain of one.
func links(cfg store.AutomodSettings, content string, pattern *regexp.Regexp) bool {
	if !cfg.LinksEnabled || content == "" {
		return false
	}
	for _, raw := range pattern.FindAllString(content, maxLinks) {
		u, err := url.Parse(raw)
		if err != nil {
			return true
		}
		if !allowed(strings.TrimSuffix(strings.ToLower(u.Hostname()), "."), cfg.AllowedDomains) {
			return true
		}
	}
	return false
}

func allowed(host string, domains []string) bool {
	if host == "" {
		return false
	}
	for _, d := range domains {
		if host == d || strings.HasSuffix(host, "."+d) {
			return true
		}
	}
	return false
}

// linkPattern matches http and https links, including ones wrapped in <> or markdown.
func linkPattern() *regexp.Regexp {
	return regexp.MustCompile(`(?i)https?://[^\s<>()\[\]"']+`)
}
