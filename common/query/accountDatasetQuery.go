package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	// AccountDatasetQuery fields that must have
	AccountDatasetQuery struct {
		Fields    []string
		TableName string
	}

	// AccountDatasetQueryInterface methods must have
	AccountDatasetQueryInterface interface {
		GetLatestAccountDataset(setterAccountAddress, recipientAccountAddress, property string) (str string, args []interface{})
		InsertAccountDataset(dataset *model.AccountDataset) (str string, args []interface{})
		RemoveAccountDataset(dataset *model.AccountDataset) [][]interface{}
		GetAccountDatasetEscrowApproval(
			recipientAccountAddress string,
		) (qStr string, args []interface{})
		ExtractModel(dataset *model.AccountDataset) []interface{}
		BuildModel(datasets []*model.AccountDataset, rows *sql.Rows) ([]*model.AccountDataset, error)
		Scan(dataset *model.AccountDataset, row *sql.Row) error
	}
)

// NewAccountDatasetsQuery will create a new AccountDatasetQuery
func NewAccountDatasetsQuery() *AccountDatasetQuery {
	return &AccountDatasetQuery{
		Fields: []string{
			"setter_account_address",
			"recipient_account_address",
			"property",
			"value",
			"is_active",
			"latest",
			"height",
		},
		TableName: "account_dataset",
	}
}

/*
InsertAccountDataset represents a query builder to insert new record and set old version as latest is false and active is false
to account_dataset table
*/
func (adq *AccountDatasetQuery) InsertAccountDataset(dataset *model.AccountDataset) (str string, args []interface{}) {
	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		adq.getTableName(),
		strings.Join(adq.Fields, ", "),
		fmt.Sprintf("?%s", strings.Repeat(", ?", len(adq.Fields)-1)),
	), adq.ExtractModel(dataset)

}

// GetLatestAccountDataset represents query builder to get the latest record of account_dataset
func (adq *AccountDatasetQuery) GetLatestAccountDataset(setterAccountAddress, recipientAccountAddress, property string) (str string, args []interface{}) {
	return fmt.Sprintf(
			"SELECT %s FROM %s WHERE setter_account_address = ? AND recipient_account_address = ? AND property = ? AND latest = ?",
			strings.Join(adq.Fields, ", "),
			adq.getTableName(),
		),
		[]interface{}{
			setterAccountAddress,
			recipientAccountAddress,
			property,
			true,
		}
}

/*
RemoveAccountDataset represents a query builder to insert new record and set old version as latest is false and active is false
to account_dataset table. Perhaps *model.AccountDataset.IsActive already set to false
*/
func (adq *AccountDatasetQuery) RemoveAccountDataset(dataset *model.AccountDataset) (str [][]interface{}) {

	return [][]interface{}{
		{
			fmt.Sprintf(
				"UPDATE %s set latest = ? WHERE setter_account_address = ? AND recipient_account_address = ? "+
					"AND property = ? AND is_active = ?",
				adq.getTableName(),
			),
			false,
			dataset.GetSetterAccountAddress(),
			dataset.GetRecipientAccountAddress(),
			dataset.GetProperty(),
			true,
		},
		append([]interface{}{
			fmt.Sprintf(
				"INSERT INTO %s (%s) VALUES(%s)",
				adq.getTableName(),
				strings.Join(adq.Fields, ", "),
				fmt.Sprintf("?%s", strings.Repeat(", ?", len(adq.Fields)-1)),
			),
		}, adq.ExtractModel(dataset)...),
	}
}

// GetAccountDatasetEscrowApproval represents query for get account dataset for AccountDatasetEscrowApproval property
func (adq *AccountDatasetQuery) GetAccountDatasetEscrowApproval(
	recipientAccountAddress string,
) (qStr string, args []interface{}) {
	return fmt.Sprintf(
			"SELECT %s FROM %s WHERE recipient_account_address = ? AND property = ? AND latest = ?",
			strings.Join(adq.Fields, ", "),
			adq.getTableName(),
		), []interface{}{
			recipientAccountAddress,
			"AccountDatasetEscrowApproval",
			1,
		}
}

func (adq *AccountDatasetQuery) getTableName() string {
	return adq.TableName
}

// ExtractModel allowing to extracting the values
func (adq *AccountDatasetQuery) ExtractModel(dataset *model.AccountDataset) []interface{} {
	return []interface{}{
		dataset.GetSetterAccountAddress(),
		dataset.GetRecipientAccountAddress(),
		dataset.GetProperty(),
		dataset.GetValue(),
		dataset.GetIsActive(),
		dataset.GetLatest(),
		dataset.GetHeight(),
	}
}

// BuildModel allowing to extract *rows into list of model.AccountDataset
func (adq *AccountDatasetQuery) BuildModel(
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
			&dataset.Value,
			&dataset.IsActive,
			&dataset.Latest,
			&dataset.Height,
		)
		if err != nil {
			return nil, err
		}
		datasets = append(datasets, &dataset)
	}
	return datasets, nil
}

// Scan represents *sql.Scan
func (*AccountDatasetQuery) Scan(dataset *model.AccountDataset, row *sql.Row) error {
	return row.Scan(
		&dataset.SetterAccountAddress,
		&dataset.RecipientAccountAddress,
		&dataset.Property,
		&dataset.Value,
		&dataset.IsActive,
		&dataset.Latest,
		&dataset.Height,
	)
}

func (adq *AccountDatasetQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE height > ?", adq.TableName),
			height,
		},
		{
			fmt.Sprintf(`
				UPDATE %s SET latest = ?
				WHERE latest = ? AND (setter_account_address, recipient_account_address, property, height) IN (
					SELECT setter_account_address, recipient_account_address, property, MAX(height)
					FROM %s
					GROUP BY setter_account_address, recipient_account_address, property
				)`,
				adq.getTableName(),
				adq.getTableName(),
			),
			1, 0,
		},
	}
}

func (adq *AccountDatasetQuery) SelectDataForSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`
			SELECT %s FROM %s
			WHERE (setter_account_address, recipient_account_address, property, height) IN (
				SELECT setter_account_address, recipient_account_address, property, MAX(height) FROM %s
				WHERE height >= %d AND height <= %d
				GROUP BY setter_account_address, recipient_account_address, property
			) ORDER BY height`,
		strings.Join(adq.Fields, ", "),
		adq.getTableName(),
		adq.getTableName(),
		fromHeight,
		toHeight,
	)
}

// TrimDataBeforeSnapshot delete entries to assure there are no duplicates before applying a snapshot
func (adq *AccountDatasetQuery) TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`DELETE FROM %s WHERE height >= %d AND height <= %d`,
		adq.TableName, fromHeight, toHeight)
}
