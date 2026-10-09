package cases

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDetailsValid(t *testing.T) {
	ok := []Details{{}, {Source: FromCommand, DeleteSeconds: 604800}, {Source: FromAutomod, Rule: strings.Repeat("r", MaxRule)},
		{Source: FromRaid}, {Source: FromEscalation, WarnCase: 9}, {Source: FromTimer, TimerID: 1}, {Rule: "mention spam"},
		{Rule: strings.Repeat("\u00e9", MaxRule)}}
	for _, d := range ok {
		if !d.Valid() {
			t.Fatalf("%+v refused", d)
		}
	}
	bad := []Details{{Source: "admin"}, {Source: "Command"}, {DeleteSeconds: -1}, {DeleteSeconds: 604801},
		{WarnCase: -1}, {TimerID: -5}, {Rule: strings.Repeat("r", MaxRule+1)}, {Rule: "a\x00b"}, {Rule: "a\nb"},
		{Rule: "\x7f"}, {Rule: "\xff\xfe"}}
	for _, d := range bad {
		if d.Valid() {
			t.Fatalf("%+v accepted", d)
		}
	}
}

// Keys are camelCase and empty fields are left out, so a command case stores a tiny object.
func TestDetailsJSON(t *testing.T) {
	b, _ := json.Marshal(Details{})
	if string(b) != "{}" {
		t.Fatalf("%s", b)
	}
	b, _ = json.Marshal(Details{Source: FromEscalation, WarnCase: 3, DeleteSeconds: 60, Rule: "x", TimerID: 4})
	if string(b) != `{"source":"escalation","deleteSeconds":60,"rule":"x","warnCase":3,"timerId":4}` {
		t.Fatalf("%s", b)
	}
}

// The largest valid details still fit the database's 1024 byte cap.
func TestDetailsMaxSizeFitsColumn(t *testing.T) {
	d := Details{Source: FromEscalation, DeleteSeconds: 604800, Rule: strings.Repeat("\U0001F600", MaxRule),
		WarnCase: 1 << 62, TimerID: 1 << 62}
	b, _ := json.Marshal(d)
	if !d.Valid() || len(b) > 1024 {
		t.Fatalf("valid %v len %d", d.Valid(), len(b))
	}
}

func FuzzDetails(f *testing.F) {
	f.Add("automod", "spam", 0, int64(1))
	f.Add("x", "\x00", -1, int64(-1))
	f.Fuzz(func(t *testing.T, src, rule string, del int, n int64) {
		d := Details{Source: Source(src), Rule: rule, DeleteSeconds: del, WarnCase: n, TimerID: n}
		if !d.Valid() {
			return
		}
		b, err := json.Marshal(d)
		if err != nil || len(b) > 1024 {
			t.Fatalf("len %d err %v", len(b), err)
		}
		var back Details
		if err := json.Unmarshal(b, &back); err != nil || back != d {
			t.Fatalf("%+v became %+v", d, back)
		}
	})
}
