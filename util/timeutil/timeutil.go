package timeutil

import (
	"fmt"
	"time"
)

// FormatBlockTime parses block time into format '2006-01-02 15:04:05'.
func FormatBlockTime(blockTime uint64) string {
	ts := int64(blockTime) / 1000
	t := time.Unix(ts, 0)

	return t.Format("2006-01-02 15:04:05")
}

// ParseSeconds returns human readable time format of the given seconds.
func ParseSeconds(totalSeconds uint64) string {
	var hours, minutes, seconds uint64 = 0, 0, 0

	if totalSeconds >= 3600 {
		hours = totalSeconds / 3600
		totalSeconds -= hours * 3600
	}
	if totalSeconds >= 60 {
		minutes = totalSeconds / 60
		totalSeconds -= minutes * 60
	}
	seconds = totalSeconds

	if hours > 0 {
		return fmt.Sprintf("%02dh %02dm %02ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%02dm %02ds", minutes, seconds)
	} else if seconds > 0 {
		return fmt.Sprintf("%02ds", seconds)
	}

	return ""
}
