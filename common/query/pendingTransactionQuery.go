package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	PendingTransactionQueryInterface interface {
		GetPendingTransactionByHash(
			txHash []byte,
			status model.PendingTransactionStatus,
			currentHeight, limit uint32,
		) (str string, args []interface{})
		GetPendingTransactionsBySenderAddress(
			multisigAddress string,
			status model.PendingTransactionStatus,
			currentHeight, limit uint32,
		) (
			str string, args []interface{},
		)
		InsertPendingTransaction(pendingTx *model.PendingTransaction) [][]interface{}
		Scan(pendingTx *model.PendingTransaction, row *sql.Row) error
		ExtractModel(pendingTx *model.PendingTransaction) []interface{}
		BuildModel(pendingTxs []*model.PendingTransaction, rows *sql.Rows) ([]*model.PendingTransaction, error)
	}

	PendingTransactionQuery struct {
		Fields    []string
		TableName string
	}
)

// NewPendingTransactionQuery returns PendingTransactionQuery instance
func NewPendingTransactionQuery() *PendingTransactionQuery {
	return &PendingTransactionQuery{
		Fields: []string{
			"sender_address",
			"transaction_hash",
			"transaction_bytes",
			"status",
			"block_height",
			"latest",
		},
		TableName: "pending_transaction",
	}
}

func (ptq *PendingTransactionQuery) getTableName() string {
	return ptq.TableName
}

func (ptq *PendingTransactionQuery) GetPendingTransactionByHash(
	txHash []byte,
	status model.PendingTransactionStatus,
	currentHeight, limit uint32,
) (str string, args []interface{}) {
	var (
		blockHeight uint32
	)
	if currentHeight > limit {
		blockHeight = currentHeight - limit
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE transaction_hash = ? AND status = ? AND block_height >= ? "+
		"AND latest = true", strings.Join(ptq.Fields, ", "), ptq.getTableName())
	return query, []interface{}{
		txHash,
		status,
		blockHeight,
	}
}

func (ptq *PendingTransactionQuery) GetPendingTransactionsBySenderAddress(
	multisigAddress string,
	status model.PendingTransactionStatus,
	currentHeight, limit uint32,
) (str string, args []interface{}) {
	var (
		blockHeight uint32
	)
	if currentHeight > limit {
		blockHeight = currentHeight - limit
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE sender_address = ? AND status = ? AND block_height >= ? "+
		"AND latest = true ORDER BY block_height ASC",
		strings.Join(ptq.Fields, ", "), ptq.getTableName())
	return query, []interface{}{
		multisigAddress,
		status,
		blockHeight,
	}
}

// InsertPendingTransaction inserts a new pending transaction into DB
func (ptq *PendingTransactionQuery) InsertPendingTransaction(pendingTx *model.PendingTransaction) [][]interface{} {
	var queries [][]interface{}
	insertQuery := fmt.Sprintf("INSERT OR REPLACE INTO %s (%s) VALUES(%s)",
		ptq.getTableName(),
		strings.Join(ptq.Fields, ", "),
		fmt.Sprintf("? %s", strings.Repeat(", ? ", len(ptq.Fields)-1)),
	)
	updateQuery := fmt.Sprintf("UPDATE %s SET latest = false WHERE transaction_hash = ? AND block_height != %d AND latest = true",
		ptq.getTableName(),
		pendingTx.BlockHeight,
	)
	queries = append(queries,
		append([]interface{}{insertQuery}, ptq.ExtractModel(pendingTx)...),
		[]interface{}{
			updateQuery, pendingTx.TransactionHash,
		},
	)
	return queries
}

func (*PendingTransactionQuery) Scan(pendingTx *model.PendingTransaction, row *sql.Row) error {
	err := row.Scan(
		&pendingTx.SenderAddress,
		&pendingTx.TransactionHash,
		&pendingTx.TransactionBytes,
		&pendingTx.Status,
		&pendingTx.BlockHeight,
		&pendingTx.Latest,
	)
	return err
}

func (*PendingTransactionQuery) ExtractModel(pendingTx *model.PendingTransaction) []interface{} {
	return []interface{}{
		&pendingTx.SenderAddress,
		&pendingTx.TransactionHash,
		&pendingTx.TransactionBytes,
		&pendingTx.Status,
		&pendingTx.BlockHeight,
		&pendingTx.Latest,
	}
}

func (ptq *PendingTransactionQuery) BuildModel(
	pts []*model.PendingTransaction, rows *sql.Rows,
) ([]*model.PendingTransaction, error) {
	for rows.Next() {
		var pendingTx model.PendingTransaction
		err := rows.Scan(
			&pendingTx.SenderAddress,
			&pendingTx.TransactionHash,
			&pendingTx.TransactionBytes,
			&pendingTx.Status,
			&pendingTx.BlockHeight,
			&pendingTx.Latest,
		)
		if err != nil {
			return nil, err
		}
		pts = append(pts, &pendingTx)
	}
	return pts, nil
}

// Rollback delete records `WHERE block_height > "height"`
func (ptq *PendingTransactionQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", ptq.TableName),
			height,
		},
		{
			fmt.Sprintf("UPDATE %s SET latest = ? WHERE latest = ? AND (block_height || '_' || "+
				"transaction_hash) IN (SELECT (MAX(block_height) || '_' || transaction_hash) as con "+
				"FROM %s GROUP BY transaction_hash)",
				ptq.getTableName(),
				ptq.getTableName(),
			),
			1, 0,
		},
	}
}

func (ptq *PendingTransactionQuery) SelectDataForSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`SELECT %s FROM %s WHERE latest = 1 AND block_height >= %d AND block_height <= %d ORDER BY block_height DESC`,
		strings.Join(ptq.Fields, ","), ptq.TableName, fromHeight, toHeight)
}

// TrimDataBeforeSnapshot delete entries to assure there are no duplicates before applying a snapshot
func (ptq *PendingTransactionQuery) TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`DELETE FROM %s WHERE block_height >= %d AND block_height <= %d`,
		ptq.TableName, fromHeight, toHeight)
}
