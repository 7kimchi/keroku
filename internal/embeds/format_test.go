package embeds

import (
	"testing"
	"time"
)

func TestDuration(t *testing.T) {
	for d, want := range map[time.Duration]string{
		0:                            "0s",
		-time.Hour:                   "0s",
		500 * time.Millisecond:       "0s",
		time.Second:                  "1s",
		90 * time.Second:             "1m 30s",
		time.Hour:                    "1h",
		28 * time.Hour:               "1d 4h",
		28 * 24 * time.Hour:          "28d",
		24*time.Hour + 5*time.Minute: "1d",
		time.Hour + 30*time.Minute + 15*time.Second: "1h 30m",
		400 * 24 * time.Hour:                        "400d",
	} {
		if got := Duration(d); got != want {
			t.Errorf("%v: got %q want %q", d, got, want)
		}
	}
}

func TestMentionsAndTimestamps(t *testing.T) {
	if got := Relative(time.Unix(1700000000, 999)); got != "<t:1700000000:R>" {
		t.Fatalf("got %s", got)
	}
	if got := User("42"); got != "<@42> (42)" {
		t.Fatalf("got %s", got)
	}
	if got := Channel("7"); got != "<#7>" {
		t.Fatalf("got %s", got)
	}
}

func TestEscape(t *testing.T) {
	for in, want := range map[string]string{
		"plain":                  "plain",
		"**bold**":               `\*\*bold\*\*`,
		"[click](http://x.y)":    `\[click\]\(http://x.y\)`,
		"<@123> <@&9> <#1>":      `\<@123\> \<@&9\> \<\#1\>`,
		"`code` ||spoiler|| ~s~": "\\`code\\` \\|\\|spoiler\\|\\| \\~s\\~",
		`back\slash`:             `back\\slash`,
		"# head - item > quote":  `\# head \- item \> quote`,
		"__u__":                  `\_\_u\_\_`,
		"":                       "",
	} {
		if got := Escape(in); got != want {
			t.Errorf("Escape(%q) = %q want %q", in, got, want)
		}
	}
}
