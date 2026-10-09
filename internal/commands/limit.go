package commands

import (
	"time"

	"github.com/7kimchi/keroku/internal/perms"
)

func (x *Dispatcher) limited(r *Request) (time.Duration, bool) {
	if ok, wait := x.d.UserLimit.Allow("u" + r.UserID); !ok {
		x.d.Metrics.RateLimited.WithLabelValues("user").Inc()
		return wait, true
	}
	if ok, wait := x.d.GuildLimit.Allow("g" + r.GuildID); !ok {
		x.d.Metrics.RateLimited.WithLabelValues("guild").Inc()
		return wait, true
	}
	return 0, false
}

// missingPermission checks the invoker holds the command's default permission. Discord
// hides commands from members without it, but admins can override that and payloads are
// not trusted, so it is checked again here every time.
func missingPermission(cmd Command, r *Request) (string, bool) {
	need := cmd.Definition().DefaultMemberPermissions
	if need == nil || perms.Has(r.Permissions, *need) {
		return "", false
	}
	return perms.Name(*need), true
}
