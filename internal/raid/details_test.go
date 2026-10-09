package raid

import (
	"testing"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/validate"
)

// Raid bans record their source and the message deletion window.
func TestRaidCasesRecordSource(t *testing.T) {
	c := raidCfg()
	c.Action = "ban"
	v := setup(t, c)
	var ids []string
	for i := range 3 {
		ids = append(ids, v.join(i, old))
	}
	v.drain()
	for _, id := range ids {
		uid, _ := validate.Snowflake(id)
		list, err := cases.ListByKind(t.Context(), v.st.Pool(), guild, uid, cases.Ban, 5)
		if err != nil || len(list) != 1 {
			t.Fatalf("%s: %d cases %v", id, len(list), err)
		}
		if d := list[0].Details; d != (cases.Details{Source: cases.FromRaid, DeleteSeconds: 3600}) {
			t.Fatalf("%+v", d)
		}
	}
}
