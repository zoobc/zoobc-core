package query

import (
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
)

func TestGetTransaction(t *testing.T) {
	transactionQuery := NewTransactionQuery(chaintype.GetChainType(0))

	type paramsStruct struct {
		ID int64
	}

	tests := []struct {
		name   string
		params *paramsStruct
		want   string
	}{
		{
			name:   "transaction query without condition",
			params: &paramsStruct{},
			want: "SELECT id, block_id, block_height, sender_account_type, sender_account_address, " +
				"recipient_account_type, recipient_account_address, transaction_type, fee, timestamp, " +
				"transaction_hash, transaction_body_length, transaction_body_bytes, signature, version from \"transaction\"",
		},
		{
			name: "transaction query with ID param only",
			params: &paramsStruct{
				ID: 1,
			},
			want: "SELECT id, block_id, block_height, sender_account_type, sender_account_address, " +
				"recipient_account_type, recipient_account_address, transaction_type, fee, timestamp, " +
				"transaction_hash, transaction_body_length, transaction_body_bytes, signature, version from \"transaction\" " +
				"WHERE id = 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := transactionQuery.GetTransaction(tt.params.ID)
			if query != tt.want {
				t.Errorf("GetTransactionError() \ngot = %v \nwant = %v", query, tt.want)
				return
			}
		})
	}
}

func TestGetTransactions(t *testing.T) {
	transactionQuery := NewTransactionQuery(chaintype.GetChainType(0))

	type paramsStruct struct {
		Limit  uint32
		Offset uint64
	}

	tests := []struct {
		name   string
		params *paramsStruct
		want   string
	}{
		{
			name:   "transactions query without condition",
			params: &paramsStruct{},
			want: "SELECT id, block_id, block_height, sender_account_type, sender_account_address, " +
				"recipient_account_type, recipient_account_address, transaction_type, fee, timestamp, " +
				"transaction_hash, transaction_body_length, transaction_body_bytes, signature, version from " +
				"\"transaction\" ORDER BY block_height, timestamp LIMIT 0,10",
		},
		{
			name: "transactions query with limit",
			params: &paramsStruct{
				Limit: 10,
			},
			want: "SELECT id, block_id, block_height, sender_account_type, sender_account_address, " +
				"recipient_account_type, recipient_account_address, transaction_type, fee, timestamp, " +
				"transaction_hash, transaction_body_length, transaction_body_bytes, signature, version from " +
				"\"transaction\" ORDER BY block_height, timestamp LIMIT 0,10",
		},
		{
			name: "transactions query with offset",
			params: &paramsStruct{
				Offset: 20,
			},
			want: "SELECT id, block_id, block_height, sender_account_type, sender_account_address, " +
				"recipient_account_type, recipient_account_address, transaction_type, fee, timestamp, " +
				"transaction_hash, transaction_body_length, transaction_body_bytes, signature, version from " +
				"\"transaction\" ORDER BY block_height, timestamp LIMIT 20,10",
		},
		{
			name: "transactions query with all the params",
			params: &paramsStruct{
				Limit:  10,
				Offset: 20,
			},
			want: "SELECT id, block_id, block_height, sender_account_type, sender_account_address, " +
				"recipient_account_type, recipient_account_address, transaction_type, fee, timestamp, " +
				"transaction_hash, transaction_body_length, transaction_body_bytes, signature, version from " +
				"\"transaction\" ORDER BY block_height, timestamp LIMIT 20,10",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := transactionQuery.GetTransactions(tt.params.Limit, tt.params.Offset)
			if query != tt.want {
				t.Errorf("GetTransactionError() \ngot = %v \nwant = %v", query, tt.want)
				return
			}
		})
	}
}
