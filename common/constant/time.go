package constant

import "time"

var (
	// OneHour 1 hour in seconds
	OneHour = int64(time.Hour.Seconds())
	// OneDay 1 day in seconds
	OneDay = 24 * OneHour
	// OneYear 1 year in seconds
	OneYear = 365 * OneDay
)
