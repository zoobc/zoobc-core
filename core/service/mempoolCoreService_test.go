package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dgraph-io/badger"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
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
	sender := "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE"
	recipient := "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN"
	tx := buildTransaction(timestamp, sender, recipient)

	txBytes, _ := (&transaction.Util{}).GetTransactionBytes(tx, false)
	signature := (&crypto.Signature{}).Sign(txBytes, constant.SignatureTypeDefault,
		"concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved")
	tx.Signature = signature
	txBytes, _ = (&transaction.Util{}).GetTransactionBytes(tx, true)
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
		transactionUtil     transaction.UtilInterface
		receiptUtil         coreUtil.ReceiptUtilInterface
	}

	test := struct {
		name string
		args args
		want *MempoolService
	}{
		name: "NewMempoolService:success",
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
		test.args.transactionUtil,
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
		test.args.receiptUtil,
	)

	mempoolGetter := &MempoolGetter{
		MempoolQuery:  test.args.mempoolQuery,
		QueryExecutor: test.args.queryExecutor,
	}
	test.want.MempoolServiceUtil = NewMempoolServiceUtil(test.args.transactionUtil,
		test.args.transactionQuery,
		test.args.queryExecutor,
		test.args.mempoolQuery,
		test.args.actionTypeSwitcher,
		test.args.accountBalanceQuery,
		test.args.blockQuery,
		mempoolGetter,
	)
	test.want.MempoolGetter = mempoolGetter

	if !reflect.DeepEqual(got, test.want) {
		jGot, _ := json.MarshalIndent(got, "", "  ")
		jWant, _ := json.MarshalIndent(test.want, "", "  ")
		t.Errorf("NewMempoolService() = %s, want %s", jGot, jWant)
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
	successTx1, _ := (&transaction.Util{}).ParseTransactionBytes(mockSuccessSelectMempool[0].TransactionBytes, true)
	successTx2, _ := (&transaction.Util{}).ParseTransactionBytes(mockSuccessSelectMempool[1].TransactionBytes, true)
	successTx3, _ := (&transaction.Util{}).ParseTransactionBytes(mockSuccessSelectMempool[2].TransactionBytes, true)
	successTx4, _ := (&transaction.Util{}).ParseTransactionBytes(mockSuccessSelectMempool[3].TransactionBytes, true)
	successTx5, _ := (&transaction.Util{}).ParseTransactionBytes(mockSuccessSelectMempool[4].TransactionBytes, true)
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
				TransactionUtil:     &transaction.Util{},
				Chaintype:           tt.fields.Chaintype,
				QueryExecutor:       tt.fields.QueryExecutor,
				MempoolQuery:        tt.fields.MempoolQuery,
				ActionTypeSwitcher:  tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				MempoolGetter: &MempoolGetter{
					QueryExecutor: tt.fields.QueryExecutor,
					MempoolQuery:  tt.fields.MempoolQuery,
				},
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

// type (
// 	ReceivedTransactionListenerMockTypeAction struct {
// 		transaction.SendMoney
// 	}
// 	ReceivedTransactionListenerMockTypeActionSuccess struct {
// 		ReceivedTransactionListenerMockTypeAction
// 	}
// )

// // mockTypeAction
// func (*ReceivedTransactionListenerMockTypeAction) ApplyConfirmed() error {
// 	return nil
// }
// func (*ReceivedTransactionListenerMockTypeAction) Validate(bool) error {
// 	return nil
// }
// func (*ReceivedTransactionListenerMockTypeAction) GetAmount() int64 {
// 	return 10
// }

// func (*ReceivedTransactionListenerMockTypeAction) ApplyUnconfirmed() error {
// 	return nil
// }

// func (*ReceivedTransactionListenerMockTypeActionSuccess) GetTransactionType(tx *model.Transaction) (transaction.TypeAction, error) {
// 	return &ReceivedTransactionListenerMockTypeAction{}, nil
// }

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
				TransactionUtil:     &transaction.Util{},
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

func TestMempoolService_ProcessReceivedTransaction(t *testing.T) {
	chainTypeUsed := &chaintype.MainChain{}
	type conditions struct {
		parseTransactionResult                 *model.Transaction
		parseTransactionError                  error
		getReceiptKeyResult                    []byte
		getReceiptKeyError                     error
		validateMempoolError                   error
		kvdbGetResult                          []byte
		kvdbGetError                           error
		generateBatchReceiptWithReminderResult *model.BatchReceipt
		generateBatchReceiptWithReminderError  error
	}
	type args struct {
		mempoolTx *model.MempoolTransaction
	}
	type want struct {
		batchReceipt *model.BatchReceipt
		transaction  *model.Transaction
		err          bool
	}
	tests := []struct {
		name       string
		conditions conditions
		args       args
		want       want
	}{
		{
			name: "Fail:ParseTransaction_error",
			conditions: conditions{
				parseTransactionError: errors.New(""),
			},
			args: args{mempoolTx: &model.MempoolTransaction{}},
			want: want{
				err: true,
			},
		},
		{
			name: "Fail:GetReceiptKey_error",
			conditions: conditions{
				parseTransactionResult: &model.Transaction{},
				getReceiptKeyError:     errors.New(""),
			},
			args: args{mempoolTx: &model.MempoolTransaction{}},
			want: want{
				err: true,
			},
		},
		{
			name: "Fail:ValidateMempoolTransaction_error_non_duplicate",
			conditions: conditions{
				parseTransactionResult: &model.Transaction{},
				getReceiptKeyError:     errors.New(""),
			},
			args: args{mempoolTx: &model.MempoolTransaction{}},
			want: want{
				err: true,
			},
		},
		{
			name: "Fail:ValidateMempoolTransaction_error_duplicate_and_kv_executor_get_error_non_err_key_not_found",
			conditions: conditions{
				parseTransactionResult: &model.Transaction{},
				validateMempoolError:   blocker.NewBlocker(blocker.DuplicateMempoolErr, ""),
				kvdbGetError:           errors.New(""),
			},
			args: args{mempoolTx: &model.MempoolTransaction{}},
			want: want{
				err: true,
			},
		},
		{
			name: "Fail:ValidateMempoolTransaction_error_duplicate_and_kv_executor_found_the_record_the_sender_has_received_receipt_for_this_data",
			conditions: conditions{
				parseTransactionResult: &model.Transaction{},
				kvdbGetError:           nil,
				validateMempoolError:   blocker.NewBlocker(blocker.DuplicateMempoolErr, ""),
				kvdbGetResult:          []byte{1},
			},
			args: args{mempoolTx: &model.MempoolTransaction{}},
			want: want{
				err: true,
			},
		},
		{
			name: "Fail:GenerateBatchReceiptWithReminder_error",
			conditions: conditions{
				parseTransactionResult:                &model.Transaction{},
				generateBatchReceiptWithReminderError: errors.New(""),
			},
			args: args{mempoolTx: &model.MempoolTransaction{}},
			want: want{
				err: true,
			},
		},
		{
			name: "Success:expected_returns_and_no_errors",
			conditions: conditions{
				parseTransactionResult:                 &model.Transaction{},
				parseTransactionError:                  nil,
				getReceiptKeyError:                     nil,
				validateMempoolError:                   nil,
				generateBatchReceiptWithReminderResult: &model.BatchReceipt{},
				generateBatchReceiptWithReminderError:  nil,
			},
			args: args{mempoolTx: &model.MempoolTransaction{}},
			want: want{
				batchReceipt: &model.BatchReceipt{},
				transaction:  &model.Transaction{},
			},
		},
		{
			name: "Success:duplicate_mempool_and_sender_has_not_got_received_for_this_data",
			conditions: conditions{
				parseTransactionResult:                 &model.Transaction{},
				parseTransactionError:                  nil,
				getReceiptKeyError:                     nil,
				kvdbGetResult:                          nil,
				kvdbGetError:                           badger.ErrKeyNotFound,
				validateMempoolError:                   blocker.NewBlocker(blocker.DuplicateMempoolErr, ""),
				generateBatchReceiptWithReminderResult: &model.BatchReceipt{},
				generateBatchReceiptWithReminderError:  nil,
			},
			args: args{mempoolTx: &model.MempoolTransaction{}},
			want: want{
				batchReceipt: &model.BatchReceipt{},
				transaction:  &model.Transaction{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mps := &MempoolService{
				Chaintype:        chainTypeUsed,
				TransactionQuery: query.NewTransactionQuery(chainTypeUsed),
				MerkleTreeQuery:  query.NewMerkleTreeQuery(),
				MempoolQuery:     query.NewMempoolQuery(chainTypeUsed),
				TransactionUtil: &MockTransactionUtil{
					ParseTransactionError:  tt.conditions.parseTransactionError,
					ParseTransactionResult: &model.Transaction{},
				},
				KVExecutor: &MockKVExecutor{
					KvdbGetResult: tt.conditions.kvdbGetResult,
					KvdbGetError:  tt.conditions.kvdbGetError,
				},
				MempoolServiceUtil: &MockMempoolServiceUtil{
					ValidatidateMempoolError: tt.conditions.validateMempoolError,
				},
				ReceiptUtil: &MockReceiptUtil{
					GenerateBatchReceiptWithReminderResult: tt.conditions.generateBatchReceiptWithReminderResult,
					GenerateBatchReceiptWithReminderError:  tt.conditions.generateBatchReceiptWithReminderError,
					GetReceiptKeyResult:                    tt.conditions.getReceiptKeyResult,
					GetReceiptKeyError:                     tt.conditions.getReceiptKeyError,
				},
			}
			batchReceipt, tx, err := mps.ProcessReceivedTransaction([]byte{}, []byte{}, &model.Block{}, "")
			if (err != nil) != tt.want.err {
				t.Errorf("ProcessReceivedTransaction() error = %v, wantErr %v", err, tt.want.err)
				return
			}
			if !reflect.DeepEqual(batchReceipt, tt.want.batchReceipt) {
				t.Errorf("ProcessReceivedTransaction() batchReceipt = \n%v, want \n%v", batchReceipt, tt.want.batchReceipt)
			}
			if !reflect.DeepEqual(tx, tt.want.transaction) {
				t.Errorf("ProcessReceivedTransaction() transaction = \n%v, want \n%v", tx, tt.want.transaction)
			}
		})
	}
}

func TestMempoolService_ReceivedTransaction(t *testing.T) {
	chainTypeUsed := &chaintype.MainChain{}
	type conditions struct {
		validateMempoolError                   error
		beginTxError                           error
		parseTransactionResult                 *model.Transaction
		parseTransactionError                  error
		actionSwitcherGetTransactionTypeError  error
		typeActionApplyUnconfirmedError        error
		addMempoolTransactionError             error
		getReceiptKeyResult                    []byte
		getReceiptKeyError                     error
		generateBatchReceiptWithReminderResult *model.BatchReceipt
		generateBatchReceiptWithReminderError  error
		commitTxError                          error
	}
	type (
		args struct {
			mempoolTx *model.MempoolTransaction
		}
		want struct {
			batchReceipt *model.BatchReceipt
			err          bool
		}
	)

	batchReceipt := &model.BatchReceipt{}

	tests := []struct {
		name       string
		conditions conditions
		args       args
		want       want
	}{
		{
			name: "wantSuccess",
			conditions: conditions{
				validateMempoolError:                   nil,
				beginTxError:                           nil,
				parseTransactionResult:                 &model.Transaction{},
				parseTransactionError:                  nil,
				typeActionApplyUnconfirmedError:        nil,
				actionSwitcherGetTransactionTypeError:  nil,
				addMempoolTransactionError:             nil,
				commitTxError:                          nil,
				generateBatchReceiptWithReminderResult: batchReceipt,
				generateBatchReceiptWithReminderError:  nil,
			},
			args: args{mempoolTx: &model.MempoolTransaction{}},
			want: want{
				batchReceipt: batchReceipt,
				err:          false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mps := &MempoolService{
				Chaintype:        chainTypeUsed,
				TransactionQuery: query.NewTransactionQuery(chainTypeUsed),
				MerkleTreeQuery:  query.NewMerkleTreeQuery(),
				MempoolQuery:     query.NewMempoolQuery(chainTypeUsed),
				TransactionUtil: &MockTransactionUtil{
					ParseTransactionResult: tt.conditions.parseTransactionResult,
					ParseTransactionError:  tt.conditions.parseTransactionError,
				},
				KVExecutor: nil,
				MempoolServiceUtil: &MockMempoolServiceUtil{
					ValidatidateMempoolError:   tt.conditions.validateMempoolError,
					AddMempoolTransactionError: tt.conditions.addMempoolTransactionError,
				},
				QueryExecutor: &MockQueryExecutor{
					BeginTxError:  tt.conditions.beginTxError,
					CommitTxError: tt.conditions.commitTxError,
				},
				ActionTypeSwitcher: &MockActionTypeSwitcher{
					GetTransactionTypeResult: &MockTypeAction{
						ApplyUnconfirmedError: tt.conditions.typeActionApplyUnconfirmedError,
					},
					GetTransactionTypeError: tt.conditions.actionSwitcherGetTransactionTypeError,
				},
				ReceiptUtil: &MockReceiptUtil{
					GenerateBatchReceiptWithReminderResult: tt.conditions.generateBatchReceiptWithReminderResult,
					GenerateBatchReceiptWithReminderError:  tt.conditions.generateBatchReceiptWithReminderError,
					GetReceiptKeyResult:                    tt.conditions.getReceiptKeyResult,
					GetReceiptKeyError:                     tt.conditions.getReceiptKeyError,
				},
				Observer: &observer.Observer{},
			}
			got, err := mps.ReceivedTransaction([]byte{}, []byte{}, &model.Block{}, "")
			if (err != nil) != tt.want.err {
				t.Errorf("ReceivedTransaction() error = %v, wantErr %v", err, tt.want.err)
				return
			}
			if !reflect.DeepEqual(got, tt.want.batchReceipt) {
				t.Errorf("ReceivedTransaction() batchReceipt = \n%v, want \n%v", batchReceipt, tt.want.batchReceipt)
			}
		})
	}
}
