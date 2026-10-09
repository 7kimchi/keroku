package workers

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCloseDrainsQueuedWork(t *testing.T) {
	p := newPool(t, 4, 100)
	var ran atomic.Int64
	for i := range 300 {
		p.Submit(strconv.Itoa(i), func(context.Context) {
			time.Sleep(time.Millisecond)
			ran.Add(1)
		})
	}
	if err := p.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	if ran.Load() != 300 {
		t.Fatalf("drained %d of 300", ran.Load())
	}
}

func TestCloseDeadlineCancelsRunningJobs(t *testing.T) {
	p := newPool(t, 2, 10)
	var cancelled atomic.Int64
	for i := range 2 {
		p.Submit(strconv.Itoa(i), func(ctx context.Context) {
			select {
			case <-ctx.Done():
				cancelled.Add(1)
			case <-time.After(time.Minute):
			}
		})
	}
	time.Sleep(20 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	if err := p.Close(ctx); err == nil {
		t.Fatal("deadline overrun not reported")
	}
	if time.Since(start) > 2*time.Second || cancelled.Load() == 0 {
		t.Fatalf("took %v, cancelled %d", time.Since(start), cancelled.Load())
	}
}

func TestConcurrentSubmitAndClose(t *testing.T) {
	p := newPool(t, 8, 64)
	var wg sync.WaitGroup
	var accepted, ran atomic.Int64
	for w := range 50 {
		wg.Go(func() {
			for i := range 200 {
				if p.Submit(strconv.Itoa(w*1000+i), func(context.Context) { ran.Add(1) }) {
					accepted.Add(1)
				}
			}
		})
	}
	time.Sleep(time.Millisecond)
	_ = p.Close(context.Background())
	wg.Wait()
	if ran.Load() != accepted.Load() {
		t.Fatalf("accepted %d but ran %d", accepted.Load(), ran.Load())
	}
}

func TestThousandsOfGuildsInParallel(t *testing.T) {
	p := newPool(t, 32, 1000)
	var ran atomic.Int64
	start := time.Now()
	for g := range 5000 {
		for range 4 {
			for !p.Submit("guild"+strconv.Itoa(g), func(context.Context) { ran.Add(1) }) {
				time.Sleep(time.Microsecond)
			}
		}
	}
	_ = p.Close(context.Background())
	if ran.Load() != 20000 || time.Since(start) > 10*time.Second {
		t.Fatalf("ran %d in %v", ran.Load(), time.Since(start))
	}
}
