package settings

import (
	"strings"
	"testing"

	"github.com/7kimchi/keroku/internal/commands"
)

func TestDefinitionRegisters(t *testing.T) {
	if _, err := commands.NewRegistry(ConfigCommand{}); err != nil {
		t.Fatal(err)
	}
}

func TestSetAndClearChannels(t *testing.T) {
	c, _ := setup(t)
	ctx := t.Context()
	e, err := c.Handle(ctx, request(t, []string{"modlog"}, opt("channel", tChan, "100000000000000010")))
	if err != nil || e.Title != "Modlog channel set" {
		t.Fatalf("%+v %v", e, err)
	}
	if logs, err := c.Handle(ctx, request(t, []string{"logs"}, opt("channel", tChan, "100000000000000010"))); err != nil || logs.Title != "Log channel set" {
		t.Fatal(err)
	}
	view, _ := c.Handle(ctx, request(t, []string{"view"}))
	if view.Fields[0].Value != "<#100000000000000010>" || view.Fields[1].Value != "<#100000000000000010>" || view.Fields[2].Value != "Off" {
		t.Fatalf("view %+v", view.Fields)
	}
	if e, _ = c.Handle(ctx, request(t, []string{"modlog"})); e.Title != "Modlog channel cleared" {
		t.Fatal("clear")
	}
	if view, _ = c.Handle(ctx, request(t, []string{"view"})); view.Fields[0].Value != "Off" {
		t.Fatal("cleared channel still shown")
	}
}

func TestChannelChecks(t *testing.T) {
	c, f := setup(t)
	for channel, want := range map[string]string{
		"100000000000000011": "Pick a text channel in this server.",
		"100000000000000012": "Keroku needs View Channel",
		"100000000000000013": "Keroku cannot see that channel.",
	} {
		_, err := c.Handle(t.Context(), request(t, []string{"modlog"}, opt("channel", tChan, channel)))
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: got %v", channel, err)
		}
	}
	if f.Calls("send") != 0 {
		t.Fatal("posted while configuring")
	}
}

func TestEscalationCommands(t *testing.T) {
	c, _ := setup(t)
	ctx := t.Context()
	add := func(opts ...string) error {
		args := request(t, []string{"escalation", "add"}, opt("warnings", tInt, 3.0), opt("action", tStr, opts[0]))
		if len(opts) > 1 {
			args = request(t, []string{"escalation", "add"}, opt("warnings", tInt, 3.0), opt("action", tStr, opts[0]), opt("duration", tStr, opts[1]))
		}
		_, err := c.Handle(ctx, args)
		return err
	}
	if err := add("timeout"); err == nil || !strings.Contains(err.Error(), "need a duration") {
		t.Fatalf("timeout without duration: %v", err)
	}
	if err := add("kick", "1h"); err == nil || !strings.Contains(err.Error(), "no duration") {
		t.Fatalf("kick with duration: %v", err)
	}
	if err := add("explode"); err == nil {
		t.Fatal("bad action accepted")
	}
	if err := add("timeout", "1d"); err != nil {
		t.Fatal(err)
	}
	view, _ := c.Handle(ctx, request(t, []string{"view"}))
	if view.Fields[2].Value != "3 warnings: timeout for 1d" {
		t.Fatalf("view %q", view.Fields[2].Value)
	}
	if _, err := c.Handle(ctx, request(t, []string{"escalation", "remove"}, opt("warnings", tInt, 3.0))); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Handle(ctx, request(t, []string{"escalation", "remove"}, opt("warnings", tInt, 3.0))); err == nil {
		t.Fatal("removed twice")
	}
	if _, err := c.Handle(ctx, request(t, []string{"nope"})); err == nil {
		t.Fatal("unknown subcommand")
	}
}
