package service

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"

	"github.com/DATA-DOG/go-sqlmock"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
)

type (
	mockNaiQueryExecutorSuccess struct {
		query.Executor
	}
	mockNaiQueryExecutorFailBeginTx struct {
		query.Executor
	}
	mockNaiQueryExecutorFailRollbackTx struct {
		query.Executor
	}
	mockNaiQueryExecutorFailCommitTx struct {
		query.Executor
	}
	mockNaiQueryExecutorFailExecuteSelect struct {
		query.Executor
	}
	mockNaiQueryExecutorFailExecuteTransactions struct {
		query.Executor
	}
	mockNaiQueryExecutorFailExecuteSelectRow struct {
		query.Executor
	}
	mockNaiQueryExecutorFailScan struct {
		query.Executor
	}
	mockNaiQueryBuildFailed struct {
		query.Executor
	}
	mockNaiStorageSuccess struct {
		storage.NodeAddressInfoStorage
	}
	mockNaiStorageFail struct {
		storage.NodeAddressInfoStorage
	}
	mockNaiStorageFailSetItem struct {
		storage.NodeAddressInfoStorage
	}
	mockNaiStorageFailRemoveItem struct {
		storage.NodeAddressInfoStorage
	}
	mockNaiStorageEmpty struct {
		storage.NodeAddressInfoStorage
	}
	mockActiveNaiStorageSuccess struct {
		storage.NodeRegistryCacheStorage
	}
	mockActiveNaiStorageFail struct {
		storage.NodeRegistryCacheStorage
	}
	mockMainBlockStateStorageSuccess struct {
		storage.CacheStorageInterface
	}
	mockMainBlockStateStorageFail struct {
		storage.CacheStorageInterface
	}
	mockNaiSignature struct {
		crypto.Signature
		success bool
	}
)

func (*mockNaiStorageSuccess) SetItem(_, item interface{}) error {
	return nil
}
func (*mockNaiStorageSuccess) GetItem(_, item interface{}) error {
	assert := item.(*[]*model.NodeAddressInfo)
	*assert = []*model.NodeAddressInfo{
		naiNode1,
	}
	return nil
}
func (*mockNaiStorageSuccess) GetAllItems(items interface{}) error {
	nai, ok := items.(*[]*model.NodeAddressInfo)
	if !ok {
		return errors.New("wrongtype")
	}
	*nai = append(*nai, naiNode1)
	return nil
}
func (*mockNaiStorageSuccess) Begin() error {
	return nil
}
func (*mockNaiStorageSuccess) Commit() error {
	return nil
}
func (*mockNaiStorageSuccess) ClearCache() error {
	return nil
}
func (*mockNaiStorageSuccess) Rollback() error {
	return nil
}

func (*mockNaiStorageFail) GetItem(interface{}, interface{}) error {
	return errors.New("error")
}
func (*mockNaiStorageFail) GetAllItems(interface{}) error {
	return errors.New("error")
}
func (*mockNaiStorageFail) Begin() error {
	return errors.New("error")
}
func (*mockNaiStorageFail) Commit() error {
	return errors.New("error")
}
func (*mockNaiStorageFail) ClearCache() error {
	return errors.New("error")
}
func (*mockNaiStorageFail) Rollback() error {
	return errors.New("error")
}
func (*mockNaiStorageFail) SetItem(_, item interface{}) error {
	return errors.New("error")
}
func (*mockNaiStorageFail) RemoveItem(idx interface{}) error {
	return nil
}

func (*mockNaiStorageFailSetItem) SetItem(_, item interface{}) error {
	return errors.New("error")
}

func (*mockNaiStorageEmpty) GetItem(interface{}, interface{}) error {
	return nil
}
func (*mockNaiStorageEmpty) GetAllItems(interface{}) error {
	return nil
}

func (*mockActiveNaiStorageSuccess) TxSetItem(idx, item interface{}) error {
	return nil
}

func (*mockActiveNaiStorageSuccess) TxSetItems(items interface{}) error {
	return nil
}

func (*mockActiveNaiStorageSuccess) TxRemoveItem(idx interface{}) error {
	return nil
}

func (*mockActiveNaiStorageSuccess) RemoveItem(idx interface{}) error {
	return nil
}

func (*mockActiveNaiStorageSuccess) GetItem(_, item interface{}) error {
	return nil
}

func (*mockActiveNaiStorageSuccess) GetAllItems(items interface{}) error {
	return nil
}

func (*mockActiveNaiStorageSuccess) SetItem(idx, item interface{}) error {
	return nil
}

func (*mockActiveNaiStorageSuccess) SetItems(items interface{}) error {
	return nil
}

func (*mockMainBlockStateStorageSuccess) GetItem(_, item interface{}) error {
	return nil
}
func (*mockMainBlockStateStorageFail) GetItem(interface{}, interface{}) error {
	return errors.New("error")
}

func (*mockActiveNaiStorageFail) TxSetItem(idx, item interface{}) error {
	return errors.New("error")
}
func (*mockActiveNaiStorageFail) TxSetItems(items interface{}) error {
	return errors.New("error")
}
func (*mockActiveNaiStorageFail) TxRemoveItem(idx interface{}) error {
	return errors.New("error")
}
func (*mockActiveNaiStorageFail) RemoveItem(idx interface{}) error {
	return errors.New("error")
}
func (*mockActiveNaiStorageFail) GetItem(idx, item interface{}) error {
	return errors.New("error")
}
func (*mockActiveNaiStorageFail) GetAllItems(items interface{}) error {
	return errors.New("error")
}
func (*mockActiveNaiStorageFail) SetItem(idx, item interface{}) error {
	return errors.New("error")
}
func (*mockActiveNaiStorageFail) SetItems(items interface{}) error {
	return errors.New("error")
}

func (*mockNaiStorageFailRemoveItem) RemoveItem(idx interface{}) error {
	return errors.New("error")
}

func (*mockNaiQueryExecutorSuccess) BeginTx() error {
	return nil
}
func (*mockNaiQueryExecutorSuccess) RollbackTx() error {
	return nil
}
func (*mockNaiQueryExecutorSuccess) CommitTx() error {
	return nil
}

func (*mockNaiQueryExecutorFailBeginTx) BeginTx() error {
	return errors.New("error")
}

func (*mockNaiQueryExecutorFailRollbackTx) BeginTx() error {
	return nil
}
func (*mockNaiQueryExecutorFailRollbackTx) RollbackTx() error {
	return errors.New("error")
}

func (*mockNaiQueryExecutorFailCommitTx) BeginTx() error {
	return nil
}
func (*mockNaiQueryExecutorFailCommitTx) RollbackTx() error {
	return nil
}
func (*mockNaiQueryExecutorFailCommitTx) CommitTx() error {
	return errors.New("error")
}

func (*mockNaiQueryExecutorSuccess) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch query {
	case "INSERT OR REPLACE INTO node_address_info (node_id, address, port, block_height, block_hash, " +
		"signature, status) VALUES(? , ? , ? , ? , ? , ? , ? )":
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
			"node_id",
			"address",
			"port",
			"block_height",
			"block_hash",
			"signature",
			"status",
		}).AddRow(
			naiNode1.NodeID,
			naiNode1.Address,
			naiNode1.Port,
			naiNode1.BlockHeight,
			naiNode1.BlockHash,
			naiNode1.Signature,
			naiNode1.Status,
		))
	case "UPDATE node_address_info SET address = ?, port = ?, block_height = ?, block_hash = ?, signature = ?, status = ? WHERE node_id = ?":
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
			"node_id",
			"address",
			"port",
			"block_height",
			"block_hash",
			"signature",
			"status",
		}).AddRow(
			naiNode1.NodeID,
			naiNode1.Address,
			naiNode1.Port,
			naiNode1.BlockHeight,
			naiNode1.BlockHash,
			naiNode1.Signature,
			naiNode1.Status,
		))
	case "SELECT node_id, address, port, block_height, block_hash, signature, status FROM node_address_info ORDER BY node_id, status ASC":
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
			"node_id",
			"address",
			"port",
			"block_height",
			"block_hash",
			"signature",
			"status",
		}).AddRow(
			naiNode1.NodeID,
			naiNode1.Address,
			naiNode1.Port,
			naiNode1.BlockHeight,
			naiNode1.BlockHash,
			naiNode1.Signature,
			naiNode1.Status,
		))
	case "DELETE FROM node_address_info WHERE node_id = ? AND status IN":
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
			"node_id",
			"address",
			"port",
			"block_height",
			"block_hash",
			"signature",
			"status",
		}).AddRow(
			naiNode1.NodeID,
			naiNode1.Address,
			naiNode1.Port,
			naiNode1.BlockHeight,
			naiNode1.BlockHash,
			naiNode1.Signature,
			naiNode1.Status,
		))
	default:
		return nil, errors.New("MockErr")
	}
	return db.Query("")
}
func (*mockNaiQueryExecutorSuccess) ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	var blockFields = []string{
		"height",
		"id",
		"block_hash",
		"previous_block_hash",
		"timestamp",
		"block_seed",
		"block_signature",
		"cumulative_difficulty",
		"payload_length",
		"payload_hash",
		"blocksmith_public_key",
		"total_amount",
		"total_fee",
		"total_coinbase",
		"version",
		"merkle_root",
		"merkle_tree",
		"reference_block_height",
	}
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows(blockFields).AddRow(
			10,
			1,
			mockBlockData.BlockHash,
			mockBlockData.PreviousBlockHash,
			mockBlockData.Timestamp,
			mockBlockData.BlockSeed,
			mockBlockData.BlockSignature,
			mockBlockData.CumulativeDifficulty,
			mockBlockData.PayloadLength,
			mockBlockData.PayloadHash,
			mockBlockData.BlocksmithPublicKey,
			mockBlockData.TotalAmount,
			mockBlockData.TotalFee,
			mockBlockData.TotalCoinBase,
			mockBlockData.Version,
			mockBlockData.MerkleRoot,
			mockBlockData.MerkleTree,
			mockBlockData.ReferenceBlockHeight,
		))
	return db.QueryRow(query), nil
}

func (*mockNaiQueryExecutorSuccess) ExecuteTransaction(string, ...interface{}) error {
	return nil
}
func (*mockNaiQueryExecutorSuccess) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}

func (*mockNaiQueryExecutorFailExecuteSelect) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("MockErr")
}

func (*mockNaiQueryExecutorFailExecuteSelectRow) ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error) {
	return nil, errors.New("MockErr")
}
func (*mockNaiQueryExecutorFailScan) ExecuteSelectRow(query string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	var blockFields = []string{
		"id",
		"block_hash",
		"previous_block_hash",
		"timestamp",
		"block_seed",
		"block_signature",
		"cumulative_difficulty",
		"payload_length",
		"payload_hash",
		"blocksmith_public_key",
		"total_amount",
		"total_fee",
		"total_coinbase",
		"version",
		"merkle_root",
		"merkle_tree",
		"reference_block_height",
	}
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows(blockFields).AddRow(
			1,
			mockBlockData.BlockHash,
			mockBlockData.PreviousBlockHash,
			mockBlockData.Timestamp,
			mockBlockData.BlockSeed,
			mockBlockData.BlockSignature,
			mockBlockData.CumulativeDifficulty,
			mockBlockData.PayloadLength,
			mockBlockData.PayloadHash,
			mockBlockData.BlocksmithPublicKey,
			mockBlockData.TotalAmount,
			mockBlockData.TotalFee,
			mockBlockData.TotalCoinBase,
			mockBlockData.Version,
			mockBlockData.MerkleRoot,
			mockBlockData.MerkleTree,
			mockBlockData.ReferenceBlockHeight,
		))
	return db.QueryRow(query), nil
}
func (*mockNaiQueryExecutorFailScan) Scan(block *model.Block, row *sql.Row) error {
	return errors.New("error")
}

func (*mockNaiQueryExecutorFailExecuteTransactions) BeginTx() error {
	return nil
}
func (*mockNaiQueryExecutorFailExecuteTransactions) RollbackTx() error {
	return errors.New("error")
}
func (*mockNaiQueryExecutorFailExecuteTransactions) ExecuteTransactions([][]interface{}) error {
	return errors.New("MockErr")
}

func (*mockNaiQueryBuildFailed) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch query {
	case "SELECT node_id, address, port, block_height, block_hash, signature, status FROM node_address_info ORDER BY node_id, status ASC":
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
			"node_id",
			"address",
			"port",
			"block_height",
			"block_hash",
			"signature",
			"status",
		}).AddRow(
			nil,
			naiNode1.Address,
			naiNode1.Port,
			naiNode1.BlockHeight,
			naiNode1.BlockHash,
			naiNode1.Signature,
			naiNode1.Status,
		))
	default:
		return nil, errors.New("error")
	}
	return db.Query("")
}
func (*mockNaiQueryBuildFailed) BuildModel([]*model.NodeAddressInfo, *sql.Rows) ([]*model.NodeAddressInfo, error) {
	return nil, errors.New("mockedError")
}

func (*mockNaiQueryExecutorFailRollbackTx) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch query {
	case "INSERT OR REPLACE INTO node_address_info (node_id, address, port, block_height, block_hash, " +
		"signature, status) VALUES(? , ? , ? , ? , ? , ? , ? )":
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
			"node_id",
			"address",
			"port",
			"block_height",
			"block_hash",
			"signature",
			"status",
		}).AddRow(
			naiNode1.NodeID,
			naiNode1.Address,
			naiNode1.Port,
			naiNode1.BlockHeight,
			naiNode1.BlockHash,
			naiNode1.Signature,
			naiNode1.Status,
		))
	case "UPDATE node_address_info SET address = ?, port = ?, block_height = ?, block_hash = ?, signature = ?, status = ? WHERE node_id =":
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
			"node_id",
			"address",
			"port",
			"block_height",
			"block_hash",
			"signature",
			"status",
		}).AddRow(
			naiNode1.NodeID,
			naiNode1.Address,
			naiNode1.Port,
			naiNode1.BlockHeight,
			naiNode1.BlockHash,
			naiNode1.Signature,
			naiNode1.Status,
		))
	case "SELECT node_id, address, port, block_height, block_hash, signature, status FROM node_address_info ORDER BY node_id, status ASC":
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
			"node_id",
			"address",
			"port",
			"block_height",
			"block_hash",
			"signature",
			"status",
		}).AddRow(
			naiNode1.NodeID,
			naiNode1.Address,
			naiNode1.Port,
			naiNode1.BlockHeight,
			naiNode1.BlockHash,
			naiNode1.Signature,
			naiNode1.Status,
		))
	case "DELETE FROM node_address_info WHERE node_id = ? AND status IN":
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
			"node_id",
			"address",
			"port",
			"block_height",
			"block_hash",
			"signature",
			"status",
		}).AddRow(
			naiNode1.NodeID,
			naiNode1.Address,
			naiNode1.Port,
			naiNode1.BlockHeight,
			naiNode1.BlockHash,
			naiNode1.Signature,
			naiNode1.Status,
		))
	default:
		return nil, errors.New("MockErr")
	}
	return db.Query("")
}
func (*mockNaiQueryExecutorFailExecuteSelect) ExecuteTransaction(string, ...interface{}) error {
	return errors.New("error")
}
func (*mockNaiQueryExecutorFailExecuteSelect) BeginTx() error {
	return nil
}
func (*mockNaiQueryExecutorFailExecuteSelect) RollbackTx() error {
	return errors.New("error")
}

func (*mockNaiQueryExecutorFailRollbackTx) ExecuteTransaction(string, ...interface{}) error {
	return errors.New("error")
}
func (*mockNaiQueryExecutorFailRollbackTx) ExecuteTransactions(queries [][]interface{}) error {
	return errors.New("error")
}

func (*mockNaiQueryExecutorFailCommitTx) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if query == "INSERT OR REPLACE INTO node_address_info (node_id, address, port, block_height, block_hash, "+
		"signature, status) VALUES(? , ? , ? , ? , ? , ? , ? )" {
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{
			"node_id",
			"address",
			"port",
			"block_height",
			"block_hash",
			"signature",
			"status",
		}).AddRow(
			naiNode1.NodeID,
			naiNode1.Address,
			naiNode1.Port,
			naiNode1.BlockHeight,
			naiNode1.BlockHash,
			naiNode1.Signature,
			naiNode1.Status,
		))
	}
	return db.Query("")
}
func (*mockNaiQueryExecutorFailCommitTx) ExecuteTransaction(string, ...interface{}) error {
	return nil
}
func (*mockNaiQueryExecutorFailCommitTx) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}

var (
	naiNode1 = &model.NodeAddressInfo{
		NodeID:      int64(111),
		Address:     "127.0.0.1",
		Port:        uint32(3000),
		BlockHeight: uint32(10),
		BlockHash:   []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		Signature: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
			1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		Status: model.NodeAddressStatus_NodeAddressConfirmed,
	}
	naiNode2 = &model.NodeAddressInfo{
		NodeID:      int64(111),
		Address:     "127.0.0.1",
		Port:        uint32(3000),
		BlockHeight: uint32(10),
		BlockHash:   []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		Signature: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
			1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		Status: model.NodeAddressStatus_NodeAddressPending,
	}
)

func TestNewNodeAddressInfoService(t *testing.T) {
	type args struct {
		executor                 query.ExecutorInterface
		nodeAddressInfoQuery     query.NodeAddressInfoQueryInterface
		nodeRegistrationQuery    query.NodeRegistrationQueryInterface
		blockQuery               query.BlockQueryInterface
		signature                crypto.SignatureInterface
		nodeAddressesInfoStorage storage.CacheStorageInterface
		mainBlockStateStorage    storage.CacheStorageInterface
		activeNodeRegistryCache  storage.CacheStorageInterface
		logger                   *log.Logger
	}
	tests := []struct {
		name string
		args args
		want *NodeAddressInfoService
	}{
		{
			name: "NewNodeAddressInfoService:Success",
			args: args{
				executor:                 nil,
				nodeAddressInfoQuery:     nil,
				nodeRegistrationQuery:    nil,
				blockQuery:               nil,
				signature:                nil,
				nodeAddressesInfoStorage: nil,
				mainBlockStateStorage:    nil,
				activeNodeRegistryCache:  nil,
				logger:                   nil,
			},
			want: &NodeAddressInfoService{
				QueryExecutor:           nil,
				NodeAddressInfoQuery:    nil,
				NodeRegistrationQuery:   nil,
				BlockQuery:              nil,
				Signature:               nil,
				NodeAddressInfoStorage:  nil,
				MainBlockStateStorage:   nil,
				ActiveNodeRegistryCache: nil,
				Logger:                  nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewNodeAddressInfoService(tt.args.executor, tt.args.nodeAddressInfoQuery, tt.args.nodeRegistrationQuery,
				tt.args.blockQuery, tt.args.signature, tt.args.nodeAddressesInfoStorage, tt.args.mainBlockStateStorage,
				tt.args.activeNodeRegistryCache, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewNodeAddressInfoService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeAddressInfoService_BeginCacheTransaction(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		NodeAddressInfoQuery    query.NodeAddressInfoQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		Signature               crypto.SignatureInterface
		NodeAddressInfoStorage  storage.CacheStorageInterface
		MainBlockStateStorage   storage.CacheStorageInterface
		ActiveNodeRegistryCache storage.CacheStorageInterface
		Logger                  *log.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "BeginCache:Success",
			fields: fields{
				NodeAddressInfoStorage: &mockNaiStorageSuccess{},
			},
			wantErr: false,
		},
		{
			name: "BeginCache:Fail",
			fields: fields{
				NodeAddressInfoStorage: &mockNaiStorageFail{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nru := &NodeAddressInfoService{
				QueryExecutor:           tt.fields.QueryExecutor,
				NodeAddressInfoQuery:    tt.fields.NodeAddressInfoQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				Signature:               tt.fields.Signature,
				NodeAddressInfoStorage:  tt.fields.NodeAddressInfoStorage,
				MainBlockStateStorage:   tt.fields.MainBlockStateStorage,
				ActiveNodeRegistryCache: tt.fields.ActiveNodeRegistryCache,
				Logger:                  tt.fields.Logger,
			}
			if err := nru.BeginCacheTransaction(); (err != nil) != tt.wantErr {
				t.Errorf("BeginCacheTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeAddressInfoService_ClearUpdateNodeAddressInfoCache(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		NodeAddressInfoQuery    query.NodeAddressInfoQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		Signature               crypto.SignatureInterface
		NodeAddressInfoStorage  storage.CacheStorageInterface
		MainBlockStateStorage   storage.CacheStorageInterface
		ActiveNodeRegistryCache storage.CacheStorageInterface
		Logger                  *log.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "ClearUpdateNodeAddressInfoCache:Success",
			fields: fields{
				QueryExecutor:          &mockNaiQueryExecutorSuccess{},
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage: &mockNaiStorageSuccess{},
				Logger:                 log.New(),
			},
			wantErr: false,
		},
		{
			name: "ClearUpdateNodeAddressInfoCache:FailExecuteSelect",
			fields: fields{
				QueryExecutor:          &mockNaiQueryExecutorFailExecuteSelect{},
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage: &mockNaiStorageSuccess{},
				Logger:                 log.New(),
			},
			wantErr: true,
		},
		{
			name: "ClearUpdateNodeAddressInfoCache:FailBuildModel",
			fields: fields{
				QueryExecutor:          &mockNaiQueryBuildFailed{},
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage: &mockNaiStorageFail{},
				Logger:                 log.New(),
			},
			wantErr: true,
		},
		{
			name: "ClearUpdateNodeAddressInfoCache:FailClearCache",
			fields: fields{
				QueryExecutor:          &mockNaiQueryExecutorSuccess{},
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage: &mockNaiStorageFailSetItem{},
				Logger:                 log.New(),
			},
			wantErr: true,
		},
		{
			name: "ClearUpdateNodeAddressInfoCache:FailSetItem",
			fields: fields{
				QueryExecutor:          &mockNaiQueryExecutorSuccess{},
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage: &mockNaiStorageFail{},
				Logger:                 log.New(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nru := &NodeAddressInfoService{
				QueryExecutor:           tt.fields.QueryExecutor,
				NodeAddressInfoQuery:    tt.fields.NodeAddressInfoQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				Signature:               tt.fields.Signature,
				NodeAddressInfoStorage:  tt.fields.NodeAddressInfoStorage,
				MainBlockStateStorage:   tt.fields.MainBlockStateStorage,
				ActiveNodeRegistryCache: tt.fields.ActiveNodeRegistryCache,
				Logger:                  tt.fields.Logger,
			}
			if err := nru.ClearUpdateNodeAddressInfoCache(); (err != nil) != tt.wantErr {
				t.Errorf("ClearUpdateNodeAddressInfoCache() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeAddressInfoService_CommitCacheTransaction(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		NodeAddressInfoQuery    query.NodeAddressInfoQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		Signature               crypto.SignatureInterface
		NodeAddressInfoStorage  storage.CacheStorageInterface
		MainBlockStateStorage   storage.CacheStorageInterface
		ActiveNodeRegistryCache storage.CacheStorageInterface
		Logger                  *log.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "CommitCacheTransaction:Success",
			fields: fields{
				NodeAddressInfoStorage: &mockNaiStorageSuccess{},
			},
			wantErr: false,
		},
		{
			name: "CommitCacheTransaction:Failed",
			fields: fields{
				NodeAddressInfoStorage: &mockNaiStorageFail{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nru := &NodeAddressInfoService{
				QueryExecutor:           tt.fields.QueryExecutor,
				NodeAddressInfoQuery:    tt.fields.NodeAddressInfoQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				Signature:               tt.fields.Signature,
				NodeAddressInfoStorage:  tt.fields.NodeAddressInfoStorage,
				MainBlockStateStorage:   tt.fields.MainBlockStateStorage,
				ActiveNodeRegistryCache: tt.fields.ActiveNodeRegistryCache,
				Logger:                  tt.fields.Logger,
			}
			if err := nru.CommitCacheTransaction(); (err != nil) != tt.wantErr {
				t.Errorf("CommitCacheTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeAddressInfoService_ConfirmNodeAddressInfo(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		NodeAddressInfoQuery    query.NodeAddressInfoQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		Signature               crypto.SignatureInterface
		NodeAddressInfoStorage  storage.CacheStorageInterface
		MainBlockStateStorage   storage.CacheStorageInterface
		ActiveNodeRegistryCache storage.CacheStorageInterface
		Logger                  *log.Logger
	}
	type args struct {
		pendingNodeAddressInfo *model.NodeAddressInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ConfirmNodeAddressInfo:Success",
			fields: fields{
				QueryExecutor:          &mockNaiQueryExecutorSuccess{},
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage: &mockNaiStorageSuccess{},
			},
			args: args{
				pendingNodeAddressInfo: naiNode2,
			},
			wantErr: false,
		},
		{
			name: "ConfirmNodeAddressInfo:FailBeginTx",
			fields: fields{
				QueryExecutor:          &mockNaiQueryExecutorFailBeginTx{},
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage: &mockNaiStorageFail{},
			},
			args: args{
				pendingNodeAddressInfo: naiNode2,
			},
			wantErr: true,
		},
		{
			name: "ConfirmNodeAddressInfo:FailExecuteTransaction",
			fields: fields{
				QueryExecutor:          &mockNaiQueryExecutorFailExecuteTransactions{},
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage: &mockNaiStorageFail{},
			},
			args: args{
				pendingNodeAddressInfo: naiNode2,
			},
			wantErr: true,
		},
		{
			name: "ConfirmNodeAddressInfo:FailCommitTx",
			fields: fields{
				QueryExecutor:          &mockNaiQueryExecutorFailCommitTx{},
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage: &mockNaiStorageFail{},
			},
			args: args{
				pendingNodeAddressInfo: naiNode2,
			},
			wantErr: true,
		},
		{
			name: "ConfirmNodeAddressInfo:FailRemoveItem",
			fields: fields{
				QueryExecutor:          &mockNaiQueryExecutorSuccess{},
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage: &mockNaiStorageFailRemoveItem{},
			},
			args: args{
				pendingNodeAddressInfo: naiNode2,
			},
			wantErr: true,
		},
		{
			name: "ConfirmNodeAddressInfo:FailSetItem",
			fields: fields{
				QueryExecutor:          &mockNaiQueryExecutorSuccess{},
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage: &mockNaiStorageFailSetItem{},
			},
			args: args{
				pendingNodeAddressInfo: naiNode2,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nru := &NodeAddressInfoService{
				QueryExecutor:           tt.fields.QueryExecutor,
				NodeAddressInfoQuery:    tt.fields.NodeAddressInfoQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				Signature:               tt.fields.Signature,
				NodeAddressInfoStorage:  tt.fields.NodeAddressInfoStorage,
				MainBlockStateStorage:   tt.fields.MainBlockStateStorage,
				ActiveNodeRegistryCache: tt.fields.ActiveNodeRegistryCache,
				Logger:                  tt.fields.Logger,
			}
			if err := nru.ConfirmNodeAddressInfo(tt.args.pendingNodeAddressInfo); (err != nil) != tt.wantErr {
				t.Errorf("ConfirmNodeAddressInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeAddressInfoService_CountNodesAddressByStatus(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		NodeAddressInfoQuery    query.NodeAddressInfoQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		Signature               crypto.SignatureInterface
		NodeAddressInfoStorage  storage.CacheStorageInterface
		MainBlockStateStorage   storage.CacheStorageInterface
		ActiveNodeRegistryCache storage.CacheStorageInterface
		Logger                  *log.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[model.NodeAddressStatus]int
		wantErr bool
	}{
		{
			name: "CountNodesAddressByStatus:Success",
			fields: fields{
				NodeAddressInfoStorage: &mockNaiStorageSuccess{},
			},
			want: map[model.NodeAddressStatus]int{
				model.NodeAddressStatus_NodeAddressConfirmed: 1,
			},
			wantErr: false,
		},
		{
			name: "CountNodesAddressByStatus:Error",
			fields: fields{
				NodeAddressInfoStorage: &mockNaiStorageFail{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nru := &NodeAddressInfoService{
				QueryExecutor:           tt.fields.QueryExecutor,
				NodeAddressInfoQuery:    tt.fields.NodeAddressInfoQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				Signature:               tt.fields.Signature,
				NodeAddressInfoStorage:  tt.fields.NodeAddressInfoStorage,
				MainBlockStateStorage:   tt.fields.MainBlockStateStorage,
				ActiveNodeRegistryCache: tt.fields.ActiveNodeRegistryCache,
				Logger:                  tt.fields.Logger,
			}
			got, err := nru.CountNodesAddressByStatus()
			if (err != nil) != tt.wantErr {
				t.Errorf("CountNodesAddressByStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CountNodesAddressByStatus() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeAddressInfoService_CountRegisteredNodeAddressWithAddressInfo(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		NodeAddressInfoQuery    query.NodeAddressInfoQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		Signature               crypto.SignatureInterface
		NodeAddressInfoStorage  storage.CacheStorageInterface
		MainBlockStateStorage   storage.CacheStorageInterface
		ActiveNodeRegistryCache storage.CacheStorageInterface
		Logger                  *log.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		want    int
		wantErr bool
	}{
		{
			name: "CountRegisteredNodeAddressWithAddressInfo:Success",
			fields: fields{
				QueryExecutor:          &mockNaiQueryExecutorSuccess{},
				NodeAddressInfoStorage: &mockNaiStorageSuccess{},
				NodeRegistrationQuery:  query.NewNodeRegistrationQuery(),
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
			},
			want:    0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nru := &NodeAddressInfoService{
				QueryExecutor:           tt.fields.QueryExecutor,
				NodeAddressInfoQuery:    tt.fields.NodeAddressInfoQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				Signature:               tt.fields.Signature,
				NodeAddressInfoStorage:  tt.fields.NodeAddressInfoStorage,
				MainBlockStateStorage:   tt.fields.MainBlockStateStorage,
				ActiveNodeRegistryCache: tt.fields.ActiveNodeRegistryCache,
				Logger:                  tt.fields.Logger,
			}
			got, err := nru.CountRegistredNodeAddressWithAddressInfo()
			if (err != nil) != tt.wantErr {
				t.Errorf("CountRegistredNodeAddressWithAddressInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CountRegistredNodeAddressWithAddressInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeAddressInfoService_DeleteNodeAddressInfoByNodeIDInDBTx(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		NodeAddressInfoQuery    query.NodeAddressInfoQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		Signature               crypto.SignatureInterface
		NodeAddressInfoStorage  storage.CacheStorageInterface
		MainBlockStateStorage   storage.CacheStorageInterface
		ActiveNodeRegistryCache storage.CacheStorageInterface
		Logger                  *log.Logger
	}
	type args struct {
		nodeID int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "DeleteNodeAddressInfoByNodeIDInDBTx:Success",
			fields: fields{
				QueryExecutor:          &mockNaiQueryExecutorSuccess{},
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage: &mockNaiStorageSuccess{},
			},
			args: args{
				nodeID: 111,
			},
			wantErr: false,
		},
		{
			name: "DeleteNodeAddressInfoByNodeIDInDBTx:Fail",
			fields: fields{
				QueryExecutor:          &mockNaiQueryBuildFailed{},
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage: &mockNaiStorageFail{},
			},
			args: args{
				nodeID: 111,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nru := &NodeAddressInfoService{
				QueryExecutor:           tt.fields.QueryExecutor,
				NodeAddressInfoQuery:    tt.fields.NodeAddressInfoQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				Signature:               tt.fields.Signature,
				NodeAddressInfoStorage:  tt.fields.NodeAddressInfoStorage,
				MainBlockStateStorage:   tt.fields.MainBlockStateStorage,
				ActiveNodeRegistryCache: tt.fields.ActiveNodeRegistryCache,
				Logger:                  tt.fields.Logger,
			}
			if err := nru.DeleteNodeAddressInfoByNodeIDInDBTx(tt.args.nodeID); (err != nil) != tt.wantErr {
				t.Errorf("DeleteNodeAddressInfoByNodeIDInDBTx() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeAddressInfoService_DeletePendingNodeAddressInfo(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		NodeAddressInfoQuery    query.NodeAddressInfoQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		Signature               crypto.SignatureInterface
		NodeAddressInfoStorage  storage.CacheStorageInterface
		MainBlockStateStorage   storage.CacheStorageInterface
		ActiveNodeRegistryCache storage.CacheStorageInterface
		Logger                  *log.Logger
	}
	type args struct {
		nodeID int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "DeletePendingNodeAddressInfo:Success",
			fields: fields{
				QueryExecutor:          &mockNaiQueryExecutorSuccess{},
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage: &mockNaiStorageSuccess{},
			},
			args: args{
				nodeID: 111,
			},
			wantErr: false,
		},
		{
			name: "DeletePendingNodeAddressInfo:FailBeginTx",
			fields: fields{
				QueryExecutor:          &mockNaiQueryExecutorFailBeginTx{},
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage: &mockNaiStorageFail{},
			},
			args: args{
				nodeID: 111,
			},
			wantErr: true,
		},
		{
			name: "DeletePendingNodeAddressInfo:FailRollbackTx",
			fields: fields{
				QueryExecutor:          &mockNaiQueryExecutorFailExecuteSelect{},
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage: &mockNaiStorageFail{},
				Logger:                 log.New(),
			},
			args: args{
				nodeID: 111,
			},
			wantErr: true,
		},
		{
			name: "DeletePendingNodeAddressInfo:FailCommitTx",
			fields: fields{
				QueryExecutor:          &mockNaiQueryExecutorFailCommitTx{},
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage: &mockNaiStorageFail{},
				Logger:                 log.New(),
			},
			args: args{
				nodeID: 111,
			},
			wantErr: true,
		},
		{
			name: "DeletePendingNodeAddressInfo:FailRemoveItem",
			fields: fields{
				QueryExecutor:          &mockNaiQueryExecutorSuccess{},
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage: &mockNaiStorageFailRemoveItem{},
			},
			args: args{
				nodeID: 111,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nru := &NodeAddressInfoService{
				QueryExecutor:           tt.fields.QueryExecutor,
				NodeAddressInfoQuery:    tt.fields.NodeAddressInfoQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				Signature:               tt.fields.Signature,
				NodeAddressInfoStorage:  tt.fields.NodeAddressInfoStorage,
				MainBlockStateStorage:   tt.fields.MainBlockStateStorage,
				ActiveNodeRegistryCache: tt.fields.ActiveNodeRegistryCache,
				Logger:                  tt.fields.Logger,
			}
			if err := nru.DeletePendingNodeAddressInfo(tt.args.nodeID); (err != nil) != tt.wantErr {
				t.Errorf("DeletePendingNodeAddressInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeAddressInfoService_GenerateNodeAddressInfo(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		NodeAddressInfoQuery    query.NodeAddressInfoQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		Signature               crypto.SignatureInterface
		NodeAddressInfoStorage  storage.CacheStorageInterface
		MainBlockStateStorage   storage.CacheStorageInterface
		ActiveNodeRegistryCache storage.CacheStorageInterface
		Logger                  *log.Logger
	}
	type args struct {
		nodeID           int64
		nodeAddress      string
		port             uint32
		nodeSecretPhrase string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.NodeAddressInfo
		wantErr bool
	}{
		{
			name: "GenerateNodeAddressInfo:Success",
			fields: fields{
				QueryExecutor:           &mockNaiQueryExecutorSuccess{},
				NodeAddressInfoQuery:    query.NewNodeAddressInfoQuery(),
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				BlockQuery:              query.NewBlockQuery(&chaintype.MainChain{}),
				NodeAddressInfoStorage:  &mockNaiStorageSuccess{},
				MainBlockStateStorage:   &mockMainBlockStateStorageSuccess{},
				ActiveNodeRegistryCache: &mockActiveNaiStorageSuccess{},
				Signature:               &mockNaiSignature{success: true},
				Logger:                  log.New(),
			},
			args: args{
				nodeID:           111,
				nodeAddress:      "127.0.0.1",
				port:             3000,
				nodeSecretPhrase: "test",
			},
			want: &model.NodeAddressInfo{
				NodeID:      111,
				Address:     "127.0.0.1",
				Port:        3000,
				BlockHeight: 10,
				BlockHash:   make([]byte, 32),
				Signature: []byte{144, 164, 51, 115, 40, 63, 10, 163, 38, 202, 110, 18, 65, 35, 139, 233, 226, 215,
					176, 164, 153, 180, 239, 222, 252, 63, 94, 168, 201, 59, 143, 152, 192, 142, 243, 6, 43, 60, 129, 138,
					29, 188, 128, 52, 33, 209, 241, 113, 119, 95, 21, 56, 162, 192, 111, 76, 50, 163, 20, 84, 72,
					141, 232, 8},
			},
			wantErr: false,
		},
		{
			name: "GenerateNodeAddressInfo:FailGetItemMainBlock",
			fields: fields{
				MainBlockStateStorage: &mockMainBlockStateStorageFail{},
			},
			args: args{
				nodeID:           111,
				nodeAddress:      "127.0.0.1",
				port:             3000,
				nodeSecretPhrase: "test",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GenerateNodeAddressInfo:FailExecuteSelectRow",
			fields: fields{
				QueryExecutor:         &mockNaiQueryExecutorFailExecuteSelectRow{},
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				MainBlockStateStorage: &mockMainBlockStateStorageSuccess{},
			},
			args: args{
				nodeID:           111,
				nodeAddress:      "127.0.0.1",
				port:             3000,
				nodeSecretPhrase: "test",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GenerateNodeAddressInfo:FailScan",
			fields: fields{
				QueryExecutor:         &mockNaiQueryExecutorFailScan{},
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				MainBlockStateStorage: &mockMainBlockStateStorageSuccess{},
			},
			args: args{
				nodeID:           111,
				nodeAddress:      "127.0.0.1",
				port:             3000,
				nodeSecretPhrase: "test",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nru := &NodeAddressInfoService{
				QueryExecutor:           tt.fields.QueryExecutor,
				NodeAddressInfoQuery:    tt.fields.NodeAddressInfoQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				Signature:               tt.fields.Signature,
				NodeAddressInfoStorage:  tt.fields.NodeAddressInfoStorage,
				MainBlockStateStorage:   tt.fields.MainBlockStateStorage,
				ActiveNodeRegistryCache: tt.fields.ActiveNodeRegistryCache,
				Logger:                  tt.fields.Logger,
			}
			got, err := nru.GenerateNodeAddressInfo(tt.args.nodeID, tt.args.nodeAddress, tt.args.port, tt.args.nodeSecretPhrase)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateNodeAddressInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateNodeAddressInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeAddressInfoService_GetAddressInfoByAddressPort(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		NodeAddressInfoQuery    query.NodeAddressInfoQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		Signature               crypto.SignatureInterface
		NodeAddressInfoStorage  storage.CacheStorageInterface
		MainBlockStateStorage   storage.CacheStorageInterface
		ActiveNodeRegistryCache storage.CacheStorageInterface
		Logger                  *log.Logger
	}
	type args struct {
		address             string
		port                uint32
		nodeAddressStatuses []model.NodeAddressStatus
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.NodeAddressInfo
		wantErr bool
	}{
		{
			name: "GetAddressInfoByAddressPort:Success",
			fields: fields{
				NodeAddressInfoStorage: &mockNaiStorageSuccess{},
			},
			args: args{
				address: "127.0.0.1",
				port:    3000,
				nodeAddressStatuses: []model.NodeAddressStatus{
					model.NodeAddressStatus_NodeAddressConfirmed,
				},
			},
			want: []*model.NodeAddressInfo{
				naiNode1,
			},
			wantErr: false,
		},
		{
			name: "GetAddressInfoByAddressPort:Failed",
			fields: fields{
				NodeAddressInfoStorage: &mockNaiStorageFail{},
			},
			args: args{
				address: "127.0.0.2",
				port:    3001,
				nodeAddressStatuses: []model.NodeAddressStatus{
					model.NodeAddressStatus_NodeAddressConfirmed,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nru := &NodeAddressInfoService{
				QueryExecutor:           tt.fields.QueryExecutor,
				NodeAddressInfoQuery:    tt.fields.NodeAddressInfoQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				Signature:               tt.fields.Signature,
				NodeAddressInfoStorage:  tt.fields.NodeAddressInfoStorage,
				MainBlockStateStorage:   tt.fields.MainBlockStateStorage,
				ActiveNodeRegistryCache: tt.fields.ActiveNodeRegistryCache,
				Logger:                  tt.fields.Logger,
			}
			got, err := nru.GetAddressInfoByAddressPort(tt.args.address, tt.args.port, tt.args.nodeAddressStatuses)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAddressInfoByAddressPort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				log.Println(got)
				t.Errorf("GetAddressInfoByAddressPort() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeAddressInfoService_GetAddressInfoByNodeID(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		NodeAddressInfoQuery    query.NodeAddressInfoQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		Signature               crypto.SignatureInterface
		NodeAddressInfoStorage  storage.CacheStorageInterface
		MainBlockStateStorage   storage.CacheStorageInterface
		ActiveNodeRegistryCache storage.CacheStorageInterface
		Logger                  *log.Logger
	}
	type args struct {
		nodeID          int64
		addressStatuses []model.NodeAddressStatus
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.NodeAddressInfo
		wantErr bool
	}{
		{
			name: "GetAddressInfoByNodeID:Success",
			args: args{
				nodeID: int64(111),
				addressStatuses: []model.NodeAddressStatus{
					model.NodeAddressStatus_NodeAddressConfirmed,
				},
			},
			fields: fields{
				NodeAddressInfoStorage: &mockNaiStorageSuccess{},
			},
			want: []*model.NodeAddressInfo{
				naiNode1,
			},
			wantErr: false,
		},
		{
			name: "GetAddressInfoByNodeID:Fail",
			args: args{
				nodeID: int64(111),
				addressStatuses: []model.NodeAddressStatus{
					model.NodeAddressStatus_NodeAddressConfirmed,
				},
			},
			fields: fields{
				NodeAddressInfoStorage: &mockNaiStorageFail{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nru := &NodeAddressInfoService{
				QueryExecutor:           tt.fields.QueryExecutor,
				NodeAddressInfoQuery:    tt.fields.NodeAddressInfoQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				Signature:               tt.fields.Signature,
				NodeAddressInfoStorage:  tt.fields.NodeAddressInfoStorage,
				MainBlockStateStorage:   tt.fields.MainBlockStateStorage,
				ActiveNodeRegistryCache: tt.fields.ActiveNodeRegistryCache,
				Logger:                  tt.fields.Logger,
			}
			got, err := nru.GetAddressInfoByNodeID(tt.args.nodeID, tt.args.addressStatuses)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAddressInfoByNodeID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAddressInfoByNodeID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeAddressInfoService_GetAddressInfoByNodeIDWithPreferredStatus(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		NodeAddressInfoQuery    query.NodeAddressInfoQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		Signature               crypto.SignatureInterface
		NodeAddressInfoStorage  storage.CacheStorageInterface
		MainBlockStateStorage   storage.CacheStorageInterface
		ActiveNodeRegistryCache storage.CacheStorageInterface
		Logger                  *log.Logger
	}
	type args struct {
		nodeID          int64
		preferredStatus model.NodeAddressStatus
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.NodeAddressInfo
		wantErr bool
	}{
		{
			name: "GetAddressInfoByNodeIDWithPreferredStatus:Success",
			fields: fields{
				NodeAddressInfoStorage: &mockNaiStorageSuccess{},
			},
			args: args{
				nodeID:          111,
				preferredStatus: 2,
			},
			want:    naiNode1,
			wantErr: false,
		},
		{
			name: "GetAddressInfoByNodeIDWithPreferredStatus:Error",
			fields: fields{
				NodeAddressInfoStorage: &mockNaiStorageFail{},
			},
			args: args{
				nodeID:          111,
				preferredStatus: 2,
			},
			wantErr: true,
		},
		{
			name: "GetAddressInfoByNodeIDWithPreferredStatus:Nil",
			fields: fields{
				NodeAddressInfoStorage: &mockNaiStorageEmpty{},
			},
			args: args{
				nodeID:          111,
				preferredStatus: 0,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nru := &NodeAddressInfoService{
				QueryExecutor:           tt.fields.QueryExecutor,
				NodeAddressInfoQuery:    tt.fields.NodeAddressInfoQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				Signature:               tt.fields.Signature,
				NodeAddressInfoStorage:  tt.fields.NodeAddressInfoStorage,
				MainBlockStateStorage:   tt.fields.MainBlockStateStorage,
				ActiveNodeRegistryCache: tt.fields.ActiveNodeRegistryCache,
				Logger:                  tt.fields.Logger,
			}
			got, err := nru.GetAddressInfoByNodeIDWithPreferredStatus(tt.args.nodeID, tt.args.preferredStatus)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAddressInfoByNodeIDWithPreferredStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAddressInfoByNodeIDWithPreferredStatus() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeAddressInfoService_GetAddressInfoByNodeIDs(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		NodeAddressInfoQuery    query.NodeAddressInfoQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		Signature               crypto.SignatureInterface
		NodeAddressInfoStorage  storage.CacheStorageInterface
		MainBlockStateStorage   storage.CacheStorageInterface
		ActiveNodeRegistryCache storage.CacheStorageInterface
		Logger                  *log.Logger
	}
	type args struct {
		nodeIDs         []int64
		addressStatuses []model.NodeAddressStatus
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.NodeAddressInfo
		wantErr bool
	}{
		{
			name: "GetAddressInfoByNodeIDs:Success",
			fields: fields{
				NodeAddressInfoStorage: &mockNaiStorageSuccess{},
			},
			args: args{
				nodeIDs: []int64{
					111,
				},
				addressStatuses: []model.NodeAddressStatus{
					model.NodeAddressStatus_NodeAddressConfirmed,
				},
			},
			want:    []*model.NodeAddressInfo{naiNode1},
			wantErr: false,
		},
		{
			name: "GetAddressInfoByNodeIDs:Error",
			fields: fields{
				NodeAddressInfoStorage: &mockNaiStorageFail{},
			},
			args: args{
				nodeIDs: []int64{
					111,
				},
				addressStatuses: []model.NodeAddressStatus{
					model.NodeAddressStatus_NodeAddressConfirmed,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nru := &NodeAddressInfoService{
				QueryExecutor:           tt.fields.QueryExecutor,
				NodeAddressInfoQuery:    tt.fields.NodeAddressInfoQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				Signature:               tt.fields.Signature,
				NodeAddressInfoStorage:  tt.fields.NodeAddressInfoStorage,
				MainBlockStateStorage:   tt.fields.MainBlockStateStorage,
				ActiveNodeRegistryCache: tt.fields.ActiveNodeRegistryCache,
				Logger:                  tt.fields.Logger,
			}
			got, err := nru.GetAddressInfoByNodeIDs(tt.args.nodeIDs, tt.args.addressStatuses)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAddressInfoByNodeIDs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAddressInfoByNodeIDs() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeAddressInfoService_GetAddressInfoByStatus(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		NodeAddressInfoQuery    query.NodeAddressInfoQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		Signature               crypto.SignatureInterface
		NodeAddressInfoStorage  storage.CacheStorageInterface
		MainBlockStateStorage   storage.CacheStorageInterface
		ActiveNodeRegistryCache storage.CacheStorageInterface
		Logger                  *log.Logger
	}
	type args struct {
		nodeAddressStatuses []model.NodeAddressStatus
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.NodeAddressInfo
		wantErr bool
	}{
		{
			name: "GetAddressInfoByStatus:Success",
			fields: fields{
				NodeAddressInfoStorage: &mockNaiStorageSuccess{},
			},
			args: args{
				nodeAddressStatuses: []model.NodeAddressStatus{
					model.NodeAddressStatus_NodeAddressConfirmed,
				},
			},
			want: []*model.NodeAddressInfo{
				naiNode1,
			},
			wantErr: false,
		},
		{
			name: "GetAddressInfoByStatus:Error",
			fields: fields{
				NodeAddressInfoStorage: &mockNaiStorageFail{},
			},
			args: args{
				nodeAddressStatuses: []model.NodeAddressStatus{
					model.NodeAddressStatus_NodeAddressConfirmed,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nru := &NodeAddressInfoService{
				QueryExecutor:           tt.fields.QueryExecutor,
				NodeAddressInfoQuery:    tt.fields.NodeAddressInfoQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				Signature:               tt.fields.Signature,
				NodeAddressInfoStorage:  tt.fields.NodeAddressInfoStorage,
				MainBlockStateStorage:   tt.fields.MainBlockStateStorage,
				ActiveNodeRegistryCache: tt.fields.ActiveNodeRegistryCache,
				Logger:                  tt.fields.Logger,
			}
			got, err := nru.GetAddressInfoByStatus(tt.args.nodeAddressStatuses)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAddressInfoByStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAddressInfoByStatus() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeAddressInfoService_GetAddressInfoTableWithConsolidatedAddresses(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		NodeAddressInfoQuery    query.NodeAddressInfoQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		Signature               crypto.SignatureInterface
		NodeAddressInfoStorage  storage.CacheStorageInterface
		MainBlockStateStorage   storage.CacheStorageInterface
		ActiveNodeRegistryCache storage.CacheStorageInterface
		Logger                  *log.Logger
	}
	type args struct {
		preferredStatus model.NodeAddressStatus
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.NodeAddressInfo
		wantErr bool
	}{
		{
			name: "GetAddressInfoTableWithConsolidatedAddresses:Success",
			fields: fields{
				NodeAddressInfoStorage: &mockNaiStorageSuccess{},
			},
			args: args{
				preferredStatus: 2,
			},
			want: []*model.NodeAddressInfo{
				naiNode1,
			},
			wantErr: false,
		},
		{
			name: "GetAddressInfoTableWithConsolidatedAddresses:Error",
			fields: fields{
				NodeAddressInfoStorage: &mockNaiStorageFail{},
			},
			args: args{
				preferredStatus: 2,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nru := &NodeAddressInfoService{
				QueryExecutor:           tt.fields.QueryExecutor,
				NodeAddressInfoQuery:    tt.fields.NodeAddressInfoQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				Signature:               tt.fields.Signature,
				NodeAddressInfoStorage:  tt.fields.NodeAddressInfoStorage,
				MainBlockStateStorage:   tt.fields.MainBlockStateStorage,
				ActiveNodeRegistryCache: tt.fields.ActiveNodeRegistryCache,
				Logger:                  tt.fields.Logger,
			}
			got, err := nru.GetAddressInfoTableWithConsolidatedAddresses(tt.args.preferredStatus)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAddressInfoTableWithConsolidatedAddresses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAddressInfoTableWithConsolidatedAddresses() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeAddressInfoService_GetUnsignedNodeAddressInfoBytes(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		NodeAddressInfoQuery    query.NodeAddressInfoQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		Signature               crypto.SignatureInterface
		NodeAddressInfoStorage  storage.CacheStorageInterface
		MainBlockStateStorage   storage.CacheStorageInterface
		ActiveNodeRegistryCache storage.CacheStorageInterface
		Logger                  *log.Logger
	}
	type args struct {
		nodeAddressMessage *model.NodeAddressInfo
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []byte
	}{
		{
			name: "GetUnsignedNodeAddressInfoBytes:Success",
			fields: fields{
				NodeAddressInfoStorage: &mockNaiStorageSuccess{},
			},
			args: args{
				nodeAddressMessage: naiNode1,
			},
			want: []byte{111, 0, 0, 0, 0, 0, 0, 0, 9, 0, 0, 0, 49, 50, 55, 46, 48, 46, 48, 46, 49, 184, 11, 0, 0, 10,
				0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nru := &NodeAddressInfoService{
				QueryExecutor:           tt.fields.QueryExecutor,
				NodeAddressInfoQuery:    tt.fields.NodeAddressInfoQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				Signature:               tt.fields.Signature,
				NodeAddressInfoStorage:  tt.fields.NodeAddressInfoStorage,
				MainBlockStateStorage:   tt.fields.MainBlockStateStorage,
				ActiveNodeRegistryCache: tt.fields.ActiveNodeRegistryCache,
				Logger:                  tt.fields.Logger,
			}
			if got := nru.GetUnsignedNodeAddressInfoBytes(tt.args.nodeAddressMessage); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUnsignedNodeAddressInfoBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeAddressInfoService_InsertAddressInfo(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		NodeAddressInfoQuery    query.NodeAddressInfoQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		Signature               crypto.SignatureInterface
		NodeAddressInfoStorage  storage.CacheStorageInterface
		MainBlockStateStorage   storage.CacheStorageInterface
		ActiveNodeRegistryCache storage.CacheStorageInterface
		Logger                  *log.Logger
	}
	type args struct {
		nodeAddressInfo *model.NodeAddressInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "InsertAddressInfo:Success",
			fields: fields{
				QueryExecutor:          &mockNaiQueryExecutorSuccess{},
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage: &mockNaiStorageSuccess{},
			},
			args: args{
				nodeAddressInfo: naiNode1,
			},
			wantErr: false,
		},
		{
			name: "InsertAddressInfo:FailBeginTx",
			fields: fields{
				QueryExecutor:          &mockNaiQueryExecutorFailBeginTx{},
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage: &mockNaiStorageFail{},
			},
			args: args{
				nodeAddressInfo: nil,
			},
			wantErr: true,
		},
		{
			name: "InsertAddressInfo:FailRollbackTx",
			fields: fields{
				QueryExecutor:          &mockNaiQueryExecutorFailRollbackTx{},
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage: &mockNaiStorageFail{},
				Logger:                 log.New(),
			},
			args: args{
				nodeAddressInfo: naiNode1,
			},
			wantErr: true,
		},
		{
			name: "InsertAddressInfo:FailCommitTx",
			fields: fields{
				QueryExecutor:          &mockNaiQueryExecutorFailCommitTx{},
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage: &mockNaiStorageFail{},
			},
			args: args{
				nodeAddressInfo: naiNode1,
			},
			wantErr: true,
		},
		{
			name: "InsertAddressInfo:FailSetItem",
			fields: fields{
				QueryExecutor:          &mockNaiQueryExecutorSuccess{},
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage: &mockNaiStorageFail{},
			},
			args: args{
				nodeAddressInfo: naiNode1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nru := &NodeAddressInfoService{
				QueryExecutor:           tt.fields.QueryExecutor,
				NodeAddressInfoQuery:    tt.fields.NodeAddressInfoQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				Signature:               tt.fields.Signature,
				NodeAddressInfoStorage:  tt.fields.NodeAddressInfoStorage,
				MainBlockStateStorage:   tt.fields.MainBlockStateStorage,
				ActiveNodeRegistryCache: tt.fields.ActiveNodeRegistryCache,
				Logger:                  tt.fields.Logger,
			}
			if err := nru.InsertAddressInfo(tt.args.nodeAddressInfo); (err != nil) != tt.wantErr {
				t.Errorf("InsertAddressInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeAddressInfoService_RollbackCacheTransaction(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		NodeAddressInfoQuery    query.NodeAddressInfoQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		Signature               crypto.SignatureInterface
		NodeAddressInfoStorage  storage.CacheStorageInterface
		MainBlockStateStorage   storage.CacheStorageInterface
		ActiveNodeRegistryCache storage.CacheStorageInterface
		Logger                  *log.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "RollbackCacheTransaction:Success",
			fields: fields{
				NodeAddressInfoStorage:  &mockNaiStorageSuccess{},
				ActiveNodeRegistryCache: &mockActiveNaiStorageSuccess{},
			},
			wantErr: false,
		},
		{
			name: "RollbackCacheTransaction:Fail",
			fields: fields{
				NodeAddressInfoStorage:  &mockNaiStorageFail{},
				ActiveNodeRegistryCache: &mockActiveNaiStorageFail{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nru := &NodeAddressInfoService{
				QueryExecutor:           tt.fields.QueryExecutor,
				NodeAddressInfoQuery:    tt.fields.NodeAddressInfoQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				Signature:               tt.fields.Signature,
				NodeAddressInfoStorage:  tt.fields.NodeAddressInfoStorage,
				MainBlockStateStorage:   tt.fields.MainBlockStateStorage,
				ActiveNodeRegistryCache: tt.fields.ActiveNodeRegistryCache,
				Logger:                  tt.fields.Logger,
			}
			if err := nru.RollbackCacheTransaction(); (err != nil) != tt.wantErr {
				t.Errorf("RollbackCacheTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeAddressInfoService_UpdateAddrressInfo(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		NodeAddressInfoQuery    query.NodeAddressInfoQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		Signature               crypto.SignatureInterface
		NodeAddressInfoStorage  storage.CacheStorageInterface
		MainBlockStateStorage   storage.CacheStorageInterface
		ActiveNodeRegistryCache storage.CacheStorageInterface
		Logger                  *log.Logger
	}
	type args struct {
		nodeAddressInfo *model.NodeAddressInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "UpdateAddressInfo:Success",
			fields: fields{
				QueryExecutor:          &mockNaiQueryExecutorSuccess{},
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage: &mockNaiStorageSuccess{},
			},
			args: args{
				nodeAddressInfo: naiNode1,
			},
			wantErr: false,
		},
		{
			name: "UpdateAddressInfo:FailBeginTx",
			fields: fields{
				QueryExecutor:          &mockNaiQueryExecutorFailBeginTx{},
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage: &mockNaiStorageFail{},
			},
			args: args{
				nodeAddressInfo: nil,
			},
			wantErr: true,
		},
		{
			name: "InsertAddressInfo:FailRollbackTx",
			fields: fields{
				QueryExecutor:          &mockNaiQueryExecutorFailRollbackTx{},
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage: &mockNaiStorageFail{},
				Logger:                 log.New(),
			},
			args: args{
				nodeAddressInfo: naiNode1,
			},
			wantErr: true,
		},
		{
			name: "InsertAddressInfo:FailCommitTx",
			fields: fields{
				QueryExecutor:          &mockNaiQueryExecutorFailCommitTx{},
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage: &mockNaiStorageFail{},
			},
			args: args{
				nodeAddressInfo: naiNode1,
			},
			wantErr: true,
		},
		{
			name: "InsertAddressInfo:FailSetItem",
			fields: fields{
				QueryExecutor:          &mockNaiQueryExecutorSuccess{},
				NodeAddressInfoQuery:   query.NewNodeAddressInfoQuery(),
				NodeAddressInfoStorage: &mockNaiStorageFail{},
			},
			args: args{
				nodeAddressInfo: naiNode1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nru := &NodeAddressInfoService{
				QueryExecutor:           tt.fields.QueryExecutor,
				NodeAddressInfoQuery:    tt.fields.NodeAddressInfoQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				Signature:               tt.fields.Signature,
				NodeAddressInfoStorage:  tt.fields.NodeAddressInfoStorage,
				MainBlockStateStorage:   tt.fields.MainBlockStateStorage,
				ActiveNodeRegistryCache: tt.fields.ActiveNodeRegistryCache,
				Logger:                  tt.fields.Logger,
			}
			if err := nru.UpdateAddrressInfo(tt.args.nodeAddressInfo); (err != nil) != tt.wantErr {
				t.Errorf("UpdateAddrressInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeAddressInfoService_UpdateOrInsertAddressInfo(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		NodeAddressInfoQuery    query.NodeAddressInfoQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		Signature               crypto.SignatureInterface
		NodeAddressInfoStorage  storage.CacheStorageInterface
		MainBlockStateStorage   storage.CacheStorageInterface
		ActiveNodeRegistryCache storage.CacheStorageInterface
		Logger                  *log.Logger
	}
	type args struct {
		nodeAddressInfo *model.NodeAddressInfo
		updatedStatus   model.NodeAddressStatus
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantUpdated bool
		wantErr     bool
	}{
		{
			name: "UpdateOrInsertAddressInfo:Success",
			fields: fields{
				QueryExecutor:           &mockNaiQueryExecutorSuccess{},
				NodeAddressInfoQuery:    query.NewNodeAddressInfoQuery(),
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				NodeAddressInfoStorage:  &mockNaiStorageSuccess{},
				ActiveNodeRegistryCache: &mockActiveNaiStorageSuccess{},
				Logger:                  log.New(),
			},
			args: args{
				nodeAddressInfo: naiNode2,
				updatedStatus:   2,
			},
			wantUpdated: false,
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nru := &NodeAddressInfoService{
				QueryExecutor:           tt.fields.QueryExecutor,
				NodeAddressInfoQuery:    tt.fields.NodeAddressInfoQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				Signature:               tt.fields.Signature,
				NodeAddressInfoStorage:  tt.fields.NodeAddressInfoStorage,
				MainBlockStateStorage:   tt.fields.MainBlockStateStorage,
				ActiveNodeRegistryCache: tt.fields.ActiveNodeRegistryCache,
				Logger:                  tt.fields.Logger,
			}
			gotUpdated, err := nru.UpdateOrInsertAddressInfo(tt.args.nodeAddressInfo, tt.args.updatedStatus)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateOrInsertAddressInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUpdated != tt.wantUpdated {
				t.Errorf("UpdateOrInsertAddressInfo() gotUpdated = %v, want %v", gotUpdated, tt.wantUpdated)
			}
		})
	}
}

func TestNodeAddressInfoService_ValidateNodeAddressInfo(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		NodeAddressInfoQuery    query.NodeAddressInfoQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		BlockQuery              query.BlockQueryInterface
		Signature               crypto.SignatureInterface
		NodeAddressInfoStorage  storage.CacheStorageInterface
		MainBlockStateStorage   storage.CacheStorageInterface
		ActiveNodeRegistryCache storage.CacheStorageInterface
		Logger                  *log.Logger
	}
	type args struct {
		nodeAddressInfo *model.NodeAddressInfo
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantFound bool
		wantErr   bool
	}{
		{
			name: "ValidateNodeAddressInfo:Success",
			fields: fields{
				NodeAddressInfoStorage:  &mockNaiStorageSuccess{},
				MainBlockStateStorage:   &mockMainBlockStateStorageSuccess{},
				ActiveNodeRegistryCache: &mockActiveNaiStorageSuccess{},
			},
			args: args{
				nodeAddressInfo: naiNode1,
			},
			wantFound: true,
			wantErr:   false,
		},
		{
			name: "ValidateNodeAddressInfo:NotFound",
			fields: fields{
				NodeAddressInfoStorage:  &mockNaiStorageFail{},
				MainBlockStateStorage:   &mockMainBlockStateStorageSuccess{},
				ActiveNodeRegistryCache: &mockActiveNaiStorageFail{},
			},
			args: args{
				nodeAddressInfo: nil,
			},
			wantFound: false,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nru := &NodeAddressInfoService{
				QueryExecutor:           tt.fields.QueryExecutor,
				NodeAddressInfoQuery:    tt.fields.NodeAddressInfoQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				BlockQuery:              tt.fields.BlockQuery,
				Signature:               tt.fields.Signature,
				NodeAddressInfoStorage:  tt.fields.NodeAddressInfoStorage,
				MainBlockStateStorage:   tt.fields.MainBlockStateStorage,
				ActiveNodeRegistryCache: tt.fields.ActiveNodeRegistryCache,
				Logger:                  tt.fields.Logger,
			}
			gotFound, err := nru.ValidateNodeAddressInfo(tt.args.nodeAddressInfo)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNodeAddressInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotFound != tt.wantFound {
				t.Errorf("ValidateNodeAddressInfo() gotFound = %v, want %v", gotFound, tt.wantFound)
			}
		})
	}
}
