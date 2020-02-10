package service

import (
	"errors"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/observer"
)

type (
	mockQueryExecutorPostApprovalEscrowTX struct {
		query.Executor
	}
	mockMempoolServicePostApprovalEscrowTX struct {
		service.MempoolService
	}
	mockMempoolServicePostApprovalEscrowTXSuccess struct {
		service.MempoolService
	}
)

func (*mockQueryExecutorPostApprovalEscrowTX) BeginTx() error {
	return nil
}
func (*mockQueryExecutorPostApprovalEscrowTX) CommitTx() error {
	return nil
}
func (*mockQueryExecutorPostApprovalEscrowTX) RollbackTx() error {
	return nil
}
func (*mockQueryExecutorPostApprovalEscrowTX) ExecuteTransaction(query string, args ...interface{}) error {
	return nil
}

func (*mockMempoolServicePostApprovalEscrowTX) ValidateMempoolTransaction(mpTx *model.MempoolTransaction) error {
	return errors.New("test")
}
func (*mockMempoolServicePostApprovalEscrowTXSuccess) ValidateMempoolTransaction(mpTx *model.MempoolTransaction) error {
	return nil
}
func (*mockMempoolServicePostApprovalEscrowTXSuccess) AddMempoolTransaction(mpTx *model.MempoolTransaction) error {
	return nil
}

func TestEscrowApprovalService_PostApprovalEscrowTransaction(t *testing.T) {
	escrowApprovalTX, _ := transaction.GetFixtureForSpecificTransaction(
		8391609053770132621,
		1581301507,
		"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
		"",
		12,
		model.TransactionType_ApprovalEscrowTransaction,
		&model.ApprovalEscrowTransactionBody{
			Approval:      0,
			TransactionID: 0,
		},
		false,
		true,
	)

	type fields struct {
		Query              query.ExecutorInterface
		Signature          crypto.SignatureInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
		MempoolService     service.MempoolServiceInterface
		Observer           *observer.Observer
	}
	type args struct {
		chainType chaintype.ChainType
		request   *model.PostEscrowApprovalRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Transaction
		wantErr bool
	}{
		{
			name:    "WantError:ParseFailed",
			fields:  fields{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "WantError:ValidateMempoolFail",
			fields: fields{
				Query:              nil,
				Signature:          nil,
				ActionTypeSwitcher: &transaction.TypeSwitcher{},
				MempoolService:     &mockMempoolServicePostApprovalEscrowTX{},
				Observer:           nil,
			},
			args: args{
				chainType: &chaintype.MainChain{},
				request: &model.PostEscrowApprovalRequest{
					// CMD generator result
					ApprovalBytes: []byte{4, 0, 0, 0, 1, 18, 65, 64, 94, 0, 0, 0, 0, 44, 0, 0, 0, 66, 67,
						90, 69, 71, 79, 98, 51, 87, 78, 120, 51, 102, 68, 79, 86, 102, 57, 90, 83, 52,
						69, 106, 118, 79, 73, 118, 95, 85, 101, 87, 52, 84, 86, 66, 81, 74, 95, 54,
						116, 72, 75, 108, 69, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 12, 0, 0, 0, 0, 0,
						0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
						0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 127, 43, 121, 224, 188, 251, 104, 32, 33,
						31, 233, 85, 251, 4, 101, 152, 113, 200, 234, 74, 103, 93, 70, 65, 138, 156,
						228, 219, 116, 71, 142, 247, 143, 76, 10, 40, 12, 28, 239, 80, 139, 189, 131,
						3, 195, 107, 231, 115, 91, 102, 140, 99, 202, 60, 54, 52, 239, 107, 229, 161,
						26, 197, 71, 6},
				},
			},
			wantErr: true,
		},
		{
			name: "WantError:ValidateMempoolFail1",
			fields: fields{
				Query:     &mockQueryExecutorPostApprovalEscrowTX{},
				Signature: nil,
				ActionTypeSwitcher: &transaction.TypeSwitcher{
					Executor: &mockQueryExecutorPostApprovalEscrowTX{},
				},
				MempoolService: &mockMempoolServicePostApprovalEscrowTXSuccess{},
				Observer:       observer.NewObserver(),
			},
			args: args{
				chainType: &chaintype.MainChain{},
				request: &model.PostEscrowApprovalRequest{
					ApprovalBytes: []byte{4, 0, 0, 0, 1, 3, 191, 64, 94, 0, 0, 0, 0, 44, 0, 0, 0, 66, 67,
						90, 69, 71, 79, 98, 51, 87, 78, 120, 51, 102, 68, 79, 86, 102, 57, 90, 83, 52,
						69, 106, 118, 79, 73, 118, 95, 85, 101, 87, 52, 84, 86, 66, 81, 74, 95, 54,
						116, 72, 75, 108, 69, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 12, 0, 0, 0, 0, 0,
						0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
						0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 76, 37, 113, 191, 231, 41, 103, 98, 57, 67,
						169, 205, 172, 140, 249, 170, 166, 46, 82, 179, 192, 127, 37, 244, 251, 113,
						230, 236, 118, 172, 62, 37, 88, 24, 121, 3, 105, 200, 185, 224, 142, 161, 63,
						6, 209, 55, 7, 108, 96, 59, 240, 182, 151, 95, 41, 202, 157, 149, 39, 144,
						135, 240, 25, 6},
				},
			},
			want: escrowApprovalTX,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eas := &EscrowApprovalService{
				Query:              tt.fields.Query,
				Signature:          tt.fields.Signature,
				ActionTypeSwitcher: tt.fields.ActionTypeSwitcher,
				MempoolService:     tt.fields.MempoolService,
				Observer:           tt.fields.Observer,
			}
			got, err := eas.PostApprovalEscrowTransaction(tt.args.chainType, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("PostApprovalEscrowTransaction() error = \n%v, wantErr \n%v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PostApprovalEscrowTransaction() got = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}
