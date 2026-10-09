package workers

import (
	"context"
	"errors"
	"sync"
)

// Close stops new submissions and waits for queued jobs. If ctx ends first, running jobs
// get a cancelled context and Close waits for them to return.
func (p *Pool) Close(ctx context.Context) error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	for _, lane := range p.lanes {
		close(lane)
	}
	p.mu.Unlock()
	done := make(chan struct{})
	var waiter sync.WaitGroup
	p.guard.Go(&waiter, p.name+"Close", func() {
		p.wg.Wait()
		close(done)
	})
	defer waiter.Wait()
	select {
	case <-done:
		p.cancel()
		return nil
	case <-ctx.Done():
		p.cancel()
		<-done
		return errors.New("workers: " + p.name + " drain deadline passed, remaining jobs were cancelled")
	}
}
