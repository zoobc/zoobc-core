package query

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

func TestAtomicTransactionQuery_InsertAtomicTransactions(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		atomics []*model.Atomic
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantStr  string
		wantArgs []interface{}
	}{
		{
			name:   "WantSuccess",
			fields: fields(*NewAtomicTransactionQuery()),
			args: args{
				atomics: []*model.Atomic{
					{
						ID:                  123456789,
						TransactionID:       1234567890,
						SenderAddress:       []byte{},
						BlockHeight:         1,
						UnsignedTransaction: []byte{},
						Signature:           []byte{},
						AtomicIndex:         0,
					},
				},
			},
			wantStr: "INSERT INTO atomic_transaction (id, transaction_id, sender_address, block_height, unsigned_transaction, signature, atomic_index) " +
				"VALUES (?, ?, ?, ?, ?, ?, ?)",
			wantArgs: []interface{}{
				int64(123456789),
				int64(1234567890),
				[]byte{},
				uint32(1),
				[]byte{},
				[]byte{},
				uint32(0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AtomicTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := a.InsertAtomicTransactions(tt.args.atomics)
			if gotStr != tt.wantStr {
				t.Errorf("InsertAtomicTransactions() gotStr = \n%v, want \n%v", gotStr, tt.wantStr)
				return
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("InsertAtomicTransactions() gotArgs = \n%v, want \n%v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestAtomicTransactionQuery_ExtractModel(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		atomic *model.Atomic
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []interface{}
	}{
		{
			name:   "WantSuccess",
			fields: fields(*NewAtomicTransactionQuery()),
			args: args{
				atomic: &model.Atomic{
					ID:                  123456789,
					TransactionID:       1234567890,
					SenderAddress:       []byte{},
					BlockHeight:         1,
					UnsignedTransaction: []byte{},
					Signature:           []byte{},
					AtomicIndex:         0,
				},
			},
			want: []interface{}{
				int64(123456789),
				int64(1234567890),
				[]byte{},
				uint32(1),
				[]byte{},
				[]byte{},
				uint32(0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AtomicTransactionQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := a.ExtractModel(tt.args.atomic); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractModel() = %v, want %v", got, tt.want)
			}
		})
	}
}
