package records

import (
	"context"
	"strconv"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/embeds"
	"github.com/7kimchi/keroku/internal/validate"
)

// WarningsCommand is /warnings.
type WarningsCommand struct{ D Deps }

// Definition describes /warnings.
func (WarningsCommand) Definition() *discordgo.ApplicationCommand {
	return definition("warnings", "List a member's warnings", userOption("Member to look up"))
}

// Handle runs /warnings.
func (c WarningsCommand) Handle(ctx context.Context, r *commands.Request) (*discordgo.MessageEmbed, error) {
	target, ok, err := r.User("user")
	if err != nil || !ok {
		return nil, orFail(err, "Warnings failed", "Pick a member.")
	}
	gid, _ := validate.Snowflake(r.GuildID)
	tid, _ := validate.Snowflake(target)
	total, err := cases.CountByKind(ctx, c.D.Store.Pool(), gid, tid, cases.Warn)
	if err != nil {
		return nil, err
	}
	list, err := cases.ListByKind(ctx, c.D.Store.Pool(), gid, tid, cases.Warn, pageSize)
	if err != nil {
		return nil, err
	}
	body := lines(list)
	if body == "" {
		body = "No warnings."
	}
	footer := count(total, "warning")
	if total > len(list) {
		footer += ", newest " + strconv.Itoa(len(list)) + " shown"
	}
	return embeds.New("Warnings").Description(embeds.User(target) + "\n\n" + body).Footer(footer).Build(), nil
}
