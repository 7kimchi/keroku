package config

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

const maxShards = 4096

// shards reads SHARD_COUNT (0 means use the gateway recommendation) and SHARD_IDS, a comma
// list of ids or ranges like "0-3,8". Running a subset requires a fixed count on every instance.
func shards(src Source) (int, []int, error) {
	count, err := intVar(src, "SHARD_COUNT", 0, 0, maxShards)
	if err != nil {
		return 0, nil, err
	}
	raw := strings.TrimSpace(src.Getenv("SHARD_IDS"))
	if raw == "" {
		return count, nil, nil
	}
	if count == 0 {
		return 0, nil, errors.New("config: SHARD_IDS needs SHARD_COUNT set")
	}
	var ids []int
	for _, part := range strings.Split(raw, ",") {
		lo, hi, err := shardRange(strings.TrimSpace(part), count)
		if err != nil {
			return 0, nil, err
		}
		for id := lo; id <= hi; id++ {
			ids = append(ids, id)
		}
	}
	slices.Sort(ids)
	if len(slices.Compact(slices.Clone(ids))) != len(ids) {
		return 0, nil, errors.New("config: SHARD_IDS lists a shard twice")
	}
	return count, ids, nil
}

func shardRange(part string, count int) (int, int, error) {
	bad := fmt.Errorf("config: SHARD_IDS entry %q must be an id or range below %d", part, count)
	loRaw, hiRaw, isRange := strings.Cut(part, "-")
	lo, err := shardID(loRaw, count)
	if err != nil {
		return 0, 0, bad
	}
	if !isRange {
		return lo, lo, nil
	}
	hi, err := shardID(hiRaw, count)
	if err != nil || hi < lo {
		return 0, 0, bad
	}
	return lo, hi, nil
}

func shardID(raw string, count int) (int, error) {
	if raw == "" || (len(raw) > 1 && raw[0] == '0') {
		return 0, errors.New("bad id")
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 || n >= count {
		return 0, errors.New("bad id")
	}
	return n, nil
}
