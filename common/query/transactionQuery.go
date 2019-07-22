package query

import (
	"database/sql"
	"fmt"
	"math/big"
	"strings"

	"github.com/zoobc/zoobc-core/core/util"
	"golang.org/x/crypto/sha3"

	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
)

type (
	TransactionQueryInterface interface {
		InsertTransaction(tx *model.Transaction) (str string, args []interface{})
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
			"sender_account_type",
			"sender_account_address",
			"recipient_account_type",
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

// ExtractModel extract the model struct fields to the order of TransactionQuery.Fields
func (*TransactionQuery) ExtractModel(tx *model.Transaction) []interface{} {
	digest := sha3.New512()
	txBytes, _ := util.GetTransactionBytes(tx, true)
	_, _ = digest.Write(txBytes)
	hash := digest.Sum([]byte{})
	res := new(big.Int)
	txID := res.SetBytes([]byte{
		hash[7],
		hash[6],
		hash[5],
		hash[4],
		hash[3],
		hash[2],
		hash[1],
		hash[0],
	}).Int64()
	return []interface{}{
		txID,
		&tx.BlockID,
		&tx.Height,
		&tx.SenderAccountType,
		&tx.SenderAccountAddress,
		&tx.RecipientAccountType,
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
			&tx.SenderAccountType,
			&tx.SenderAccountAddress,
			&tx.RecipientAccountType,
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
