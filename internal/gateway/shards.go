// Package gateway owns the Discord websocket sessions: sharding, intents and event handoff.
package gateway

import (
	"errors"
	"fmt"
	"slices"
)

// Plan is the shard layout for this instance.
type Plan struct {
	Count   int
	Batches [][]int // shards identified together, 5 seconds apart
}

// PlanShards picks the shard count and splits this instance's shards into identify batches.
// configured 0 means use Discord's recommendation. ids nil means run every shard.
func PlanShards(recommended, configured int, ids []int, maxConcurrency, remainingStarts int) (Plan, error) {
	count := configured
	if count == 0 {
		count = recommended
	}
	if count < 1 {
		return Plan{}, errors.New("gateway: shard count must be at least 1")
	}
	if ids == nil {
		ids = make([]int, count)
		for i := range ids {
			ids[i] = i
		}
	}
	ids = slices.Clone(ids)
	slices.Sort(ids)
	for i, id := range ids {
		if id < 0 || id >= count || (i > 0 && ids[i-1] == id) {
			return Plan{}, fmt.Errorf("gateway: shard id %d invalid for count %d", id, count)
		}
	}
	if len(ids) == 0 {
		return Plan{}, errors.New("gateway: no shards to run")
	}
	if remainingStarts < len(ids) {
		return Plan{}, fmt.Errorf("gateway: %d session starts left, need %d", remainingStarts, len(ids))
	}
	if maxConcurrency < 1 {
		maxConcurrency = 1
	}
	var batches [][]int
	for len(ids) > 0 {
		n := min(maxConcurrency, len(ids))
		batches = append(batches, ids[:n])
		ids = ids[n:]
	}
	return Plan{Count: count, Batches: batches}, nil
}
