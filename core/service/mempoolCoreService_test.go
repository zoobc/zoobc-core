package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dgraph-io/badger/v2"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
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

func TestNewMempoolService(t *testing.T) {
	type args struct {
		ct                     chaintype.ChainType
		kvExecutor             kvdb.KVExecutorInterface
		queryExecutor          query.ExecutorInterface
		mempoolQuery           query.MempoolQueryInterface
		merkleTreeQuery        query.MerkleTreeQueryInterface
		accountBalanceQuery    query.AccountBalanceQueryInterface
		transactionQuery       query.TransactionQueryInterface
		blockQuery             query.BlockQueryInterface
		actionTypeSwitcher     transaction.TypeActionSwitcher
		obsr                   *observer.Observer
		signature              crypto.SignatureInterface
		logger                 *log.Logger
		transactionUtil        transaction.UtilInterface
		receiptUtil            coreUtil.ReceiptUtilInterface
		receiptService         ReceiptServiceInterface
		TransactionCoreService TransactionCoreServiceInterface
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
		test.args.receiptService,
		test.args.TransactionCoreService,
	)

	mempoolGetter := &MempoolGetter{
		MempoolQuery:  test.args.mempoolQuery,
		QueryExecutor: test.args.queryExecutor,
	}
	test.want.MempoolServiceUtil = NewMempoolServiceUtil(
		test.args.transactionUtil,
		test.args.transactionQuery,
		test.args.queryExecutor,
		test.args.mempoolQuery,
		test.args.actionTypeSwitcher,
		test.args.accountBalanceQuery,
		test.args.blockQuery,
		mempoolGetter,
		test.args.TransactionCoreService,
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
		ID:               1,
		FeePerByte:       1,
		ArrivalTimestamp: 1562893305,
		TransactionBytes: transaction.GetFixturesForSignedMempoolTransaction(
			1,
			1562893305,
			"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
			"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			false,
		).TransactionBytes,
		SenderAccountAddress:    "A",
		RecipientAccountAddress: "B",
	},
	{
		ID:               2,
		FeePerByte:       10,
		ArrivalTimestamp: 1562893304,
		TransactionBytes: transaction.GetFixturesForSignedMempoolTransaction(
			2,
			1562893304,
			"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
			"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			false,
		).TransactionBytes,
		SenderAccountAddress:    "A",
		RecipientAccountAddress: "B",
	},
	{
		ID:               3,
		FeePerByte:       1,
		ArrivalTimestamp: 1562893302,
		TransactionBytes: transaction.GetFixturesForSignedMempoolTransaction(
			3,
			1562893302,
			"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
			"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			false,
		).TransactionBytes,
		SenderAccountAddress:    "A",
		RecipientAccountAddress: "B",
	},
	{
		ID:               4,
		FeePerByte:       100,
		ArrivalTimestamp: 1562893306,
		TransactionBytes: transaction.GetFixturesForSignedMempoolTransaction(
			4,
			1562893306,
			"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
			"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			false,
		).TransactionBytes,
		SenderAccountAddress:    "A",
		RecipientAccountAddress: "B",
	},
	{
		ID:               5,
		FeePerByte:       5,
		ArrivalTimestamp: 1562893303,
		TransactionBytes: transaction.GetFixturesForSignedMempoolTransaction(
			5,
			1562893303,
			"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
			"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			false,
		).TransactionBytes,
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
		mtxRows.AddRow(1, 0, 1, 1562893305, transaction.GetFixturesForSignedMempoolTransaction(
			1,
			1562893305,
			"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
			"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			false,
		).TransactionBytes, "A", "B")
		mtxRows.AddRow(2, 0, 10, 1562893304, transaction.GetFixturesForSignedMempoolTransaction(
			2,
			1562893304,
			"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
			"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			false,
		).TransactionBytes, "A", "B")
		mtxRows.AddRow(3, 0, 1, 1562893302, transaction.GetFixturesForSignedMempoolTransaction(
			3,
			1562893302,
			"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
			"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			false,
		).TransactionBytes, "A", "B")
		mtxRows.AddRow(4, 0, 100, 1562893306, transaction.GetFixturesForSignedMempoolTransaction(
			4,
			1562893306,
			"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
			"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			false,
		).TransactionBytes, "A", "B")
		mtxRows.AddRow(5, 0, 5, 1562893303, transaction.GetFixturesForSignedMempoolTransaction(
			5,
			1562893303,
			"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
			"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			false,
		).TransactionBytes, "A", "B")
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(mtxRows)
	}
	rows, _ := db.Query(qe)
	return rows, nil
}
func (*mockQueryExecutorSelectTransactionsFromMempoolSuccess) ExecuteSelectRow(qe string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mockRow := sqlmock.NewRows(query.NewAccountBalanceQuery().Fields)
	mockRow.AddRow(
		"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
		1,
		100,
		10,
		0,
		true,
	)
	mock.ExpectQuery("").WillReturnRows(mockRow)

	mockedRow := db.QueryRow("")
	return mockedRow, nil
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
func (*mockQueryExecutorDeleteExpiredMempoolTransactionsEmpty) ExecuteTransaction(string, ...interface{}) error {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectPrepare("").
		ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
	_, _ = db.Exec("")
	return nil
}
func (*mockQueryExecutorDeleteExpiredMempoolTransactionsEmpty) ExecuteSelect(string, bool, ...interface{}) (*sql.Rows, error) {
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
func (*mockQueryExecutorDeleteExpiredMempoolTransactions) RollbackTx() error {
	return nil
}
func (*mockQueryExecutorDeleteExpiredMempoolTransactions) ExecuteTransaction(string, ...interface{}) error {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectPrepare("").
		ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
	_, _ = db.Exec("")
	return nil
}
func (*mockQueryExecutorDeleteExpiredMempoolTransactions) ExecuteSelect(string, bool, ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mTx := transaction.GetFixturesForSignedMempoolTransaction(
		3,
		1562893302,
		"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
		"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
		true,
	)

	mock.ExpectQuery("").WillReturnRows(
		sqlmock.NewRows(mockMempoolQuery.Fields).AddRow(
			mTx.GetID(),
			mTx.GetBlockHeight(),
			mTx.GetFeePerByte(),
			mTx.GetArrivalTimestamp(),
			mTx.GetTransactionBytes(),
			mTx.GetSenderAccountAddress(),
			mTx.GetRecipientAccountAddress(),
		),
	)
	return db.Query("")
}

func TestMempoolService_DeleteExpiredMempoolTransactions(t *testing.T) {
	type fields struct {
		Chaintype              chaintype.ChainType
		QueryExecutor          query.ExecutorInterface
		MempoolQuery           query.MempoolQueryInterface
		ActionTypeSwitcher     transaction.TypeActionSwitcher
		AccountBalanceQuery    query.AccountBalanceQueryInterface
		Signature              crypto.SignatureInterface
		TransactionQuery       query.TransactionQueryInterface
		Observer               *observer.Observer
		TransactionCoreService TransactionCoreServiceInterface
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
				TransactionCoreService: NewTransactionCoreService(
					&mockQueryExecutorDeleteExpiredMempoolTransactions{},
					nil,
					nil,
					query.NewTransactionQuery(&chaintype.MainChain{}),
					nil,
					nil,
				),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mps := &MempoolService{
				TransactionUtil:        &transaction.Util{},
				Chaintype:              tt.fields.Chaintype,
				QueryExecutor:          tt.fields.QueryExecutor,
				MempoolQuery:           tt.fields.MempoolQuery,
				ActionTypeSwitcher:     tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery:    tt.fields.AccountBalanceQuery,
				Signature:              tt.fields.Signature,
				TransactionQuery:       tt.fields.TransactionQuery,
				Observer:               tt.fields.Observer,
				TransactionCoreService: tt.fields.TransactionCoreService,
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

type (
	mockTransactionUtilSuccess struct {
		transaction.UtilInterface
	}

	mockTransactionUtilErrorParse struct {
		transaction.UtilInterface
	}

	mockReceiptUtilSuccess struct {
		coreUtil.ReceiptUtilInterface
	}

	mockReceiptUtilError struct {
		coreUtil.ReceiptUtilInterface
	}

	mockMempoolServiceUtilSuccess struct {
		MempoolServiceUtilInterface
	}

	mockMempoolServiceUtilErrorDuplicate struct {
		MempoolServiceUtilInterface
	}

	mockMempoolServiceUtilErrorNonDuplicate struct {
		MempoolServiceUtilInterface
	}

	mockReceiptServiceSucces struct {
		ReceiptServiceInterface
	}

	mockReceiptServiceError struct {
		ReceiptServiceInterface
	}

	mockKvExecutorErrKeyNotFound struct {
		kvdb.KVExecutorInterface
	}

	mockKvExecutorErrNonKeyNotFound struct {
		kvdb.KVExecutorInterface
	}

	mockKvExecutorFoundKey struct {
		kvdb.KVExecutorInterface
	}
)

func (*mockTransactionUtilSuccess) ParseTransactionBytes(transactionBytes []byte, sign bool) (*model.Transaction, error) {
	return &model.Transaction{}, nil
}

func (*mockTransactionUtilErrorParse) ParseTransactionBytes(transactionBytes []byte, sign bool) (*model.Transaction, error) {
	return nil, errors.New("")
}

func (*mockReceiptUtilSuccess) GetReceiptKey(
	dataHash, senderPublicKey []byte,
) ([]byte, error) {
	return []byte{}, nil
}

func (*mockReceiptUtilError) GetReceiptKey(
	dataHash, senderPublicKey []byte,
) ([]byte, error) {
	return nil, errors.New("")
}

func (*mockMempoolServiceUtilSuccess) ValidateMempoolTransaction(mpTx *model.MempoolTransaction) error {
	return nil
}

func (*mockMempoolServiceUtilSuccess) AddMempoolTransaction(*model.MempoolTransaction) error {
	return nil
}

func (*mockReceiptServiceSucces) GenerateBatchReceiptWithReminder(
	ct chaintype.ChainType,
	receivedDatumHash []byte,
	lastBlock *model.Block,
	senderPublicKey []byte,
	nodeSecretPhrase, receiptKey string,
	datumType uint32,
) (*model.BatchReceipt, error) {
	return &model.BatchReceipt{}, nil
}

func (*mockReceiptServiceError) GenerateBatchReceiptWithReminder(
	ct chaintype.ChainType,
	receivedDatumHash []byte,
	lastBlock *model.Block,
	senderPublicKey []byte,
	nodeSecretPhrase, receiptKey string,
	datumType uint32,
) (*model.BatchReceipt, error) {
	return nil, errors.New("")
}

func (*mockMempoolServiceUtilErrorDuplicate) ValidateMempoolTransaction(mpTx *model.MempoolTransaction) error {
	return blocker.NewBlocker(blocker.DuplicateMempoolErr, "")
}

func (*mockMempoolServiceUtilErrorNonDuplicate) ValidateMempoolTransaction(mpTx *model.MempoolTransaction) error {
	return blocker.NewBlocker(blocker.ParserErr, "")
}

func (*mockKvExecutorErrKeyNotFound) Get(key string) ([]byte, error) {
	return nil, badger.ErrKeyNotFound
}

func (*mockKvExecutorErrNonKeyNotFound) Get(key string) ([]byte, error) {
	return nil, errors.New("")
}

func (*mockKvExecutorFoundKey) Get(key string) ([]byte, error) {
	return []byte{1}, nil
}

func TestMempoolService_ProcessReceivedTransaction(t *testing.T) {
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
		TransactionUtil     transaction.UtilInterface
		ReceiptUtil         coreUtil.ReceiptUtilInterface
		MempoolServiceUtil  MempoolServiceUtilInterface
		ReceiptService      ReceiptServiceInterface
	}
	type args struct {
		senderPublicKey, receivedTxBytes []byte
		lastBlock                        *model.Block
		nodeSecretPhrase                 string
	}
	type want struct {
		batchReceipt *model.BatchReceipt
		transaction  *model.Transaction
		err          bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "Fail:ParseTransaction_error",
			fields: fields{
				TransactionUtil: &mockTransactionUtilErrorParse{},
			},
			args:    args{},
			want:    want{},
			wantErr: true,
		},
		{
			name: "Fail:GetReceiptKey_error",
			fields: fields{
				TransactionUtil: &mockTransactionUtilSuccess{},
				ReceiptUtil:     &mockReceiptUtilError{},
			},
			args:    args{},
			want:    want{},
			wantErr: true,
		},
		{
			name: "Fail:ValidateMempoolTransaction_error_non_duplicate",
			fields: fields{
				QueryExecutor:      &mockGetMempoolTransactionsByBlockHeightExecutor{},
				MempoolQuery:       query.NewMempoolQuery(chaintype.GetChainType(0)),
				TransactionUtil:    &mockTransactionUtilSuccess{},
				ReceiptUtil:        &mockReceiptUtilSuccess{},
				MempoolServiceUtil: &mockMempoolServiceUtilErrorNonDuplicate{},
				ReceiptService:     &mockReceiptServiceSucces{},
			},
			args:    args{},
			want:    want{},
			wantErr: true,
		},
		{
			name: "Fail:ValidateMempoolTransaction_error_duplicate_and_kv_executor_get_error_non_err_key_not_found",
			fields: fields{
				QueryExecutor:      &mockGetMempoolTransactionsByBlockHeightExecutor{},
				MempoolQuery:       query.NewMempoolQuery(chaintype.GetChainType(0)),
				TransactionUtil:    &mockTransactionUtilSuccess{},
				ReceiptUtil:        &mockReceiptUtilSuccess{},
				MempoolServiceUtil: &mockMempoolServiceUtilErrorDuplicate{},
				ReceiptService:     &mockReceiptServiceSucces{},
				KVExecutor:         &mockKvExecutorErrNonKeyNotFound{},
			},
			args:    args{},
			want:    want{},
			wantErr: true,
		},
		{
			name: "Fail:ValidateMempoolTransaction_error_duplicate_and_kv_executor_found_the_record_the_sender_has_received_receipt_for_this_data",
			fields: fields{
				QueryExecutor:      &mockGetMempoolTransactionsByBlockHeightExecutor{},
				MempoolQuery:       query.NewMempoolQuery(chaintype.GetChainType(0)),
				TransactionUtil:    &mockTransactionUtilSuccess{},
				ReceiptUtil:        &mockReceiptUtilSuccess{},
				MempoolServiceUtil: &mockMempoolServiceUtilErrorDuplicate{},
				ReceiptService:     &mockReceiptServiceSucces{},
				KVExecutor:         &mockKvExecutorFoundKey{},
			},
			args:    args{},
			want:    want{},
			wantErr: true,
		},
		{
			name: "Fail:GenerateBatchReceiptWithReminder_error",
			fields: fields{
				QueryExecutor:      &mockGetMempoolTransactionsByBlockHeightExecutor{},
				MempoolQuery:       query.NewMempoolQuery(chaintype.GetChainType(0)),
				TransactionUtil:    &mockTransactionUtilSuccess{},
				ReceiptUtil:        &mockReceiptUtilSuccess{},
				MempoolServiceUtil: &mockMempoolServiceUtilSuccess{},
				ReceiptService:     &mockReceiptServiceError{},
			},
			args:    args{},
			want:    want{},
			wantErr: true,
		},
		{
			name: "Success:expected_returns_and_no_errors",
			fields: fields{
				QueryExecutor:      &mockGetMempoolTransactionsByBlockHeightExecutor{},
				MempoolQuery:       query.NewMempoolQuery(chaintype.GetChainType(0)),
				TransactionUtil:    &mockTransactionUtilSuccess{},
				ReceiptUtil:        &mockReceiptUtilSuccess{},
				MempoolServiceUtil: &mockMempoolServiceUtilSuccess{},
				ReceiptService:     &mockReceiptServiceSucces{},
			},
			args: args{},
			want: want{
				batchReceipt: &model.BatchReceipt{},
				transaction:  &model.Transaction{},
			},
			wantErr: false,
		},
		{
			name: "Success:duplicate_mempool_and_sender_has_not_got_received_for_this_data",
			fields: fields{
				QueryExecutor:      &mockGetMempoolTransactionsByBlockHeightExecutor{},
				MempoolQuery:       query.NewMempoolQuery(chaintype.GetChainType(0)),
				TransactionUtil:    &mockTransactionUtilSuccess{},
				ReceiptUtil:        &mockReceiptUtilSuccess{},
				MempoolServiceUtil: &mockMempoolServiceUtilErrorDuplicate{},
				ReceiptService:     &mockReceiptServiceSucces{},
				KVExecutor:         &mockKvExecutorErrKeyNotFound{},
			},
			args: args{},
			want: want{
				batchReceipt: &model.BatchReceipt{},
				transaction:  &model.Transaction{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mps := &MempoolService{
				Chaintype:           tt.fields.Chaintype,
				KVExecutor:          tt.fields.KVExecutor,
				MerkleTreeQuery:     tt.fields.MerkleTreeQuery,
				ActionTypeSwitcher:  tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				Signature:           tt.fields.Signature,
				TransactionQuery:    tt.fields.TransactionQuery,
				Observer:            tt.fields.Observer,
				Logger:              tt.fields.Logger,
				TransactionUtil:     tt.fields.TransactionUtil,
				ReceiptUtil:         tt.fields.ReceiptUtil,
				MempoolServiceUtil:  tt.fields.MempoolServiceUtil,
				ReceiptService:      tt.fields.ReceiptService,
			}
			batchReceipt, tx, err := mps.ProcessReceivedTransaction(
				tt.args.senderPublicKey,
				tt.args.receivedTxBytes,
				tt.args.lastBlock,
				tt.args.nodeSecretPhrase,
			)
			if (err != nil) != tt.wantErr {
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

type (
	mockTxTypeSuccess struct {
		transaction.TypeAction
	}
	mockActionTypeSwitcherSuccess struct {
		transaction.TypeActionSwitcher
	}
	mockEscrowTypeAction struct {
		transaction.EscrowTypeAction
	}
)

func (*mockEscrowTypeAction) EscrowApplyUnconfirmed() error {
	return nil
}
func (*mockTxTypeSuccess) ApplyUnconfirmed() error {
	return nil
}

func (*mockTxTypeSuccess) Escrowable() (transaction.EscrowTypeAction, bool) {
	return &mockEscrowTypeAction{}, true
}

func (*mockActionTypeSwitcherSuccess) GetTransactionType(tx *model.Transaction) (transaction.TypeAction, error) {
	return &mockTxTypeSuccess{}, nil
}

func TestMempoolService_ReceivedTransaction(t *testing.T) {
	type fields struct {
		Chaintype              chaintype.ChainType
		KVExecutor             kvdb.KVExecutorInterface
		QueryExecutor          query.ExecutorInterface
		MempoolQuery           query.MempoolQueryInterface
		MerkleTreeQuery        query.MerkleTreeQueryInterface
		ActionTypeSwitcher     transaction.TypeActionSwitcher
		AccountBalanceQuery    query.AccountBalanceQueryInterface
		Signature              crypto.SignatureInterface
		TransactionQuery       query.TransactionQueryInterface
		Observer               *observer.Observer
		Logger                 *log.Logger
		TransactionUtil        transaction.UtilInterface
		ReceiptUtil            coreUtil.ReceiptUtilInterface
		MempoolServiceUtil     MempoolServiceUtilInterface
		ReceiptService         ReceiptServiceInterface
		TransactionCoreService TransactionCoreServiceInterface
	}
	type args struct {
		senderPublicKey, receivedTxBytes []byte
		lastBlock                        *model.Block
		nodeSecretPhrase                 string
	}
	type want struct {
		batchReceipt *model.BatchReceipt
		err          bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "Fail:ProcessReceivedTransaction_fails",
			fields: fields{
				QueryExecutor:      &mockQueryExecutorSuccess{},
				MempoolQuery:       query.NewMempoolQuery(chaintype.GetChainType(0)),
				TransactionUtil:    &mockTransactionUtilErrorParse{},
				ReceiptUtil:        &mockReceiptUtilSuccess{},
				MempoolServiceUtil: &mockMempoolServiceUtilSuccess{},
				ReceiptService:     &mockReceiptServiceSucces{},
				ActionTypeSwitcher: &mockActionTypeSwitcherSuccess{},
				Observer:           observer.NewObserver(),
			},
			args:    args{},
			want:    want{},
			wantErr: true,
		},
		{
			name: "Success:No_errors",
			fields: fields{
				QueryExecutor:      &mockQueryExecutorSuccess{},
				MempoolQuery:       query.NewMempoolQuery(chaintype.GetChainType(0)),
				TransactionUtil:    &mockTransactionUtilSuccess{},
				ReceiptUtil:        &mockReceiptUtilSuccess{},
				MempoolServiceUtil: &mockMempoolServiceUtilSuccess{},
				ReceiptService:     &mockReceiptServiceSucces{},
				ActionTypeSwitcher: &mockActionTypeSwitcherSuccess{},
				Observer:           observer.NewObserver(),
				TransactionCoreService: NewTransactionCoreService(
					&mockQueryExecutorDeleteExpiredMempoolTransactions{},
					nil,
					nil,
					query.NewTransactionQuery(&chaintype.MainChain{}),
					nil,
					nil,
				),
			},
			args: args{},
			want: want{
				batchReceipt: &model.BatchReceipt{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mps := &MempoolService{
				Chaintype:              tt.fields.Chaintype,
				QueryExecutor:          tt.fields.QueryExecutor,
				KVExecutor:             tt.fields.KVExecutor,
				MerkleTreeQuery:        tt.fields.MerkleTreeQuery,
				ActionTypeSwitcher:     tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery:    tt.fields.AccountBalanceQuery,
				Signature:              tt.fields.Signature,
				TransactionQuery:       tt.fields.TransactionQuery,
				Observer:               tt.fields.Observer,
				Logger:                 tt.fields.Logger,
				TransactionUtil:        tt.fields.TransactionUtil,
				ReceiptUtil:            tt.fields.ReceiptUtil,
				MempoolServiceUtil:     tt.fields.MempoolServiceUtil,
				ReceiptService:         tt.fields.ReceiptService,
				TransactionCoreService: tt.fields.TransactionCoreService,
			}
			batchReceipt, err := mps.ReceivedTransaction(
				tt.args.senderPublicKey,
				tt.args.receivedTxBytes,
				tt.args.lastBlock,
				tt.args.nodeSecretPhrase,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReceivedTransaction() error = %v, wantErr %v", err, tt.want.err)
				return
			}
			if !reflect.DeepEqual(batchReceipt, tt.want.batchReceipt) {
				t.Errorf("ReceivedTransaction() batchReceipt = \n%v, want \n%v", batchReceipt, tt.want.batchReceipt)
			}
		})
	}
}
