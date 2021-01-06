package transaction

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

// Atomic TX
var _, innerTX = GetFixtureForSpecificTransaction(
	0, 123456789,
	senderAddress1, nil,
	8, model.TransactionType_SendMoneyTransaction,
	&model.SendMoneyTransactionBody{
		Amount: 10,
	},
	false, false,
)
var atomicBody, _ = GetFixtureForAtomicTransaction(
	map[string][][]byte{
		senderAddress1PassPhrase: {
			innerTX,
		},
	},
)

type (
	mockTypeActionAtomicTransaction struct {
		TypeAction
		WantErr bool
	}
)

func (m *mockTypeActionAtomicTransaction) ApplyConfirmed(int64) error {
	if m.WantErr {
		return errors.New("just want to")
	}
	return nil
}
func (m *mockTypeActionAtomicTransaction) Validate(bool) error {
	if m.WantErr {
		return errors.New("just want to")
	}
	return nil
}

type (
	mockTypeActionSwitcherAtomicTransaction struct {
		WantErr bool
		TypeActionSwitcher
	}
)

func (m *mockTypeActionSwitcherAtomicTransaction) GetTransactionType(*model.Transaction) (TypeAction, error) {
	return &mockTypeActionAtomicTransaction{
		WantErr: m.WantErr,
	}, nil
}

type (
	mockQueryExecutorAtomicTransactionApplyConfirmed struct {
		query.Executor
		WantErr bool
	}
)

func (m *mockQueryExecutorAtomicTransactionApplyConfirmed) ExecuteTransactions([][]interface{}) error {
	if m.WantErr {
		return sql.ErrConnDone
	}
	return nil
}
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
		{
			name: "WantErr:ApplyingInnerFail",
			fields: fields{
				AccountBalanceHelper: &mockAccountBalanceHelperSuccess{},
				Body:                 atomicBody,
				TransactionUtil:      &Util{},
				TypeActionSwitcher:   &mockTypeActionSwitcherAtomicTransaction{WantErr: true},
			},
			wantErr: true,
		},
		{
			name: "WantErr:ExecuteTransactionFail",
			fields: fields{
				AccountBalanceHelper:   &mockAccountBalanceHelperSuccess{},
				Body:                   atomicBody,
				TransactionUtil:        &Util{},
				TypeActionSwitcher:     &mockTypeActionSwitcherAtomicTransaction{},
				QueryExecutor:          &mockQueryExecutorAtomicTransactionApplyConfirmed{WantErr: true},
				TransactionQuery:       query.NewTransactionQuery(&chaintype.MainChain{}),
				AtomicTransactionQuery: query.NewAtomicTransactionQuery(),
			},
			wantErr: true,
		},
		{
			name: "WantSuccess",
			fields: fields{
				AccountBalanceHelper:   &mockAccountBalanceHelperSuccess{},
				Body:                   atomicBody,
				TransactionUtil:        &Util{},
				TypeActionSwitcher:     &mockTypeActionSwitcherAtomicTransaction{},
				QueryExecutor:          &mockQueryExecutorAtomicTransactionApplyConfirmed{},
				TransactionQuery:       query.NewTransactionQuery(&chaintype.MainChain{}),
				AtomicTransactionQuery: query.NewAtomicTransactionQuery(),
			},
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

func TestAtomicTransaction_Validate(t *testing.T) {
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
		dbTx bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "WantErr:EmptyAtomicInnerTransaction",
			fields: fields{
				Body: &model.AtomicTransactionBody{},
			},
			wantErr: true,
		},
		{
			name: "WantErr:InnerValidateFail",
			fields: fields{
				Body:                 atomicBody,
				AccountBalanceHelper: &mockAccountBalanceHelperSuccess{},
				TypeActionSwitcher:   &mockTypeActionSwitcherAtomicTransaction{WantErr: true},
				TransactionUtil:      &Util{},
			},
			wantErr: true,
		},
		{
			name: "WantSuccess",
			fields: fields{
				Body:                 atomicBody,
				AccountBalanceHelper: &mockAccountBalanceHelperSuccess{},
				TypeActionSwitcher:   &mockTypeActionSwitcherAtomicTransaction{},
				TransactionUtil:      &Util{},
				Signature:            crypto.NewSignature(),
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
			if err := tx.Validate(tt.args.dbTx); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
