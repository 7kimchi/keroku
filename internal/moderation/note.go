package moderation

import (
	"context"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/perms"
	"github.com/7kimchi/keroku/internal/validate"
)

// NoteCommand is /note. Notes are private records: the member is never told.
type NoteCommand struct{ S *Service }

// Definition describes /note.
func (NoteCommand) Definition() *discordgo.ApplicationCommand {
	return definition("note", "Add a private note to a member's record", perms.ModerateMembers,
		userOption("Member the note is about"),
		&discordgo.ApplicationCommandOption{Type: discordgo.ApplicationCommandOptionString, Name: "reason",
			Description: "Note text, visible to moderators only", Required: true, MaxLength: validate.MaxReason})
}

// Handle runs /note.
func (c NoteCommand) Handle(ctx context.Context, r *commands.Request) (*discordgo.MessageEmbed, error) {
	return simple{s: c.S, kind: cases.Note, reasonRequired: true}.handle(ctx, r)
}
