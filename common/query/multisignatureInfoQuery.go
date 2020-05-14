package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	MultisignatureInfoQueryInterface interface {
		GetMultisignatureInfoByAddress(
			multisigAddress string,
			currentHeight, limit uint32,
		) (str string, args []interface{})
		InsertMultisignatureInfo(multisigInfo *model.MultiSignatureInfo) [][]interface{}
		Scan(multisigInfo *model.MultiSignatureInfo, row *sql.Row) error
		ExtractModel(multisigInfo *model.MultiSignatureInfo) []interface{}
		BuildModel(multisigInfos []*model.MultiSignatureInfo, rows *sql.Rows) ([]*model.MultiSignatureInfo, error)
	}

	MultisignatureInfoQuery struct {
		Fields    []string
		TableName string
	}
)

// NewMultisignatureInfoQuery returns PendingTransactionQuery instance
func NewMultisignatureInfoQuery() *MultisignatureInfoQuery {
	return &MultisignatureInfoQuery{
		Fields: []string{
			"multisig_address",
			"minimum_signatures",
			"nonce",
			"addresses",
			"block_height",
			"latest",
		},
		TableName: "multisignature_info",
	}
}

func (msi *MultisignatureInfoQuery) getTableName() string {
	return msi.TableName
}

func (msi *MultisignatureInfoQuery) GetMultisignatureInfoByAddress(
	multisigAddress string,
	currentHeight, limit uint32,
) (str string, args []interface{}) {
	var (
		blockHeight uint32
	)
	if currentHeight > limit {
		blockHeight = currentHeight - limit
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE multisig_address = ? AND block_height >= ? AND latest = true",
		strings.Join(msi.Fields, ", "), msi.getTableName())
	return query, []interface{}{
		multisigAddress,
		blockHeight,
	}
}

// InsertMultisignatureInfo inserts a new pending transaction into DB
func (msi *MultisignatureInfoQuery) InsertMultisignatureInfo(multisigInfo *model.MultiSignatureInfo) [][]interface{} {
	var queries [][]interface{}
	insertQuery := fmt.Sprintf("INSERT OR REPLACE INTO %s (%s) VALUES(%s)",
		msi.getTableName(),
		strings.Join(msi.Fields, ", "),
		fmt.Sprintf("? %s", strings.Repeat(", ? ", len(msi.Fields)-1)),
	)
	updateQuery := fmt.Sprintf("UPDATE %s SET latest = false WHERE multisig_address = ? "+
		"AND block_height != %d AND latest = true",
		msi.getTableName(),
		multisigInfo.BlockHeight,
	)
	queries = append(queries,
		append([]interface{}{insertQuery}, msi.ExtractModel(multisigInfo)...),
		[]interface{}{
			updateQuery, multisigInfo.MultisigAddress,
		},
	)
	return queries
}

func (*MultisignatureInfoQuery) Scan(multisigInfo *model.MultiSignatureInfo, row *sql.Row) error {
	var addresses string
	err := row.Scan(
		&multisigInfo.MultisigAddress,
		&multisigInfo.MinimumSignatures,
		&multisigInfo.Nonce,
		&addresses,
		&multisigInfo.BlockHeight,
		&multisigInfo.Latest,
	)
	multisigInfo.Addresses = strings.Split(addresses, ", ")
	return err
}

func (*MultisignatureInfoQuery) ExtractModel(multisigInfo *model.MultiSignatureInfo) []interface{} {
	addresses := strings.Join(multisigInfo.Addresses, ", ")
	return []interface{}{
		&multisigInfo.MultisigAddress,
		&multisigInfo.MinimumSignatures,
		&multisigInfo.Nonce,
		addresses,
		&multisigInfo.BlockHeight,
		&multisigInfo.Latest,
	}
}

func (msi *MultisignatureInfoQuery) BuildModel(
	mss []*model.MultiSignatureInfo, rows *sql.Rows,
) ([]*model.MultiSignatureInfo, error) {
	for rows.Next() {
		var (
			multisigInfo model.MultiSignatureInfo
			addresses    string
		)
		err := rows.Scan(
			&multisigInfo.MultisigAddress,
			&multisigInfo.MinimumSignatures,
			&multisigInfo.Nonce,
			&addresses,
			&multisigInfo.BlockHeight,
			&multisigInfo.Latest,
		)
		multisigInfo.Addresses = strings.Split(addresses, ", ")
		if err != nil {
			return nil, err
		}
		mss = append(mss, &multisigInfo)
	}
	return mss, nil
}

// Rollback delete records `WHERE block_height > "height"`
func (msi *MultisignatureInfoQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", msi.getTableName()),
			height,
		},
		{
			fmt.Sprintf("UPDATE %s SET latest = ? WHERE latest = ? AND (multisig_address, block_height"+
				") IN (SELECT t2.multisig_address, MAX(t2.block_height) "+
				"FROM %s as t2 GROUP BY t2.multisig_address)",
				msi.getTableName(),
				msi.getTableName(),
			),
			1, 0,
		},
	}
}

func (msi *MultisignatureInfoQuery) SelectDataForSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE ("+
		"multisig_address, block_height) IN (SELECT t2.multisig_address, MAX("+
		"t2.block_height) FROM %s as t2 WHERE t2.block_height >= %d AND t2.block_height <= %d GROUP BY t2.multisig_address) ORDER BY"+
		" block_height",
		strings.Join(msi.Fields, ","), msi.TableName, msi.TableName, fromHeight, toHeight)
}

// TrimDataBeforeSnapshot delete entries to assure there are no duplicates before applying a snapshot
func (msi *MultisignatureInfoQuery) TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`DELETE FROM %s WHERE block_height >= %d AND block_height <= %d`,
		msi.TableName, fromHeight, toHeight)
}
