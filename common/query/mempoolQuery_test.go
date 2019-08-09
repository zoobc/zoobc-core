package query

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"

	"github.com/zoobc/zoobc-core/common/chaintype"

	"github.com/zoobc/zoobc-core/common/contract"
)

var mockMempoolQuery = &MempoolQuery{
	ChainType: &chaintype.MainChain{},
	TableName: "mempool",
	Fields: []string{
		"id", "fee_per_byte", "arrival_timestamp", "transaction_bytes",
		"sender_account_id", "recipient_account_id",
	},
}

var mockMempool = &model.MempoolTransaction{
	ID:                      1,
	ArrivalTimestamp:        1000,
	FeePerByte:              10,
	TransactionBytes:        []byte{1, 2, 3, 4, 5},
	SenderAccountAddress:    "BCZ",
	RecipientAccountAddress: "ZCB",
}

func TestNewMempoolQuery(t *testing.T) {
	type args struct {
		chaintype contract.ChainType
	}
	tests := []struct {
		name string
		args args
		want *MempoolQuery
	}{
		{
			name: "NewMempoolQuery",
			args: args{
				chaintype: &chaintype.MainChain{},
			},
			want: mockMempoolQuery,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMempoolQuery(tt.args.chaintype); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMempoolQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMempoolQuery_getTableName(t *testing.T) {
	t.Run("MempoolQuery-getTableName:mainchain", func(t *testing.T) {
		tableName := mockMempoolQuery.getTableName()
		if tableName != mockMempoolQuery.TableName {
			t.Errorf("getTableName mainchain should return mempool")
		}
	})
}

func TestMempoolQuery_GetMempoolTransactions(t *testing.T) {
	t.Run("GetMempoolTransactions:success", func(t *testing.T) {
		q := mockMempoolQuery.GetMempoolTransactions()
		wantQ := "SELECT id, fee_per_byte, arrival_timestamp, transaction_bytes, sender_account_id" +
			", recipient_account_id FROM mempool"
		if q != wantQ {
			t.Errorf("query returned wrong: get: %s\nwant: %s", q, wantQ)
		}
	})
}

func TestMempoolQuery_GetMempoolTransaction(t *testing.T) {
	t.Run("GetMempoolTransaction:success", func(t *testing.T) {
		q := mockMempoolQuery.GetMempoolTransaction()
		wantQ := "SELECT id, fee_per_byte, arrival_timestamp, transaction_bytes, sender_account_id," +
			" recipient_account_id FROM mempool WHERE id = :id"
		if q != wantQ {
			t.Errorf("query returned wrong: get: %s\nwant: %s", q, wantQ)
		}
	})
}

func TestMempoolQuery_InsertMempoolTransaction(t *testing.T) {
	t.Run("InsertMempoolTransaction:success", func(t *testing.T) {
		q := mockMempoolQuery.InsertMempoolTransaction()
		wantQ := "INSERT INTO mempool (id, fee_per_byte, arrival_timestamp, transaction_bytes, sender_account_id," +
			" recipient_account_id) VALUES(:id, :fee_per_byte, :arrival_timestamp, :transaction_bytes," +
			" :sender_account_id, :recipient_account_id)"
		if q != wantQ {
			t.Errorf("query returned wrong: get: %s\nwant: %s", q, wantQ)
		}
	})
}

func TestMempoolQuery_DeleteMempoolTransaction(t *testing.T) {
	t.Run("DeleteMempoolTransaction:success", func(t *testing.T) {
		q := mockMempoolQuery.DeleteMempoolTransaction()
		wantQ := "DELETE FROM mempool WHERE id = :id"
		if q != wantQ {
			t.Errorf("query returned wrong: get: %s\nwant: %s", q, wantQ)
		}
	})
}

func TestMempoolQuery_DeleteMempoolTransactions(t *testing.T) {
	t.Run("DeleteMempoolTransactions:success", func(t *testing.T) {
		q := mockMempoolQuery.DeleteMempoolTransactions()
		wantQ := "DELETE FROM mempool WHERE id IN (:ids)"
		if q != wantQ {
			t.Errorf("query returned wrong: get: %s\nwant: %s", q, wantQ)
		}
	})
}

func TestMempoolQuery_ExtractModel(t *testing.T) {
	t.Run("MempoolQuery-ExtractModel:success", func(t *testing.T) {
		res := mockMempoolQuery.ExtractModel(mockMempool)
		want := []interface{}{
			mockMempool.ID,
			mockMempool.FeePerByte,
			mockMempool.ArrivalTimestamp,
			mockMempool.TransactionBytes,
			mockMempool.SenderAccountAddress,
			mockMempool.RecipientAccountAddress,
		}
		if !reflect.DeepEqual(res, want) {
			t.Errorf("arguments returned wrong: get: %v\nwant: %v", res, want)
		}
	})
}
