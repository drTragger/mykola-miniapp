package sysutil

import (
	"fmt"
	"time"
)

func HumanBytes(v uint64) string {
	const unit = 1024
	if v < unit {
		return fmt.Sprintf("%d B", v)
	}

	div, exp := uint64(unit), 0
	for n := v / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	value := float64(v) / float64(div)
	suffixes := []string{"KiB", "MiB", "GiB", "TiB"}
	return fmt.Sprintf("%.2f %s", value, suffixes[exp])
}

func HumanizeDurationShort(d time.Duration) string {
	if d < 0 {
		return "н/д"
	}

	seconds := int(d.Seconds())
	if seconds < 60 {
		return fmt.Sprintf("%dс", seconds)
	}

	minutes := seconds / 60
	if minutes < 60 {
		return fmt.Sprintf("%dхв", minutes)
	}

	hours := minutes / 60
	minutes = minutes % 60
	if hours < 24 {
		return fmt.Sprintf("%dг %dхв", hours, minutes)
	}

	days := hours / 24
	hours = hours % 24
	return fmt.Sprintf("%dд %dг", days, hours)
}
