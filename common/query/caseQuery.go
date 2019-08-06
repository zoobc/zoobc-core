package query

import (
	"bytes"
	"fmt"
	"strings"
)

type (
	// CaseQueryInterface is interface includes `func` that handle Operand
	// CaseQueryInterface for complex query builder
	CaseQueryInterface interface {
		Select(tableName string, columns ...string)
		Where(args map[string]interface{})
		Or(args map[string]interface{})
		Conjunct(firstSeparator Operand, args map[string]interface{})
		Done(limit uint32, offset uint64) (query string, args []interface{})
		// TODO: WhereIn, When etc... if needed
	}
	// CaseQuery would be as swap `Query` and `Args` those can save the query and args values
	// until `Done` called
	Operand   string
	CaseQuery struct {
		Query *bytes.Buffer
		Args  []interface{}
	}
)

const (
	OperandAnd Operand = " AND "
	OperandOr  Operand = " OR "
)

// Select build buffer query string
func (fq *CaseQuery) Select(tableName string, columns ...string) {
	fq.Query.WriteString(fmt.Sprintf(
		"SELECT %s FROM %s ",
		strings.Join(columns, ", "),
		tableName,
	))
}

// Where build buffer query string where operand
func (fq *CaseQuery) Where(args map[string]interface{}) {
	i := 0
	fq.Query.WriteString("WHERE ")
	for k, v := range args {
		fq.Query.WriteString(fmt.Sprintf("%s = ? ", k))
		if len(args)-1 > i {
			fq.Query.WriteString("AND ")
		}
		fq.Args = append(fq.Args, v)
		i++
	}
}

// Or build buffer query string `OR` operand
func (fq *CaseQuery) Or(args map[string]interface{}) {
	i := 0
	fq.Query.WriteString("OR ")
	for k, v := range args {
		fq.Query.WriteString(fmt.Sprintf("%s = ? ", k))
		if len(args)-1 > i {
			fq.Query.WriteString("AND ")
		}
		fq.Args = append(fq.Args, v)
		i++
	}
}

// Conjunct build buffer query string, firstSep would be as separator at first query
// and continue with AND
func (fq *CaseQuery) Conjunct(firstSep Operand, args map[string]interface{}) {
	fq.Query.WriteString(string(firstSep))
	i := 0
	for k, v := range args {
		fq.Query.WriteString(fmt.Sprintf("%s = ? ", k))
		if len(args)-1 > i {
			fq.Query.WriteString("AND ")
		}
		fq.Args = append(fq.Args, v)
		i++
	}
}

// Done should be called in the end of `CaseQuery` circular.
// And build buffer query string into string
func (fq *CaseQuery) Done(limit uint32, offset uint64) (query string, args []interface{}) {
	if limit <= 0 {
		limit = 30
	}
	fq.Query.WriteString("limit ? offset ? ")
	fq.Args = append(fq.Args, limit, offset)
	return fq.Query.String(), fq.Args
}
