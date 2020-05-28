package util

import (
	"time"
)

// GetDayOfMonthUTC to return the day of month in UTC from timestamp
func GetDayOfMonthUTC(timestamp int64) int {
	return time.Unix(timestamp, 0).UTC().Day()
}
