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

// HistoryCommand is /history.
type HistoryCommand struct{ D Deps }

// Definition describes /history.
func (HistoryCommand) Definition() *discordgo.ApplicationCommand {
	return definition("history", "List a user's cases, newest first", userOption("User to look up"),
		&discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionInteger, Name: "page",
			Description: "Page number, 10 cases per page", MinValue: ptr(1), MaxValue: 1000})
}

// Handle runs /history.
func (c HistoryCommand) Handle(ctx context.Context, r *commands.Request) (*discordgo.MessageEmbed, error) {
	target, ok, err := r.User("user")
	if err != nil || !ok {
		return nil, orFail(err, "History failed", "Pick a user.")
	}
	page, ok, err := r.Int("page", 1, 1000)
	if err != nil {
		return nil, err
	}
	if !ok {
		page = 1
	}
	gid, _ := validate.Snowflake(r.GuildID)
	tid, _ := validate.Snowflake(target)
	total, err := cases.Count(ctx, c.D.Store.Pool(), gid, tid)
	if err != nil {
		return nil, err
	}
	list, err := cases.History(ctx, c.D.Store.Pool(), gid, tid, pageSize, int(page-1)*pageSize)
	if err != nil {
		return nil, err
	}
	body := lines(list)
	if body == "" {
		body = "No cases on this page."
	}
	pages := max(1, (total+pageSize-1)/pageSize)
	return embeds.New("History").Description(embeds.User(target) + "\n\n" + body).
		Footer("Page " + strconv.FormatInt(page, 10) + " of " + strconv.Itoa(pages) + ", " + count(total, "case")).Build(), nil
}

func orFail(err error, title, detail string) error {
	if err != nil {
		return err
	}
	return commands.Fail(title, detail)
}
