package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	PendingTransactionQueryInterface interface {
		GetPendingTransactionByHash(txHash []byte) (str string, args []interface{})
		InsertPendingTransaction(pendingTx *model.PendingTransaction) (str string, args []interface{})
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
			"transaction_hash",
			"transaction_bytes",
			"status",
			"block_height",
		},
		TableName: "pending_transaction",
	}
}

func (ptq *PendingTransactionQuery) getTableName() string {
	return ptq.TableName
}

func (ptq *PendingTransactionQuery) GetPendingTransactionByHash(txHash []byte) (str string, args []interface{}) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE transaction_hash = ?", strings.Join(ptq.Fields, ", "), ptq.getTableName())
	return query, []interface{}{
		txHash,
	}
}

// InsertPendingTransaction inserts a new pending transaction into DB
func (ptq *PendingTransactionQuery) InsertPendingTransaction(pendingTx *model.PendingTransaction) (str string, args []interface{}) {
	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES(%s)",
		ptq.getTableName(),
		strings.Join(ptq.Fields, ", "),
		fmt.Sprintf("? %s", strings.Repeat(", ? ", len(ptq.Fields)-1)),
	), ptq.ExtractModel(pendingTx)
}

func (*PendingTransactionQuery) Scan(pendingTx *model.PendingTransaction, row *sql.Row) error {
	err := row.Scan(
		&pendingTx.TransactionHash,
		&pendingTx.TransactionBytes,
		&pendingTx.Status,
		&pendingTx.BlockHeight,
	)
	return err
}

func (*PendingTransactionQuery) ExtractModel(pendingTx *model.PendingTransaction) []interface{} {
	return []interface{}{
		&pendingTx.TransactionHash,
		&pendingTx.TransactionBytes,
		&pendingTx.Status,
		&pendingTx.BlockHeight,
	}
}

func (ptq *PendingTransactionQuery) BuildModel(
	pts []*model.PendingTransaction, rows *sql.Rows,
) ([]*model.PendingTransaction, error) {
	for rows.Next() {
		var pendingTx model.PendingTransaction
		err := rows.Scan(
			&pendingTx.TransactionHash,
			&pendingTx.TransactionBytes,
			&pendingTx.Status,
			&pendingTx.BlockHeight,
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
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", ptq.getTableName()),
			height,
		},
	}
}
