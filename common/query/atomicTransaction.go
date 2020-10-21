package query

import (
	"database/sql"
	"fmt"
	"strings"

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
