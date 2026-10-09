package validate

import (
	"strconv"
	"time"
)

// discordEpochMs is the first second of 2015 in Unix milliseconds.
const discordEpochMs = 1420070400000

// Snowflake parses a Discord ID strictly: digits only, no sign, no leading zero, fits in int64.
func Snowflake(s string) (int64, error) {
	if len(s) == 0 || len(s) > 19 || s[0] == '0' {
		return 0, ErrSnowflake
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, ErrSnowflake
		}
	}
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		return 0, ErrSnowflake
	}
	return id, nil
}

// FormatSnowflake renders an ID the way Discord sends it.
func FormatSnowflake(id int64) string {
	return strconv.FormatInt(id, 10)
}

// SnowflakeTime returns the creation time encoded in an ID.
func SnowflakeTime(id int64) time.Time {
	return time.UnixMilli((id >> 22) + discordEpochMs).UTC()
}
