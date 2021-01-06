package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	AtomicTransactionQuery struct {
		Fields    []string
		TableName string
	}
	AtomicTransactionQueryInterface interface {
		InsertAtomicTransactions(atomics []*model.Atomic) (str string, args []interface{})
		ExtractModel(atomic *model.Atomic) []interface{}
		BuildModel(atomics []*model.Atomic, rows *sql.Rows) ([]*model.Atomic, error)
		Scan(atomic *model.Atomic, row *sql.Row) error
	}
)

func NewAtomicTransactionQuery() *AtomicTransactionQuery {
	return &AtomicTransactionQuery{
		Fields: []string{
			"id",
			"transaction_id",
			"sender_address",
			"block_height",
			"unsigned_transaction",
			"signature",
			"atomic_index",
		},
		TableName: "atomic_transaction",
	}
}

func (a *AtomicTransactionQuery) InsertAtomicTransactions(atomics []*model.Atomic) (str string, args []interface{}) {
	if len(atomics) > 0 {
		str = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES ",
			a.getTableName(),
			strings.Join(a.Fields, ", "),
		)
		for k, atomic := range atomics {
			str += fmt.Sprintf(
				"(?%s)",
				strings.Repeat(", ?", len(a.Fields)-1),
			)
			if k < len(atomics)-1 {
				str += ","
			}

			args = append(args, a.ExtractModel(atomic)...)
		}
	}
	return str, args
}

func (a *AtomicTransactionQuery) ExtractModel(atomic *model.Atomic) []interface{} {
	return []interface{}{
		atomic.GetID(),
		atomic.GetTransactionID(),
		atomic.GetSenderAddress(),
		atomic.GetBlockHeight(),
		atomic.GetUnsignedTransaction(),
		atomic.GetSignature(),
		atomic.GetAtomicIndex(),
	}
}

func (a *AtomicTransactionQuery) BuildModel(atomics []*model.Atomic, rows *sql.Rows) ([]*model.Atomic, error) {
	for rows.Next() {
		var (
			atomic model.Atomic
			err    error
		)
		err = rows.Scan(
			&atomic.ID,
			&atomic.TransactionID,
			&atomic.SenderAddress,
			&atomic.BlockHeight,
			&atomic.UnsignedTransaction,
			&atomic.Signature,
			&atomic.AtomicIndex,
		)
		if err != nil {
			return atomics, err
		}
		atomics = append(atomics, &atomic)
	}
	return atomics, nil
}

func (a *AtomicTransactionQuery) Scan(atomic *model.Atomic, row *sql.Row) error {
	return row.Scan(
		&atomic.ID,
		&atomic.TransactionID,
		&atomic.SenderAddress,
		&atomic.BlockHeight,
		&atomic.UnsignedTransaction,
		&atomic.Signature,
		&atomic.AtomicIndex,
	)
}

func (a *AtomicTransactionQuery) getTableName() string {
	return a.TableName
}

func (a *AtomicTransactionQuery) Rollback(height uint32) (queries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE height > ?", a.getTableName()),
			height,
		},
	}
}

func (a *AtomicTransactionQuery) SelectDataForSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(
		"SELECT %s FROM %s WHERE (transaction_id, block_height) IN "+
			"(SELECT tx.transaction_id, MAX(tx.block_height) FROM %s tx "+
			"WHERE tx.block_height >= %d AND tx.block_height <= %d AND tx.block_height != 0 GROUP BY tx.transaction_id) "+
			"ORDER BY block_height",
		strings.Join(a.Fields, ", "),
		a.getTableName(),
		a.getTableName(),
		fromHeight,
		toHeight,
	)
}

func (a *AtomicTransactionQuery) TrimDataBeforeSnapshot(fromHeight, toHeight uint32) string {
	return fmt.Sprintf(
		"DELETE FROM %s WHERE block_height >= %d AND block_height <= %d AND block_height != 0",
		a.getTableName(),
		fromHeight,
		toHeight,
	)

}

func (a *AtomicTransactionQuery) ImportSnapshot(payload interface{}) ([][]interface{}, error) {
	var (
		queries [][]interface{}
	)

	atomicTXs, ok := payload.([]*model.Atomic)
	if !ok {
		return nil, blocker.NewBlocker(blocker.DBErr, "ImportSnapshotCannotCastTo"+a.TableName)
	}
	if len(atomicTXs) > 0 {
		var (
			qry  string
			args []interface{}
		)
		recordsPerPeriod, rounds, remaining := CalculateBulkSize(len(a.Fields), len(atomicTXs))
		for i := 0; i < rounds; i++ {
			qry, args = a.InsertAtomicTransactions(atomicTXs[i*recordsPerPeriod : (i*recordsPerPeriod)+recordsPerPeriod])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
		if remaining > 0 {
			qry, args = a.InsertAtomicTransactions(atomicTXs[len(atomicTXs)-remaining:])
			queries = append(queries, append([]interface{}{qry}, args...))
		}
	}
	return queries, nil
}

func (a *AtomicTransactionQuery) RecalibrateVersionedTable() []string {
	return []string{}
}
