package query

import (
	"bytes"
	"fmt"
	"regexp"
)

// GetTotalRecordOfSelect transform a select query that is used for getting data
// into one for getting total record
func GetTotalRecordOfSelect(selectQuery string, whereArgs map[string]interface{}) (str string, args []interface{}) {
	var (
		rgx  = regexp.MustCompile(`(?i)SELECT(.*)FROM`)
		buff *bytes.Buffer
		i    int
	)

	buff = bytes.NewBufferString(rgx.ReplaceAllLiteralString(selectQuery, `SELECT count() as total_record FROM`))

	buff.WriteString("WHERE")
	for k, v := range whereArgs {
		buff.WriteString(fmt.Sprintf("%s = ?", k))
		if i < len(whereArgs) {
			buff.WriteString(" AND")
		}
		i++
		args = append(args, v)
	}
	return buff.String(), args
}
