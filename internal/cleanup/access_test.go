package cleanup

import (
	"strings"
	"testing"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/perms"
)

const tChan = 7 // channel option type

// Holding Manage Channels where the command runs must not grant it in another channel.
func TestOtherChannelNeedsPermissionThere(t *testing.T) {
	s, f, _ := setup(t)
	const other = "100000000000000020"
	f.AddChannel(gs, other)
	_ = f.SetRoleOverwrite(t.Context(), other, gs, 0, perms.ManageChannels, "")
	f.AddChannel("100000000000000099", "100000000000000021")
	cmds := []commands.Command{SlowmodeCommand{s}, LockCommand{s}, UnlockCommand{s}}
	names := []string{"slowmode", "lock", "unlock"}
	for i, c := range cmds {
		r := request(t, names[i], "", opt("channel", tChan, other), opt("interval", tStr, "5s"))
		r.Member.Roles = nil
		if _, err := c.Handle(t.Context(), r); err == nil || !strings.Contains(err.Error(), "You need Manage Channels in that channel.") {
			t.Fatalf("%s: got %v", names[i], err)
		}
		foreign := request(t, names[i], "", opt("channel", tChan, "100000000000000021"), opt("interval", tStr, "5s"))
		if _, err := c.Handle(t.Context(), foreign); err == nil || !strings.Contains(err.Error(), "not in this server") {
			t.Fatalf("%s foreign: got %v", names[i], err)
		}
	}
	if f.Calls("slowmode") != 0 || f.Calls("setOverwrite") != 1 {
		t.Fatal("acted in a channel without permission")
	}
}
