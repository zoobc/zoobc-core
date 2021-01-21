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
	"database/sql"
	"fmt"
	"github.com/zoobc/zoobc-core/common/blocker"
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
		GetLatestAccountDataset(setterAccountAddress, recipientAccountAddress []byte, property string) (str string, args []interface{})
		InsertAccountDatasets(datasets []*model.AccountDataset) (str string, args []interface{})
		InsertAccountDataset(dataset *model.AccountDataset) [][]interface{}
		GetAccountDatasetEscrowApproval(accountAddress []byte) (qStr string, args []interface{})
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

// InsertAccountDatasets represents query builder to insert multiple record in single query
func (adq *AccountDatasetQuery) InsertAccountDatasets(datasets []*model.AccountDataset) (str string, args []interface{}) {
	if len(datasets) > 0 {
		str = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES ",
			adq.getTableName(),
			strings.Join(adq.Fields, ", "),
		)
		for k, dataset := range datasets {
			str += fmt.Sprintf(
				"(?%s)",
				strings.Repeat(", ?", len(adq.Fields)-1),
			)
			if k < len(datasets)-1 {
				str += ","
			}
			args = append(args, adq.ExtractModel(dataset)...)
		}
	}
	return str, args
}

// ImportSnapshot takes payload from downloaded snapshot and insert them into database
func (adq *AccountDatasetQuery) ImportSnapshot(payload interface{}) ([][]interface{}, error) {
	var (
		queries [][]interface{}
	)
	accountDatasets, ok := payload.([]*model.AccountDataset)
	if !ok {
		return nil, blocker.NewBlocker(blocker.DBErr, "ImportSnapshotCannotCastTo"+adq.TableName)
	}
	if len(accountDatasets) > 0 {
		recordsPerPeriod, rounds, remaining := CalculateBulkSize(len(adq.Fields), len(accountDatasets))
		for i := 0; i < rounds; i++ {
			qry, args := adq.InsertAccountDatasets(accountDatasets[i*recordsPerPeriod : (i*recordsPerPeriod)+recordsPerPeriod])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
		if remaining > 0 {
			qry, args := adq.InsertAccountDatasets(accountDatasets[len(accountDatasets)-remaining:])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
	}
	return queries, nil
}

// RecalibrateVersionedTable recalibrate table to clean up multiple latest rows due to import function
func (adq *AccountDatasetQuery) RecalibrateVersionedTable() []string {
	return []string{
		fmt.Sprintf(
			"update %s set latest = false where latest = true AND (setter_account_address, recipient_account_address, property, height) NOT IN "+
				"(select t2.setter_account_address, t2.recipient_account_address, t2.property, max(t2.height) from %s t2 "+
				"group by t2.setter_account_address, t2.recipient_account_address, t2.property)",
			adq.getTableName(), adq.getTableName()),
		fmt.Sprintf(
			"update %s set latest = true where latest = false AND (setter_account_address, recipient_account_address, property, height) IN "+
				"(select t2.setter_account_address, t2.recipient_account_address, t2.property, max(t2.height) from %s t2 "+
				"group by t2.setter_account_address, t2.recipient_account_address, t2.property)",
			adq.getTableName(), adq.getTableName()),
	}
}

// GetLatestAccountDataset represents query builder to get the latest record of account_dataset
func (adq *AccountDatasetQuery) GetLatestAccountDataset(
	setterAccountAddress, recipientAccountAddress []byte, property string,
) (str string, args []interface{}) {
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
InsertAccountDataset represents a query builder to insert new record and set old version as latest is false and active is false
to account_dataset table. Perhaps *model.AccountDataset.IsActive already set to false
*/
func (adq *AccountDatasetQuery) InsertAccountDataset(dataset *model.AccountDataset) (str [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf(
				"UPDATE %s SET latest = ? WHERE setter_account_address = ? AND recipient_account_address = ? "+
					"AND property = ? AND height < ? AND latest = ?",
				adq.getTableName(),
			),
			false,
			dataset.GetSetterAccountAddress(),
			dataset.GetRecipientAccountAddress(),
			dataset.GetProperty(),
			dataset.GetHeight(),
			true,
		},
		append(
			[]interface{}{
				fmt.Sprintf(
					"INSERT INTO %s (%s) VALUES(%s) "+
						"ON CONFLICT(setter_account_address, recipient_account_address, property, height) "+
						"DO UPDATE SET value = ?, is_active = ?, latest = ?",
					adq.getTableName(),
					strings.Join(adq.Fields, ", "),
					fmt.Sprintf("?%s", strings.Repeat(", ?", len(adq.Fields)-1)),
				),
			},
			append(
				adq.ExtractModel(dataset),
				dataset.GetValue(),
				dataset.GetIsActive(),
				dataset.GetLatest(),
			)...,
		),
	}
}

// GetAccountDatasetEscrowApproval represents query for get account dataset for AccountDatasetEscrowApproval property
// SetterAccountAddress and RecipientAccountAddress must be the same person
func (adq *AccountDatasetQuery) GetAccountDatasetEscrowApproval(accountAddress []byte) (qStr string, args []interface{}) {
	return fmt.Sprintf(
			"SELECT %s FROM %s WHERE setter_account_address = ? AND recipient_account_address = ? AND property = ? AND latest = ?",
			strings.Join(adq.Fields, ", "),
			adq.getTableName(),
		), []interface{}{
			accountAddress,
			accountAddress,
			model.AccountDatasetProperty_AccountDatasetEscrowApproval.String(),
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
				WHERE height >= %d AND height <= %d AND height != 0
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
	return fmt.Sprintf(`DELETE FROM %s WHERE height >= %d AND height <= %d AND height != 0`,
		adq.TableName, fromHeight, toHeight)
}
