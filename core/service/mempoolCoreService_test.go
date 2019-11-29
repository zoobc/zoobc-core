package service

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/observer"
)

var (
	getTxByIDQuery = "SELECT id, block_height, fee_per_byte, arrival_timestamp, transaction_bytes, sender_account_address, " +
		"recipient_account_address FROM mempool WHERE id = :id"
	mockMempoolQuery       = query.NewMempoolQuery(chaintype.GetChainType(0))
	mockMempoolTransaction = &model.MempoolTransaction{
		ID:                      1,
		BlockHeight:             0,
		ArrivalTimestamp:        1000,
		FeePerByte:              10,
		TransactionBytes:        []byte{1, 2, 3, 4, 5},
		SenderAccountAddress:    "BCZ",
		RecipientAccountAddress: "ZCB",
	}
)

var _ = mockMempoolTransaction

type mockMempoolQueryExecutorFail struct {
	query.Executor
}

func (*mockMempoolQueryExecutorFail) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	// before adding mempool transactions to db we check for duplicate transactions
	case getTxByIDQuery:
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows(
			query.NewMempoolQuery(chaintype.GetChainType(0)).Fields,
		).AddRow(3, 0, 1, 1562893302, []byte{}, []byte{1}, []byte{2}))
	default:
		return nil, errors.New("MockedError")
	}

	rows, _ := db.Query(qe)
	return rows, nil
}

func (*mockMempoolQueryExecutorFail) ExecuteStatement(qe string, args ...interface{}) (sql.Result, error) {
	return nil, errors.New("MockedError")
}

func (*mockMempoolQueryExecutorFail) ExecuteTransaction(qe string, args ...interface{}) error {
	return errors.New("MockedError")
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
		Signature:               make([]byte, 68),
	}
}

func getTestSignedMempoolTransaction(id, timestamp int64) *model.MempoolTransaction {
	tx := buildTransaction(timestamp, "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE", "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN")

	txBytes, _ := transaction.GetTransactionBytes(tx, false)
	signature := (&crypto.Signature{}).Sign(txBytes, constant.SignatureTypeDefault,
		"concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved")
	tx.Signature = signature
	txBytes, _ = transaction.GetTransactionBytes(tx, true)
	return &model.MempoolTransaction{
		ID:                      id,
		BlockHeight:             0,
		FeePerByte:              1,
		ArrivalTimestamp:        timestamp,
		TransactionBytes:        txBytes,
		SenderAccountAddress:    "A",
		RecipientAccountAddress: "B",
	}
}

func TestNewMempoolService(t *testing.T) {
	type args struct {
		ct                  chaintype.ChainType
		kvExecutor          kvdb.KVExecutorInterface
		queryExecutor       query.ExecutorInterface
		mempoolQuery        query.MempoolQueryInterface
		merkleTreeQuery     query.MerkleTreeQueryInterface
		accountBalanceQuery query.AccountBalanceQueryInterface
		transactionQuery    query.TransactionQueryInterface
		blockQuery          query.BlockQueryInterface
		actionTypeSwitcher  transaction.TypeActionSwitcher
		obsr                *observer.Observer
		signature           crypto.SignatureInterface
		logger              *log.Logger
	}

	test := struct {
		name string
		args args
		want *MempoolService
	}{
		name: "NewBlockService:success",
		args: args{
			ct:   &chaintype.MainChain{},
			obsr: observer.NewObserver(),
		},
		want: &MempoolService{
			Chaintype: &chaintype.MainChain{},
			Observer:  observer.NewObserver(),
		},
	}

	got := NewMempoolService(
		test.args.ct,
		test.args.kvExecutor,
		test.args.queryExecutor,
		test.args.mempoolQuery,
		test.args.merkleTreeQuery,
		test.args.actionTypeSwitcher,
		test.args.accountBalanceQuery,
		test.args.blockQuery,
		test.args.transactionQuery,
		test.args.signature,
		test.args.obsr,
		test.args.logger,
	)
	if !reflect.DeepEqual(got, test.want) {
		t.Errorf("NewMempoolService() = %v, want %v", got, test.want)
	}
}

type (
	mockQueryExecutorGetMempoolTransactionsSuccess struct {
		query.Executor
	}
	mockQueryExecutorGetMempoolTransactionsFail struct {
		query.Executor
	}
)

func (*mockQueryExecutorGetMempoolTransactionsSuccess) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mockedRows := sqlmock.NewRows(query.NewMempoolQuery(chaintype.GetChainType(0)).Fields)
	mockedRows.AddRow(1, 0, 1, 1562893305, getTestSignedMempoolTransaction(1, 1562893305).TransactionBytes, "A", "B")
	mockedRows.AddRow(2, 0, 10, 1562893304, getTestSignedMempoolTransaction(2, 1562893304).TransactionBytes, "A", "B")
	mockedRows.AddRow(3, 0, 1, 1562893302, getTestSignedMempoolTransaction(3, 1562893302).TransactionBytes, "A", "B")
	mockedRows.AddRow(4, 0, 100, 1562893306, getTestSignedMempoolTransaction(4, 1562893306).TransactionBytes, "A", "B")
	mockedRows.AddRow(5, 0, 5, 1562893303, getTestSignedMempoolTransaction(5, 1562893303).TransactionBytes, "A", "B")
	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(mockedRows)
	rows, _ := db.Query(qe)
	return rows, nil
}

func (*mockQueryExecutorGetMempoolTransactionsFail) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("mockError:executeSelectFail")
}

func TestMempoolService_GetMempoolTransactions(t *testing.T) {
	type fields struct {
		Chaintype           chaintype.ChainType
		QueryExecutor       query.ExecutorInterface
		MempoolQuery        query.MempoolQueryInterface
		AccountBalanceQuery query.AccountBalanceQueryInterface
	}
	tests := []struct {
		name    string
		fields  fields
		want    []*model.MempoolTransaction
		wantErr bool
	}{
		{
			name: "GetMempoolTransactions:Success",
			fields: fields{
				Chaintype:           &chaintype.MainChain{},
				MempoolQuery:        query.NewMempoolQuery(&chaintype.MainChain{}),
				QueryExecutor:       &mockQueryExecutorGetMempoolTransactionsSuccess{},
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
			},
			want: []*model.MempoolTransaction{
				{
					ID:                      1,
					FeePerByte:              1,
					ArrivalTimestamp:        1562893305,
					TransactionBytes:        getTestSignedMempoolTransaction(1, 1562893305).TransactionBytes,
					SenderAccountAddress:    "A",
					RecipientAccountAddress: "B",
				},
				{
					ID:                      2,
					FeePerByte:              10,
					ArrivalTimestamp:        1562893304,
					TransactionBytes:        getTestSignedMempoolTransaction(2, 1562893304).TransactionBytes,
					SenderAccountAddress:    "A",
					RecipientAccountAddress: "B",
				},
				{
					ID:                      3,
					FeePerByte:              1,
					ArrivalTimestamp:        1562893302,
					TransactionBytes:        getTestSignedMempoolTransaction(3, 1562893302).TransactionBytes,
					SenderAccountAddress:    "A",
					RecipientAccountAddress: "B",
				},
				{
					ID:                      4,
					FeePerByte:              100,
					ArrivalTimestamp:        1562893306,
					TransactionBytes:        getTestSignedMempoolTransaction(4, 1562893306).TransactionBytes,
					SenderAccountAddress:    "A",
					RecipientAccountAddress: "B",
				},
				{
					ID:                      5,
					FeePerByte:              5,
					ArrivalTimestamp:        1562893303,
					TransactionBytes:        getTestSignedMempoolTransaction(5, 1562893303).TransactionBytes,
					SenderAccountAddress:    "A",
					RecipientAccountAddress: "B",
				},
			},
			wantErr: false,
		},
		{
			name: "GetMempoolTransactions:Fail",
			fields: fields{
				Chaintype:           &chaintype.MainChain{},
				MempoolQuery:        query.NewMempoolQuery(&chaintype.MainChain{}),
				QueryExecutor:       &mockQueryExecutorGetMempoolTransactionsFail{},
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mps := &MempoolService{
				Chaintype:           tt.fields.Chaintype,
				QueryExecutor:       tt.fields.QueryExecutor,
				MempoolQuery:        tt.fields.MempoolQuery,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
			}
			got, err := mps.GetMempoolTransactions()
			if (err != nil) != tt.wantErr {
				t.Errorf("MempoolService.GetMempoolTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MempoolService.GetMempoolTransactions() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockMempoolQueryExecutorSuccess struct {
		query.Executor
	}
)

func (*mockMempoolQueryExecutorSuccess) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT account_address,block_height,spendable_balance,balance,pop_revenue,latest FROM account_balance " +
		"WHERE account_address = ? AND latest = 1":
		mockedRows := sqlmock.NewRows([]string{"account_address", "block_height", "spendable_balance", "balance", "pop_revenue", "latest"})
		mockedRows.AddRow("BCZ", 1, 1000, 10000, nil, 1)
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(mockedRows)
	default:
		mockedRows := sqlmock.NewRows(mockMempoolQuery.Fields)
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(mockedRows)
	}

	rows, _ := db.Query(qe)
	return rows, nil
}
func (*mockMempoolQueryExecutorSuccess) ExecuteStatement(qe string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}
func (*mockMempoolQueryExecutorSuccess) ExecuteTransaction(qe string, args ...interface{}) error {
	return nil
}
func (*mockMempoolQueryExecutorSuccess) ExecuteSelectRow(qe string, tx bool, args ...interface{}) (*sql.Row, error) {
	// While getting last block
	db, mock, _ := sqlmock.New()
	mockedRow := sqlmock.NewRows(query.NewBlockQuery(chaintype.GetChainType(0)).Fields)
	mockedRow.AddRow(
		mockBlockData.GetID(),
		mockBlockData.GetBlockHash(),
		mockBlockData.GetPreviousBlockHash(),
		mockBlockData.GetHeight(),
		mockBlockData.GetTimestamp(),
		mockBlockData.GetBlockSeed(),
		mockBlockData.GetBlockSignature(),
		mockBlockData.GetCumulativeDifficulty(),
		mockBlockData.GetSmithScale(),
		mockBlockData.GetPayloadLength(),
		mockBlockData.GetPayloadHash(),
		mockBlockData.GetBlocksmithPublicKey(),
		mockBlockData.GetTotalAmount(),
		mockBlockData.GetTotalFee(),
		mockBlockData.GetTotalCoinBase(),
		mockBlockData.GetVersion(),
	)
	mock.ExpectQuery("").WillReturnRows(mockedRow)
	return db.QueryRow(""), nil
}
func (*mockMempoolQueryExecutorSuccess) BeginTx() error {
	return nil
}
func (*mockMempoolQueryExecutorSuccess) CommitTx() error {
	return nil
}

func TestMempoolService_AddMempoolTransaction(t *testing.T) {
	type fields struct {
		Chaintype          chaintype.ChainType
		QueryExecutor      query.ExecutorInterface
		MempoolQuery       query.MempoolQueryInterface
		BlockQuery         query.BlockQueryInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
		Observer           *observer.Observer
	}
	type args struct {
		mpTx *model.MempoolTransaction
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "AddMempoolTransaction:Success",
			fields: fields{
				Chaintype:          &chaintype.MainChain{},
				MempoolQuery:       query.NewMempoolQuery(&chaintype.MainChain{}),
				BlockQuery:         query.NewBlockQuery(chaintype.GetChainType(0)),
				QueryExecutor:      &mockMempoolQueryExecutorSuccess{},
				ActionTypeSwitcher: &transaction.TypeSwitcher{},
				Observer:           observer.NewObserver(),
			},
			args: args{
				mpTx: getTestSignedMempoolTransaction(3, 1562893302),
			},
			wantErr: false,
		},
		{
			name: "AddMempoolTransaction:DuplicateTransaction",
			fields: fields{
				Chaintype:          &chaintype.MainChain{},
				MempoolQuery:       query.NewMempoolQuery(&chaintype.MainChain{}),
				QueryExecutor:      &mockMempoolQueryExecutorFail{},
				ActionTypeSwitcher: &transaction.TypeSwitcher{},
				Observer:           observer.NewObserver(),
			},
			args: args{
				mpTx: getTestSignedMempoolTransaction(3, 1562893303),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mps := &MempoolService{
				Chaintype:          tt.fields.Chaintype,
				QueryExecutor:      tt.fields.QueryExecutor,
				MempoolQuery:       tt.fields.MempoolQuery,
				BlockQuery:         tt.fields.BlockQuery,
				ActionTypeSwitcher: tt.fields.ActionTypeSwitcher,
				Observer:           tt.fields.Observer,
			}
			if err := mps.AddMempoolTransaction(tt.args.mpTx); (err != nil) != tt.wantErr {
				t.Errorf("MempoolService.AddMempoolTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockQueryExecutorSelectTransactionsFromMempoolSuccess struct {
		query.Executor
	}
)

var mockSuccessSelectMempool = []*model.MempoolTransaction{
	{
		ID:                      1,
		FeePerByte:              1,
		ArrivalTimestamp:        1562893305,
		TransactionBytes:        getTestSignedMempoolTransaction(1, 1562893305).TransactionBytes,
		SenderAccountAddress:    "A",
		RecipientAccountAddress: "B",
	},
	{
		ID:                      2,
		FeePerByte:              10,
		ArrivalTimestamp:        1562893304,
		TransactionBytes:        getTestSignedMempoolTransaction(2, 1562893304).TransactionBytes,
		SenderAccountAddress:    "A",
		RecipientAccountAddress: "B",
	},
	{
		ID:                      3,
		FeePerByte:              1,
		ArrivalTimestamp:        1562893302,
		TransactionBytes:        getTestSignedMempoolTransaction(3, 1562893302).TransactionBytes,
		SenderAccountAddress:    "A",
		RecipientAccountAddress: "B",
	},
	{
		ID:                      4,
		FeePerByte:              100,
		ArrivalTimestamp:        1562893306,
		TransactionBytes:        getTestSignedMempoolTransaction(4, 1562893306).TransactionBytes,
		SenderAccountAddress:    "A",
		RecipientAccountAddress: "B",
	},
	{
		ID:                      5,
		FeePerByte:              5,
		ArrivalTimestamp:        1562893303,
		TransactionBytes:        getTestSignedMempoolTransaction(5, 1562893303).TransactionBytes,
		SenderAccountAddress:    "A",
		RecipientAccountAddress: "B",
	},
}

func (*mockQueryExecutorSelectTransactionsFromMempoolSuccess) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT account_address,block_height,spendable_balance,balance,pop_revenue,latest FROM account_balance " +
		"WHERE account_address = ? AND latest = 1":
		abRows := sqlmock.NewRows(query.NewAccountBalanceQuery().Fields)
		abRows.AddRow([]byte{1}, 1, 10000, 10000, 0, 1)
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(abRows)
	default:
		mtxRows := sqlmock.NewRows(query.NewMempoolQuery(chaintype.GetChainType(0)).Fields)
		mtxRows.AddRow(1, 0, 1, 1562893305, getTestSignedMempoolTransaction(1, 1562893305).TransactionBytes, "A", "B")
		mtxRows.AddRow(2, 0, 10, 1562893304, getTestSignedMempoolTransaction(2, 1562893304).TransactionBytes, "A", "B")
		mtxRows.AddRow(3, 0, 1, 1562893302, getTestSignedMempoolTransaction(3, 1562893302).TransactionBytes, "A", "B")
		mtxRows.AddRow(4, 0, 100, 1562893306, getTestSignedMempoolTransaction(4, 1562893306).TransactionBytes, "A", "B")
		mtxRows.AddRow(5, 0, 5, 1562893303, getTestSignedMempoolTransaction(5, 1562893303).TransactionBytes, "A", "B")
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(mtxRows)
	}
	rows, _ := db.Query(qe)
	return rows, nil
}
func TestMempoolService_SelectTransactionsFromMempool(t *testing.T) {
	successTx1, _ := transaction.ParseTransactionBytes(mockSuccessSelectMempool[0].TransactionBytes, true)
	successTx2, _ := transaction.ParseTransactionBytes(mockSuccessSelectMempool[1].TransactionBytes, true)
	successTx3, _ := transaction.ParseTransactionBytes(mockSuccessSelectMempool[2].TransactionBytes, true)
	successTx4, _ := transaction.ParseTransactionBytes(mockSuccessSelectMempool[3].TransactionBytes, true)
	successTx5, _ := transaction.ParseTransactionBytes(mockSuccessSelectMempool[4].TransactionBytes, true)
	type fields struct {
		Chaintype           chaintype.ChainType
		QueryExecutor       query.ExecutorInterface
		MempoolQuery        query.MempoolQueryInterface
		ActionTypeSwitcher  transaction.TypeActionSwitcher
		AccountBalanceQuery query.AccountBalanceQueryInterface
	}
	type args struct {
		blockTimestamp int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Transaction
		wantErr bool
	}{
		{
			name: "SelectTransactionsFromMempool:Success",
			fields: fields{
				Chaintype:           &chaintype.MainChain{},
				MempoolQuery:        query.NewMempoolQuery(&chaintype.MainChain{}),
				QueryExecutor:       &mockQueryExecutorSelectTransactionsFromMempoolSuccess{},
				ActionTypeSwitcher:  &transaction.TypeSwitcher{},
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
			},
			args: args{
				blockTimestamp: 1562893106,
			},
			want: []*model.Transaction{
				successTx2, successTx1, successTx4, successTx3, successTx5,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mps := &MempoolService{
				Chaintype:           tt.fields.Chaintype,
				QueryExecutor:       tt.fields.QueryExecutor,
				MempoolQuery:        tt.fields.MempoolQuery,
				ActionTypeSwitcher:  tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
			}
			got, err := mps.SelectTransactionsFromMempool(tt.args.blockTimestamp)
			if (err != nil) != tt.wantErr {
				t.Errorf("MempoolService.SelectTransactionsFromMempool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MempoolService.SelectTransactionsFromMempool() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	ReceivedTransactionListenerMockTypeAction struct {
		transaction.SendMoney
	}
	ReceivedTransactionListenerMockTypeActionSuccess struct {
		ReceivedTransactionListenerMockTypeAction
	}
)

// mockTypeAction
func (*ReceivedTransactionListenerMockTypeAction) ApplyConfirmed() error {
	return nil
}
func (*ReceivedTransactionListenerMockTypeAction) Validate(bool) error {
	return nil
}
func (*ReceivedTransactionListenerMockTypeAction) GetAmount() int64 {
	return 10
}

func (*ReceivedTransactionListenerMockTypeAction) ApplyUnconfirmed() error {
	return nil
}

func (*ReceivedTransactionListenerMockTypeActionSuccess) GetTransactionType(tx *model.Transaction) (transaction.TypeAction, error) {
	return &ReceivedTransactionListenerMockTypeAction{}, nil
}

type (
	mockExecutorValidateMempoolTransactionSuccess struct {
		query.Executor
	}
	mockExecutorValidateMempoolTransactionSuccessNoRow struct {
		query.Executor
	}
	mockExecutorValidateMempoolTransactionFail struct {
		query.Executor
	}
)

func (*mockExecutorValidateMempoolTransactionSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
		sqlmock.NewRows(query.NewTransactionQuery(&chaintype.MainChain{}).Fields).AddRow(
			1,
			2774809487,
			1,
			1,
			"BCZ-Sender",
			"BCZ-Recipient",
			0,
			1,
			23445959,
			make([]byte, 32),
			0,
			make([]byte, 0),
			nil,
			make([]byte, 64),
		),
	)
	return db.QueryRow(qStr), nil
}

func (*mockExecutorValidateMempoolTransactionSuccessNoRow) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
		sqlmock.NewRows(query.NewTransactionQuery(&chaintype.MainChain{}).Fields),
	)
	return db.QueryRow(qStr), nil
}
func (*mockExecutorValidateMempoolTransactionSuccessNoRow) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	balanceQ := query.NewAccountBalanceQuery()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(
		sqlmock.NewRows(balanceQ.Fields).AddRow(
			"BCZ",
			100,
			1000,
			10000,
			100,
			1,
		),
	)
	return db.Query(qStr)
}

func (*mockExecutorValidateMempoolTransactionFail) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("").WillReturnError(errors.New("mocked err"))
	return db.QueryRow(qStr), nil
}

func (*mockExecutorValidateMempoolTransactionFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("mockExecutorValidateMempoolTransactionFail : mocked Err")
}

func TestMempoolService_ValidateMempoolTransaction(t *testing.T) {
	type fields struct {
		Chaintype           chaintype.ChainType
		QueryExecutor       query.ExecutorInterface
		MempoolQuery        query.MempoolQueryInterface
		ActionTypeSwitcher  transaction.TypeActionSwitcher
		AccountBalanceQuery query.AccountBalanceQueryInterface
		TransactionQuery    query.TransactionQueryInterface
		Observer            *observer.Observer
	}
	type args struct {
		mpTx *model.MempoolTransaction
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				Chaintype:           &chaintype.MainChain{},
				QueryExecutor:       &mockExecutorValidateMempoolTransactionSuccessNoRow{},
				ActionTypeSwitcher:  &transaction.TypeSwitcher{},
				MempoolQuery:        query.NewMempoolQuery(&chaintype.MainChain{}),
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				TransactionQuery:    query.NewTransactionQuery(&chaintype.MainChain{}),
			},
			args: args{
				mpTx: getTestSignedMempoolTransaction(3, 1562893302),
			},
			wantErr: false,
		},
		{
			name: "wantErr:TransactionExisted",
			fields: fields{
				Chaintype:          &chaintype.MainChain{},
				QueryExecutor:      &mockExecutorValidateMempoolTransactionSuccess{},
				MempoolQuery:       query.NewMempoolQuery(&chaintype.MainChain{}),
				ActionTypeSwitcher: &transaction.TypeSwitcher{},
				TransactionQuery:   query.NewTransactionQuery(&chaintype.MainChain{}),
			},
			args: args{
				mpTx: getTestSignedMempoolTransaction(3, 1562893302),
			},
			wantErr: true,
		},
		{
			name: "wantErr:TransactionExisted",
			fields: fields{
				Chaintype:          &chaintype.MainChain{},
				QueryExecutor:      &mockExecutorValidateMempoolTransactionFail{},
				TransactionQuery:   query.NewTransactionQuery(&chaintype.MainChain{}),
				MempoolQuery:       query.NewMempoolQuery(&chaintype.MainChain{}),
				ActionTypeSwitcher: &transaction.TypeSwitcher{},
			},
			args: args{
				mpTx: getTestSignedMempoolTransaction(3, 1562893302),
			},
			wantErr: true,
		},
		{
			name: "wantErr:ParseFail",
			fields: fields{
				Chaintype:          &chaintype.MainChain{},
				QueryExecutor:      &mockExecutorValidateMempoolTransactionSuccessNoRow{},
				TransactionQuery:   query.NewTransactionQuery(&chaintype.MainChain{}),
				MempoolQuery:       query.NewMempoolQuery(&chaintype.MainChain{}),
				ActionTypeSwitcher: &transaction.TypeSwitcher{},
			},
			args: args{
				mpTx: &model.MempoolTransaction{
					ID: 12,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mps := &MempoolService{
				Chaintype:           tt.fields.Chaintype,
				QueryExecutor:       tt.fields.QueryExecutor,
				MempoolQuery:        tt.fields.MempoolQuery,
				ActionTypeSwitcher:  tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				TransactionQuery:    tt.fields.TransactionQuery,
				Observer:            tt.fields.Observer,
			}
			if err := mps.ValidateMempoolTransaction(tt.args.mpTx); (err != nil) != tt.wantErr {
				t.Errorf("MempoolService.ValidateMempoolTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockQueryExecutorDeleteExpiredMempoolTransactions struct {
		query.Executor
	}
	mockQueryExecutorDeleteExpiredMempoolTransactionsEmpty struct {
		query.Executor
	}
)

func (*mockQueryExecutorDeleteExpiredMempoolTransactionsEmpty) BeginTx() error {
	return nil
}
func (*mockQueryExecutorDeleteExpiredMempoolTransactionsEmpty) CommitTx() error {
	return nil
}
func (*mockQueryExecutorDeleteExpiredMempoolTransactionsEmpty) ExecuteTransaction(query string, args ...interface{}) error {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectPrepare("").
		ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
	_, _ = db.Exec("")
	return nil
}
func (*mockQueryExecutorDeleteExpiredMempoolTransactionsEmpty) ExecuteSelect(
	query string,
	tx bool,
	args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("").WillReturnRows(
		sqlmock.NewRows(mockMempoolQuery.Fields),
	)
	return db.Query("")
}

// Not Empty mempool
func (*mockQueryExecutorDeleteExpiredMempoolTransactions) BeginTx() error {
	return nil
}
func (*mockQueryExecutorDeleteExpiredMempoolTransactions) CommitTx() error {
	return nil
}
func (*mockQueryExecutorDeleteExpiredMempoolTransactions) ExecuteTransaction(
	query string,
	args ...interface{},
) error {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectPrepare("").
		ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
	_, _ = db.Exec("")
	return nil
}
func (*mockQueryExecutorDeleteExpiredMempoolTransactions) ExecuteSelect(
	query string,
	tx bool,
	args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("").WillReturnRows(
		sqlmock.NewRows(mockMempoolQuery.Fields).AddRow(
			1,
			0,
			1000,
			10,
			[]byte{1, 0, 0, 0, 1, 210, 225, 121, 93, 0, 0, 0, 0, 44, 0, 0, 0, 110, 75, 95, 111, 117, 120, 100, 68,
				68, 119, 117, 74, 105, 111, 103, 105, 68, 65, 105, 95, 122, 115, 49, 76, 113, 101, 78, 55, 102,
				53, 90, 115, 88, 98, 70, 116, 88, 71, 113, 71, 99, 48, 80, 100, 44, 0, 0, 0, 118, 66, 75, 98,
				114, 82, 53, 89, 57, 83, 71, 68, 74, 51, 50, 49, 76, 119, 53, 53, 50, 119, 53, 106, 85, 50, 76,
				109, 79, 81, 67, 68, 120, 81, 114, 114, 118, 74, 48, 67, 85, 107, 101, 70, 160, 134, 1, 0, 0,
				0, 0, 0, 8, 0, 0, 0, 0, 225, 245, 5, 0, 0, 0, 0, 0, 0, 0, 0, 109, 6, 82, 80, 77, 171, 32, 88,
				211, 199, 11, 114, 75, 229, 243, 98, 167, 159, 225, 11, 40, 125, 221, 252, 44, 131, 136, 13,
				104, 109, 228, 40, 27, 177, 175, 128, 223, 154, 19, 71, 18, 134, 177, 77, 96, 157, 187, 91,
				152, 160, 78, 140, 96, 81, 116, 38, 164, 105, 149, 50, 138, 236, 209, 11},
			"BCZ",
			"ZCB",
		),
	)
	return db.Query("")
}

func TestMempoolService_DeleteExpiredMempoolTransactions(t *testing.T) {
	type fields struct {
		Chaintype           chaintype.ChainType
		QueryExecutor       query.ExecutorInterface
		MempoolQuery        query.MempoolQueryInterface
		ActionTypeSwitcher  transaction.TypeActionSwitcher
		AccountBalanceQuery query.AccountBalanceQueryInterface
		Signature           crypto.SignatureInterface
		TransactionQuery    query.TransactionQueryInterface
		Observer            *observer.Observer
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "wantSuccess:EmptyMempool",
			fields: fields{
				QueryExecutor: &mockQueryExecutorDeleteExpiredMempoolTransactionsEmpty{},
				MempoolQuery:  mockMempoolQuery,
			},
			wantErr: false,
		},
		{
			name: "wantSuccess:PruneExpiredMempool",
			fields: fields{
				QueryExecutor: &mockQueryExecutorDeleteExpiredMempoolTransactions{},
				MempoolQuery:  mockMempoolQuery,
				ActionTypeSwitcher: &transaction.TypeSwitcher{
					Executor: &mockQueryExecutorDeleteExpiredMempoolTransactions{},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mps := &MempoolService{
				Chaintype:           tt.fields.Chaintype,
				QueryExecutor:       tt.fields.QueryExecutor,
				MempoolQuery:        tt.fields.MempoolQuery,
				ActionTypeSwitcher:  tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				Signature:           tt.fields.Signature,
				TransactionQuery:    tt.fields.TransactionQuery,
				Observer:            tt.fields.Observer,
			}
			if err := mps.DeleteExpiredMempoolTransactions(); (err != nil) != tt.wantErr {
				t.Errorf("DeleteExpiredMempoolTransactions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockGetMempoolTransactionsByBlockHeightExecutor struct {
		query.Executor
	}
)

func (*mockGetMempoolTransactionsByBlockHeightExecutor) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("").WillReturnRows(
		sqlmock.NewRows(query.NewMempoolQuery(chaintype.GetChainType(0)).Fields).AddRow(
			1,
			0,
			10,
			1000,
			[]byte{1, 2, 3, 4, 5},
			"BCZ",
			"ZCB",
		),
	)
	return db.Query("")
}

func TestMempoolService_GetMempoolTransactionsByBlockHeight(t *testing.T) {
	type fields struct {
		Chaintype           chaintype.ChainType
		KVExecutor          kvdb.KVExecutorInterface
		QueryExecutor       query.ExecutorInterface
		MempoolQuery        query.MempoolQueryInterface
		MerkleTreeQuery     query.MerkleTreeQueryInterface
		ActionTypeSwitcher  transaction.TypeActionSwitcher
		AccountBalanceQuery query.AccountBalanceQueryInterface
		Signature           crypto.SignatureInterface
		TransactionQuery    query.TransactionQueryInterface
		Observer            *observer.Observer
		Logger              *log.Logger
	}
	type args struct {
		height uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.MempoolTransaction
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				QueryExecutor: &mockGetMempoolTransactionsByBlockHeightExecutor{},
				MempoolQuery:  query.NewMempoolQuery(chaintype.GetChainType(0)),
			},
			args: args{height: 0},
			want: []*model.MempoolTransaction{mockMempoolTransaction},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mps := &MempoolService{
				Chaintype:           tt.fields.Chaintype,
				KVExecutor:          tt.fields.KVExecutor,
				QueryExecutor:       tt.fields.QueryExecutor,
				MempoolQuery:        tt.fields.MempoolQuery,
				MerkleTreeQuery:     tt.fields.MerkleTreeQuery,
				ActionTypeSwitcher:  tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				Signature:           tt.fields.Signature,
				TransactionQuery:    tt.fields.TransactionQuery,
				Observer:            tt.fields.Observer,
				Logger:              tt.fields.Logger,
			}
			got, err := mps.GetMempoolTransactionsWantToBackup(tt.args.height)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMempoolTransactionsWantToBackup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMempoolTransactionsWantToBackup() got = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}
