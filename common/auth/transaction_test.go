package auth

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
	"regexp"
	"testing"
)

type mockQueryExecutorSuccess struct {
	query.Executor
}

func (*mockQueryExecutorSuccess) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()

	getAccountBalanceByAccountID := "SELECT account_address,block_height,spendable_balance,balance,pop_revenue," +
		"latest FROM account_balance WHERE account_address = ? AND latest = 1"
	defer db.Close()
	switch qe {
	case getAccountBalanceByAccountID:
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"account_address", "block_height", "spendable_balance", "balance", "pop_revenue", "latest"},
		).AddRow("BCZ", 1, 10000, 10000, 0, 1))
	default:
		return nil, nil
	}

	rows, _ := db.Query(qe)
	return rows, nil
}

func TestValidateTransaction(t *testing.T) {
	type args struct {
		tx                  *model.Transaction
		queryExecutor       query.ExecutorInterface
		accountBalanceQuery query.AccountBalanceQueryInterface
		verifySignature     bool
	}
	tx := buildTransaction(
		1562893303, "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
		"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
	)
	txBytes, _ := util.GetTransactionBytes(tx, false)
	signature := (&crypto.Signature{}).Sign(txBytes, constant.NodeSignatureTypeDefault,
		"concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved")
	tx.Signature = signature
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "TestValidateTransaction:success",
			args: args{
				tx: buildTransaction(1562893303, "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
					"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE"),
				queryExecutor:       &mockQueryExecutorSuccess{},
				accountBalanceQuery: query.NewAccountBalanceQuery(),
				verifySignature:     false,
			},
			wantErr: false,
		},
		{
			name: "TestValidateTransaction:success - verify signature",
			args: args{
				tx:                  tx,
				queryExecutor:       &mockQueryExecutorSuccess{},
				accountBalanceQuery: query.NewAccountBalanceQuery(),
				verifySignature:     true,
			},
			wantErr: false,
		},
		{
			name: "ValidateTransaction:Fee<0",
			args: args{
				tx: &model.Transaction{
					Height: 1,
					Fee:    0,
				},
			},
			wantErr: true,
		},
		{
			name: "ValidateTransaction:SenderAddressEmpty",
			args: args{
				tx: &model.Transaction{
					Height: 1,
					Fee:    1,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateTransaction(tt.args.tx, tt.args.queryExecutor, tt.args.accountBalanceQuery,
				tt.args.verifySignature); (err != nil) != tt.wantErr {
				t.Errorf("ValidateTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func buildTransaction(timestamp int64, sender, recipient string) *model.Transaction {
	return &model.Transaction{
		Version:                 1,
		ID:                      2774809487,
		BlockID:                 1,
		Height:                  1,
		SenderAccountAddress:    sender,
		RecipientAccountAddress: recipient,
		TransactionType:         0,
		Fee:                     1,
		Timestamp:               timestamp,
		TransactionHash:         make([]byte, 32),
		TransactionBodyLength:   0,
		TransactionBodyBytes:    make([]byte, 0),
		TransactionBody:         nil,
		Signature:               make([]byte, 64),
	}
}
