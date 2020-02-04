package constant

import "time"

var (
	// SQLMaxIdleConnections Represent number of maximum idle connetion in sql pool connection
	SQLMaxIdleConnections = 0
	// SQLMaxConnectionLifetime Reprensent the expiration of idle database connetion
	SQLMaxConnectionLifetime = time.Microsecond
)
