package query

import (
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/contract"
)

type (
	TransactionQueryInterface interface {
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
			"sender_account_id",
			"recipient_account_id",
			"transaction_type",
			"fee",
			"timestamp",
			"transaction_hash",
			"transaction_body_length",
			"transaction_body_bytes",
			"signature",
		},
		TableName: "transaction",
		ChainType: chaintype,
	}
}

func (tq *TransactionQuery) getTableName() string {
	return tq.TableName
}

// GetTransaction get a single transaction from DB
func (tq *TransactionQuery) GetTransaction(ID int64) string {
	query := fmt.Sprintf("SELECT %s from %s", strings.Join(tq.Fields, ", "), tq.TableName)

	var queryParam []string
	if ID != 0 {
		queryParam = append(queryParam, fmt.Sprintf("id = %d", ID))
	}

	if len(queryParam) > 0 {
		query = query + " WHERE " + strings.Join(queryParam, " AND ")

	}
	return query
}

// GetTransactions get a set of transaction that satisfies the params from DB
func (tq *TransactionQuery) GetTransactions(limit uint32, offset uint64, senderAccountID, recipientAccountID string) string {
	query := fmt.Sprintf("SELECT %s from %s", strings.Join(tq.Fields, ", "), tq.TableName)

	var queryParam []string
	if senderAccountID != "" {
		queryParam = append(queryParam, fmt.Sprintf("sender_account_id = \"%s\"", senderAccountID))
	}
	if recipientAccountID != "" {
		queryParam = append(queryParam, fmt.Sprintf("recipient_account_id = \"%s\"", recipientAccountID))
	}

	if len(queryParam) > 0 {
		query = query + " WHERE " + strings.Join(queryParam, " AND ")

	}

	newLimit := limit
	if limit == 0 {
		newLimit = uint32(10)
	}

	query = query + " ORDER BY block_height, timestamp" + fmt.Sprintf(" LIMIT %d,%d", offset, newLimit)

	return query
}
