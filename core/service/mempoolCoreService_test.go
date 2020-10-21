package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/transaction"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
	"github.com/zoobc/zoobc-core/observer"
	"golang.org/x/crypto/sha3"
)

var (
	mockMempoolQuery       = query.NewMempoolQuery(chaintype.GetChainType(0))
	mockMempoolTransaction = &model.MempoolTransaction{
		ID:               1,
		BlockHeight:      0,
		ArrivalTimestamp: 1000,
		FeePerByte:       10,
		TransactionBytes: []byte{1, 2, 3, 4, 5},
		SenderAccountAddress: []byte{0, 0, 0, 0, 4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
			45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
		RecipientAccountAddress: []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
			81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
	}
)

var _ = mockMempoolTransaction

func TestNewMempoolService(t *testing.T) {
	type args struct {
		ct                     chaintype.ChainType
		queryExecutor          query.ExecutorInterface
		mempoolQuery           query.MempoolQueryInterface
		merkleTreeQuery        query.MerkleTreeQueryInterface
		accountBalanceQuery    query.AccountBalanceQueryInterface
		transactionQuery       query.TransactionQueryInterface
		actionTypeSwitcher     transaction.TypeActionSwitcher
		obsr                   *observer.Observer
		signature              crypto.SignatureInterface
		logger                 *log.Logger
		transactionUtil        transaction.UtilInterface
		receiptUtil            coreUtil.ReceiptUtilInterface
		receiptService         ReceiptServiceInterface
		TransactionCoreService TransactionCoreServiceInterface
		BlockStateStorage      storage.CacheStorageInterface
		MempoolCacheStorage    storage.CacheStorageInterface
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
		test.args.queryExecutor,
		test.args.mempoolQuery,
		test.args.merkleTreeQuery,
		test.args.actionTypeSwitcher,
		test.args.accountBalanceQuery,
		test.args.transactionQuery,
		test.args.signature,
		test.args.obsr,
		test.args.logger,
		test.args.receiptUtil,
		test.args.receiptService,
		test.args.TransactionCoreService,
		test.args.BlockStateStorage,
		test.args.MempoolCacheStorage,
		nil,
	)

	if !reflect.DeepEqual(got, test.want) {
		jGot, _ := json.MarshalIndent(got, "", "  ")
		jWant, _ := json.MarshalIndent(test.want, "", "  ")
		t.Errorf("NewMempoolService() = %s, want %s", jGot, jWant)
	}
}

var mempoolSenderAccountAddress1 = []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56,
	139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169}
var mempoolRecipientAccountAddress1 = []byte{0, 0, 0, 0, 31, 61, 150, 75, 69, 179, 131, 81, 155, 32, 54, 19, 63, 225, 154, 35, 152, 215, 161,
	242, 32, 28, 136, 189, 16, 27, 197, 211, 161, 252, 211, 195}
var mempoolRecipientAccountAddress2 = []byte{0, 0, 0, 0, 99, 4, 26, 226, 105, 29, 218, 56, 18, 173, 152, 185, 58, 97, 189, 6, 16, 1, 126,
	159, 75, 224, 91, 137, 93, 206, 174, 151, 229, 184, 214, 80}
var mempoolRecipientAccountAddress3 = []byte{0, 0, 0, 0, 201, 123, 194, 252, 228, 24, 9, 99, 127, 53, 38, 126, 200, 49, 227, 202, 245, 49,
	82, 41, 93, 168, 92, 182, 52, 79, 35, 103, 76, 244, 60, 127}
var mempoolRecipientAccountAddress4 = []byte{0, 0, 0, 0, 1, 237, 22, 193, 217, 33, 254, 63, 28, 91, 184, 164, 172, 170, 248, 68, 130, 162,
	185, 7, 36, 130, 151, 239, 148, 255, 206, 49, 26, 82, 17, 49}
var mempoolRecipientAccountAddress5 = []byte{0, 0, 0, 0, 74, 132, 47, 111, 228, 161, 96, 163, 111, 165, 204, 196, 54, 89, 167, 156, 227,
	191, 195, 212, 254, 211, 54, 195, 204, 23, 49, 22, 89, 135, 29, 243}

var mockSuccessSelectMempool = []*model.MempoolTransaction{
	{
		ID:               1,
		FeePerByte:       1,
		ArrivalTimestamp: 1562893305,
		TransactionBytes: transaction.GetFixturesForSignedMempoolTransaction(
			1,
			1562893305,
			mempoolSenderAccountAddress1,
			mempoolRecipientAccountAddress1,
			false,
		).TransactionBytes,
		SenderAccountAddress:    mempoolSenderAccountAddress1,
		RecipientAccountAddress: mempoolRecipientAccountAddress1,
	},
	{
		ID:               2,
		FeePerByte:       10,
		ArrivalTimestamp: 1562893304,
		TransactionBytes: transaction.GetFixturesForSignedMempoolTransaction(
			2,
			1562893304,
			mempoolSenderAccountAddress1,
			mempoolRecipientAccountAddress2,
			false,
		).TransactionBytes,
		SenderAccountAddress:    mempoolSenderAccountAddress1,
		RecipientAccountAddress: mempoolRecipientAccountAddress2,
	},
	{
		ID:               3,
		FeePerByte:       1,
		ArrivalTimestamp: 1562893302,
		TransactionBytes: transaction.GetFixturesForSignedMempoolTransaction(
			3,
			1562893302,
			mempoolSenderAccountAddress1,
			mempoolRecipientAccountAddress3,
			false,
		).TransactionBytes,
		SenderAccountAddress:    mempoolSenderAccountAddress1,
		RecipientAccountAddress: mempoolRecipientAccountAddress3,
	},
	{
		ID:               4,
		FeePerByte:       100,
		ArrivalTimestamp: 1562893306,
		TransactionBytes: transaction.GetFixturesForSignedMempoolTransaction(
			4,
			1562893306,
			mempoolSenderAccountAddress1,
			mempoolRecipientAccountAddress4,
			false,
		).TransactionBytes,
		SenderAccountAddress:    mempoolSenderAccountAddress1,
		RecipientAccountAddress: mempoolRecipientAccountAddress4,
	},
	{
		ID:               5,
		FeePerByte:       5,
		ArrivalTimestamp: 1562893303,
		TransactionBytes: transaction.GetFixturesForSignedMempoolTransaction(
			5,
			1562893303,
			mempoolSenderAccountAddress1,
			mempoolRecipientAccountAddress5,
			false,
		).TransactionBytes,
		SenderAccountAddress:    mempoolSenderAccountAddress1,
		RecipientAccountAddress: mempoolRecipientAccountAddress5,
	},
}

type (
	mockSelectTransactionFromMempoolFeeScaleServiceSuccessCache struct {
		fee.FeeScaleServiceInterface
	}
	mockCacheStorageSelectMempoolSuccess struct {
		storage.MempoolCacheStorage
	}
)

func (*mockCacheStorageSelectMempoolSuccess) GetAllItems(item interface{}) error {
	successTx1, _ := (&transaction.Util{
		MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
	}).ParseTransactionBytes(mockSuccessSelectMempool[0].TransactionBytes, true)
	successTx2, _ := (&transaction.Util{
		MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
	}).ParseTransactionBytes(mockSuccessSelectMempool[1].TransactionBytes, true)
	successTx3, _ := (&transaction.Util{
		MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
	}).ParseTransactionBytes(mockSuccessSelectMempool[2].TransactionBytes, true)
	successTx4, _ := (&transaction.Util{
		MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
	}).ParseTransactionBytes(mockSuccessSelectMempool[3].TransactionBytes, true)
	successTx5, _ := (&transaction.Util{
		MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
	}).ParseTransactionBytes(mockSuccessSelectMempool[4].TransactionBytes, true)
	itemCopy, ok := item.(storage.MempoolMap)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongTypeItem")
	}
	itemCopy[successTx1.ID] = storage.MempoolCacheObject{
		Tx:                  *successTx1,
		ArrivalTimestamp:    mockSuccessSelectMempool[0].ArrivalTimestamp,
		FeePerByte:          mockSuccessSelectMempool[0].FeePerByte,
		TransactionByteSize: uint32(len(mockSuccessSelectMempool[0].TransactionBytes)),
	}
	itemCopy[successTx2.ID] = storage.MempoolCacheObject{
		Tx:                  *successTx2,
		ArrivalTimestamp:    mockSuccessSelectMempool[1].ArrivalTimestamp,
		FeePerByte:          mockSuccessSelectMempool[1].FeePerByte,
		TransactionByteSize: uint32(len(mockSuccessSelectMempool[1].TransactionBytes)),
	}
	itemCopy[successTx3.ID] = storage.MempoolCacheObject{
		Tx:                  *successTx3,
		ArrivalTimestamp:    mockSuccessSelectMempool[2].ArrivalTimestamp,
		FeePerByte:          mockSuccessSelectMempool[2].FeePerByte,
		TransactionByteSize: uint32(len(mockSuccessSelectMempool[2].TransactionBytes)),
	}
	itemCopy[successTx4.ID] = storage.MempoolCacheObject{
		Tx:                  *successTx4,
		ArrivalTimestamp:    mockSuccessSelectMempool[3].ArrivalTimestamp,
		FeePerByte:          mockSuccessSelectMempool[3].FeePerByte,
		TransactionByteSize: uint32(len(mockSuccessSelectMempool[3].TransactionBytes)),
	}
	itemCopy[successTx5.ID] = storage.MempoolCacheObject{
		Tx:                  *successTx5,
		ArrivalTimestamp:    mockSuccessSelectMempool[4].ArrivalTimestamp,
		FeePerByte:          mockSuccessSelectMempool[4].FeePerByte,
		TransactionByteSize: uint32(len(mockSuccessSelectMempool[4].TransactionBytes)),
	}
	return nil
}

func (*mockSelectTransactionFromMempoolFeeScaleServiceSuccessCache) GetLatestFeeScale(feeScale *model.FeeScale) error {
	*feeScale = model.FeeScale{
		FeeScale:    constant.OneZBC,
		BlockHeight: 0,
		Latest:      true,
	}
	return nil
}
func (*mockSelectTransactionFromMempoolFeeScaleServiceSuccessCache) InsertFeeScale(feeScale *model.FeeScale) error {
	return nil
}

type (
	mockAccountDatasetQueryMempoolCoreService struct {
		query.AccountDatasetQuery
		wantNoRow bool
	}
)

func (*mockAccountDatasetQueryMempoolCoreService) GetAccountDatasetEscrowApproval([]byte) (qry string, args []interface{}) {
	return
}
func (m *mockAccountDatasetQueryMempoolCoreService) Scan(dataset *model.AccountDataset, _ *sql.Row) error {
	if m.wantNoRow {
		return sql.ErrNoRows
	}
	*dataset = model.AccountDataset{
		SetterAccountAddress: []byte{0, 0, 0, 0, 4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
			45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
		RecipientAccountAddress: []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
			81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},

		Property: "Admin",
		Value:    "You're Welcome",
		IsActive: true,
		Latest:   true,
		Height:   5,
	}

	return nil
}

type mockQueryExecutoMempoolCoreService struct {
	query.Executor
	wantErr     bool
	wantErrType error
}

func (m *mockQueryExecutoMempoolCoreService) ExecuteSelectRow(qu string, tx bool, args ...interface{}) (*sql.Row, error) {
	if m.wantErr {
		if m.wantErrType == sql.ErrNoRows {
			db, mock, _ := sqlmock.New()
			mock.ExpectQuery(regexp.QuoteMeta(qu)).WillReturnError(sql.ErrNoRows)
			return db.QueryRow(qu), nil
		}
		return nil, m.wantErrType
	}

	db, mock, _ := sqlmock.New()
	mock.ExpectQuery(regexp.QuoteMeta(qu)).WillReturnRows(sqlmock.NewRows([]string{"column"}))
	return db.QueryRow(qu), nil
}

func TestMempoolService_SelectTransactionsFromMempool(t *testing.T) {
	successTx1, _ := (&transaction.Util{
		MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
	}).ParseTransactionBytes(mockSuccessSelectMempool[0].TransactionBytes, true)
	successTx2, _ := (&transaction.Util{
		MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
	}).ParseTransactionBytes(mockSuccessSelectMempool[1].TransactionBytes, true)
	successTx3, _ := (&transaction.Util{
		MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
	}).ParseTransactionBytes(mockSuccessSelectMempool[2].TransactionBytes, true)
	successTx4, _ := (&transaction.Util{
		MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
	}).ParseTransactionBytes(mockSuccessSelectMempool[3].TransactionBytes, true)
	successTx5, _ := (&transaction.Util{
		MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
	}).ParseTransactionBytes(mockSuccessSelectMempool[4].TransactionBytes, true)
	type fields struct {
		Chaintype           chaintype.ChainType
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
				ActionTypeSwitcher:  &transaction.TypeSwitcher{},
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
			},
			args: args{
				blockTimestamp: 1562893106,
			},
			want: []*model.Transaction{
				successTx4, successTx2, successTx5, successTx3, successTx1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mps := &MempoolService{
				TransactionUtil: &transaction.Util{
					FeeScaleService:     &mockSelectTransactionFromMempoolFeeScaleServiceSuccessCache{},
					AccountDatasetQuery: &mockAccountDatasetQueryMempoolCoreService{wantNoRow: true},
					QueryExecutor:       &mockQueryExecutoMempoolCoreService{},
				},
				Chaintype:           tt.fields.Chaintype,
				ActionTypeSwitcher:  tt.fields.ActionTypeSwitcher,
				MempoolCacheStorage: &mockCacheStorageSelectMempoolSuccess{},
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
			}
			got, err := mps.SelectTransactionsFromMempool(tt.args.blockTimestamp, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("MempoolService.SelectTransactionsFromMempool() error = \n%v, wantErr \n%v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MempoolService.SelectTransactionsFromMempool() = \n%v, want \n%v", got, tt.want)
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
		[]byte{0, 0, 0, 0, 4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
			45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
		[]byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
			81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
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

type (
	mockMempoolCacheStorageEmpty struct {
		storage.MempoolCacheStorage
	}
	mockMempoolCacheStorageExpiryExist struct {
		storage.MempoolCacheStorage
	}
)

func (*mockMempoolCacheStorageEmpty) GetAllItems(item interface{}) error {
	return nil
}

func (*mockMempoolCacheStorageExpiryExist) GetAllItems(item interface{}) error {
	itemCopy, ok := item.(storage.MempoolMap)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongTypeItem")
	}
	mTx := transaction.GetFixturesForTransaction(
		1562893302,
		[]byte{0, 0, 0, 0, 4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
			45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
		[]byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
			81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
		true,
	)
	itemCopy[1111] = storage.MempoolCacheObject{
		ArrivalTimestamp: 10,
		Tx:               *mTx,
	}

	return nil
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
		MempoolCacheStorage    storage.CacheStorageInterface
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
				QueryExecutor:       &mockQueryExecutorDeleteExpiredMempoolTransactionsEmpty{},
				MempoolQuery:        mockMempoolQuery,
				MempoolCacheStorage: &mockMempoolCacheStorageEmpty{},
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
				MempoolCacheStorage: &mockMempoolCacheStorageExpiryExist{},
				TransactionCoreService: NewTransactionCoreService(
					log.New(),
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
				TransactionUtil: &transaction.Util{
					MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
				},
				Chaintype:              tt.fields.Chaintype,
				QueryExecutor:          tt.fields.QueryExecutor,
				MempoolQuery:           tt.fields.MempoolQuery,
				ActionTypeSwitcher:     tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery:    tt.fields.AccountBalanceQuery,
				Signature:              tt.fields.Signature,
				TransactionQuery:       tt.fields.TransactionQuery,
				Observer:               tt.fields.Observer,
				TransactionCoreService: tt.fields.TransactionCoreService,
				MempoolCacheStorage:    tt.fields.MempoolCacheStorage,
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

type (
	mockMempoolCacheStorageGetMempoolTransactionsByBlockHeightSuccess struct {
		storage.MempoolCacheStorage
	}
	mockMempoolCacheStorageGetMempoolTransactionsByBlockHeightSuccessReturnExpired struct {
		storage.MempoolCacheStorage
	}
)

func (*mockMempoolCacheStorageGetMempoolTransactionsByBlockHeightSuccess) GetAllItems(item interface{}) error {
	itemCopy := item.(storage.MempoolMap)
	itemCopy[mockTransaction.ID] = storage.MempoolCacheObject{
		Tx:                  *mockTransaction,
		ArrivalTimestamp:    mockMempoolTransaction.ArrivalTimestamp,
		FeePerByte:          mockMempoolTransaction.FeePerByte,
		TransactionByteSize: uint32(len(mockMempoolTransaction.TransactionBytes)),
		BlockHeight:         mockTransaction.Height,
	}
	return nil
}

func (*mockMempoolCacheStorageGetMempoolTransactionsByBlockHeightSuccessReturnExpired) GetAllItems(item interface{}) error {
	itemCopy := item.(storage.MempoolMap)
	itemCopy[mockTransaction.ID] = storage.MempoolCacheObject{
		Tx:                  *mockTransactionExpired,
		ArrivalTimestamp:    mockMempoolTransaction.ArrivalTimestamp,
		FeePerByte:          mockMempoolTransaction.FeePerByte,
		TransactionByteSize: uint32(len(mockMempoolTransaction.TransactionBytes)),
		BlockHeight:         mockTransactionExpired.Height,
	}
	return nil
}

func TestMempoolService_GetMempoolTransactionsByBlockHeight(t *testing.T) {
	type fields struct {
		Chaintype           chaintype.ChainType
		QueryExecutor       query.ExecutorInterface
		MempoolQuery        query.MempoolQueryInterface
		MerkleTreeQuery     query.MerkleTreeQueryInterface
		ActionTypeSwitcher  transaction.TypeActionSwitcher
		AccountBalanceQuery query.AccountBalanceQueryInterface
		Signature           crypto.SignatureInterface
		TransactionQuery    query.TransactionQueryInterface
		Observer            *observer.Observer
		MempoolCacheStorage storage.CacheStorageInterface
		Logger              *log.Logger
	}
	type args struct {
		height uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.Transaction
		wantErr bool
	}{
		{
			name: "wantSuccess - no expired",
			fields: fields{
				MempoolQuery:        query.NewMempoolQuery(chaintype.GetChainType(0)),
				MempoolCacheStorage: &mockMempoolCacheStorageGetMempoolTransactionsByBlockHeightSuccess{},
			},
			args: args{height: 0},
			want: make([]*model.Transaction, 0),
		},
		{
			name: "wantSuccess - with expired",
			fields: fields{
				MempoolQuery:        query.NewMempoolQuery(chaintype.GetChainType(0)),
				MempoolCacheStorage: &mockMempoolCacheStorageGetMempoolTransactionsByBlockHeightSuccessReturnExpired{},
			},
			args: args{height: 0},
			want: []*model.Transaction{mockTransactionExpired},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mps := &MempoolService{
				Chaintype:           tt.fields.Chaintype,
				QueryExecutor:       tt.fields.QueryExecutor,
				MempoolQuery:        tt.fields.MempoolQuery,
				MerkleTreeQuery:     tt.fields.MerkleTreeQuery,
				ActionTypeSwitcher:  tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				Signature:           tt.fields.Signature,
				TransactionQuery:    tt.fields.TransactionQuery,
				Observer:            tt.fields.Observer,
				Logger:              tt.fields.Logger,
				MempoolCacheStorage: tt.fields.MempoolCacheStorage,
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

	mockReceiptServiceSucces struct {
		ReceiptServiceInterface
		WantErr        bool
		WantDuplicated bool
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

func (*mockReceiptServiceSucces) GenerateReceipt(
	chaintype.ChainType, []byte, *model.Block, []byte, string, uint32,
) (*model.Receipt, error) {
	return &model.Receipt{}, nil
}

func (mrs *mockReceiptServiceSucces) CheckDuplication(publicKey, datumHash []byte) error {
	if mrs.WantErr {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"FailedGetReceiptKey",
		)
	}
	if mrs.WantDuplicated {
		return blocker.NewBlocker(blocker.DuplicateReceiptErr, "ReceiptExistsOnReminder")
	}
	return nil
}

type (
	mockMempoolCacheStorageFailGetItem struct {
		storage.MempoolCacheStorage
	}
)

func (*mockMempoolCacheStorageFailGetItem) GetItem(key, item interface{}) error {
	return errors.New("mocked error")
}

func TestMempoolService_ProcessReceivedTransaction(t *testing.T) {
	type fields struct {
		Chaintype           chaintype.ChainType
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
		ReceiptService      ReceiptServiceInterface
		MempoolCacheStorage storage.CacheStorageInterface
	}
	type args struct {
		senderPublicKey, receivedTxBytes []byte
		lastBlock                        *model.Block
		nodeSecretPhrase                 string
	}
	type want struct {
		batchReceipt *model.Receipt
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
				TransactionUtil:     &mockTransactionUtilErrorParse{},
				MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
			},
			args:    args{},
			want:    want{},
			wantErr: true,
		},
		{
			name: "Fail:ValidateMempoolTransaction_error_non_duplicate",
			fields: fields{
				QueryExecutor:       &mockGetMempoolTransactionsByBlockHeightExecutor{},
				MempoolQuery:        query.NewMempoolQuery(chaintype.GetChainType(0)),
				TransactionUtil:     &mockTransactionUtilSuccess{},
				ReceiptUtil:         &mockReceiptUtilSuccess{},
				ReceiptService:      &mockReceiptServiceSucces{},
				MempoolCacheStorage: &mockMempoolCacheStorageFailGetItem{},
			},
			args:    args{},
			want:    want{},
			wantErr: true,
		},
		{
			name: "Fail:ValidateMempoolTransaction_error_duplicate_and_kv_executor_get_error_non_err_key_not_found",
			fields: fields{
				QueryExecutor:       &mockGetMempoolTransactionsByBlockHeightExecutor{},
				MempoolQuery:        query.NewMempoolQuery(chaintype.GetChainType(0)),
				TransactionUtil:     &mockTransactionUtilSuccess{},
				ReceiptUtil:         &mockReceiptUtilSuccess{},
				ReceiptService:      &mockReceiptServiceSucces{WantErr: true},
				MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
			},
			args:    args{},
			want:    want{},
			wantErr: true,
		},
		{
			name: "Fail:ValidateMempoolTransaction_error_duplicate_and_kv_executor_found_the_record_the_sender_has_received_receipt_for_this_data",
			fields: fields{
				QueryExecutor:       &mockGetMempoolTransactionsByBlockHeightExecutor{},
				MempoolQuery:        query.NewMempoolQuery(chaintype.GetChainType(0)),
				TransactionUtil:     &mockTransactionUtilSuccess{},
				ReceiptUtil:         &mockReceiptUtilSuccess{},
				ReceiptService:      &mockReceiptServiceSucces{WantDuplicated: true},
				MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
			},
			args:    args{},
			want:    want{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mps := &MempoolService{
				Chaintype:           tt.fields.Chaintype,
				MerkleTreeQuery:     tt.fields.MerkleTreeQuery,
				ActionTypeSwitcher:  tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				Signature:           tt.fields.Signature,
				TransactionQuery:    tt.fields.TransactionQuery,
				Observer:            tt.fields.Observer,
				Logger:              tt.fields.Logger,
				TransactionUtil:     tt.fields.TransactionUtil,
				ReceiptUtil:         tt.fields.ReceiptUtil,
				ReceiptService:      tt.fields.ReceiptService,
				QueryExecutor:       tt.fields.QueryExecutor,
				MempoolCacheStorage: tt.fields.MempoolCacheStorage,
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
				return
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
		mockBlockData.GetHeight(),
		mockBlockData.GetID(),
		mockBlockData.GetBlockHash(),
		mockBlockData.GetPreviousBlockHash(),
		mockBlockData.GetTimestamp(),
		mockBlockData.GetBlockSeed(),
		mockBlockData.GetBlockSignature(),
		mockBlockData.GetCumulativeDifficulty(),
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

type (
	mockCacheStorageAlwaysSuccess struct {
		storage.CacheStorageInterface
	}
)

func (*mockCacheStorageAlwaysSuccess) SetItem(key, item interface{}) error { return nil }
func (*mockCacheStorageAlwaysSuccess) GetItem(key, item interface{}) error { return nil }
func (*mockCacheStorageAlwaysSuccess) GetAllItems(item interface{}) error  { return nil }
func (*mockCacheStorageAlwaysSuccess) RemoveItem(key interface{}) error    { return nil }
func (*mockCacheStorageAlwaysSuccess) GetSize() int64                      { return 0 }
func (*mockCacheStorageAlwaysSuccess) ClearCache() error                   { return nil }

func TestMempoolService_AddMempoolTransaction(t *testing.T) {
	type fields struct {
		QueryExecutor      query.ExecutorInterface
		MempoolQuery       query.MempoolQueryInterface
		BlockQuery         query.BlockQueryInterface
		ActionTypeSwitcher transaction.TypeActionSwitcher
		BlockStateStorage  storage.CacheStorageInterface
		MempoolStorage     storage.CacheStorageInterface
		Observer           *observer.Observer
	}
	type args struct {
		mpTx *model.Transaction
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
				MempoolQuery:       query.NewMempoolQuery(&chaintype.MainChain{}),
				BlockQuery:         query.NewBlockQuery(chaintype.GetChainType(0)),
				QueryExecutor:      &mockMempoolQueryExecutorSuccess{},
				ActionTypeSwitcher: &transaction.TypeSwitcher{},
				BlockStateStorage:  &mockCacheStorageAlwaysSuccess{},
				MempoolStorage:     &mockCacheStorageAlwaysSuccess{},
			},
			args: args{
				mpTx: transaction.GetFixturesForTransaction(
					1562893302,
					[]byte{0, 0, 0, 0, 4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
						45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
					[]byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
						81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
					false,
				),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mps := &MempoolService{
				QueryExecutor:       tt.fields.QueryExecutor,
				MempoolQuery:        tt.fields.MempoolQuery,
				ActionTypeSwitcher:  tt.fields.ActionTypeSwitcher,
				BlockStateStorage:   tt.fields.BlockStateStorage,
				MempoolCacheStorage: tt.fields.MempoolStorage,
			}
			if err := mps.AddMempoolTransaction(tt.args.mpTx, nil); (err != nil) != tt.wantErr {
				t.Errorf("MempoolService.AddMempoolTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
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
			false,
		),
	)
	return db.QueryRow(qStr), nil
}

func (*mockExecutorValidateMempoolTransactionSuccessNoRow) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	switch strings.Contains(qStr, "FROM account_balance") {
	case true:
		mockedRow := mock.NewRows(query.NewAccountBalanceQuery().Fields)
		mockedRow.AddRow(
			"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
			1,
			100,
			10,
			0,
			true,
		)
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(mockedRow)
	default:
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(
			sqlmock.NewRows(query.NewTransactionQuery(&chaintype.MainChain{}).Fields),
		)
	}
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

type (
	mockValidateMempoolTransactionScaleServiceSuccessCache struct {
		fee.FeeScaleServiceInterface
	}
)

func (*mockValidateMempoolTransactionScaleServiceSuccessCache) GetLatestFeeScale(feeScale *model.FeeScale) error {
	*feeScale = model.FeeScale{
		FeeScale:    constant.OneZBC,
		BlockHeight: 0,
		Latest:      true,
	}
	return nil
}
func (*mockValidateMempoolTransactionScaleServiceSuccessCache) InsertFeeScale(feeScale *model.FeeScale) error {
	return nil
}

func TestMempoolService_ValidateMempoolTransaction(t *testing.T) {
	var (
		senderAccountAddress = []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
			81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169}
		recipientAccountAddress = []byte{0, 0, 0, 0, 4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
			45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135}
	)
	successTx := transaction.GetFixturesForTransaction(
		1562893302,
		senderAccountAddress,
		recipientAccountAddress,
		false,
	)
	txBytes, _ := transactionUtil.GetTransactionBytes(successTx, false)
	txBytesHash := sha3.Sum256(txBytes)
	successTx.Signature, _ = (&crypto.Signature{}).Sign(txBytesHash[:], model.SignatureType_DefaultSignature,
		"concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved")
	type fields struct {
		Chaintype              chaintype.ChainType
		QueryExecutor          query.ExecutorInterface
		MempoolQuery           query.MempoolQueryInterface
		ActionTypeSwitcher     transaction.TypeActionSwitcher
		AccountBalanceQuery    query.AccountBalanceQueryInterface
		TransactionQuery       query.TransactionQueryInterface
		Observer               *observer.Observer
		TransactionCoreService TransactionCoreServiceInterface
		MempoolCacheStorage    storage.CacheStorageInterface
	}
	type args struct {
		mpTx *model.Transaction
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
				TransactionCoreService: NewTransactionCoreService(
					log.New(), &mockExecutorValidateMempoolTransactionSuccessNoRow{},
					nil,
					nil,
					query.NewTransactionQuery(&chaintype.MainChain{}),
					nil,
					nil,
				),
				MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
			},
			args: args{
				mpTx: successTx,
			},
			wantErr: false,
		},
		{
			name: "wantErr:TransactionExisted",
			fields: fields{
				Chaintype:           &chaintype.MainChain{},
				QueryExecutor:       &mockExecutorValidateMempoolTransactionSuccess{},
				MempoolQuery:        query.NewMempoolQuery(&chaintype.MainChain{}),
				ActionTypeSwitcher:  &transaction.TypeSwitcher{},
				TransactionQuery:    query.NewTransactionQuery(&chaintype.MainChain{}),
				MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
			},
			args: args{
				mpTx: transaction.GetFixturesForTransaction(
					1562893302,
					[]byte{0, 0, 0, 0, 4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
						45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
					[]byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
						81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
					false,
				),
			},
			wantErr: true,
		},
		{
			name: "wantErr:TransactionExisted",
			fields: fields{
				Chaintype:           &chaintype.MainChain{},
				QueryExecutor:       &mockExecutorValidateMempoolTransactionFail{},
				TransactionQuery:    query.NewTransactionQuery(&chaintype.MainChain{}),
				MempoolQuery:        query.NewMempoolQuery(&chaintype.MainChain{}),
				ActionTypeSwitcher:  &transaction.TypeSwitcher{},
				MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
			},
			args: args{
				mpTx: transaction.GetFixturesForTransaction(
					1562893302,
					[]byte{0, 0, 0, 0, 4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
						45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
					[]byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
						81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
					false,
				),
			},
			wantErr: true,
		},
		{
			name: "wantErr:ParseFail",
			fields: fields{
				Chaintype:           &chaintype.MainChain{},
				QueryExecutor:       &mockExecutorValidateMempoolTransactionSuccessNoRow{},
				TransactionQuery:    query.NewTransactionQuery(&chaintype.MainChain{}),
				MempoolQuery:        query.NewMempoolQuery(&chaintype.MainChain{}),
				ActionTypeSwitcher:  &transaction.TypeSwitcher{},
				MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
			},
			args: args{
				mpTx: &model.Transaction{
					ID: 12,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mps := &MempoolService{
				QueryExecutor:       tt.fields.QueryExecutor,
				MempoolQuery:        tt.fields.MempoolQuery,
				ActionTypeSwitcher:  tt.fields.ActionTypeSwitcher,
				AccountBalanceQuery: tt.fields.AccountBalanceQuery,
				TransactionQuery:    tt.fields.TransactionQuery,
				TransactionUtil: &transaction.Util{
					FeeScaleService:     &mockValidateMempoolTransactionScaleServiceSuccessCache{},
					MempoolCacheStorage: &mockCacheStorageAlwaysSuccess{},
					QueryExecutor:       &mockQueryExecutoMempoolCoreService{},
					AccountDatasetQuery: &mockAccountDatasetQueryMempoolCoreService{wantNoRow: true},
				},
				TransactionCoreService: tt.fields.TransactionCoreService,
				MempoolCacheStorage:    tt.fields.MempoolCacheStorage,
			}
			if err := mps.ValidateMempoolTransaction(tt.args.mpTx); (err != nil) != tt.wantErr {
				t.Errorf("MempoolServiceUtil.ValidateMempoolTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
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
	sender := []byte{0, 0, 0, 0, 4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49, 45, 118, 97, 219, 80, 242, 244, 100, 134, 144,
		246, 37, 144, 213, 135}
	recipient := []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229,
		184, 77, 80, 80, 39, 254, 173, 28, 169}

	mockedRows := sqlmock.NewRows(query.NewMempoolQuery(chaintype.GetChainType(0)).Fields)
	mockedRows.AddRow(1, 0, 1, 1562893305, transaction.GetFixturesForSignedMempoolTransaction(1, 1562893305,
		sender, recipient, false).TransactionBytes, "A", "B")
	mockedRows.AddRow(2, 0, 10, 1562893304, transaction.GetFixturesForSignedMempoolTransaction(2, 1562893304,
		sender, recipient, false).TransactionBytes, "A", "B")
	mockedRows.AddRow(3, 0, 1, 1562893302, transaction.GetFixturesForSignedMempoolTransaction(3, 1562893302,
		sender, recipient, false).TransactionBytes, "A", "B")
	mockedRows.AddRow(4, 0, 100, 1562893306, transaction.GetFixturesForSignedMempoolTransaction(4, 1562893306,
		sender, recipient, false).TransactionBytes, "A", "B")
	mockedRows.AddRow(5, 0, 5, 1562893303, transaction.GetFixturesForSignedMempoolTransaction(5, 1562893303,
		sender, recipient, false).TransactionBytes, "A", "B")
	mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(mockedRows)
	rows, _ := db.Query(qe)
	return rows, nil
}

func (*mockQueryExecutorGetMempoolTransactionsFail) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("mockError:executeSelectFail")
}

type (
	mockMempoolCacheGetMempoolTransactionsSuccess struct {
		storage.MempoolCacheStorage
	}
	mockCacheStorageGetAllItemsError struct {
		storage.MempoolCacheStorage
	}
)

func (*mockCacheStorageGetAllItemsError) GetAllItems(items interface{}) error {
	return errors.New("mockedError")
}

var (
	sender = []byte{0, 0, 0, 0, 4, 5, 6, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49, 45, 118, 97, 219, 80, 242, 244, 100, 134, 144,
		246, 37, 144, 213, 135}
	recipient = []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229,
		184, 77, 80, 80, 39, 254, 173, 28, 169}

	mockMempoolObjectsMap = storage.MempoolMap{
		1: {
			Tx: model.Transaction{
				ID:                      1,
				SenderAccountAddress:    sender,
				RecipientAccountAddress: recipient,
			},
			FeePerByte:       1,
			ArrivalTimestamp: 1562893305,
			TransactionByteSize: uint32(len(transaction.GetFixturesForSignedMempoolTransaction(1, 1562893305,
				sender, recipient, false).TransactionBytes)),
		},
		2: {
			Tx: model.Transaction{
				ID: 2,

				SenderAccountAddress:    sender,
				RecipientAccountAddress: recipient,
			},
			FeePerByte:       10,
			ArrivalTimestamp: 1562893304,
			TransactionByteSize: uint32(len(transaction.GetFixturesForSignedMempoolTransaction(2, 1562893304,
				sender, recipient, false).TransactionBytes)),
		},
		3: {
			Tx: model.Transaction{
				ID:                      3,
				SenderAccountAddress:    sender,
				RecipientAccountAddress: recipient,
			},
			FeePerByte:       1,
			ArrivalTimestamp: 1562893302,
			TransactionByteSize: uint32(len(transaction.GetFixturesForSignedMempoolTransaction(3, 1562893302,
				sender, recipient, false).TransactionBytes)),
		},
		4: {
			Tx: model.Transaction{
				ID:                      4,
				SenderAccountAddress:    sender,
				RecipientAccountAddress: recipient,
			},
			FeePerByte:       100,
			ArrivalTimestamp: 1562893306,
			TransactionByteSize: uint32(len(transaction.GetFixturesForSignedMempoolTransaction(4, 1562893306,
				sender, recipient, false).TransactionBytes)),
		},
		5: {
			Tx: model.Transaction{
				ID:                      5,
				SenderAccountAddress:    sender,
				RecipientAccountAddress: recipient,
			},
			FeePerByte:       5,
			ArrivalTimestamp: 1562893303,
			TransactionByteSize: uint32(len(transaction.GetFixturesForSignedMempoolTransaction(5, 1562893303,
				sender, recipient, false).TransactionBytes)),
		},
	}
)

func (m *mockMempoolCacheGetMempoolTransactionsSuccess) GetAllItems(item interface{}) error {
	itemCopy, ok := item.(storage.MempoolMap)
	if !ok {
		return blocker.NewBlocker(blocker.ValidationErr, "WrongTypeItem")
	}
	for i := 1; i <= 5; i++ {
		itemCopy[int64(i)] = mockMempoolObjectsMap[int64(i)]
	}
	return nil
}

func TestMempoolService_GetMempoolTransactions(t *testing.T) {
	type fields struct {
		Chaintype           chaintype.ChainType
		QueryExecutor       query.ExecutorInterface
		MempoolQuery        query.MempoolQueryInterface
		AccountBalanceQuery query.AccountBalanceQueryInterface
		MempoolCacheStorage storage.CacheStorageInterface
	}
	tests := []struct {
		name    string
		fields  fields
		want    []storage.MempoolCacheObject
		wantErr bool
	}{
		{
			name: "GetMempoolTransactions:Success",
			fields: fields{
				MempoolQuery:        query.NewMempoolQuery(&chaintype.MainChain{}),
				QueryExecutor:       &mockQueryExecutorGetMempoolTransactionsSuccess{},
				MempoolCacheStorage: &mockMempoolCacheGetMempoolTransactionsSuccess{},
			},
			want: []storage.MempoolCacheObject{
				mockMempoolObjectsMap[1],
				mockMempoolObjectsMap[2],
				mockMempoolObjectsMap[3],
				mockMempoolObjectsMap[4],
				mockMempoolObjectsMap[5],
			},
			wantErr: false,
		},
		{
			name: "GetMempoolTransactions:Fail",
			fields: fields{
				MempoolQuery:        query.NewMempoolQuery(&chaintype.MainChain{}),
				QueryExecutor:       &mockQueryExecutorGetMempoolTransactionsFail{},
				MempoolCacheStorage: &mockCacheStorageGetAllItemsError{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mg := &MempoolService{
				QueryExecutor:       tt.fields.QueryExecutor,
				MempoolQuery:        tt.fields.MempoolQuery,
				MempoolCacheStorage: tt.fields.MempoolCacheStorage,
			}
			got, err := mg.GetMempoolTransactions()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMempoolTransactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("GetMempoolTransactions() error different length")
			}
		})
	}
}
