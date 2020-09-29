package service

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
)

type (
	nrsMockQueryExecutorSuccess struct {
		query.Executor
	}
	nrsMockQueryExecutorFailNoNodeRegistered struct {
		query.Executor
	}
	nrsMockQueryExecutorFailNodeRegistryListener struct {
		query.Executor
	}
	nrsMockQueryExecutorFailActiveNodeRegistrations struct {
		query.Executor
	}
)

var (
	nrsAddress1    = "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN"
	nrsNodePubKey1 = []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
		45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135}
	nrsQueuedNode1 = &model.NodeRegistration{
		NodeID:             int64(1),
		NodePublicKey:      nrsNodePubKey1,
		AccountAddress:     nrsAddress1,
		RegistrationHeight: 10,
		LockedBalance:      100000000,
		RegistrationStatus: uint32(model.NodeRegistrationState_NodeQueued),
		Latest:             true,
		Height:             100,
	}
	nrsRegisteredNode1 = &model.NodeRegistration{
		NodeID:             int64(1),
		NodePublicKey:      nrsNodePubKey1,
		AccountAddress:     nrsAddress1,
		RegistrationHeight: 10,
		LockedBalance:      100000000,
		RegistrationStatus: uint32(model.NodeRegistrationState_NodeRegistered),
		Latest:             true,
		Height:             200,
	}
)

func (*nrsMockQueryExecutorFailNodeRegistryListener) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("MockedError")
}

func (*nrsMockQueryExecutorFailActiveNodeRegistrations) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("MockedError")
}

func (*nrsMockQueryExecutorFailNoNodeRegistered) ExecuteSelectRow(qe string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT id, node_public_key, account_address, registration_height, locked_balance, registration_status, " +
		"latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		},
		))
	default:
		return nil, errors.New("InvalidQuery")
	}
	row := db.QueryRow(qe)
	return row, nil
}
func (*nrsMockQueryExecutorFailNoNodeRegistered) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT id, node_public_key, account_address, registration_height, locked_balance, registration_status, " +
		"latest, height FROM node_registry WHERE locked_balance > 0 AND registration_status = 1 AND latest=1 " +
		"ORDER BY locked_balance DESC LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		},
		))
	case "SELECT id, node_public_key, account_address, registration_height, locked_balance, registration_status, " +
		"latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		},
		))
	case "SELECT A.id, A.node_public_key, A.account_address, A.registration_height, A.locked_balance, " +
		"A.registration_status, A.latest, A.height FROM node_registry as A INNER JOIN participation_score as B ON A.id = B.node_id " +
		"WHERE B.score = 0 AND A.latest=1 AND A.registration_status=0 AND B.latest=1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		},
		))
	default:
		return nil, errors.New("InvalidQuery")
	}
	rows, _ := db.Query(qe)
	return rows, nil
}

func (*nrsMockQueryExecutorSuccess) ExecuteSelectRow(qe string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	blockQ := query.NewBlockQuery(&chaintype.MainChain{})

	switch qe {
	case "SELECT MAX(height), id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, cumulative_difficulty, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version FROM main_block":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(
			sqlmock.NewRows(blockQ.Fields).AddRow(
				uint32(1000),
				mockGoodBlock.GetID(),
				mockGoodBlock.GetBlockHash(),
				mockGoodBlock.GetPreviousBlockHash(),
				mockGoodBlock.GetTimestamp(),
				mockGoodBlock.GetBlockSeed(),
				mockGoodBlock.GetBlockSignature(),
				mockGoodBlock.GetCumulativeDifficulty(),
				mockGoodBlock.GetPayloadLength(),
				mockGoodBlock.GetPayloadHash(),
				mockGoodBlock.GetBlocksmithPublicKey(),
				mockGoodBlock.GetTotalAmount(),
				mockGoodBlock.GetTotalFee(),
				mockGoodBlock.GetTotalCoinBase(),
				mockGoodBlock.GetVersion(),
			),
		)
	case fmt.Sprintf("SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, cumulative_difficulty, "+
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, "+
		"version FROM main_block WHERE height = %d", 1000-constant.MinRollbackBlocks):
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"Height", "ID", "BlockHash", "PreviousBlockHash", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
			"PayloadLength", "PayloadHash", "BlocksmithPublicKey", "TotalAmount", "TotalFee", "TotalCoinBase",
			"Version"},
		).AddRow(280, 1, make([]byte, 32), []byte{}, 10000, []byte{}, []byte{}, "", 2, []byte{}, bcsNodePubKey1, 0, 0, 0,
			1))
	case "SELECT id, node_public_key, account_address, registration_height, locked_balance, registration_status, " +
		"latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		},
		).AddRow(1, nrsNodePubKey1, nrsAddress1, 10, 100000000, uint32(model.NodeRegistrationState_NodeQueued), true, 100))
	default:
		return nil, errors.New("InvalidQueryRow")

	}
	return db.QueryRow(qe), nil
}

func (*nrsMockQueryExecutorSuccess) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT id, node_public_key, account_address, registration_height, locked_balance, registration_status, " +
		"latest, height FROM node_registry WHERE locked_balance > 0 AND registration_status = 1 AND latest=1 " +
		"ORDER BY locked_balance DESC LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		},
		).AddRow(1, nrsNodePubKey1, nrsAddress1, 10, 100000000, uint32(model.NodeRegistrationState_NodeQueued), true, 100))
	case "SELECT id, node_public_key, account_address, registration_height, locked_balance, registration_status, " +
		"latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		},
		).AddRow(1, nrsNodePubKey1, nrsAddress1, 10, 100000000, uint32(model.NodeRegistrationState_NodeQueued), true, 100))
	case "SELECT A.id, A.node_public_key, A.account_address, A.registration_height, A.locked_balance, " +
		"A.registration_status, A.latest, A.height FROM node_registry as A INNER JOIN participation_score as B ON A.id = B.node_id " +
		"WHERE B.score <= 0 AND A.latest=1 AND A.registration_status=0 AND B.latest=1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		},
		).AddRow(1, nrsNodePubKey1, nrsAddress1, 10, 100000000, uint32(model.NodeRegistrationState_NodeQueued), true, 100))
	case "SELECT nr.id AS nodeID, nr.node_public_key AS node_public_key, ps.score AS participation_score " +
		"FROM node_registry AS nr INNER JOIN participation_score AS ps ON nr.id = ps.node_id " +
		"WHERE nr.registration_status = 0 AND nr.latest = 1 AND ps.score > 0 AND ps.latest = 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"participation_score",
		},
		).AddRow(1, nrsNodePubKey1, 8000))
	case "SELECT id, node_public_key, account_address, registration_height, locked_balance, " +
		"registration_status, latest, height FROM node_registry INNER JOIN node_address_info AS t2 ON id = t2.node_id " +
		"WHERE registration_status = 0 AND (id,height) in (SELECT t1.id,MAX(t1.height) " +
		"FROM node_registry AS t1 WHERE t1.height <= 1 GROUP BY t1.id) ORDER BY height DESC":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		},
		).AddRow(1, nrsNodePubKey1, nrsAddress1, 10, 100000000,
			uint32(model.NodeRegistrationState_NodeRegistered), true, 200))
	default:
		return nil, errors.New("InvalidQuery")
	}
	rows, _ := db.Query(qe)
	return rows, nil
}

func (*nrsMockQueryExecutorSuccess) ExecuteStatement(qe string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}

func (*nrsMockQueryExecutorSuccess) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}

func (*nrsMockQueryExecutorSuccess) BeginTx() error { return nil }

func (*nrsMockQueryExecutorSuccess) RollbackTx() error { return nil }

func (*nrsMockQueryExecutorSuccess) CommitTx() error { return nil }

type (
	mockNodeRegistryCacheSelectNodesToBeAdmittedSuccess struct {
		storage.CacheStorageInterface
	}
	mockNodeRegistryCacheSelectNodesToBeAdmittedEmpty struct {
		storage.CacheStorageInterface
	}
	mockNodeRegistryCacheSelectNodesToBeAdmittedError struct {
		storage.CacheStorageInterface
	}
)

func (*mockNodeRegistryCacheSelectNodesToBeAdmittedSuccess) GetAllItems(items interface{}) error {
	nodeRegistries, ok := items.(*[]storage.NodeRegistry)
	if !ok {
		return errors.New("wrongtype")
	}
	*nodeRegistries = append(*nodeRegistries, storage.NodeRegistry{
		Node:               *nrsQueuedNode1,
		ParticipationScore: 0,
	})
	return nil
}

func (*mockNodeRegistryCacheSelectNodesToBeAdmittedEmpty) GetAllItems(items interface{}) error {
	return nil
}

func (*mockNodeRegistryCacheSelectNodesToBeAdmittedError) GetAllItems(items interface{}) error {
	return errors.New("mockedError")
}

func TestNodeRegistrationService_SelectNodesToBeAdmitted(t *testing.T) {
	type fields struct {
		QueryExecutor            query.ExecutorInterface
		AccountBalanceQuery      query.AccountBalanceQueryInterface
		NodeRegistrationQuery    query.NodeRegistrationQueryInterface
		PendingNodeRegistryCache storage.CacheStorageInterface
	}
	type args struct {
		limit uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.NodeRegistration
		wantErr bool
	}{
		{
			name: "SelectNodesToBeAdmitted:success",
			fields: fields{
				QueryExecutor:            &nrsMockQueryExecutorSuccess{},
				AccountBalanceQuery:      query.NewAccountBalanceQuery(),
				NodeRegistrationQuery:    query.NewNodeRegistrationQuery(),
				PendingNodeRegistryCache: &mockNodeRegistryCacheSelectNodesToBeAdmittedSuccess{},
			},
			args: args{
				limit: 1,
			},
			want: []*model.NodeRegistration{
				nrsQueuedNode1,
			},
			wantErr: false,
		},
		{
			name: "SelectNodesToBeAdmitted:fail-{NoNodeRegistered}",
			fields: fields{
				QueryExecutor:            &nrsMockQueryExecutorFailNoNodeRegistered{},
				AccountBalanceQuery:      query.NewAccountBalanceQuery(),
				NodeRegistrationQuery:    query.NewNodeRegistrationQuery(),
				PendingNodeRegistryCache: &mockNodeRegistryCacheSelectNodesToBeAdmittedEmpty{},
			},
			args: args{
				limit: 1,
			},
			wantErr: false,
			want:    []*model.NodeRegistration{},
		},
		{
			name: "SelectNodesToBeAdmitted:fail-{read cache fail}",
			fields: fields{
				QueryExecutor:            &nrsMockQueryExecutorFailNoNodeRegistered{},
				AccountBalanceQuery:      query.NewAccountBalanceQuery(),
				NodeRegistrationQuery:    query.NewNodeRegistrationQuery(),
				PendingNodeRegistryCache: &mockNodeRegistryCacheSelectNodesToBeAdmittedError{},
			},
			args: args{
				limit: 1,
			},
			wantErr: true,
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrs := &NodeRegistrationService{
				QueryExecutor:                   tt.fields.QueryExecutor,
				AccountBalanceQuery:             tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery:           tt.fields.NodeRegistrationQuery,
				PendingNodeRegistryCacheStorage: tt.fields.PendingNodeRegistryCache,
			}
			got, err := nrs.SelectNodesToBeAdmitted(tt.args.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistrationService.SelectNodesToBeAdmitted() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeRegistrationService.SelectNodesToBeAdmitted() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockActiveNodeRegistryCacheSuccess struct {
		storage.NodeRegistryCacheStorage
	}
	mockPendingNodeRegistryCacheSuccess struct {
		storage.NodeRegistryCacheStorage
	}
)

func (*mockActiveNodeRegistryCacheSuccess) GetAllItems(items interface{}) error {
	return nil
}

func (*mockActiveNodeRegistryCacheSuccess) SetItems(items interface{}) error {
	return nil
}

func (*mockActiveNodeRegistryCacheSuccess) TxSetItems(items interface{}) error {
	return nil
}

func (*mockPendingNodeRegistryCacheSuccess) GetAllItems(items interface{}) error {
	nodeRegistries, ok := items.(*[]storage.NodeRegistry)
	if !ok {
		return errors.New("wrongtype")
	}
	*nodeRegistries = append(*nodeRegistries, storage.NodeRegistry{
		Node:               *nrsQueuedNode1,
		ParticipationScore: 0,
	})
	return nil
}

func (*mockPendingNodeRegistryCacheSuccess) RemoveItem(index interface{}) error {
	return nil
}

func (*mockPendingNodeRegistryCacheSuccess) TxRemoveItem(index interface{}) error {
	return nil
}

func TestNodeRegistrationService_AdmitNodes(t *testing.T) {
	type fields struct {
		QueryExecutor            query.ExecutorInterface
		AccountBalanceQuery      query.AccountBalanceQueryInterface
		NodeRegistrationQuery    query.NodeRegistrationQueryInterface
		ParticipationScoreQuery  query.ParticipationScoreQueryInterface
		ActiveNodeRegistryCache  storage.CacheStorageInterface
		PendingNodeRegistryCache storage.CacheStorageInterface
	}
	type args struct {
		nodeRegistrations []*model.NodeRegistration
		height            uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "AdmitNodes:success",
			fields: fields{
				QueryExecutor:            &nrsMockQueryExecutorSuccess{},
				AccountBalanceQuery:      query.NewAccountBalanceQuery(),
				NodeRegistrationQuery:    query.NewNodeRegistrationQuery(),
				ParticipationScoreQuery:  query.NewParticipationScoreQuery(),
				ActiveNodeRegistryCache:  &mockActiveNodeRegistryCacheSuccess{},
				PendingNodeRegistryCache: &mockPendingNodeRegistryCacheSuccess{},
			},
			args: args{
				nodeRegistrations: []*model.NodeRegistration{
					nrsQueuedNode1,
				},
				height: 200,
			},
			wantErr: false,
		},
		{
			name: "AdmitNodes:fail-{NoNodesToBeAdmitted}",
			fields: fields{
				QueryExecutor:            &nrsMockQueryExecutorSuccess{},
				AccountBalanceQuery:      query.NewAccountBalanceQuery(),
				NodeRegistrationQuery:    query.NewNodeRegistrationQuery(),
				ParticipationScoreQuery:  query.NewParticipationScoreQuery(),
				ActiveNodeRegistryCache:  &mockActiveNodeRegistryCacheSuccess{},
				PendingNodeRegistryCache: &mockPendingNodeRegistryCacheSuccess{},
			},
			args: args{
				nodeRegistrations: []*model.NodeRegistration{},
				height:            200,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrs := &NodeRegistrationService{
				QueryExecutor:                   tt.fields.QueryExecutor,
				AccountBalanceQuery:             tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery:           tt.fields.NodeRegistrationQuery,
				ParticipationScoreQuery:         tt.fields.ParticipationScoreQuery,
				ActiveNodeRegistryCacheStorage:  tt.fields.ActiveNodeRegistryCache,
				PendingNodeRegistryCacheStorage: tt.fields.PendingNodeRegistryCache,
			}
			if err := nrs.AdmitNodes(tt.args.nodeRegistrations, tt.args.height); (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistrationService.AdmitNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockExpelNodesNodeAddressInfoSuccess struct {
		NodeAddressInfoServiceInterface
	}
)

func (*mockExpelNodesNodeAddressInfoSuccess) DeleteNodeAddressInfoByNodeIDInDBTx(nodeID int64) error {
	return nil
}

type (
	mockActiveNodeRegistryCacheExpelNodeSuccess struct {
		storage.NodeRegistryCacheStorage
	}

	mockPendingNodeRegistryCacheExpelNodeSuccess struct {
		storage.NodeRegistryCacheStorage
	}
)

func (*mockActiveNodeRegistryCacheExpelNodeSuccess) GetAllItems(items interface{}) error {
	nodeRegistries, ok := items.(*[]storage.NodeRegistry)
	if !ok {
		return errors.New("wrongtype")
	}
	*nodeRegistries = append(*nodeRegistries, storage.NodeRegistry{
		Node:               *nrsRegisteredNode1,
		ParticipationScore: 0,
	})
	return nil
}

func (*mockActiveNodeRegistryCacheExpelNodeSuccess) RemoveItem(index interface{}) error {
	return nil
}

func (*mockActiveNodeRegistryCacheExpelNodeSuccess) TxRemoveItem(index interface{}) error {
	return nil
}

func TestNodeRegistrationService_ExpelNodes(t *testing.T) {
	type fields struct {
		QueryExecutor            query.ExecutorInterface
		AccountBalanceQuery      query.AccountBalanceQueryInterface
		NodeRegistrationQuery    query.NodeRegistrationQueryInterface
		ParticipationScoreQuery  query.ParticipationScoreQueryInterface
		NodeAddressInfoService   NodeAddressInfoServiceInterface
		NodeAdmittanceCycle      uint32
		PendingNodeRegistryCache storage.CacheStorageInterface
		ActiveNodeRegistryCache  storage.CacheStorageInterface
	}
	type args struct {
		nodeRegistrations []*model.NodeRegistration
		height            uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ExpelNodes:success",
			fields: fields{
				QueryExecutor:            &nrsMockQueryExecutorSuccess{},
				AccountBalanceQuery:      query.NewAccountBalanceQuery(),
				NodeRegistrationQuery:    query.NewNodeRegistrationQuery(),
				ParticipationScoreQuery:  query.NewParticipationScoreQuery(),
				NodeAddressInfoService:   &mockExpelNodesNodeAddressInfoSuccess{},
				PendingNodeRegistryCache: &mockPendingNodeRegistryCacheExpelNodeSuccess{},
				ActiveNodeRegistryCache:  &mockActiveNodeRegistryCacheExpelNodeSuccess{},
			},
			args: args{
				nodeRegistrations: []*model.NodeRegistration{nrsRegisteredNode1},
				height:            300,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrs := &NodeRegistrationService{
				QueryExecutor:                   tt.fields.QueryExecutor,
				AccountBalanceQuery:             tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery:           tt.fields.NodeRegistrationQuery,
				ParticipationScoreQuery:         tt.fields.ParticipationScoreQuery,
				NodeAddressInfoService:          tt.fields.NodeAddressInfoService,
				PendingNodeRegistryCacheStorage: tt.fields.PendingNodeRegistryCache,
				ActiveNodeRegistryCacheStorage:  tt.fields.ActiveNodeRegistryCache,
			}
			if err := nrs.ExpelNodes(tt.args.nodeRegistrations, tt.args.height); (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistrationService.ExpelNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeRegistrationService_GetNodeRegistrationByNodePublicKey(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		NodeAdmittanceCycle     uint32
	}
	type args struct {
		nodePublicKey []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.NodeRegistration
		wantErr bool
	}{
		{
			name: "GetNodeRegistrationByNodePublicKey:success",
			fields: fields{
				QueryExecutor:         &nrsMockQueryExecutorSuccess{},
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			args: args{
				nodePublicKey: nrsNodePubKey1,
			},
			want: &model.NodeRegistration{
				NodeID:             int64(1),
				NodePublicKey:      nrsNodePubKey1,
				AccountAddress:     nrsAddress1,
				RegistrationHeight: 10,
				LockedBalance:      100000000,
				RegistrationStatus: uint32(model.NodeRegistrationState_NodeQueued),
				Latest:             true,
				Height:             100,
			},
			wantErr: false,
		},
		{
			name: "GetNodeRegistrationByNodePublicKey:fail-{NoNodeRegistered}",
			fields: fields{
				QueryExecutor:         &nrsMockQueryExecutorFailNoNodeRegistered{},
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			args: args{
				nodePublicKey: nrsNodePubKey1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrs := &NodeRegistrationService{
				QueryExecutor:           tt.fields.QueryExecutor,
				AccountBalanceQuery:     tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				ParticipationScoreQuery: tt.fields.ParticipationScoreQuery,
			}
			got, err := nrs.GetNodeRegistrationByNodePublicKey(tt.args.nodePublicKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistrationService.GetNodeRegistrationByNodePublicKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeRegistrationService.GetNodeRegistrationByNodePublicKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

var (
	mockNodeRegistrySelectNodesToBeExpelledSuccess = &model.NodeRegistration{
		NodeID:             int64(1),
		NodePublicKey:      nrsNodePubKey1,
		AccountAddress:     nrsAddress1,
		RegistrationHeight: 10,
		LockedBalance:      100000000,
		RegistrationStatus: uint32(model.NodeRegistrationState_NodeQueued),
		Latest:             true,
		Height:             100,
	}
)

type (
	mockActiveNodeRegistryCacheSelectNodesToBeExpelled struct {
		storage.CacheStorageInterface
	}
	mockActiveNodeRegistryCacheSelectNodesToBeExpelledEmpty struct {
		storage.CacheStorageInterface
	}
	mockActiveNodeRegistryCacheSelectNodesToBeExpelledError struct {
		storage.CacheStorageInterface
	}
)

func (*mockActiveNodeRegistryCacheSelectNodesToBeExpelled) GetAllItems(items interface{}) error {
	nodeRegistries, ok := items.(*[]storage.NodeRegistry)
	if !ok {
		return errors.New("wrongtype")
	}
	*nodeRegistries = append(*nodeRegistries, storage.NodeRegistry{
		Node:               *mockNodeRegistrySelectNodesToBeExpelledSuccess,
		ParticipationScore: 0,
	})
	return nil
}

func (*mockActiveNodeRegistryCacheSelectNodesToBeExpelledEmpty) GetAllItems(items interface{}) error {
	return nil
}

func (*mockActiveNodeRegistryCacheSelectNodesToBeExpelledError) GetAllItems(items interface{}) error {
	return errors.New("mockedError")
}

func TestNodeRegistrationService_SelectNodesToBeExpelled(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		NodeAdmittanceCycle     uint32
		ActiveNodeRegistryCache storage.CacheStorageInterface
	}
	tests := []struct {
		name    string
		fields  fields
		want    []*model.NodeRegistration
		wantErr bool
	}{
		{
			name: "SelectNodesToBeExpelled:success",
			fields: fields{
				QueryExecutor:           &nrsMockQueryExecutorSuccess{},
				AccountBalanceQuery:     query.NewAccountBalanceQuery(),
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				ActiveNodeRegistryCache: &mockActiveNodeRegistryCacheSelectNodesToBeExpelled{},
			},
			want: []*model.NodeRegistration{
				mockNodeRegistrySelectNodesToBeExpelledSuccess,
			},
			wantErr: false,
		},
		{
			name: "SelectNodesToBeExpelled:fail-{Empty}",
			fields: fields{
				QueryExecutor:           &nrsMockQueryExecutorFailNoNodeRegistered{},
				AccountBalanceQuery:     query.NewAccountBalanceQuery(),
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				ActiveNodeRegistryCache: &mockActiveNodeRegistryCacheSelectNodesToBeExpelledEmpty{},
			},
			wantErr: false,
			want:    []*model.NodeRegistration{},
		},
		{
			name: "SelectNodesToBeExpelled:fail-{NoNodeRegistered}",
			fields: fields{
				QueryExecutor:           &nrsMockQueryExecutorFailNoNodeRegistered{},
				AccountBalanceQuery:     query.NewAccountBalanceQuery(),
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				ActiveNodeRegistryCache: &mockActiveNodeRegistryCacheSelectNodesToBeExpelledError{},
			},
			wantErr: true,
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrs := &NodeRegistrationService{
				QueryExecutor:                  tt.fields.QueryExecutor,
				AccountBalanceQuery:            tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery:          tt.fields.NodeRegistrationQuery,
				ParticipationScoreQuery:        tt.fields.ParticipationScoreQuery,
				ActiveNodeRegistryCacheStorage: tt.fields.ActiveNodeRegistryCache,
			}
			got, err := nrs.SelectNodesToBeExpelled()
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistrationService.SelectNodesToBeExpelled() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeRegistrationService.SelectNodesToBeExpelled() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockQueryExecutorAddParticipationScorePsNotFound struct {
		query.Executor
	}
	mockQueryExecutorAddParticipationScoreSuccess struct {
		query.Executor
		prevScore int64
	}
)

func (*mockQueryExecutorAddParticipationScorePsNotFound) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	return nil, nil
}

func (mk *mockQueryExecutorAddParticipationScoreSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	psQ := query.NewParticipationScoreQuery()
	mock.MatchExpectationsInOrder(false)
	mock.ExpectQuery("").WillReturnRows(
		sqlmock.NewRows(psQ.Fields).AddRow(
			int64(1111),
			mk.prevScore,
			true,
			uint32(0),
		),
	)
	return db.QueryRow(""), nil
}

func (*mockQueryExecutorAddParticipationScoreSuccess) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}

type (
	mockActiveNodeRegistryCacheAddParticipationScoreSuccess struct {
		storage.NodeRegistryCacheStorage
		prevScore int64
	}
	mockActiveNodeRegistryCacheAddParticipationScoreNotFound struct {
		storage.NodeRegistryCacheStorage
	}
)

func (ma *mockActiveNodeRegistryCacheAddParticipationScoreSuccess) GetItem(id, item interface{}) error {
	registry := item.(*storage.NodeRegistry)
	*registry = storage.NodeRegistry{
		Node: model.NodeRegistration{
			NodeID: 1111,
		},
		ParticipationScore: float64(ma.prevScore),
	}
	return nil
}

func (ma *mockActiveNodeRegistryCacheAddParticipationScoreNotFound) GetItem(id, item interface{}) error {
	return errors.New("mockedError")
}

func TestNodeRegistrationService_AddParticipationScore(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		Logger                  *log.Logger
		ActiveNodeRegistryCache storage.CacheStorageInterface
	}
	type args struct {
		nodeID     int64
		scoreDelta int64
		height     uint32
		dbTx       bool
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantNewScore int64
		wantErr      bool
	}{
		{
			name: "fail-{ParticipationScoreNotFound}",
			fields: fields{
				QueryExecutor:           &mockQueryExecutorAddParticipationScorePsNotFound{},
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				Logger:                  log.New(),
				ActiveNodeRegistryCache: &mockActiveNodeRegistryCacheAddParticipationScoreNotFound{},
			},
			args: args{
				nodeID:     -1,
				scoreDelta: 10,
				height:     1,
			},
			wantErr: true,
		},
		{
			name: "wantSuccess-{AlreadyMaxScore}",
			fields: fields{
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				Logger:                  log.New(),
				ActiveNodeRegistryCache: &mockActiveNodeRegistryCacheAddParticipationScoreSuccess{
					prevScore: constant.MaxParticipationScore,
				},
			},
			args: args{
				nodeID:     1111,
				scoreDelta: 10,
				height:     1,
			},
			wantNewScore: constant.MaxParticipationScore,
		},
		{
			name: "wantSuccess-{AlreadyZeroScore}",
			fields: fields{
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				Logger:                  log.New(),
				ActiveNodeRegistryCache: &mockActiveNodeRegistryCacheAddParticipationScoreSuccess{
					prevScore: 0,
				},
			},
			args: args{
				nodeID:     1111,
				scoreDelta: -10,
				height:     1,
			},
			wantNewScore: 0,
		},
		{
			name: "wantSuccess-{ToMaxScore}",
			fields: fields{
				QueryExecutor:           &mockQueryExecutorAddParticipationScoreSuccess{},
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				Logger:                  log.New(),
				ActiveNodeRegistryCache: &mockActiveNodeRegistryCacheAddParticipationScoreSuccess{
					prevScore: constant.MaxParticipationScore - 10,
				},
			},
			args: args{
				nodeID:     1111,
				scoreDelta: 10,
				height:     1,
			},
			wantNewScore: constant.MaxParticipationScore,
		},
		{
			name: "wantSuccess-{ToMinScore}",
			fields: fields{
				QueryExecutor:           &mockQueryExecutorAddParticipationScoreSuccess{},
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				Logger:                  log.New(),
				ActiveNodeRegistryCache: &mockActiveNodeRegistryCacheAddParticipationScoreSuccess{
					prevScore: 5,
				},
			},
			args: args{
				nodeID:     1111,
				scoreDelta: -10,
				height:     1,
			},
			wantNewScore: 0,
		},
		{
			name: "wantSuccess-{IncreaseScore}",
			fields: fields{
				QueryExecutor: &mockQueryExecutorAddParticipationScoreSuccess{
					prevScore: constant.MaxParticipationScore,
				},
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				Logger:                  log.New(),
				ActiveNodeRegistryCache: &mockActiveNodeRegistryCacheAddParticipationScoreSuccess{
					prevScore: constant.MaxParticipationScore - 11,
				},
			},
			args: args{
				nodeID:     1111,
				scoreDelta: 10,
				height:     1,
			},
			wantNewScore: constant.MaxParticipationScore - 1,
		},
		{
			name: "wantSuccess-{DecreaseScore}",
			fields: fields{
				QueryExecutor: &mockQueryExecutorAddParticipationScoreSuccess{
					prevScore: constant.MaxParticipationScore,
				},
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				Logger:                  log.New(),
				ActiveNodeRegistryCache: &mockActiveNodeRegistryCacheAddParticipationScoreSuccess{
					prevScore: 11,
				},
			},
			args: args{
				nodeID:     1111,
				scoreDelta: -10,
				height:     1,
			},
			wantNewScore: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrs := &NodeRegistrationService{
				QueryExecutor:                  tt.fields.QueryExecutor,
				ParticipationScoreQuery:        tt.fields.ParticipationScoreQuery,
				Logger:                         tt.fields.Logger,
				ActiveNodeRegistryCacheStorage: tt.fields.ActiveNodeRegistryCache,
			}
			gotNewScore, err := nrs.AddParticipationScore(tt.args.nodeID, tt.args.scoreDelta, tt.args.height, tt.args.dbTx)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistrationService.AddParticipationScore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotNewScore != tt.wantNewScore {
				t.Errorf("NodeRegistrationService.AddParticipationScore() = %v, want %v", gotNewScore, tt.wantNewScore)
			}
		})
	}
}

type (
	mockGetNextNodeAdmissionTimestampNextNodeAdmissionStorageFail struct {
		storage.CacheStorageInterface
	}
	mockGetNextNodeAdmissionTimestampNextNodeAdmissionStorageSuccess struct {
		storage.CacheStorageInterface
	}
)

func (*mockGetNextNodeAdmissionTimestampNextNodeAdmissionStorageFail) GetItem(
	lastChange, item interface{}) error {
	return errors.New("mockedError")
}

func (*mockGetNextNodeAdmissionTimestampNextNodeAdmissionStorageSuccess) GetItem(
	lastChange, item interface{}) error {

	return nil
}

func TestNodeRegistrationService_GetNextNodeAdmissionTimestamp(t *testing.T) {
	type fields struct {
		QueryExecutor               query.ExecutorInterface
		AccountBalanceQuery         query.AccountBalanceQueryInterface
		NodeRegistrationQuery       query.NodeRegistrationQueryInterface
		ParticipationScoreQuery     query.ParticipationScoreQueryInterface
		NodeAdmissionTimestampQuery query.NodeAdmissionTimestampQueryInterface
		NextNodeAdmissionStorage    storage.CacheStorageInterface
		Logger                      *log.Logger
		BlockchainStatusService     BlockchainStatusServiceInterface
		CurrentNodePublicKey        []byte
		NodeAddressInfoService      NodeAddressInfoServiceInterface
	}
	tests := []struct {
		name    string
		fields  fields
		want    *model.NodeAdmissionTimestamp
		wantErr bool
	}{
		{
			name: "wantFail:GetItem",
			fields: fields{
				NextNodeAdmissionStorage: &mockGetNextNodeAdmissionTimestampNextNodeAdmissionStorageFail{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				NextNodeAdmissionStorage: &mockGetNextNodeAdmissionTimestampNextNodeAdmissionStorageSuccess{},
			},
			want:    &model.NodeAdmissionTimestamp{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrs := &NodeRegistrationService{
				QueryExecutor:               tt.fields.QueryExecutor,
				AccountBalanceQuery:         tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery:       tt.fields.NodeRegistrationQuery,
				ParticipationScoreQuery:     tt.fields.ParticipationScoreQuery,
				NodeAdmissionTimestampQuery: tt.fields.NodeAdmissionTimestampQuery,
				NextNodeAdmissionStorage:    tt.fields.NextNodeAdmissionStorage,
				Logger:                      tt.fields.Logger,
				BlockchainStatusService:     tt.fields.BlockchainStatusService,
				CurrentNodePublicKey:        tt.fields.CurrentNodePublicKey,
				NodeAddressInfoService:      tt.fields.NodeAddressInfoService,
			}
			got, err := nrs.GetNextNodeAdmissionTimestamp()
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistrationService.GetNextNodeAdmissionTimestamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeRegistrationService.GetNextNodeAdmissionTimestamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockQueryExecutorInsertNextNodeAdmissionTimestampFail struct {
		query.Executor
	}
	mockQueryExecutorInsertNextNodeAdmissionTimestampSuccess struct {
		query.Executor
	}
	mockQueryExecutorInsertNextNodeAdmissionTimestampFailExecuteTransactions struct {
		query.Executor
	}
	mockNodeRegistrationQueryInsertNextNodeAdmissionTimestampFail struct {
		query.NodeRegistrationQuery
	}
	mockNodeRegistrationQueryInsertNextNodeAdmissionTimestampSuccess struct {
		query.NodeRegistrationQuery
	}
)

func (*mockQueryExecutorInsertNextNodeAdmissionTimestampFail) ExecuteSelect(
	query string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	return nil, errors.New("mockedErr")
}

func (*mockQueryExecutorInsertNextNodeAdmissionTimestampSuccess) ExecuteSelect(
	qry string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	var (
		db, mock, _ = sqlmock.New()
		mockRows    = mock.NewRows(query.NewNodeAdmissionTimestampQuery().Fields)
	)
	defer db.Close()
	mock.ExpectQuery("").WillReturnRows(mockRows)
	return db.Query("")
}

func (*mockQueryExecutorInsertNextNodeAdmissionTimestampSuccess) ExecuteTransactions(
	queries [][]interface{},
) error {
	return nil
}
func (*mockQueryExecutorInsertNextNodeAdmissionTimestampFailExecuteTransactions) ExecuteSelect(
	qry string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	var (
		db, mock, _ = sqlmock.New()
		mockRows    = mock.NewRows(query.NewNodeAdmissionTimestampQuery().Fields)
	)
	defer db.Close()
	mock.ExpectQuery("").WillReturnRows(mockRows)
	return db.Query("")
}

func (*mockQueryExecutorInsertNextNodeAdmissionTimestampFailExecuteTransactions) ExecuteTransactions(
	queries [][]interface{},
) error {
	return errors.New("mockedErr")
}
func (*mockNodeRegistrationQueryInsertNextNodeAdmissionTimestampFail) BuildBlocksmith(
	blocksmiths []*model.Blocksmith, rows *sql.Rows,
) ([]*model.Blocksmith, error) {
	return nil, errors.New("mockedErrs")
}
func (*mockNodeRegistrationQueryInsertNextNodeAdmissionTimestampSuccess) BuildBlocksmith(
	blocksmiths []*model.Blocksmith, rows *sql.Rows,
) ([]*model.Blocksmith, error) {
	return []*model.Blocksmith{
		{
			NodeID: 1,
		},
		{
			NodeID: 2,
		},
	}, nil
}

func TestNodeRegistrationService_InsertNextNodeAdmissionTimestamp(t *testing.T) {
	type fields struct {
		QueryExecutor               query.ExecutorInterface
		AccountBalanceQuery         query.AccountBalanceQueryInterface
		NodeRegistrationQuery       query.NodeRegistrationQueryInterface
		ParticipationScoreQuery     query.ParticipationScoreQueryInterface
		NodeAdmissionTimestampQuery query.NodeAdmissionTimestampQueryInterface
		NextNodeAdmissionStorage    storage.CacheStorageInterface
		Logger                      *log.Logger
		BlockchainStatusService     BlockchainStatusServiceInterface
		CurrentNodePublicKey        []byte
	}
	type args struct {
		lastAdmissionTimestamp int64
		blockHeight            uint32
		dbTx                   bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.NodeAdmissionTimestamp
		wantErr bool
	}{
		{
			name: "wantFail:ExecuteSelect",
			fields: fields{
				QueryExecutor:         &mockQueryExecutorInsertNextNodeAdmissionTimestampFail{},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			args: args{
				lastAdmissionTimestamp: 1,
				blockHeight:            1,
				dbTx:                   false,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:BuildModel",
			fields: fields{
				QueryExecutor:         &mockQueryExecutorInsertNextNodeAdmissionTimestampSuccess{},
				NodeRegistrationQuery: &mockNodeRegistrationQueryInsertNextNodeAdmissionTimestampFail{},
			},
			args: args{
				lastAdmissionTimestamp: 1,
				blockHeight:            1,
				dbTx:                   false,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:ExecuteTransactions",
			fields: fields{
				QueryExecutor:               &mockQueryExecutorInsertNextNodeAdmissionTimestampFailExecuteTransactions{},
				NodeRegistrationQuery:       &mockNodeRegistrationQueryInsertNextNodeAdmissionTimestampSuccess{},
				NodeAdmissionTimestampQuery: query.NewNodeAdmissionTimestampQuery(),
			},
			args: args{
				lastAdmissionTimestamp: 1,
				blockHeight:            1,
				dbTx:                   false,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantSuccess",
			fields: fields{
				QueryExecutor:               &mockQueryExecutorInsertNextNodeAdmissionTimestampSuccess{},
				NodeRegistrationQuery:       &mockNodeRegistrationQueryInsertNextNodeAdmissionTimestampSuccess{},
				NodeAdmissionTimestampQuery: query.NewNodeAdmissionTimestampQuery(),
			},
			args: args{
				lastAdmissionTimestamp: 1,
				blockHeight:            1,
				dbTx:                   false,
			},
			want: &model.NodeAdmissionTimestamp{
				Timestamp:   1801,
				BlockHeight: 1,
				Latest:      true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrs := &NodeRegistrationService{
				QueryExecutor:               tt.fields.QueryExecutor,
				AccountBalanceQuery:         tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery:       tt.fields.NodeRegistrationQuery,
				ParticipationScoreQuery:     tt.fields.ParticipationScoreQuery,
				NodeAdmissionTimestampQuery: tt.fields.NodeAdmissionTimestampQuery,
				NextNodeAdmissionStorage:    tt.fields.NextNodeAdmissionStorage,
				Logger:                      tt.fields.Logger,
				BlockchainStatusService:     tt.fields.BlockchainStatusService,
				CurrentNodePublicKey:        tt.fields.CurrentNodePublicKey,
			}
			got, err := nrs.InsertNextNodeAdmissionTimestamp(tt.args.lastAdmissionTimestamp, tt.args.blockHeight, tt.args.dbTx)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistrationService.InsertNextNodeAdmissionTimestamp() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeRegistrationService.InsertNextNodeAdmissionTimestamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	validateNodeAddressInfoExecutorMock struct {
		query.Executor
		nodeIDNotFound bool
		blockNotFound  bool
		nodePublicKey  []byte
		blockHash      []byte
	}
	validateNodeAddressInfoSignatureMock struct {
		crypto.SignatureInterface
		isValid bool
	}
	nodeAddressInfoServiceMock struct {
		NodeAddressInfoServiceInterface
		nodeAddressInfoBytes []byte
	}
)

func (nrMock *nodeAddressInfoServiceMock) GetUnsignedNodeAddressInfoBytes(nodeAddressMessage *model.NodeAddressInfo) []byte {
	return nrMock.nodeAddressInfoBytes
}

func (nrMock *validateNodeAddressInfoSignatureMock) VerifyNodeSignature(payload, signature, nodePublicKey []byte) bool {
	return nrMock.isValid
}

func (nrMock *validateNodeAddressInfoSignatureMock) SignByNode(payload []byte, nodeSeed string) []byte {
	return make([]byte, 64)
}

func (nrMock *validateNodeAddressInfoExecutorMock) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	var (
		sqlRows *sqlmock.Rows
	)
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if (nrMock.nodeIDNotFound && qStr == "SELECT id, node_public_key, account_address, registration_height, locked_balance, "+
		"registration_status, latest, height FROM node_registry WHERE id = ? AND latest=1") ||
		(nrMock.blockNotFound && qStr == "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, "+
			"cumulative_difficulty, payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version "+
			"FROM main_block WHERE height = 10") {
		mock.ExpectQuery("SELECT").WillReturnError(sql.ErrNoRows)
		row := db.QueryRow(qStr)
		return row, nil
	}

	switch qStr {
	case "SELECT id, node_public_key, account_address, registration_height, locked_balance, registration_status, " +
		"latest, height FROM node_registry WHERE id = ? AND latest=1":
		sqlRows = sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		},
		).AddRow(
			0, nrMock.nodePublicKey, "accountA", 0, 0, 0, true, 0,
		)
		mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlRows)
	case "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, cumulative_difficulty, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version " +
		"FROM main_block WHERE height = 10":
		sqlRows = sqlmock.NewRows([]string{
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
		},
		).AddRow(
			0, 0, nrMock.blockHash, nil, 0, nil, nil, "", 0, nil, nil, 0, 0, 0, 0,
		)
	case "SELECT height, id, block_hash, previous_block_hash, timestamp, block_seed, block_signature, cumulative_difficulty, " +
		"payload_length, payload_hash, blocksmith_public_key, total_amount, total_fee, total_coinbase, version " +
		"FROM main_block WHERE height = 11":
		sqlRows = sqlmock.NewRows([]string{
			"id",
			"block_hash",
			"previous_block_hash",
			"height",
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
		},
		).AddRow(
			0, 0, nrMock.blockHash, nil, 0, nil, nil, "", 0, nil, nil, 0, 0, 0, 0,
		)
	default:
		return nil, errors.New("InvalidQuery")
	}
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlRows)
	return db.QueryRow(qStr), nil
}

func (nrMock *validateNodeAddressInfoExecutorMock) ExecuteSelect(qe string, tx bool,
	args ...interface{}) (*sql.Rows,
	error) {
	var (
		sqlRows *sqlmock.Rows
	)
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT node_id, address, port, block_height, block_hash, signature, status FROM node_address_info WHERE node_id IN (1111) " +
		"AND status IN (2, 1) ORDER BY node_id, status ASC":
		sqlRows = sqlmock.NewRows([]string{
			"node_id",
			"address",
			"port",
			"block_height",
			"block_hash",
			"signature",
			"status",
		},
		)
		sqlRows.AddRow(
			int64(1111),
			"192.168.1.1",
			uint32(8080),
			uint32(10),
			make([]byte, 32),
			make([]byte, 64),
			model.NodeAddressStatus_NodeAddressPending,
		)
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlRows)
	default:
		return nil, errors.New("InvalidQuery")
	}
	rows, _ := db.Query(qe)
	return rows, nil
}

type (
	mockValidateNodeAddressInfoNodeAddressInfoServiceSuccess struct {
		NodeAddressInfoServiceInterface
	}
)

func (*mockValidateNodeAddressInfoNodeAddressInfoServiceSuccess) GetUnsignedNodeAddressInfoBytes(
	nodeAddressMessage *model.NodeAddressInfo) []byte {
	return make([]byte, 64)
}

var (
	mockValidateNodeAddressInfoValidBlockHash = []byte{
		3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	}
	mockValidateNodeAddressInfoNodeAddressInfoValid = &model.NodeAddressInfo{
		NodeID:      int64(1111),
		Address:     "192.168.1.2",
		Port:        uint32(8080),
		BlockHeight: uint32(11),
		BlockHash:   mockValidateNodeAddressInfoValidBlockHash,
	}
)

func (*mockValidateNodeAddressInfoNodeAddressInfoServiceSuccess) GetAddressInfoByNodeID(
	nodeID int64,
	addressStatuses []model.NodeAddressStatus,
) ([]*model.NodeAddressInfo, error) {
	return []*model.NodeAddressInfo{
		mockValidateNodeAddressInfoNodeAddressInfoValid,
	}, nil
}
