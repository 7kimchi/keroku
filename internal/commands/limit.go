package commands

import "time"

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
