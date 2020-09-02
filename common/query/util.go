package query

import "regexp"

// GetTotalRecordOfSelect transacform a select query that is used for getting data
// into one for getting total record
func GetTotalRecordOfSelect(selectQuery string) string {
	re := regexp.MustCompile(`(?i)SELECT(.*)FROM`)
	return "SELECT count() as total_record FROM (" + re.ReplaceAllString(selectQuery, `SELECT NULL FROM`) + ")"
}
