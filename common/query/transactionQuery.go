package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	TransactionQueryInterface interface {
		InsertTransaction(tx *model.Transaction) (str string, args []interface{})
		GetTransaction(id int64) string
		GetTransactions(limit uint32, offset uint64) string
		GetTransactionsByBlockID(blockID int64) (str string, argss []interface{})
		ExtractModel(tx *model.Transaction) []interface{}
		BuildModel(transactions []*model.Transaction, rows *sql.Rows) []*model.Transaction
	}

	TransactionQuery struct {
		Fields    []string
		TableName string
		ChainType contract.ChainType
	}
)

// NewTransactionQuery returns TransactionQuery instance
func NewTransactionQuery(chaintype contract.ChainType) *TransactionQuery {
	return &TransactionQuery{
		Fields: []string{
			"id",
			"block_id",
			"block_height",
			"sender_account_address",
			"recipient_account_address",
			"transaction_type",
			"fee",
			"timestamp",
			"transaction_hash",
			"transaction_body_length",
			"transaction_body_bytes",
			"signature",
			"version",
		},
		TableName: "\"transaction\"",
		ChainType: chaintype,
	}
}

func (tq *TransactionQuery) getTableName() string {
	return tq.TableName
}

// GetTransaction get a single transaction from DB
func (tq *TransactionQuery) GetTransaction(id int64) string {
	query := fmt.Sprintf("SELECT %s from %s", strings.Join(tq.Fields, ", "), tq.getTableName())

	var queryParam []string
	if id != 0 {
		queryParam = append(queryParam, fmt.Sprintf("id = %d", id))
	}

	if len(queryParam) > 0 {
		query = query + " WHERE " + strings.Join(queryParam, " AND ")

	}
	return query
}

// GetTransactions get a set of transaction that satisfies the params from DB
func (tq *TransactionQuery) GetTransactions(limit uint32, offset uint64) string {
	query := fmt.Sprintf("SELECT %s from %s", strings.Join(tq.Fields, ", "), tq.getTableName())

	newLimit := limit
	if limit == 0 {
		newLimit = uint32(10)
	}

	query = query + " ORDER BY block_height, timestamp" + fmt.Sprintf(" LIMIT %d,%d", offset, newLimit)

	return query
}

// InsertTransaction inserts a new transaction into DB
func (tq *TransactionQuery) InsertTransaction(tx *model.Transaction) (str string, args []interface{}) {
	var value = fmt.Sprintf("? %s", strings.Repeat(", ?", len(tq.Fields)-1))
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s)",
		tq.getTableName(), strings.Join(tq.Fields, ", "), value)
	return query, tq.ExtractModel(tx)
}

func (tq *TransactionQuery) GetTransactionsByBlockID(blockID int64) (str string, argss []interface{}) {
	query := fmt.Sprintf("SELECT %s from %s WHERE block_id = ?", strings.Join(tq.Fields, ", "), tq.getTableName())
	return query, []interface{}{blockID}
}

// ExtractModel extract the model struct fields to the order of TransactionQuery.Fields
func (*TransactionQuery) ExtractModel(tx *model.Transaction) []interface{} {
	return []interface{}{
		&tx.ID,
		&tx.BlockID,
		&tx.Height,
		&tx.SenderAccountAddress,
		&tx.RecipientAccountAddress,
		&tx.TransactionType,
		&tx.Fee,
		&tx.Timestamp,
		&tx.TransactionHash,
		&tx.TransactionBodyLength,
		&tx.TransactionBodyBytes,
		&tx.Signature,
		tx.Version,
	}
}

func (*TransactionQuery) BuildModel(txs []*model.Transaction, rows *sql.Rows) []*model.Transaction {
	for rows.Next() {
		var tx model.Transaction
		_ = rows.Scan(
			&tx.ID,
			&tx.BlockID,
			&tx.Height,
			&tx.SenderAccountAddress,
			&tx.RecipientAccountAddress,
			&tx.TransactionType,
			&tx.Fee,
			&tx.Timestamp,
			&tx.TransactionHash,
			&tx.TransactionBodyLength,
			&tx.TransactionBodyBytes,
			&tx.Signature,
			&tx.Version,
		)
		txs = append(txs, &tx)
	}
	return txs
}
