package query

import (
	"database/sql"
	"fmt"
	"github.com/zoobc/zoobc-core/common/blocker"
	"strings"

	"github.com/zoobc/zoobc-core/common/constant"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	PendingTransactionQueryInterface interface {
		GetPendingTransactionByHash(
			txHash []byte,
			statuses []model.PendingTransactionStatus,
			currentHeight, limit uint32,
		) (str string, args []interface{})
		GetPendingTransactionsBySenderAddress(
			multisigAddress []byte,
			status model.PendingTransactionStatus,
			currentHeight, limit uint32,
		) (
			str string, args []interface{},
		)
		GetPendingTransactionsExpireByHeight(blockHeight uint32) (str string, args []interface{})
		InsertPendingTransaction(pendingTx *model.PendingTransaction) [][]interface{}
		InsertPendingTransactions(pendingTXs []*model.PendingTransaction) (str string, args []interface{})
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
	statuses []model.PendingTransactionStatus,
	currentHeight, limit uint32,
) (str string, args []interface{}) {
	var (
		blockHeight uint32
		query       string
	)
	if currentHeight > limit {
		blockHeight = currentHeight - limit
	}
	args = []interface{}{
		txHash,
	}
	if len(statuses) > 0 {
		var statusesFilter = fmt.Sprintf("AND status IN (?%s)", strings.Repeat(", ?", len(statuses)-1))
		query = fmt.Sprintf("SELECT %s FROM %s WHERE transaction_hash = ? %s AND block_height >= ? "+
			"AND latest = true", strings.Join(ptq.Fields, ", "), ptq.getTableName(), statusesFilter)
		for _, status := range statuses {
			args = append(args, status)
		}
	} else {
		query = fmt.Sprintf("SELECT %s FROM %s WHERE transaction_hash = ? AND block_height >= ? "+
			"AND latest = true", strings.Join(ptq.Fields, ", "), ptq.getTableName())
	}
	return query, append(args, blockHeight)
}

func (ptq *PendingTransactionQuery) GetPendingTransactionsBySenderAddress(
	multisigAddress []byte,
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

// GetPendingTransactionsExpireByHeight presents query to get pending_transactions that was expire by block_height
func (ptq *PendingTransactionQuery) GetPendingTransactionsExpireByHeight(currentHeight uint32) (str string, args []interface{}) {
	return fmt.Sprintf(
			"SELECT %s FROM %s WHERE block_height = ? AND status = ? AND latest = ?",
			strings.Join(ptq.Fields, ", "),
			ptq.getTableName(),
		),
		[]interface{}{
			currentHeight - constant.MinRollbackBlocks,
			model.PendingTransactionStatus_PendingTransactionPending,
			true,
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

// InsertPendingTransactions represents query builder to insert multiple record in single query
func (ptq *PendingTransactionQuery) InsertPendingTransactions(pendingTXs []*model.PendingTransaction) (str string, args []interface{}) {
	if len(pendingTXs) > 0 {
		str = fmt.Sprintf(
			"INSERT OR REPLACE INTO %s (%s) VALUES ",
			ptq.getTableName(),
			strings.Join(ptq.Fields, ", "),
		)
		for k, pendingTX := range pendingTXs {
			str += fmt.Sprintf(
				"(?%s)",
				strings.Repeat(", ?", len(ptq.Fields)-1),
			)
			if k < len(pendingTXs)-1 {
				str += ","
			}
			args = append(args, ptq.ExtractModel(pendingTX)...)
		}
	}
	return str, args
}

// ImportSnapshot takes payload from downloaded snapshot and insert them into database
func (ptq *PendingTransactionQuery) ImportSnapshot(payload interface{}) ([][]interface{}, error) {
	var (
		queries [][]interface{}
	)
	pendingTransactions, ok := payload.([]*model.PendingTransaction)
	if !ok {
		return nil, blocker.NewBlocker(blocker.DBErr, "ImportSnapshotCannotCastTo"+ptq.TableName)
	}
	if len(pendingTransactions) > 0 {
		recordsPerPeriod, rounds, remaining := CalculateBulkSize(len(ptq.Fields), len(pendingTransactions))
		for i := 0; i < rounds; i++ {
			qry, args := ptq.InsertPendingTransactions(pendingTransactions[i*recordsPerPeriod : (i*recordsPerPeriod)+recordsPerPeriod])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
		if remaining > 0 {
			qry, args := ptq.InsertPendingTransactions(pendingTransactions[len(pendingTransactions)-remaining:])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
	}
	return queries, nil
}

// RecalibrateVersionedTable recalibrate table to clean up multiple latest rows due to import function
func (ptq *PendingTransactionQuery) RecalibrateVersionedTable() []string {
	return []string{
		fmt.Sprintf(
			"update %s set latest = false where latest = true AND (transaction_hash, block_height) NOT IN "+
				"(select t2.transaction_hash, max(t2.block_height) from %s t2 group by t2.transaction_hash)",
			ptq.getTableName(), ptq.getTableName()),
		fmt.Sprintf(
			"update %s set latest = true where latest = false AND (transaction_hash, block_height) IN "+
				"(select t2.transaction_hash, max(t2.block_height) from %s t2 group by t2.transaction_hash)",
			ptq.getTableName(), ptq.getTableName()),
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
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", ptq.TableName),
			height,
		},
		{
			fmt.Sprintf("UPDATE %s SET latest = ? WHERE latest = ? AND (transaction_hash, block_height"+
				") IN (SELECT t2.transaction_hash, MAX(t2.block_height) "+
				"FROM %s as t2 GROUP BY t2.transaction_hash)",
				ptq.getTableName(),
				ptq.getTableName(),
			),
			1, 0,
		},
	}
}

func (ptq *PendingTransactionQuery) SelectDataForSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(
		"SELECT %s FROM %s WHERE (transaction_hash, block_height) IN (SELECT t2.transaction_hash, MAX(t2.block_height) FROM %s as t2 "+
			"WHERE t2.block_height >= %d AND t2.block_height <= %d AND t2.block_height != 0 GROUP BY t2.transaction_hash) ORDER BY block_height",
		strings.Join(ptq.Fields, ","),
		ptq.TableName,
		ptq.TableName,
		fromHeight,
		toHeight,
	)
}

// TrimDataBeforeSnapshot delete entries to assure there are no duplicates before applying a snapshot
func (ptq *PendingTransactionQuery) TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(`DELETE FROM %s WHERE block_height >= %d AND block_height <= %d AND block_height != 0`,
		ptq.TableName, fromHeight, toHeight)
}
