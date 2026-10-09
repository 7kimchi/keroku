package discord

import (
	"strconv"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
)

func TestFakeMessagesAndBulkDelete(t *testing.T) {
	f, ctx := seeded(), t.Context()
	now := time.Now()
	f.AddMessage("c", "old", "a", now.Add(-15*24*time.Hour))
	for i := range 10 {
		f.AddMessage("c", "m"+strconv.Itoa(i), "a", now)
	}
	got, err := f.Messages(ctx, "c", 3, "")
	if err != nil || len(got) != 3 || got[0].ID != "m9" {
		t.Fatalf("newest first: %v %v", got, err)
	}
	got, _ = f.Messages(ctx, "c", 100, "m5")
	if len(got) != 6 || got[0].ID != "m4" {
		t.Fatalf("before: %d first %s", len(got), got[0].ID)
	}
	if _, err := f.Messages(ctx, "c", 101, ""); !Is(err, BadRequest) {
		t.Fatal("limit over 100 accepted")
	}
	if !Is(f.BulkDelete(ctx, "c", []string{"m1"}, ""), BadRequest) {
		t.Fatal("single message bulk delete accepted")
	}
	if !Is(f.BulkDelete(ctx, "c", []string{"m1", "old"}, ""), BadRequest) || f.MessageCount("c") != 11 {
		t.Fatal("old message bulk delete accepted or partial")
	}
	if err := f.BulkDelete(ctx, "c", []string{"m1", "m2"}, ""); err != nil || f.MessageCount("c") != 9 {
		t.Fatalf("bulk delete: %v", err)
	}
	if err := f.DeleteMessage(ctx, "c", "old", ""); err != nil || !Is(f.DeleteMessage(ctx, "c", "old", ""), NotFound) {
		t.Fatal("single delete")
	}
}

func TestFakeOverwritesAndRoles(t *testing.T) {
	f, ctx := seeded(), t.Context()
	if err := f.SetRoleOverwrite(ctx, "c", "g", 0, 2048, ""); err != nil {
		t.Fatal(err)
	}
	_ = f.SetRoleOverwrite(ctx, "c", "g", 1, 2, "")
	ch, _ := f.Channel(ctx, "c")
	if len(ch.PermissionOverwrites) != 1 || ch.PermissionOverwrites[0].Deny != 2 {
		t.Fatalf("overwrite not replaced: %+v", ch.PermissionOverwrites)
	}
	_ = f.DeleteOverwrite(ctx, "c", "g", "")
	if ch, _ = f.Channel(ctx, "c"); len(ch.PermissionOverwrites) != 0 {
		t.Fatal("overwrite not deleted")
	}
	if err := f.SetRolePermissions(ctx, "g", "g", 7, ""); err != nil {
		t.Fatal(err)
	}
	if g, _ := f.Guild(ctx, "g"); g.Roles[0].Permissions != 7 {
		t.Fatal("role permissions not set")
	}
	if !Is(f.SetRolePermissions(ctx, "g", "nope", 7, ""), NotFound) || !Is(f.SetSlowmode(ctx, "c", 21601, ""), BadRequest) {
		t.Fatal("bad role or slowmode accepted")
	}
}

func TestFakeAutoModAndCommands(t *testing.T) {
	f, ctx := seeded(), t.Context()
	if err := f.CreateAutoModRule(ctx, "g", &discordgo.AutoModerationRule{Name: "a"}, ""); err != nil {
		t.Fatal(err)
	}
	rules, _ := f.AutoModRules(ctx, "g")
	if len(rules) != 1 || rules[0].ID == "" {
		t.Fatal("rule not stored")
	}
	if err := f.EditAutoModRule(ctx, "g", rules[0].ID, &discordgo.AutoModerationRule{Name: "b"}, ""); err != nil {
		t.Fatal(err)
	}
	if err := f.DeleteAutoModRule(ctx, "g", rules[0].ID, ""); err != nil || !Is(f.DeleteAutoModRule(ctx, "g", rules[0].ID, ""), NotFound) {
		t.Fatal("delete")
	}
	_ = f.OverwriteCommands(ctx, "app", []*discordgo.ApplicationCommand{{Name: "ban"}})
	if names := f.Commands(); len(names) != 1 || names[0] != "ban" {
		t.Fatalf("commands %v", names)
	}
}
