package commands

import (
	"strconv"
	"time"

	"github.com/7kimchi/keroku/internal/embeds"
)

func itoa(n int64) string { return strconv.FormatInt(n, 10) }

func formatMax(d time.Duration) string { return embeds.Duration(d) }
