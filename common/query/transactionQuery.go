package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	TransactionQueryInterface interface {
		InsertTransaction(tx *model.Transaction) (str string, args []interface{})
		GetTransaction(id int64) string
		GetTransactions(limit uint32, offset uint64) string
		GetTransactionsByBlockID(blockID int64) (str string, args []interface{})
		ExtractModel(tx *model.Transaction) []interface{}
		BuildModel(txs []*model.Transaction, rows *sql.Rows) []*model.Transaction
		DeleteTransactions(id int64) string
		Scan(tx *model.Transaction, row *sql.Row) error
	}

	TransactionQuery struct {
		Fields    []string
		TableName string
		ChainType chaintype.ChainType
	}
)

// NewTransactionQuery returns TransactionQuery instance
func NewTransactionQuery(chaintype chaintype.ChainType) *TransactionQuery {
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
			"transaction_index",
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
	var value = fmt.Sprintf("?%s", strings.Repeat(", ?", len(tq.Fields)-1))
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s)",
		tq.getTableName(), strings.Join(tq.Fields, ", "), value)
	return query, tq.ExtractModel(tx)
}

func (tq *TransactionQuery) GetTransactionsByBlockID(blockID int64) (str string, args []interface{}) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE block_id = ?", strings.Join(tq.Fields, ", "), tq.getTableName())
	return query, []interface{}{blockID}
}

// DeleteTransactions. delete some transactions according to timestamp
func (tq *TransactionQuery) DeleteTransactions(id int64) string {
	return fmt.Sprintf("DELETE FROM %v WHERE height >= (SELECT height FROM %v WHERE ID = %v)", tq.getTableName(), tq.getTableName(), id)
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
		&tx.Version,
		&tx.TransactionIndex,
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
			&tx.TransactionIndex,
		)
		txs = append(txs, &tx)
	}
	return txs
}

func (*TransactionQuery) Scan(tx *model.Transaction, row *sql.Row) error {
	err := row.Scan(
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
		&tx.TransactionIndex,
	)
	return err
}

// Rollback delete records `WHERE height > "height"
func (tq *TransactionQuery) Rollback(height uint32) (multiQueries [][]interface{}) {
	return [][]interface{}{
		{
			fmt.Sprintf("DELETE FROM %s WHERE block_height > ?", tq.getTableName()),
			height,
		},
	}
}
