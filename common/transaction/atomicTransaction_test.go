package transaction

import (
	"testing"

	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

func TestAtomicTransaction_ApplyConfirmed(t *testing.T) {
	type fields struct {
		ID                     int64
		Fee                    int64
		SenderAddress          []byte
		Height                 uint32
		Body                   *model.AtomicTransactionBody
		AtomicTransactionQuery query.AtomicTransactionQueryInterface
		Escrow                 *model.Escrow
		EscrowQuery            query.EscrowTransactionQueryInterface
		QueryExecutor          query.ExecutorInterface
		TransactionQuery       query.TransactionQueryInterface
		TypeActionSwitcher     TypeActionSwitcher
		AccountBalanceHelper   AccountBalanceHelperInterface
		EscrowFee              fee.FeeModelInterface
		NormalFee              fee.FeeModelInterface
		TransactionUtil        UtilInterface
		Signature              crypto.SignatureInterface
	}
	type args struct {
		blockTimestamp int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "WantErr:AddAccountBalance",
			fields: fields{
				AccountBalanceHelper: &mockAccountBalanceHelperFail{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &AtomicTransaction{
				ID:                     tt.fields.ID,
				Fee:                    tt.fields.Fee,
				SenderAddress:          tt.fields.SenderAddress,
				Height:                 tt.fields.Height,
				Body:                   tt.fields.Body,
				AtomicTransactionQuery: tt.fields.AtomicTransactionQuery,
				Escrow:                 tt.fields.Escrow,
				EscrowQuery:            tt.fields.EscrowQuery,
				QueryExecutor:          tt.fields.QueryExecutor,
				TransactionQuery:       tt.fields.TransactionQuery,
				TypeActionSwitcher:     tt.fields.TypeActionSwitcher,
				AccountBalanceHelper:   tt.fields.AccountBalanceHelper,
				EscrowFee:              tt.fields.EscrowFee,
				NormalFee:              tt.fields.NormalFee,
				TransactionUtil:        tt.fields.TransactionUtil,
				Signature:              tt.fields.Signature,
			}
			if err := tx.ApplyConfirmed(tt.args.blockTimestamp); (err != nil) != tt.wantErr {
				t.Errorf("ApplyConfirmed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
