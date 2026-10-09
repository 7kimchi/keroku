package gateway

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/discord"
)

// InstallLogger routes discordgo's own logging into slog. discordgo only offers a package
// level hook, so this is called once at startup.
func InstallLogger(log *slog.Logger) {
	discordgo.Logger = func(level, _ int, format string, a ...any) {
		msg := fmt.Sprintf(format, a...)
		switch level {
		case discordgo.LogError:
			log.Error("discordgo", "detail", msg)
		case discordgo.LogWarning:
			log.Warn("discordgo", "detail", msg)
		default:
			log.Debug("discordgo", "detail", msg)
		}
	}
}

// errText describes an error without request details that could carry tokens.
func errText(err error) string {
	if k := discord.KindOf(err); k != discord.Unknown {
		return k.String()
	}
	var rest *discordgo.RESTError
	if errors.As(err, &rest) && rest.Response != nil {
		return "http status " + rest.Response.Status
	}
	return err.Error()
}
