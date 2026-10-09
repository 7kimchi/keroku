package workers

import (
	"context"
	"strconv"
	"sync/atomic"
	"testing"
)

func TestPanicDropsOnlyThatJob(t *testing.T) {
	p := newPool(t, 2, 100)
	var ran atomic.Int64
	for i := range 100 {
		p.Submit(strconv.Itoa(i%3), func(context.Context) {
			if i%10 == 0 {
				panic("bad event")
			}
			ran.Add(1)
		})
	}
	_ = p.Close(context.Background())
	if ran.Load() != 90 {
		t.Fatalf("ran %d, want 90", ran.Load())
	}
}

func TestSubmitAfterCloseIsRefused(t *testing.T) {
	p := newPool(t, 2, 10)
	if err := p.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	if p.Submit("k", func(context.Context) {}) {
		t.Fatal("accepted after close")
	}
	if err := p.Close(context.Background()); err != nil {
		t.Fatal("second close failed")
	}
}
