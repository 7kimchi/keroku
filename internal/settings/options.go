package settings

import (
	"time"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/bwmarrin/discordgo"
)

func intInto(r *commands.Request, name string, lo, hi int64, dst *int) error {
	n, ok, err := r.Int(name, lo, hi)
	if ok {
		*dst = int(n)
	}
	return err
}

func int64Into(r *commands.Request, name string, lo, hi int64, dst *int64) error {
	n, ok, err := r.Int(name, lo, hi)
	if ok {
		*dst = n
	}
	return err
}

func firstErr(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func windowed(r *commands.Request, countName string, count *int, window *time.Duration, maxCount, maxSeconds int64) error {
	if n, ok, err := r.Int(countName, 2, maxCount); err != nil {
		return err
	} else if ok {
		*count = int(n)
	}
	if s, ok, err := r.Int("seconds", 1, maxSeconds); err != nil {
		return err
	} else if ok {
		*window = time.Duration(s) * time.Second
	}
	return nil
}

func intOpt(name, desc string, lo, hi float64) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionInteger, Name: name,
		Description: desc, MinValue: &lo, MaxValue: hi}
}
