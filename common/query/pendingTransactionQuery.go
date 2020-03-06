package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	PendingTransactionQueryInterface interface {
		GetPendingTransactionByHash(txHash []byte, status model.PendingTransactionStatus) (str string, args []interface{})
		GetPendingTransactionsBySenderAddress(multisigAddress string, status model.PendingTransactionStatus) (
			str string, args []interface{},
		)
		InsertPendingTransaction(pendingTx *model.PendingTransaction) (str string, args []interface{})
		UpdatePendingTransaction(pendingTx *model.PendingTransaction) [][]interface{}
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

func (ptq *PendingTransactionQuery) GetPendingTransactionByHash(txHash []byte, status model.PendingTransactionStatus) (str string, args []interface{}) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE transaction_hash = ? AND status = ? "+
		"AND latest = true", strings.Join(ptq.Fields, ", "), ptq.getTableName())
	return query, []interface{}{
		txHash,
		status,
	}
}

func (ptq *PendingTransactionQuery) GetPendingTransactionsBySenderAddress(
	multisigAddress string, status model.PendingTransactionStatus,
) (str string, args []interface{}) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE sender_address = ? AND status = ? AND latest = true ORDER BY block_height ASC",
		strings.Join(ptq.Fields, ", "), ptq.getTableName())
	return query, []interface{}{
		multisigAddress,
		status,
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

// UpdatePendingTransaction Update status of pending transaction
func (ptq *PendingTransactionQuery) UpdatePendingTransaction(pendingTx *model.PendingTransaction) [][]interface{} {
	return [][]interface{}{
		{
			fmt.Sprintf(
				"UPDATE %s set latest = ? WHERE transaction_hash = ?",
				ptq.getTableName(),
			),
			false,
			pendingTx.GetTransactionHash(),
		},
		append(
			[]interface{}{
				fmt.Sprintf(
					"INSERT OR REPLACE INTO %s (%s) VALUES(%s)",
					ptq.getTableName(),
					strings.Join(ptq.Fields, ","),
					fmt.Sprintf("? %s", strings.Repeat(", ?", len(ptq.Fields)-1))),
			},
			ptq.ExtractModel(pendingTx)...,
		),
	}
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
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", ptq.getTableName()),
			height,
		},
	}
}
