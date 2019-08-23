package query

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	AccountDatasetsQuery struct {
		PrimaryFields  []string
		OrdinaryFields []string
		TableName      string
	}

	AccountDatasetsQueryInterface interface {
		GetLastDataset(accountSetter, accountRecipient, property string) (query string, args []interface{})
		GetDatasetsByRecipientAccountAddress(accountRecipient string) (query string, args interface{})
		AddDataset(dataset *model.AccountDataset) [][]interface{}
		RemoveDataset(dataset *model.AccountDataset) [][]interface{}
		ExtractModel(dataset *model.AccountDataset) []interface{}
		BuildModel(datasets []*model.AccountDataset, rows *sql.Rows) []*model.AccountDataset
	}
)

// NewAccountDatasetsQuery will create a new AccountDatasetsQuery
func NewAccountDatasetsQuery() *AccountDatasetsQuery {
	return &AccountDatasetsQuery{
		PrimaryFields: []string{
			"setter_account_address",
			"recipient_account_address",
			"property",
			"height",
		},
		OrdinaryFields: []string{
			"value",
			"timestamp_starts",
			"timestamp_expires",
			"latest",
		},
		TableName: "account_dataset",
	}
}

func (adq *AccountDatasetsQuery) GetDatasetsByRecipientAccountAddress(accountRecipient string) (query string, args interface{}) {
	return fmt.Sprintf("SELECT %s FROM %s WHERE recipient_account_address = ? AND latest = 1",
			strings.Join(adq.getFields(), ","),
			adq.TableName,
		),
		accountRecipient
}

func (adq *AccountDatasetsQuery) GetLastDataset(accountSetter, accountRecipient, property string) (query string, args []interface{}) {
	caseArgs := []interface{}{accountSetter, accountRecipient, property}
	cq := CaseQuery{
		Query: bytes.NewBuffer([]byte{}),
	}
	cq.Select(adq.TableName, adq.getFields()...)
	// where caluse : setter_account_address, recipient_account_address, property, lasted
	cq.Where(cq.Equal("latest", 1))
	for k, v := range adq.PrimaryFields[:3] {
		cq.And(cq.Equal(v, caseArgs[k]))
	}
	// it's removed dataset when timestamp_starts = timestamp_expires
	cq.And("timestamp_starts <> timestamp_expires ")
	cq.OrderBy("height", OrderDesc)
	cq.Limit(1)

	return cq.Build()
}

func (adq *AccountDatasetsQuery) AddDataset(dataset *model.AccountDataset) [][]interface{} {
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
		adq.TableName,
		strings.Join(adq.OrdinaryFields[:3], ","),
		dataset.GetValue(),
		dataset.GetTimestampStarts(),
		dataset.GetTimestampExpires(),
		dataset.GetTimestampStarts(),
		dataset.GetTimestampStarts(),
		adq.TableName,
		fmt.Sprintf("%s = ? ", strings.Join(adq.PrimaryFields, " = ? AND ")),
		fmt.Sprintf("%s = ? ", strings.Join(adq.PrimaryFields, " = ? AND ")),
	)

	// Insert Dataset will happen when new dataset doesn't exist in highest height
	insertDataset := fmt.Sprintf(`
		INSERT INTO %s (%s)
		SELECT %s,
			%d + IFNULL((
				SELECT CASE
					WHEN timestamp_expires - %d < 0 THEN 0
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
		adq.TableName,
		strings.Join(adq.getFields(), ", "),
		fmt.Sprintf("? %s", strings.Repeat(", ?", len(adq.getFields()[:6])-1)),
		dataset.GetTimestampExpires(),
		dataset.GetTimestampStarts(),
		dataset.GetTimestampStarts(),
		adq.TableName,
		fmt.Sprintf("%s != ? ", strings.Join(adq.PrimaryFields, " = ? AND ")),
		adq.PrimaryFields[0],
		adq.TableName,
		fmt.Sprintf("%s = ? ", strings.Join(adq.PrimaryFields, " = ? AND ")),
	)

	argumentWhere := adq.ExtractArgsWhere(dataset)
	queries = append(queries,
		append([]interface{}{updateDataset}, append(argumentWhere, argumentWhere...)...),
		append([]interface{}{insertDataset},
			append(adq.ExtractModel(dataset)[:6], append(argumentWhere, argumentWhere...)...)...),
		adq.UpdateVersion(dataset),
	)

	return queries
}

func (adq *AccountDatasetsQuery) RemoveDataset(dataset *model.AccountDataset) [][]interface{} {
	var (
		queries [][]interface{}
	)

	updateDataset := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s AND latest = true",
		adq.TableName,
		fmt.Sprintf("%s = ? ", strings.Join(adq.OrdinaryFields, " = ?, ")),
		fmt.Sprintf("%s = ? ", strings.Join(adq.PrimaryFields, " = ? AND ")),
	)

	insertDataset := fmt.Sprintf(`
		INSERT INTO %s (%s)
		SELECT %s
		WHERE NOT EXISTS (
			SELECT %s FROM %s
			WHERE %s
		)
	`,
		adq.TableName,
		strings.Join(adq.getFields(), ", "),
		fmt.Sprintf("? %s", strings.Repeat(", ?", len(adq.getFields())-1)),
		adq.PrimaryFields[0],
		adq.TableName,
		fmt.Sprintf("%s = ? ", strings.Join(adq.PrimaryFields, " = ? AND ")),
	)

	argumentWhere := adq.ExtractArgsWhere(dataset)
	queries = append(queries,
		append([]interface{}{updateDataset}, append(adq.ExtractModel(dataset)[4:], argumentWhere...)...),
		append([]interface{}{insertDataset}, append(adq.ExtractModel(dataset), argumentWhere...)...),
		adq.UpdateVersion(dataset),
	)
	return queries
}

func (adq *AccountDatasetsQuery) UpdateVersion(dataset *model.AccountDataset) []interface{} {
	updateVersionQ := fmt.Sprintf(
		"UPDATE %s SET latest = false WHERE %s AND latest = 1",
		adq.TableName,
		fmt.Sprintf("%s != ? ", strings.Join(adq.PrimaryFields, " = ? AND ")), // where clause
	)
	return append([]interface{}{updateVersionQ}, adq.ExtractArgsWhere(dataset)...)
}

func (adq *AccountDatasetsQuery) ExtractModel(dataset *model.AccountDataset) []interface{} {
	return []interface{}{
		dataset.GetSetterAccountAddress(),
		dataset.GetRecipientAccountAddress(),
		dataset.GetProperty(),
		dataset.GetHeight(),
		dataset.GetValue(),
		dataset.GetTimestampStarts(),
		dataset.GetTimestampExpires(),
		dataset.GetLatest(),
	}
}

/*
	ExtractArgsWhere represent extracted spesific field of account dataset model
	(Primary field of account dataset)
*/
func (adq *AccountDatasetsQuery) ExtractArgsWhere(dataset *model.AccountDataset) []interface{} {
	return []interface{}{
		dataset.GetSetterAccountAddress(),
		dataset.GetRecipientAccountAddress(),
		dataset.GetProperty(),
		dataset.GetHeight(),
	}
}

func (adq *AccountDatasetsQuery) BuildModel(datasets []*model.AccountDataset, rows *sql.Rows) []*model.AccountDataset {
	for rows.Next() {
		var dataset model.AccountDataset
		_ = rows.Scan(
			&dataset.SetterAccountAddress,
			&dataset.RecipientAccountAddress,
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

func (adq *AccountDatasetsQuery) getFields() []string {
	return append(
		adq.PrimaryFields,
		adq.OrdinaryFields...,
	)
}
