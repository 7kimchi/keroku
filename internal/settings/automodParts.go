package settings

import (
	"strconv"
	"strings"
	"time"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/discord"
	"github.com/7kimchi/keroku/internal/embeds"
	"github.com/7kimchi/keroku/internal/store"
	"github.com/7kimchi/keroku/internal/validate"
)

// maxDomains matches the database limit on the allowlist.
const maxDomains = 50

func allowlist(r *commands.Request, cfg *store.AutomodSettings) error {
	raw, ok, err := r.Text("allow", 1500)
	if err != nil || !ok {
		return err
	}
	domains := []string{}
	for _, part := range strings.Split(raw, ",") {
		if strings.TrimSpace(part) == "" {
			continue
		}
		d, err := validate.Domain(part)
		if err != nil {
			return commands.Fail("Config failed", "Not a valid domain: "+embeds.Escape(embeds.Truncate(strings.TrimSpace(part), 100))+".")
		}
		domains = append(domains, d)
	}
	if len(domains) > maxDomains {
		return commands.Fail("Config failed", "At most "+strconv.Itoa(maxDomains)+" domains.")
	}
	cfg.AllowedDomains = domains
	return nil
}

func timeoutSetting(r *commands.Request, cfg *store.AutomodSettings) error {
	raw, _, err := r.Text("duration", 32)
	if err != nil {
		return err
	}
	if strings.EqualFold(raw, "off") {
		cfg.Timeout = 0
		return nil
	}
	d, err := validate.Duration(raw, time.Minute, 28*24*time.Hour)
	if err != nil {
		return commands.Fail("Config failed", "Duration must be like 10m or 1h, up to 28d, or off.")
	}
	cfg.Timeout = d
	return nil
}

func nativeFailure(err error) error {
	if discord.Is(err, discord.Forbidden) {
		return commands.Fail("Config failed", "Saved, but Discord AutoMod needs Keroku to have Manage Server.")
	}
	return commands.FailWith("Config failed", "Saved, but Discord AutoMod could not be updated.", err)
}

func automodSummary(c store.AutomodSettings) string {
	onOff := func(on bool, detail string) string {
		if !on {
			return "Off"
		}
		return detail
	}
	lines := []string{
		"Spam: " + onOff(c.SpamEnabled, strconv.Itoa(c.SpamMessages)+" messages in "+embeds.Duration(c.SpamWindow)),
		"Duplicates: " + onOff(c.DuplicateEnabled, strconv.Itoa(c.DuplicateCount)+" repeats in "+embeds.Duration(c.DuplicateWindow)),
		"Links: " + onOff(c.LinksEnabled, "allowed "+orNone(strings.Join(c.AllowedDomains, ", "))),
		"Mentions: " + onOff(c.MentionLimit > 0, strconv.Itoa(c.MentionLimit)+" per message"),
		"Invites: " + onOff(c.InvitesBlocked, "Blocked"),
		"Timeout: " + onOff(c.Timeout > 0, embeds.Duration(c.Timeout)),
	}
	return strings.Join(lines, "\n")
}

func orNone(s string) string {
	if s == "" {
		return "none"
	}
	return s
}
