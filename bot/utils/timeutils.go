package utils

import (
	"fmt"
	"time"
)

func FormatTime(interval time.Duration) string {
	minutes := (interval.Milliseconds() / (1000 * 60)) % 60
	hours := (interval.Milliseconds() / (1000 * 60 * 60)) % 24

	return fmt.Sprintf("%dh %02d", hours, minutes)
}
