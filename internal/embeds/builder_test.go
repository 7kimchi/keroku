package embeds

import (
	"strings"
	"testing"
	"time"
)

func TestBuildBasic(t *testing.T) {
	ts := time.Date(2026, 1, 2, 3, 4, 5, 0, time.FixedZone("x", 3600))
	e := New("Member banned").Description("d").Field("Target", "x", true).Field("", "", false).
		Footer("Case 1").Timestamp(ts).Build()
	if e.Title != "Member banned" || e.Description != "d" || e.Color != Color {
		t.Fatalf("got %+v", e)
	}
	if len(e.Fields) != 2 || e.Fields[1].Name != "-" || e.Fields[1].Value != "None" || !e.Fields[0].Inline {
		t.Fatalf("fields %+v", e.Fields)
	}
	if e.Footer.Text != "Case 1" || e.Timestamp != "2026-01-02T02:04:05Z" {
		t.Fatalf("footer %v ts %s", e.Footer, e.Timestamp)
	}
	if e.Thumbnail != nil || e.Image != nil || e.Author != nil {
		t.Fatal("no images or authors allowed")
	}
}

func TestEmptyFooterIsOmitted(t *testing.T) {
	if New("x").Footer("").Build().Footer != nil {
		t.Fatal("empty footer kept")
	}
}

func TestBuildEnforcesEveryLimit(t *testing.T) {
	huge := strings.Repeat("\U0001F600", 10000)
	b := New(huge).Description(huge).Footer(huge)
	for range 40 {
		b.Field(huge, huge, false)
	}
	e := b.Build()
	if Length(e.Title) > MaxTitle || Length(e.Description) > MaxDescription || Length(e.Footer.Text) > MaxFooter {
		t.Fatal("per field limit broken")
	}
	if len(e.Fields) > MaxFields {
		t.Fatalf("%d fields", len(e.Fields))
	}
	for _, f := range e.Fields {
		if Length(f.Name) > MaxFieldName || Length(f.Value) > MaxFieldValue {
			t.Fatal("field limit broken")
		}
	}
	if total(e) > MaxTotal {
		t.Fatalf("total %d", total(e))
	}
}

func TestBuildDoesNotMutateBuilder(t *testing.T) {
	b := New(strings.Repeat("a", 300))
	b.Build()
	if Length(b.e.Title) != 300 {
		t.Fatal("builder state changed by Build")
	}
	e1, e2 := b.Build(), b.Build()
	if e1 == e2 {
		t.Fatal("Build returned shared embed")
	}
}

func TestFitTotalPrefersDescription(t *testing.T) {
	b := New("t").Description(strings.Repeat("d", 4000))
	for range 3 {
		b.Field("n", strings.Repeat("v", 1000), false)
	}
	e := b.Build()
	if len(e.Fields) != 3 || total(e) > MaxTotal {
		t.Fatalf("fields %d total %d", len(e.Fields), total(e))
	}
}

func FuzzBuild(f *testing.F) {
	f.Add("title", "desc", "name", "value", "footer", 3)
	f.Fuzz(func(t *testing.T, title, desc, name, value, footer string, n int) {
		b := New(title).Description(desc).Footer(footer)
		for range n % 40 {
			b.Field(name, value, n%2 == 0)
		}
		if e := b.Build(); total(e) > MaxTotal || len(e.Fields) > MaxFields {
			t.Fatal("limits broken")
		}
	})
}
