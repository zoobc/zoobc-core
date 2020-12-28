// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
package query

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	// CaseQueryInterface is interface includes `func` that handle Operand
	// CaseQueryInterface for complex query builder
	CaseQueryInterface interface {
		Select(tableName string, columns ...string) *CaseQuery
		Where(query ...string) *CaseQuery
		And(expression ...string) *CaseQuery
		Or(expression ...string) *CaseQuery
		AndOr(expression ...string) *CaseQuery
		In(column string, value ...interface{}) string
		NotIn(column string, value ...interface{}) string
		Equal(column string, value interface{}) string
		NotEqual(column string, value interface{}) string
		Between(column string, start, end interface{}) string
		NotBetween(column string, start, end interface{}) string
		GroupBy(column ...string) *CaseQuery
		OrderBy(column string, order model.OrderBy) *CaseQuery
		Limit(limit uint32) *CaseQuery
		Paginate(limit, currentPage uint32) *CaseQuery
		QueryString() string
		Build() (query string, args []interface{})
		SubBuild() (query string, args []interface{})
		As(alias string) *CaseQuery
	}
	// CaseQuery would be as swap `Query` and `Args` those can save the query and args values
	CaseQuery struct {
		Query *bytes.Buffer
		Args  []interface{}
		Alias string
	}
)

// NewCaseQuery initiate New `CaseQuery`
func NewCaseQuery() CaseQuery {
	return CaseQuery{
		Query: bytes.NewBuffer([]byte{}),
	}
}

// Select build buffer query string
func (fq *CaseQuery) Select(tableName string, columns ...string) *CaseQuery {
	fq.Query.WriteString("SELECT ")
	if columns != nil {
		fq.Query.WriteString(strings.Join(columns, ", "))
	} else {
		fq.Query.WriteString("*")
	}
	fq.Query.WriteString(fmt.Sprintf(" FROM %s ", tableName))
	return fq
}

// Where build buffer query string, can combine with `In(), NotIn() ...`
func (fq *CaseQuery) Where(query ...string) *CaseQuery {
	sliceWhere := strings.Split(fq.Query.String(), "WHERE")
	if !strings.Contains(fq.Query.String(), "WHERE") ||
		regexp.MustCompile(`AS|FROM|JOIN`).MatchString(sliceWhere[len(sliceWhere)-1]) {
		fq.Query.WriteString(fmt.Sprintf(
			"WHERE %s ",
			strings.Join(query, ""),
		))
	} else {
		fq.And(query...)
	}
	return fq
}

// And represents `expressionFoo AND expressionBar`
func (fq *CaseQuery) And(query ...string) *CaseQuery {
	if !strings.Contains(fq.Query.String(), "WHERE") {
		fq.Query.WriteString("WHERE 1=1 ")
	}
	fq.Query.WriteString(fmt.Sprintf("AND %s", strings.Join(query, "AND ")))
	return fq
}

// Or represents `expressionFoo OR expressionBar`
func (fq *CaseQuery) Or(expression ...string) *CaseQuery {
	if !strings.Contains(fq.Query.String(), "WHERE") {
		fq.Query.WriteString("WHERE 1=1 ")
	}

	fq.Query.WriteString(fmt.Sprintf("OR %s ", strings.Join(expression, "OR ")))
	return fq
}

// AndOr represents `AND (expressionFoo OR expressionBar)`
func (fq *CaseQuery) AndOr(expression ...string) *CaseQuery {
	if !strings.Contains(fq.Query.String(), "WHERE") {
		fq.Query.WriteString("WHERE 1=1 ")
	}
	fq.Query.WriteString(fmt.Sprintf("AND (%s) ", strings.Join(expression, " OR ")))
	return fq
}

// In represents `column` IN (value...)
func (fq *CaseQuery) In(column string, value ...interface{}) string {
	fq.Args = append(fq.Args, value...)
	return fmt.Sprintf("%s IN (?%s) ", column, strings.Repeat(", ?", len(value)-1))
}

// NotIn represents `column NOT IN (value...)`
func (fq *CaseQuery) NotIn(column string, value ...interface{}) string {
	fq.Args = append(fq.Args, value...)
	return fmt.Sprintf("%s NOT IN (?%s) ", column, strings.Repeat(", ?", len(value)-1))
}

// Equal represents `column` == `value`
func (fq *CaseQuery) Equal(column string, value interface{}) string {
	fq.Args = append(fq.Args, value)
	return fmt.Sprintf("%s = ? ", column)
}

// NotEqual represents `column` != `value`
func (fq *CaseQuery) NotEqual(column string, value interface{}) string {
	fq.Args = append(fq.Args, value)
	return fmt.Sprintf("%s <> ? ", column)
}

// GreaterEqual represents `column >= value`
func (fq *CaseQuery) GreaterEqual(column string, value interface{}) string {
	fq.Args = append(fq.Args, value)
	return fmt.Sprintf("%s >= ? ", column)
}

// LessEqual represents `column <= value`
func (fq *CaseQuery) LessEqual(column string, value interface{}) string {
	fq.Args = append(fq.Args, value)
	return fmt.Sprintf("%s <= ? ", column)
}

// Between represents `column BETWEEN foo AND bar`
func (fq *CaseQuery) Between(column string, start, end interface{}) string {
	fq.Args = append(fq.Args, start, end)
	return fmt.Sprintf("%s BETWEEN ? AND ? ", column)
}

// NotBetween represents `column NOT BETWEEN foo AND bar`
func (fq *CaseQuery) NotBetween(column string, start, end interface{}) string {
	fq.Args = append(fq.Args, start, end)
	return fmt.Sprintf("%s NOT BETWEEN ? AND ? ", column)
}

// OrderBy represents `... ORDER BY column DESC|ASC`
func (fq *CaseQuery) OrderBy(column string, order model.OrderBy) *CaseQuery {
	// manual sanitizing, prepare-statement don't work on column name
	valid := regexp.MustCompile("^[A-Za-z0-9_]+$") // only number & lower+upper case and underscore
	if !valid.MatchString(column) {
		// invalid column name, do not proceed in order to prevent SQL injection
		return fq
	}
	fq.Query.WriteString(fmt.Sprintf("ORDER BY %s %s ", column, order.String()))
	return fq
}

// GroupBy represents `... GROUP BY column, column ...`
func (fq *CaseQuery) GroupBy(column ...string) *CaseQuery {
	fq.Query.WriteString(fmt.Sprintf("GROUP BY %s ", strings.Join(column, ", ")))
	return fq
}

// Limit represents `... LIMIT ...`
func (fq *CaseQuery) Limit(limit uint32) *CaseQuery {
	if limit == 0 {
		limit = 1
	}
	fq.Query.WriteString("limit ? ")
	fq.Args = append(fq.Args, limit)
	return fq
}

// Paginate represents `limit = ? offset = ?`
// default limit = 30, page start from 1
func (fq *CaseQuery) Paginate(limit, currentPage uint32) *CaseQuery {
	if limit == 0 {
		limit = 30
	}
	if currentPage == 0 {
		currentPage = 1
	}
	offset := (currentPage - 1) * limit
	fq.Query.WriteString("limit ? offset ? ")
	fq.Args = append(fq.Args, limit, offset)
	return fq
}

// QueryString allow to get buffer as string, sub query separated by comma
func (fq *CaseQuery) QueryString() string {
	var query = fq.Query.String()
	if len(fq.Alias) > 0 {
		query += fmt.Sprintf(" AS %s ", fq.Alias)
	}
	return query
}

// Build should be called in the end of `CaseQuery` circular.
// And build buffer query string into string, sub query separated by comma
func (fq *CaseQuery) Build() (query string, args []interface{}) {
	query = fq.Query.String()
	if len(fq.Alias) > 0 {
		query += fmt.Sprintf(" AS %s ", fq.Alias)
	}
	args = fq.Args
	return query, args
}

// SubBuild represents sub query builder without break the struct values,
// make sure call this method in separate declaration of CaseQuery
func (fq *CaseQuery) SubBuild() (query string, args []interface{}) {
	query = fmt.Sprintf("(%s)", fq.Query.String())
	if len(fq.Alias) > 0 {
		query += fmt.Sprintf("AS %s", fq.Alias)
	}
	return query, fq.Args
}

// As represents AS ..., and it will join with query string on Build
func (fq *CaseQuery) As(alias string) *CaseQuery {
	fq.Alias = alias
	return fq
}
