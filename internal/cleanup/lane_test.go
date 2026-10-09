package cleanup

import (
	"testing"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/commands"
)

func chanOpt(v any) *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{Name: "channel", Type: discordgo.ApplicationCommandOptionChannel, Value: v}
}

func userFilter(v string) *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{Name: "user", Type: discordgo.ApplicationCommandOptionUser, Value: v}
}

const other = "100000000000000020"

// Channel commands order on the channel they change, defaulting to the one they ran in.
func TestChannelCommandLanes(t *testing.T) {
	for _, c := range []commands.Laned{LockCommand{}, UnlockCommand{}, SlowmodeCommand{}} {
		if got := c.Lane(request(t, "x", "")); got != "c"+cs {
			t.Fatalf("%T default %q", c, got)
		}
		if got := c.Lane(request(t, "x", "", chanOpt(other))); got != "c"+other {
			t.Fatalf("%T option %q", c, got)
		}
	}
}

// Lock and unlock on one channel share a lane, whether the channel was typed or implied.
func TestLockAndUnlockShareLane(t *testing.T) {
	implied := LockCommand{}.Lane(request(t, "lock", ""))
	typed := UnlockCommand{}.Lane(request(t, "unlock", "", chanOpt(cs)))
	if implied != typed {
		t.Fatalf("%q vs %q", implied, typed)
	}
}

// A malformed channel option falls back to the invoking channel instead of a hostile key.
func TestChannelLaneRejectsBadOption(t *testing.T) {
	for _, v := range []any{"abc", "", "0", "1/../2", 5, "99999999999999999999999"} {
		if got := (LockCommand{}).Lane(request(t, "lock", "", chanOpt(v))); got != "c"+cs {
			t.Fatalf("%v gave %q", v, got)
		}
	}
}

// Purge orders on the channel, so a filtered and an unfiltered purge never overlap.
func TestPurgeLane(t *testing.T) {
	plain := PurgeCommand{}.Lane(request(t, "purge", ""))
	filtered := PurgeCommand{}.Lane(request(t, "purge", "", userFilter("100000000000000005")))
	if plain != "c"+cs || plain != filtered {
		t.Fatalf("%q vs %q", plain, filtered)
	}
}

// Lockdown has no narrower key, it stays on the guild lane.
func TestLockdownHasNoLane(t *testing.T) {
	var c commands.Command = LockdownCommand{}
	if _, ok := c.(commands.Laned); ok {
		t.Fatal("lockdown should order per guild")
	}
}
