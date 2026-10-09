package info

import (
	"testing"

	"github.com/7kimchi/keroku/internal/commands"
	"github.com/7kimchi/keroku/internal/perms"
)

func all() []commands.Command {
	return []commands.Command{ServerInfoCommand{}, BotInfoCommand{}, UserInfoCommand{}, RoleInfoCommand{}}
}

func TestDefinitionsRegister(t *testing.T) {
	reg, err := commands.NewRegistry(all()...)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, d := range reg.Definitions() {
		names[d.Name] = true
		if d.DefaultMemberPermissions == nil || *d.DefaultMemberPermissions != perms.ViewChannel {
			t.Fatalf("%s permission %v", d.Name, d.DefaultMemberPermissions)
		}
	}
	for _, n := range []string{"serverinfo", "botinfo", "userinfo", "roleinfo"} {
		if !names[n] {
			t.Fatalf("%s missing", n)
		}
	}
}

// Info commands carry no lane override, so they order like any read.
func TestInfoCommandsUseDefaultLane(t *testing.T) {
	for _, c := range all() {
		if _, ok := c.(commands.Laned); ok {
			t.Fatalf("%T overrides its lane", c)
		}
	}
}
