package query

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/contract"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	MempoolQueryInterface interface {
		GetMempoolTransactions() string
		GetMempoolTransaction() string
		InsertMempoolTransaction() string
		DeleteMempoolTransaction() string
		DeleteMempoolTransactions() string
		ExtractModel(block *model.MempoolTransaction) []interface{}
		BuildModel(mempools []*model.MempoolTransaction, rows *sql.Rows) []*model.MempoolTransaction
	}

	MempoolQuery struct {
		Fields    []string
		TableName string
		ChainType contract.ChainType
	}
)

// NewMempoolQuery returns MempoolQuery instance
func NewMempoolQuery(chaintype contract.ChainType) *MempoolQuery {
	return &MempoolQuery{
		Fields: []string{
			"id", "fee_per_byte", "arrival_timestamp", "transaction_bytes", "sender_account_address",
			"recipient_account_address",
		},
		TableName: "mempool",
		ChainType: chaintype,
	}
}

func (mpq *MempoolQuery) getTableName() string {
	return mpq.TableName
}

// GetMempoolTransactions returns query string to get multiple mempool transactions
func (mpq *MempoolQuery) GetMempoolTransactions() string {
	return fmt.Sprintf("SELECT %s FROM %s", strings.Join(mpq.Fields, ", "), mpq.getTableName())
}

// GetMempoolTransactions returns query string to get multiple mempool transactions
func (mpq *MempoolQuery) GetMempoolTransaction() string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE id = :id", strings.Join(mpq.Fields, ", "), mpq.getTableName())
}

func (mpq *MempoolQuery) InsertMempoolTransaction() string {
	var value = ":" + mpq.Fields[0]
	for _, field := range mpq.Fields[1:] {
		value += (", :" + field)

	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s)",
		mpq.getTableName(), strings.Join(mpq.Fields, ", "), value)
	return query
}

// DeleteMempoolTransaction delete one mempool transaction by id
func (mpq *MempoolQuery) DeleteMempoolTransaction() string {
	return fmt.Sprintf("DELETE FROM %s WHERE id = :id", mpq.getTableName())
}

// DeleteMempoolTransaction delete one mempool transaction by id
func (mpq *MempoolQuery) DeleteMempoolTransactions() string {
	return fmt.Sprintf("DELETE FROM %s WHERE id IN (:ids)", mpq.getTableName())
}

// ExtractModel extract the model struct fields to the order of MempoolQuery.Fields
func (*MempoolQuery) ExtractModel(mempool *model.MempoolTransaction) []interface{} {
	return []interface{}{
		mempool.ID,
		mempool.FeePerByte,
		mempool.ArrivalTimestamp,
		mempool.TransactionBytes,
		mempool.SenderAccountAddress,
		mempool.RecipientAccountAddress,
	}
}

// BuildModel will only be used for mapping the result of `select` query, which will guarantee that
// the result of build model will be correctly mapped based on the modelQuery.Fields order.
func (*MempoolQuery) BuildModel(mempools []*model.MempoolTransaction, rows *sql.Rows) []*model.MempoolTransaction {
	for rows.Next() {
		var mempool model.MempoolTransaction
		_ = rows.Scan(
			&mempool.ID,
			&mempool.FeePerByte,
			&mempool.ArrivalTimestamp,
			&mempool.TransactionBytes,
			&mempool.SenderAccountAddress,
			&mempool.RecipientAccountAddress,
		)
		mempools = append(mempools, &mempool)
	}
	return mempools
}
