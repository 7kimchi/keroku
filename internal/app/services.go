package app

import (
	"time"

	"github.com/7kimchi/keroku/internal/automod"
	"github.com/7kimchi/keroku/internal/cleanup"
	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/eventlog"
	"github.com/7kimchi/keroku/internal/info"
	"github.com/7kimchi/keroku/internal/moderation"
	"github.com/7kimchi/keroku/internal/raid"
	"github.com/7kimchi/keroku/internal/records"
	"github.com/7kimchi/keroku/internal/settings"
	"github.com/7kimchi/keroku/internal/store"
	"github.com/7kimchi/keroku/internal/sweeper"
	"github.com/7kimchi/keroku/internal/validate"
)

// services builds the feature services shared by commands, automod, raid and the sweeper.
func (a *App) services() error {
	var err error
	if a.automodSettings, err = store.NewCached(a.store.Automod, maxCachedGuilds, time.Minute); err != nil {
		return err
	}
	if a.raidSettings, err = store.NewCached(a.store.Raid, maxCachedGuilds, time.Minute); err != nil {
		return err
	}
	botID, _ := validate.Snowflake(a.botID)
	a.mod = moderation.New(moderation.Deps{Store: a.store, Client: a.client, Modlog: a.modlog,
		Metrics: a.metrics, Log: a.log, Guard: a.guard, BotID: a.botID})
	a.clean = cleanup.New(a.store, a.client, a.modlog, nil)
	a.sweeper = sweeper.New(sweeper.Deps{Store: a.store, Client: a.client, Moderation: a.mod, Cleanup: a.clean,
		Modlog: a.modlog, Metrics: a.metrics, Log: a.log, BotID: botID})
	if a.automod, err = automod.New(automod.Deps{Settings: a.automodSettings, Client: a.client, Moderation: a.mod,
		Modlog: a.modlog, Pool: a.events, Metrics: a.metrics, Log: a.log, BotID: botID}); err != nil {
		return err
	}
	if a.raid, err = raid.New(raid.Deps{Settings: a.raidSettings, Moderation: a.mod, Cleanup: a.clean,
		Modlog: a.modlog, Pool: a.events, Metrics: a.metrics, Log: a.log, BotID: botID}); err != nil {
		return err
	}
	a.eventlog, err = eventlog.New(a.modlog, a.settings, a.events, a.metrics, nil)
	return err
}

// commandList returns every slash command.
func (a *App) commandList() []commands.Command {
	mod, clean := a.mod, a.clean
	rec := records.Deps{Store: a.store, Modlog: a.modlog, BotID: a.botID}
	cfg := settings.Deps{Store: a.store, Settings: a.settings, Automod: a.automodSettings, Raid: a.raidSettings,
		Client: a.client, BotID: a.botID}
	look := info.Deps{Client: a.client, Started: a.started, Version: info.Version()}
	return []commands.Command{
		moderation.BanCommand{S: mod}, moderation.UnbanCommand{S: mod}, moderation.KickCommand{S: mod},
		moderation.TimeoutCommand{S: mod}, moderation.UntimeoutCommand{S: mod}, moderation.WarnCommand{S: mod},
		moderation.NoteCommand{S: mod},
		records.CaseCommand{D: rec}, records.HistoryCommand{D: rec}, records.WarningsCommand{D: rec},
		cleanup.PurgeCommand{S: clean}, cleanup.SlowmodeCommand{S: clean}, cleanup.LockCommand{S: clean},
		cleanup.UnlockCommand{S: clean}, cleanup.LockdownCommand{S: clean},
		settings.ConfigCommand{D: cfg},
		info.ServerInfoCommand{D: look}, info.BotInfoCommand{D: look}, info.UserInfoCommand{D: look},
		info.RoleInfoCommand{D: look},
	}
}
