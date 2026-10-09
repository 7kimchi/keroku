package discord

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestFakeInteractions(t *testing.T) {
	f, ctx := seeded(), t.Context()
	i := &discordgo.Interaction{ID: "i1"}
	e := &discordgo.MessageEmbed{Title: "x"}
	if !Is(f.EditResponse(ctx, i, e), NotFound) {
		t.Fatal("edit before respond accepted")
	}
	if err := f.Respond(ctx, i, &discordgo.InteractionResponse{Type: 5}); err != nil {
		t.Fatal(err)
	}
	if !Is(f.Respond(ctx, i, &discordgo.InteractionResponse{Type: 5}), Expired) {
		t.Fatal("second ack accepted")
	}
	if err := f.EditResponse(ctx, i, e); err != nil || len(f.Replies("i1")) != 1 {
		t.Fatalf("edit: %v", err)
	}
	f.ExpireInteraction("i1")
	if !Is(f.EditResponse(ctx, i, e), Expired) {
		t.Fatal("expired token accepted")
	}
}

func TestFakeChannelsAndDMs(t *testing.T) {
	f, ctx := seeded(), t.Context()
	e := &discordgo.MessageEmbed{Title: "x"}
	if err := f.Send(ctx, "c", e); err != nil || len(f.SentTo("c")) != 1 {
		t.Fatalf("send: %v", err)
	}
	f.DeleteChannel("c")
	if !Is(f.Send(ctx, "c", e), NotFound) {
		t.Fatal("send to deleted channel")
	}
	f.BlockDMs("u")
	if !Is(f.DM(ctx, "u", e), Forbidden) || f.DMCount() != 0 {
		t.Fatal("blocked DM delivered")
	}
	if err := f.DM(ctx, "v", e); err != nil || len(f.SentTo("dm:v")) != 1 {
		t.Fatalf("dm: %v", err)
	}
}
