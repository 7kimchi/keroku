package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// Respond accepts one initial response per interaction.
func (f *Fake) Respond(ctx context.Context, i *discordgo.Interaction, resp *discordgo.InteractionResponse) error {
	if err := f.enter(ctx, "respond"); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.expired[i.ID] {
		return &Error{Op: "respond", Kind: Expired, Status: 404, Code: 10062}
	}
	if f.responded[i.ID] {
		return &Error{Op: "respond", Kind: Expired, Status: 400, Code: 40060}
	}
	f.responded[i.ID] = true
	if resp.Data != nil && len(resp.Data.Embeds) > 0 {
		f.edits = append(f.edits, Sent{To: i.ID, Embed: resp.Data.Embeds[0]})
	}
	return nil
}

// EditResponse records the final reply. It fails before the initial response, like Discord.
func (f *Fake) EditResponse(ctx context.Context, i *discordgo.Interaction, embed *discordgo.MessageEmbed) error {
	if err := f.enter(ctx, "editResponse"); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.expired[i.ID] {
		return &Error{Op: "editResponse", Kind: Expired, Status: 404, Code: 10015}
	}
	if !f.responded[i.ID] {
		return notFound("editResponse", 10008)
	}
	f.edits = append(f.edits, Sent{To: i.ID, Embed: embed})
	return nil
}

// OverwriteCommands stores the registered commands.
func (f *Fake) OverwriteCommands(ctx context.Context, _ string, cmds []*discordgo.ApplicationCommand) error {
	if err := f.enter(ctx, "overwriteCommands"); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.commands = cmds
	return nil
}
