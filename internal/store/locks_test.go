package store_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/7kimchi/keroku/internal/dbtest"
	"github.com/7kimchi/keroku/internal/store"
)

func TestChannelLocks(t *testing.T) {
	pool := dbtest.New(t)
	ctx := t.Context()
	const g, ch = int64(1300000000000000001), int64(1300000000000000002)
	l := store.ChannelLock{HadOverwrite: true, PreviousAllow: 1 << 40, PreviousDeny: 2048}
	if ok, err := store.AddChannelLock(ctx, pool, g, ch, l); !ok || err != nil {
		t.Fatal(ok, err)
	}
	if ok, _ := store.AddChannelLock(ctx, pool, g, ch, l); ok {
		t.Fatal("locked twice")
	}
	if _, ok, _ := store.TakeChannelLock(ctx, pool, g+1, ch); ok {
		t.Fatal("unlocked from another guild")
	}
	got, ok, err := store.TakeChannelLock(ctx, pool, g, ch)
	if !ok || err != nil || got != l {
		t.Fatalf("got %+v %v %v", got, ok, err)
	}
	if _, ok, _ := store.TakeChannelLock(ctx, pool, g, ch); ok {
		t.Fatal("unlocked twice")
	}
}

func TestLockdowns(t *testing.T) {
	pool := dbtest.New(t)
	ctx := t.Context()
	var won atomic.Int64
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			if ok, err := store.AddLockdown(ctx, pool, 5, 1<<50); ok && err == nil {
				won.Add(1)
			}
		})
	}
	wg.Wait()
	if won.Load() != 1 {
		t.Fatalf("%d lockdowns started", won.Load())
	}
	if _, ok, _ := store.TakeLockdown(ctx, pool, 6); ok {
		t.Fatal("ended another guild's lockdown")
	}
	if prev, ok, err := store.TakeLockdown(ctx, pool, 5); !ok || err != nil || prev != 1<<50 {
		t.Fatalf("got %d %v %v", prev, ok, err)
	}
}
