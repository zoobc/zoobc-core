package query

import (
	"fmt"
	"strings"

	"github.com/zoobc/zoobc-core/common/contract"

	"github.com/zoobc/zoobc-core/common/model"
)

type (
	MempoolQueryInterface interface {
		GetMempoolTransactions() string
		InsertMempoolTransaction() string
		ExtractModel(block model.MempoolTransaction) []interface{}
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
			"ID",
			"FeePerByte",
			"ArrivalTimestamp",
			"TransactionBytes",
		},
		TableName: "mempool",
		ChainType: chaintype,
	}
}

func (bq *MempoolQuery) getTableName() string {
	return bq.ChainType.GetTablePrefix() + "_" + bq.TableName
}

// GetMempoolTransactions returns query string to get multiple mempool transactions
func (bq *MempoolQuery) GetMempoolTransactions() string {
	return fmt.Sprintf("SELECT %s FROM %s", strings.Join(bq.Fields, ", "), bq.getTableName())
}

func (bq *MempoolQuery) InsertMempoolTransaction() string {
	var value = ":" + bq.Fields[0]
	for _, field := range bq.Fields[1:] {
		value += (", :" + field)

	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s)",
		bq.getTableName(), strings.Join(bq.Fields, ", "), value)
	return query
}

// ExtractModel extract the model struct fields to the order of MempoolQuery.Fields
func (*MempoolQuery) ExtractModel(mempool model.MempoolTransaction) []interface{} {
	return []interface{}{mempool.ID, mempool.FeePerByte, mempool.ArrivalTimestamp, mempool.TransactionBytes}
}
