// Package workers runs jobs on a fixed set of lanes. Jobs with the same key run in order on
// one lane, unrelated keys run in parallel, and full lanes refuse work instead of growing.
package workers

import (
	"context"
	"errors"
	"hash/maphash"
	"sync"

	"github.com/7kimchi/keroku/internal/safe"
)

// Job is one unit of work. ctx is cancelled if shutdown runs out of time.
type Job func(ctx context.Context)

// Pool is a fixed number of lanes, each with a bounded queue and one worker.
type Pool struct {
	name   string
	seed   maphash.Seed
	lanes  []chan Job
	guard  *safe.Guard
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
	closed bool
	wg     sync.WaitGroup
}

// New starts lanes workers, each accepting up to queue waiting jobs.
func New(name string, lanes, queue int, guard *safe.Guard) (*Pool, error) {
	if lanes < 1 || queue < 1 {
		return nil, errors.New("workers: lanes and queue must be positive")
	}
	ctx, cancel := context.WithCancel(context.Background())
	p := &Pool{name: name, seed: maphash.MakeSeed(), lanes: make([]chan Job, lanes), guard: guard, ctx: ctx, cancel: cancel}
	for i := range p.lanes {
		lane := make(chan Job, queue)
		p.lanes[i] = lane
		p.wg.Go(func() {
			for job := range lane {
				p.guard.Run(p.name, func() { job(p.ctx) })
			}
		})
	}
	return p, nil
}

// Submit queues job on key's lane without blocking. It reports false when the lane is full
// or the pool is closed.
func (p *Pool) Submit(key string, job Job) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.closed {
		return false
	}
	lane := p.lanes[maphash.String(p.seed, key)%uint64(len(p.lanes))]
	select {
	case lane <- job:
		return true
	default:
		return false
	}
}

// Depth counts queued jobs across lanes.
func (p *Pool) Depth() int {
	n := 0
	for _, lane := range p.lanes {
		n += len(lane)
	}
	return n
}
