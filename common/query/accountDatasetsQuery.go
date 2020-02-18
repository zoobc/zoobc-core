package query

import (
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
		GetAccountDatasetsForSnapshot(fromHeight, toHeight uint32) string
		GetLastDataset(accountSetter, accountRecipient, property string) (query string, args []interface{})
		GetDatasetsByRecipientAccountAddress(accountRecipient string) (query string, args interface{})
		AddDataset(dataset *model.AccountDataset) [][]interface{}
		RemoveDataset(dataset *model.AccountDataset) [][]interface{}
		ExtractModel(dataset *model.AccountDataset) []interface{}
		BuildModel(datasets []*model.AccountDataset, rows *sql.Rows) ([]*model.AccountDataset, error)
		Scan(dataset *model.AccountDataset, row *sql.Row) error
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
			strings.Join(adq.GetFields(), ","),
			adq.TableName,
		),
		accountRecipient
}

func (adq *AccountDatasetsQuery) GetAccountDatasetsForSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE height >= %d AND height <= %d AND latest = 1 ORDER BY height",
		strings.Join(adq.GetFields(), ","),
		adq.TableName,
		fromHeight,
		toHeight,
	)
}

func (adq *AccountDatasetsQuery) GetLastDataset(accountSetter, accountRecipient, property string) (query string, args []interface{}) {
	caseArgs := []interface{}{accountSetter, accountRecipient, property}
	cq := NewCaseQuery()
	cq.Select(adq.TableName, adq.GetFields()...)
	// where caluse : setter_account_address, recipient_account_address, property, lasted
	cq.Where(cq.Equal("latest", true))
	for k, v := range adq.PrimaryFields[:3] {
		cq.And(cq.Equal(v, caseArgs[k]))
	}
	// it's removed dataset when timestamp_starts = timestamp_expires
	cq.And("timestamp_starts <> timestamp_expires ")
	cq.OrderBy("height", model.OrderBy_DESC)
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
			WHERE %s AND latest = true
			ORDER BY height DESC LIMIT 1
		) 
		WHERE %s AND latest = true
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
				WHERE %s AND latest = true
				ORDER BY height DESC LIMIT 1
			), 0),
			true
		WHERE NOT EXISTS (
			SELECT %s FROM %s
			WHERE %s
		)
	`,
		adq.TableName,
		strings.Join(adq.GetFields(), ", "),
		fmt.Sprintf("? %s", strings.Repeat(", ?", len(adq.GetFields()[:6])-1)),
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
		strings.Join(adq.GetFields(), ", "),
		fmt.Sprintf("? %s", strings.Repeat(", ?", len(adq.GetFields())-1)),
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
		"UPDATE %s SET latest = false WHERE %s AND latest = true",
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

func (adq *AccountDatasetsQuery) BuildModel(
	datasets []*model.AccountDataset,
	rows *sql.Rows,
) ([]*model.AccountDataset, error) {
	for rows.Next() {
		var (
			dataset model.AccountDataset
			err     error
		)
		err = rows.Scan(
			&dataset.SetterAccountAddress,
			&dataset.RecipientAccountAddress,
			&dataset.Property,
			&dataset.Height,
			&dataset.Value,
			&dataset.TimestampStarts,
			&dataset.TimestampExpires,
			&dataset.Latest,
		)
		if err != nil {
			return nil, err
		}
		datasets = append(datasets, &dataset)
	}
	return datasets, nil
}

func (*AccountDatasetsQuery) Scan(dataset *model.AccountDataset, row *sql.Row) error {
	err := row.Scan(
		&dataset.SetterAccountAddress,
		&dataset.RecipientAccountAddress,
		&dataset.Property,
		&dataset.Height,
		&dataset.Value,
		&dataset.TimestampStarts,
		&dataset.TimestampExpires,
		&dataset.Latest,
	)
	return err
}

func (adq *AccountDatasetsQuery) GetFields() []string {
	return append(
		adq.PrimaryFields,
		adq.OrdinaryFields...,
	)
}

func (adq *AccountDatasetsQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE height > ?", adq.TableName),
			height,
		},
		{
			fmt.Sprintf(`
				UPDATE %s SET latest = ?
				WHERE latest = ? AND (%s) IN (
					SELECT (%s) as con
					FROM %s
					GROUP BY %s
				)`,
				adq.TableName,
				strings.Join(adq.PrimaryFields, " || '_' || "),
				fmt.Sprintf("%s || '_' || MAX(height)", strings.Join(adq.PrimaryFields[:3], " || '_' || ")),
				adq.TableName,
				strings.Join(adq.PrimaryFields[:3], ", "),
			),
			1, 0,
		},
	}
}
