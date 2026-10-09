package automod

import (
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/cases"
)

// An automod timeout records which rule fired.
func TestTimeoutCaseRecordsRule(t *testing.T) {
	c := cfg()
	c.Timeout = 10 * time.Minute
	v := setup(t, c)
	v.post(spammer, "https://evil.com")
	v.drain()
	list, err := cases.ListByKind(t.Context(), v.store.Pool(), 100000000000000001, 100000000000000007, cases.Timeout, 5)
	if err != nil || len(list) != 1 {
		t.Fatalf("%d cases %v", len(list), err)
	}
	d := list[0].Details
	if d.Source != cases.FromAutomod || d.Rule == "" || !d.Valid() {
		t.Fatalf("%+v", d)
	}
}
