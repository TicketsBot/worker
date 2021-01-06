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

func FormatDateTime(time time.Time) string {
	zone, _ := time.Zone()
	return fmt.Sprintf("%d/%d/%d %d:%d (%s)", time.Day(), time.Month(), time.Year(), time.Hour(), time.Minute(), zone)
}
