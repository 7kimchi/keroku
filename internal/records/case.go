package records

import (
	"context"
	"errors"
	"math"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/embeds"
	"github.com/7kimchi/keroku/internal/moderation"
	"github.com/7kimchi/keroku/internal/validate"
)

// CaseCommand is /case with view and reason subcommands.
type CaseCommand struct{ D Deps }

// Definition describes /case.
func (CaseCommand) Definition() *discordgo.ApplicationCommand {
	reason := &discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionString, Name: "reason",
		Description: "New reason", Required: true, MaxLength: validate.MaxReason}
	return definition("case", "View a case or change its reason",
		&discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionSubCommand, Name: "view",
			Description: "Show one case", Options: []*discordgo.ApplicationCommandOption{numberOption()}},
		&discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionSubCommand, Name: "reason",
			Description: "Change a case's reason. The old reason is kept in the edit history",
			Options:     []*discordgo.ApplicationCommandOption{numberOption(), reason}})
}

// Handle runs /case.
func (c CaseCommand) Handle(ctx context.Context, r *commands.Request) (*discordgo.MessageEmbed, error) {
	gid, _ := validate.Snowflake(r.GuildID)
	number, ok, err := r.Int("case", 1, math.MaxInt32)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, commands.Fail("Case failed", "A case number is required.")
	}
	switch r.Sub {
	case "view":
		return c.view(ctx, gid, number)
	case "reason":
		return c.reason(ctx, r, gid, number)
	}
	return nil, commands.Fail("Case failed", "Unknown subcommand.")
}

func (c CaseCommand) view(ctx context.Context, gid, number int64) (*discordgo.MessageEmbed, error) {
	got, err := cases.Get(ctx, c.D.Store.Pool(), gid, number)
	if errors.Is(err, cases.ErrNotFound) {
		return nil, commands.Fail("Case not found", "No case "+strconv.FormatInt(number, 10)+" in this server.")
	}
	if err != nil {
		return nil, err
	}
	e := moderation.CaseEmbed(got, c.D.BotID)
	edits, err := cases.Edits(ctx, c.D.Store.Pool(), gid, got.ID)
	if err != nil {
		return nil, err
	}
	if n := len(edits); n > 0 {
		last := edits[n-1]
		e.Fields = append(e.Fields, &discordgo.MessageEmbedField{Name: "Edits",
			Value: strconv.Itoa(n) + ", last by " + embeds.User(validate.FormatSnowflake(last.EditorID)) + " " + embeds.Relative(last.CreatedAt)})
	}
	return e, nil
}

func (c CaseCommand) reason(ctx context.Context, r *commands.Request, gid, number int64) (*discordgo.MessageEmbed, error) {
	text, ok, err := r.Text("reason", validate.MaxReason)
	if err != nil {
		return nil, err
	}
	if !ok || text == "" {
		return nil, commands.Fail("Reason update failed", "A reason is required.")
	}
	editor, _ := validate.Snowflake(r.UserID)
	var updated cases.Case
	err = c.D.Store.InTx(ctx, func(tx pgx.Tx) error {
		updated, err = cases.EditReason(ctx, tx, gid, number, editor, text, r.InteractionID())
		return err
	})
	switch {
	case errors.Is(err, cases.ErrNotFound):
		return nil, commands.Fail("Case not found", "No case "+strconv.FormatInt(number, 10)+" in this server.")
	case errors.Is(err, cases.ErrDuplicate):
		return embeds.New("No change").Description("This edit was already applied.").Build(), nil
	case err != nil:
		return nil, err
	}
	c.D.Modlog.Case(gid, embeds.New("Reason updated").Field("Case", strconv.FormatInt(number, 10), true).
		Field("Moderator", embeds.User(r.UserID), true).Field("Reason", embeds.Escape(text), false).Build())
	return moderation.CaseEmbed(updated, c.D.BotID), nil
}
