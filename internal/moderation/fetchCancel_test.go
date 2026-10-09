package moderation

import (
	"context"
	"testing"
	"time"

	"github.com/7kimchi/keroku/internal/cases"
	"github.com/7kimchi/keroku/internal/discord"
)

// One failure cancels the slower calls instead of waiting them out.
func TestFetchFailureCancelsSiblings(t *testing.T) {
	e := setup(t)
	e.fake.FailNext("guild", errDown, 1)
	e.fake.SetDelay("member", 5*time.Second)
	start := time.Now()
	if _, err := e.svc.fetch(context.Background(), act(cases.Warn)); !discord.Is(err, discord.Unavailable) {
		t.Fatal(err)
	}
	if d := time.Since(start); d > time.Second {
		t.Fatalf("waited %v", d)
	}
}
