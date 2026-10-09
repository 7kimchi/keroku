package validate

import (
	"strings"
	"time"
)

const maxDurationInput = 32

// Duration parses input like "1d4h" or "30m". Units are w, d, h, m and s, largest first,
// each at most once. The result must be within [minimum, maximum].
func Duration(s string, minimum, maximum time.Duration) (time.Duration, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" || len(s) > maxDurationInput || maximum < 0 {
		return 0, ErrDuration
	}
	var total time.Duration
	nextRank := 0
	for len(s) > 0 {
		digits := 0
		for digits < len(s) && s[digits] >= '0' && s[digits] <= '9' {
			digits++
		}
		if digits == 0 || digits == len(s) {
			return 0, ErrDuration
		}
		rank, size := unit(s[digits])
		if rank < nextRank {
			return 0, ErrDuration
		}
		var n time.Duration
		for _, c := range s[:digits] {
			n = n*10 + time.Duration(c-'0')
			// Stop before anything can overflow: past maximum is already a range error.
			if n > maximum/size {
				return 0, ErrRange
			}
		}
		total += n * size
		if total > maximum {
			return 0, ErrRange
		}
		nextRank = rank + 1
		s = s[digits+1:]
	}
	if total < minimum {
		return 0, ErrRange
	}
	return total, nil
}

// unit returns the order rank and size of a duration suffix. Unknown suffixes rank -1.
func unit(c byte) (int, time.Duration) {
	switch c {
	case 'w':
		return 0, 7 * 24 * time.Hour
	case 'd':
		return 1, 24 * time.Hour
	case 'h':
		return 2, time.Hour
	case 'm':
		return 3, time.Minute
	case 's':
		return 4, time.Second
	}
	return -1, time.Second
}
