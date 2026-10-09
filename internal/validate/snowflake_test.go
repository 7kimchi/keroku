package validate

import (
	"errors"
	"testing"
	"time"
)

func TestSnowflakeAccepts(t *testing.T) {
	for in, want := range map[string]int64{
		"1":                   1,
		"175928847299117063":  175928847299117063,
		"9223372036854775807": 9223372036854775807,
	} {
		got, err := Snowflake(in)
		if err != nil || got != want {
			t.Fatalf("%q: got %d %v", in, got, err)
		}
	}
}

func TestSnowflakeRejects(t *testing.T) {
	for _, in := range []string{
		"", "0", "00", "01", "-1", "+1", " 1", "1 ", "1.0", "1e5", "0x10", "12a",
		"9223372036854775808", "18446744073709551615", "99999999999999999999",
		"\U00000661\U00000662", "\U0000FF11\U0000FF12\U0000FF13", "1\x00", "<@123>", "123\n",
	} {
		if _, err := Snowflake(in); !errors.Is(err, ErrSnowflake) {
			t.Fatalf("%q: want ErrSnowflake, got %v", in, err)
		}
	}
}

func TestSnowflakeRoundTrip(t *testing.T) {
	const id = "175928847299117063"
	n, err := Snowflake(id)
	if err != nil {
		t.Fatal(err)
	}
	if FormatSnowflake(n) != id {
		t.Fatalf("round trip gave %s", FormatSnowflake(n))
	}
}

func TestSnowflakeTime(t *testing.T) {
	// Example from the Discord docs.
	got := SnowflakeTime(175928847299117063)
	want := time.Date(2016, 4, 30, 11, 18, 25, 796_000_000, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if !SnowflakeTime(1).Equal(time.UnixMilli(discordEpochMs).UTC()) {
		t.Fatal("tiny id should map to the epoch")
	}
}

func FuzzSnowflake(f *testing.F) {
	for _, s := range []string{"1", "0", "175928847299117063", "-5", "9223372036854775808"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		id, err := Snowflake(s)
		if err != nil {
			return
		}
		if id <= 0 || FormatSnowflake(id) != s {
			t.Fatalf("%q accepted as %d", s, id)
		}
	})
}
