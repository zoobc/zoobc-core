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
		GetTransactionsByIds(txIds []int64) (str string, args []interface{})
		GetTransactionsByBlockID(blockID int64) (str string, args []interface{})
		ExtractModel(tx *model.Transaction) []interface{}
		BuildModel(txs []*model.Transaction, rows *sql.Rows) ([]*model.Transaction, error)
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
			"multisig_child",
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
	query := fmt.Sprintf(
		"SELECT %s from %s WHERE id = %d",
		strings.Join(tq.Fields, ", "),
		tq.getTableName(),
		id,
	)
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
	query := fmt.Sprintf("SELECT %s FROM %s WHERE block_id = ? AND multisig_child = false "+
		"ORDER BY transaction_index ASC", strings.Join(tq.Fields, ", "), tq.getTableName())
	return query, []interface{}{blockID}
}

func (tq *TransactionQuery) GetTransactionsByIds(txIds []int64) (str string, args []interface{}) {

	for _, id := range txIds {
		args = append(args, id)
	}
	return fmt.Sprintf(
			"SELECT %s FROM %s WHERE multisig_child = false AND id IN(?%s)",
			strings.Join(tq.Fields, ", "),
			tq.getTableName(),
			strings.Repeat(", ?", len(txIds)-1),
		),
		args
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
		&tx.MultisigChild,
	}
}

func (*TransactionQuery) BuildModel(txs []*model.Transaction, rows *sql.Rows) ([]*model.Transaction, error) {
	for rows.Next() {
		var (
			tx  model.Transaction
			err error
		)
		err = rows.Scan(
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
			&tx.MultisigChild,
		)
		if err != nil {
			return nil, err
		}
		txs = append(txs, &tx)
	}
	return txs, nil
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
		&tx.MultisigChild,
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
