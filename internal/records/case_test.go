package records

import (
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/commands"
)

const tInt, tStr, tUser = discordgo.ApplicationCommandOptionInteger, discordgo.ApplicationCommandOptionString, discordgo.ApplicationCommandOptionUser

func TestCaseView(t *testing.T) {
	d, _ := setup(t)
	seed(t, d, guild, cases.Ban, "spam")
	e, err := CaseCommand{d}.Handle(t.Context(), request(t, "case", "view", opt("case", tInt, 1.0)))
	if err != nil || e.Title != "Member banned" || e.Footer.Text != "Case 1" {
		t.Fatalf("got %+v %v", e, err)
	}
	if _, err := (CaseCommand{d}).Handle(t.Context(), request(t, "case", "view", opt("case", tInt, 2.0))); err == nil ||
		!strings.Contains(err.Error(), "No case 2") {
		t.Fatalf("missing case: %v", err)
	}
}

func TestCaseViewOtherGuild(t *testing.T) {
	d, _ := setup(t)
	seed(t, d, 999, cases.Ban, "secret")
	_, err := CaseCommand{d}.Handle(t.Context(), request(t, "case", "view", opt("case", tInt, 1.0)))
	if err == nil {
		t.Fatal("read a case from another guild")
	}
}

func TestCaseReason(t *testing.T) {
	d, ml := setup(t)
	seed(t, d, guild, cases.Warn, "old")
	r := request(t, "case", "reason", opt("case", tInt, 1.0), opt("reason", tStr, "**new** [x](y)"))
	e, err := CaseCommand{d}.Handle(t.Context(), r)
	if err != nil || len(ml.sent) != 1 || ml.sent[0].Title != "Reason updated" {
		t.Fatalf("got %+v %v", e, err)
	}
	again, err := CaseCommand{d}.Handle(t.Context(), r)
	if err != nil || again.Title != "No change" || len(ml.sent) != 1 {
		t.Fatalf("replay: %+v %v", again, err)
	}
	view, _ := CaseCommand{d}.Handle(t.Context(), request(t, "case", "view", opt("case", tInt, 1.0)))
	last := view.Fields[len(view.Fields)-1]
	if last.Name != "Edits" || !strings.HasPrefix(last.Value, "1, last by") {
		t.Fatalf("edits field %+v", last)
	}
	for _, f := range view.Fields {
		if f.Name == "Reason" && strings.Contains(f.Value, "**new**") {
			t.Fatal("reason not escaped")
		}
	}
}

func TestCaseInputErrors(t *testing.T) {
	d, _ := setup(t)
	for _, r := range []*commands.Request{
		request(t, "case", "view"),
		request(t, "case", "view", opt("case", tInt, 0.0)),
		request(t, "case", "reason", opt("case", tInt, 1.0)),
		request(t, "case", "reason", opt("case", tInt, 1.0), opt("reason", tStr, "   ")),
		request(t, "case", "other", opt("case", tInt, 1.0)),
		request(t, "case", "reason", opt("case", tInt, 5.0), opt("reason", tStr, "x")),
	} {
		if _, err := (CaseCommand{d}).Handle(t.Context(), r); err == nil {
			t.Fatalf("accepted %s", r.Sub)
		}
	}
}
