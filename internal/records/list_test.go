package records

import (
	"strings"
	"testing"

	"github.com/7kimchi/keroku/internal/cases"
)

func TestHistoryPaging(t *testing.T) {
	d, _ := setup(t)
	for range 12 {
		seed(t, d, guild, cases.Warn, "w")
	}
	seed(t, d, 999, cases.Ban, "other guild")
	e, err := HistoryCommand{d}.Handle(t.Context(), request(t, "history", "", opt("user", tUser, ts)))
	if err != nil || strings.Count(e.Description, "Case ") != 10 || e.Footer.Text != "Page 1 of 2, 12 cases" {
		t.Fatalf("page 1: %+v %v", e, err)
	}
	e, _ = HistoryCommand{d}.Handle(t.Context(), request(t, "history", "", opt("user", tUser, ts), opt("page", tInt, 2.0)))
	if strings.Count(e.Description, "Case ") != 2 || !strings.Contains(e.Description, "Case 1,") {
		t.Fatalf("page 2: %s", e.Description)
	}
	e, _ = HistoryCommand{d}.Handle(t.Context(), request(t, "history", "", opt("user", tUser, ts), opt("page", tInt, 9.0)))
	if !strings.Contains(e.Description, "No cases on this page.") {
		t.Fatal("empty page")
	}
	if _, err := (HistoryCommand{d}).Handle(t.Context(), request(t, "history", "")); err == nil {
		t.Fatal("missing user accepted")
	}
}

func TestWarnings(t *testing.T) {
	d, _ := setup(t)
	e, _ := WarningsCommand{d}.Handle(t.Context(), request(t, "warnings", "", opt("user", tUser, ts)))
	if !strings.Contains(e.Description, "No warnings.") || e.Footer.Text != "0 warnings" {
		t.Fatalf("empty: %+v", e)
	}
	seed(t, d, guild, cases.Warn, "one")
	seed(t, d, guild, cases.Note, "not a warning")
	e, _ = WarningsCommand{d}.Handle(t.Context(), request(t, "warnings", "", opt("user", tUser, ts)))
	if e.Footer.Text != "1 warning" || strings.Contains(e.Description, "not a warning") {
		t.Fatalf("got %+v", e)
	}
	for range 11 {
		seed(t, d, guild, cases.Warn, "more")
	}
	e, _ = WarningsCommand{d}.Handle(t.Context(), request(t, "warnings", "", opt("user", tUser, ts)))
	if e.Footer.Text != "12 warnings, newest 10 shown" {
		t.Fatalf("footer %q", e.Footer.Text)
	}
}

func TestLineTruncatesHugeReasons(t *testing.T) {
	l := line(cases.Case{Number: 1, Kind: cases.Ban, Reason: strings.Repeat("a", 600)})
	if len(l) > 300 {
		t.Fatalf("line %d chars", len(l))
	}
	if !strings.Contains(line(cases.Case{Number: 2, Kind: cases.Kick}), "No reason given.") {
		t.Fatal("empty reason")
	}
}
