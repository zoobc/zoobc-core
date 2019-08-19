package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	DatasetsQuery struct {
		PrimaryFields  []string
		OrdinaryFields []string
		TableName      string
	}

	DatasetsQueryInterface interface {
		GetLastDataset(accountSetter, accountRecipient, property string) (query string, args []interface{})
		GetDatasetsByAccountRecipient(accountRecipient string) (query string, args interface{})
		AddDataset(dataset *model.Dataset) [][]interface{}
		ExtractModel(dataset *model.Dataset) []interface{}
		BuildModel(datasets []*model.Dataset, rows *sql.Rows) []*model.Dataset
		BuildModelRow(dataset *model.Dataset, row *sql.Row) (*model.Dataset, error)
	}
)

// NewDatasetsQuery will create a new DatasetsQuery
func NewDatasetsQuery() *DatasetsQuery {
	return &DatasetsQuery{
		PrimaryFields: []string{
			"account_setter",
			"account_recipient",
			"property",
			"height",
		},
		OrdinaryFields: []string{
			"value",
			"timestamp_starts",
			"timestamp_expires",
			"latest",
		},
		TableName: "datasets",
	}
}

func (dq *DatasetsQuery) GetDatasetsByAccountRecipient(accountRecipient string) (query string, args interface{}) {
	return fmt.Sprintf("SELECT %s FROM %s WHERE account_recipient = ? AND latest = 1",
			strings.Join(dq.getFields(), ","),
			dq.getTableName(),
		),
		accountRecipient
}

func (dq *DatasetsQuery) GetLastDataset(accountSetter, accountRecipient, property string) (query string, args []interface{}) {
	return fmt.Sprintf("SELECT %s FROM %s WHERE %s AND latest = 1 ORDER BY height DESC LIMIT 1",
			strings.Join(dq.getFields(), ","),
			dq.getTableName(),
			fmt.Sprintf("%s = ?", strings.Join(dq.PrimaryFields[:3], " = ? AND ")), // where clause
		),
		[]interface{}{accountSetter, accountRecipient, property}
}

func (dq *DatasetsQuery) AddDataset(dataset *model.Dataset) [][]interface{} {
	var (
		queries [][]interface{}
	)

	// Update Dataset will happen when new dataset already exist in highest height
	updateDataset := fmt.Sprintf(`
		UPDATE %s SET (%s) = 
		(
			SELECT '%s', %d, 
				%d + CASE 
					WHEN timestamp_expires - %d < 0 THEN 0
					ELSE timestamp_expires - %d END 
			FROM %s 
			WHERE %s AND latest = 1
			ORDER BY height DESC LIMIT 1
		) 
		WHERE %s AND latest = 1
	`,
		dq.getTableName(),
		strings.Join(dq.OrdinaryFields[:3], ","),
		dataset.GetValue(),
		dataset.GetTimestampStarts(),
		dataset.GetTimestampExpires(),
		dataset.GetTimestampStarts(),
		dataset.GetTimestampStarts(),
		dq.getTableName(),
		fmt.Sprintf("%s = ? ", strings.Join(dq.PrimaryFields, " = ? AND ")),
		fmt.Sprintf("%s = ? ", strings.Join(dq.PrimaryFields, " = ? AND ")),
	)

	// Insert Dataset will happen when new dataset doesn't exist in highest height
	insertDataset := fmt.Sprintf(`
		INSERT INTO %s (%s)
		SELECT %s,
			%d + IFNULL((
				SELECT CASE
					WHEN timestamp_expires - %d < 0
						THEN 0
					ELSE timestamp_expires - %d END
				FROM %s
				WHERE %s AND latest = 1
				ORDER BY height DESC LIMIT 1
			), 0),
			1
		WHERE NOT EXISTS (
			SELECT %s FROM %s
			WHERE %s
		)
	`,
		dq.getTableName(),
		strings.Join(dq.getFields(), ", "),
		fmt.Sprintf("? %s", strings.Repeat(", ?", len(dq.getFields()[:6])-1)),
		dataset.GetTimestampExpires(),
		dataset.GetTimestampStarts(),
		dataset.GetTimestampStarts(),
		dq.getTableName(),
		fmt.Sprintf("%s != ? ", strings.Join(dq.PrimaryFields, " = ? AND ")),
		dq.PrimaryFields[0],
		dq.getTableName(),
		fmt.Sprintf("%s = ? ", strings.Join(dq.PrimaryFields, " = ? AND ")),
	)

	updateVersionQuery := fmt.Sprintf(
		"UPDATE %s SET latest = false WHERE %s AND latest = 1",
		dq.getTableName(),
		fmt.Sprintf("%s != ? ", strings.Join(dq.PrimaryFields, " = ? AND ")), // where clause
	)

	queries = append(queries,
		append([]interface{}{updateDataset}, append(dq.ExtractModel(dataset)[:4], dq.ExtractModel(dataset)[:4]...)...),
		append([]interface{}{insertDataset},
			append(dq.ExtractModel(dataset)[:6], append(dq.ExtractModel(dataset)[:4], dq.ExtractModel(dataset)[:4]...)...)...),
		append([]interface{}{updateVersionQuery}, dq.ExtractModel(dataset)[:4]...),
	)

	return queries
}

func (dq *DatasetsQuery) ExtractModel(dataset *model.Dataset) []interface{} {
	return []interface{}{
		dataset.GetAccountSetter(),
		dataset.GetAccountRecipient(),
		dataset.GetProperty(),
		dataset.GetHeight(),
		dataset.GetValue(),
		dataset.GetTimestampStarts(),
		dataset.GetTimestampExpires(),
		dataset.GetLatest(),
	}
}

func (dq *DatasetsQuery) BuildModel(datasets []*model.Dataset, rows *sql.Rows) []*model.Dataset {
	for rows.Next() {
		var dataset model.Dataset
		_ = rows.Scan(
			&dataset.AccountSetter,
			&dataset.AccountRecipient,
			&dataset.Property,
			&dataset.Height,
			&dataset.Value,
			&dataset.TimestampStarts,
			&dataset.TimestampExpires,
			&dataset.Latest,
		)
		datasets = append(datasets, &dataset)
	}
	return datasets
}

func (dq *DatasetsQuery) getTableName() string {
	if dq != nil {
		return dq.TableName
	}
	return ""
}

func (dq *DatasetsQuery) getFields() []string {
	return append(
		dq.PrimaryFields,
		dq.OrdinaryFields...,
	)
}
