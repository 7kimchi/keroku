package embeds

import "testing"

// Discord trims field text and rejects what is left empty, so blank input must never reach it.
func TestBlankFieldsReplaced(t *testing.T) {
	for _, s := range []string{" ", "   ", "\t", "\n\n", " \r\n\t "} {
		e := New("x").Field(s, s, false).Build()
		if e.Fields[0].Name != "-" || e.Fields[0].Value != "None" {
			t.Fatalf("%q kept: %+v", s, e.Fields[0])
		}
	}
}

func TestPaddedFieldsKept(t *testing.T) {
	e := New("x").Field(" a ", " b ", false).Build()
	if e.Fields[0].Name != " a " || e.Fields[0].Value != " b " {
		t.Fatalf("%+v", e.Fields[0])
	}
}
