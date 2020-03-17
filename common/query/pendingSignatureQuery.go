package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	PendingSignatureQueryInterface interface {
		GetPendingSignatureByHash(
			txHash []byte,
			currentHeight, limit uint32,
		) (str string, args []interface{})
		InsertPendingSignature(pendingSig *model.PendingSignature) [][]interface{}
		Scan(pendingSig *model.PendingSignature, row *sql.Row) error
		ExtractModel(pendingSig *model.PendingSignature) []interface{}
		BuildModel(pendingSigs []*model.PendingSignature, rows *sql.Rows) ([]*model.PendingSignature, error)
		TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string
	}

	PendingSignatureQuery struct {
		Fields    []string
		TableName string
	}
)

// NewPendingSignatureQuery returns PendingTransactionQuery instance
func NewPendingSignatureQuery() *PendingSignatureQuery {
	return &PendingSignatureQuery{
		Fields: []string{
			"transaction_hash",
			"account_address",
			"signature",
			"block_height",
			"latest",
		},
		TableName: "pending_signature",
	}
}

func (psq *PendingSignatureQuery) getTableName() string {
	return psq.TableName
}

func (psq *PendingSignatureQuery) GetPendingSignatureByHash(
	txHash []byte,
	currentHeight, limit uint32,
) (str string, args []interface{}) {
	var (
		blockHeight uint32
	)
	if currentHeight > limit {
		blockHeight = currentHeight - limit
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE transaction_hash = ? AND block_height >= ? AND latest = true",
		strings.Join(psq.Fields, ", "), psq.getTableName())
	return query, []interface{}{
		txHash,
		blockHeight,
	}
}

// InsertPendingSignature inserts a new pending transaction into DB
func (psq *PendingSignatureQuery) InsertPendingSignature(pendingSig *model.PendingSignature) [][]interface{} {
	var queries [][]interface{}
	insertQuery := fmt.Sprintf("INSERT OR REPLACE INTO %s (%s) VALUES(%s)",
		psq.getTableName(),
		strings.Join(psq.Fields, ", "),
		fmt.Sprintf("? %s", strings.Repeat(", ? ", len(psq.Fields)-1)),
	)
	updateQuery := fmt.Sprintf("UPDATE %s SET latest = false WHERE account_address = ? AND transaction_hash = ? "+
		"AND block_height != %d AND latest = true",
		psq.getTableName(),
		pendingSig.BlockHeight,
	)
	queries = append(queries,
		append([]interface{}{insertQuery}, psq.ExtractModel(pendingSig)...),
		[]interface{}{
			updateQuery, pendingSig.AccountAddress, pendingSig.TransactionHash,
		},
	)
	return queries
}

func (*PendingSignatureQuery) Scan(pendingSig *model.PendingSignature, row *sql.Row) error {
	err := row.Scan(
		&pendingSig.TransactionHash,
		&pendingSig.AccountAddress,
		&pendingSig.Signature,
		&pendingSig.BlockHeight,
		&pendingSig.Latest,
	)
	return err
}

func (*PendingSignatureQuery) ExtractModel(pendingSig *model.PendingSignature) []interface{} {
	return []interface{}{
		&pendingSig.TransactionHash,
		&pendingSig.AccountAddress,
		&pendingSig.Signature,
		&pendingSig.BlockHeight,
		&pendingSig.Latest,
	}
}

func (psq *PendingSignatureQuery) BuildModel(
	pss []*model.PendingSignature, rows *sql.Rows,
) ([]*model.PendingSignature, error) {
	for rows.Next() {
		var pendingSig model.PendingSignature
		err := rows.Scan(
			&pendingSig.TransactionHash,
			&pendingSig.AccountAddress,
			&pendingSig.Signature,
			&pendingSig.BlockHeight,
			&pendingSig.Latest,
		)
		if err != nil {
			return nil, err
		}
		pss = append(pss, &pendingSig)
	}
	return pss, nil
}

// Rollback delete records `WHERE block_height > "height"`
func (psq *PendingSignatureQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", psq.TableName),
			height,
		},
		{
			fmt.Sprintf("UPDATE %s SET latest = ? WHERE latest = ? AND (block_height || '_' || "+
				"account_address || '_' || transaction_hash) IN (SELECT (MAX(block_height) || '_' || "+
				"account_address || '_' || transaction_hash) as con FROM %s GROUP BY account_address "+
				"|| '_' || transaction_hash)",
				psq.TableName,
				psq.TableName,
			),
			1, 0,
		},
	}
}

func (psq *PendingSignatureQuery) SelectDataForSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`SELECT %s FROM %s WHERE latest = 1 AND block_height >= %d AND block_height <= %d ORDER BY block_height DESC`,
		strings.Join(psq.Fields, ","), psq.TableName, fromHeight, toHeight)
}

// TrimDataBeforeSnapshot delete entries to assure there are no duplicates before applying a snapshot
func (psq *PendingSignatureQuery) TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`DELETE FROM %s WHERE block_height >= %d AND block_height <= %d`,
		psq.TableName, fromHeight, toHeight)
}
