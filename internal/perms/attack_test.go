package perms

import (
	"strconv"
	"testing"

	"github.com/bwmarrin/discordgo"
)

// A target holding thousands of roles, an invoker spoofing role ids, and huge guilds must
// not change the outcome or blow up.
func TestHierarchyAttacks(t *testing.T) {
	g := guild()
	for i := range 250 {
		g.Roles = append(g.Roles, &discordgo.Role{ID: "r" + strconv.Itoa(i), Position: 2})
	}
	r := baseRequest()
	r.Guild = g
	for i := range 5000 {
		r.TargetRoles = append(r.TargetRoles, "r"+strconv.Itoa(i))
	}
	if d := Check(r); d != nil {
		t.Fatalf("mod above position 2 denied: %s", d.Message())
	}
	r.InvokerRoles = []string{"admin-but-not-really", "ADMIN", " admin"}
	if code(Check(r)) != InvokerTooLow {
		t.Fatal("spoofed role ids passed")
	}
}
