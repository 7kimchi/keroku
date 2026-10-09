package store

import (
	"fmt"
	"strconv"
	"strings"
)

func versionOf(name string) (int, error) {
	end := 0
	for end < len(name) && name[end] >= '0' && name[end] <= '9' {
		end++
	}
	v, err := strconv.Atoi(name[:end])
	if end == 0 || err != nil || v <= 0 || !strings.HasSuffix(name, ".sql") {
		return 0, fmt.Errorf("migrate: %s has no version prefix", name)
	}
	return v, nil
}
