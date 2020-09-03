package constant

import "time"

var (
	// SQLMaxIdleConnections Represent number of maximum idle connetion in sql pool connection
	SQLMaxIdleConnections = 10
	// SQLMaxConnectionLifetime Reprensent the expiration of opened database connetion
	SQLMaxConnectionLifetime = 30 * time.Minute
	// SQLMaxOpenConnetion represent the number of maximum open connetion to the database
	SQLMaxOpenConnetion = 50
	// SQLiteLimitVariableNumber equivalent to SQLITE_LIMIT_VARIABLE_NUMBER from sqlite
	SQLiteLimitVariableNumber = 999
)
