package moderation

import (
	"context"
	"errors"
	"sync"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/discord"
	"github.com/7kimchi/keroku/internal/validate"
)

var errFetchPanic = errors.New("moderation: fetch panicked")

// fetched is the Discord state an action is checked against.
type fetched struct {
	guild  *discordgo.Guild
	member *discordgo.Member // nil when the target is not in the guild
	bot    *discordgo.Member
	banned bool
	banErr error // only matters once the permission checks pass
}

// fetch reads the guild, the target, the bot and the ban state in parallel. Run one after
// another they were most of a command's latency.
func (s *Service) fetch(ctx context.Context, a Action) (fetched, error) {
	gid, tid := validate.FormatSnowflake(a.GuildID), validate.FormatSnowflake(a.TargetID)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var (
		f     fetched
		wg    sync.WaitGroup
		once  sync.Once
		first error
	)
	fail := func(err error) {
		once.Do(func() { first = err; cancel() })
	}
	run := func(where string, fn func() error) {
		wg.Go(func() {
			var err error
			if s.guard.Run(where, func() { err = fn() }) {
				err = errFetchPanic
			}
			if err != nil {
				fail(err)
			}
		})
	}
	run("fetch guild", func() (err error) {
		f.guild, err = s.client.Guild(ctx, gid)
		return err
	})
	run("fetch members", func() error {
		// Both members share one Discord rate limit bucket, so they go in order anyway.
		m, err := s.client.Member(ctx, gid, tid)
		switch {
		case err == nil:
			f.member = m
		case !discord.Is(err, discord.NotFound):
			return err
		}
		f.bot, err = s.client.Member(ctx, gid, s.botID)
		return err
	})
	if a.Kind == cases.Ban || a.Kind == cases.Unban {
		wg.Go(func() {
			if s.guard.Run("fetch ban", func() { f.banned, f.banErr = s.client.IsBanned(ctx, gid, tid) }) {
				f.banErr = errFetchPanic
			}
		})
	}
	wg.Wait()
	return f, first
}
