package raid

import (
	"context"
	"strconv"
	"time"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/embeds"
	"github.com/7kimchi/keroku/internal/moderation"
	"github.com/7kimchi/keroku/internal/store"
	"github.com/7kimchi/keroku/internal/validate"
)

func (x *Detector) handle(ctx context.Context, guildID, userID string, joinedAt time.Time) {
	gid, err := validate.Snowflake(guildID)
	if err != nil {
		return
	}
	if _, err := validate.Snowflake(userID); err != nil {
		return
	}
	cfg, err := x.d.Settings.Get(ctx, gid)
	if err != nil {
		x.d.Log.Warn("raid settings lookup failed", "guildId", gid, "err", err)
		return
	}
	if !cfg.Enabled {
		return
	}
	now := x.d.Now()
	if joinedAt.IsZero() {
		joinedAt = now
	}
	var v verdict
	x.windows.Update(guildID, func(w *window, found bool) *window {
		if !found || w == nil {
			w = &window{}
		}
		v = w.observe(userID, now, cfg)
		return w
	})
	if v.detected {
		x.alert(gid, cfg, len(v.targets))
		if cfg.Action == "lockdown" {
			x.lockdown(ctx, gid, cfg, now)
			return
		}
	}
	if len(v.targets) > 0 && cfg.Action != "lockdown" {
		for _, target := range v.targets {
			x.remove(ctx, gid, target, joinedAt, cfg.Action, "Raid protection: join burst.")
		}
		return
	}
	uid, _ := validate.Snowflake(userID)
	if cfg.MinAccountAge > 0 && now.Sub(validate.SnowflakeTime(uid)) < cfg.MinAccountAge {
		action := cfg.Action
		if action == "lockdown" {
			action = "kick"
		}
		x.remove(ctx, gid, userID, joinedAt, action, "Raid protection: account younger than "+embeds.Duration(cfg.MinAccountAge)+".")
	}
}

// remove kicks or bans one joiner through the normal moderation path. The join time is in
// the idempotency key, so a member who rejoins later is handled again.
func (x *Detector) remove(ctx context.Context, gid int64, userID string, joinedAt time.Time, action, reason string) {
	uid, _ := validate.Snowflake(userID)
	a := moderation.Action{Kind: cases.Kick, GuildID: gid, TargetID: uid, ModeratorID: x.d.BotID, Reason: reason,
		IdempotencyKey: "raid:" + userID + ":" + strconv.FormatInt(joinedAt.UnixNano(), 10), Automated: true}
	if action == "ban" {
		a.Kind, a.DeleteSeconds = cases.Ban, 3600
	}
	if _, err := x.d.Moderation.Execute(ctx, a); err != nil {
		x.d.Log.Warn("raid removal failed", "guildId", gid, "userId", userID, "err", err)
	}
}

func (x *Detector) lockdown(ctx context.Context, gid int64, cfg store.RaidSettings, now time.Time) {
	if err := x.d.Cleanup.Lockdown(ctx, gid, "Raid protection: join burst.", now.Add(cfg.LockdownFor)); err != nil {
		x.d.Log.Warn("raid lockdown failed", "guildId", gid, "err", err)
	}
}

func (x *Detector) alert(gid int64, cfg store.RaidSettings, joins int) {
	x.d.Metrics.Raids.Inc()
	x.d.Modlog.Case(gid, embeds.New("Raid detected").
		Field("Joins", strconv.Itoa(joins)+" in "+embeds.Duration(cfg.Window), true).
		Field("Action", cfg.Action, true).Timestamp(x.d.Now()).Build())
}
