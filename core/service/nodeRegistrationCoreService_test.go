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
	"github.com/zoobc/zoobc-core/common/blocker"
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

func TestNodeRegistrationService_SelectNodesToBeAdmitted(t *testing.T) {
	type fields struct {
		QueryExecutor         query.ExecutorInterface
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
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
				QueryExecutor:         &nrsMockQueryExecutorSuccess{},
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
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
				QueryExecutor:         &nrsMockQueryExecutorFailNoNodeRegistered{},
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			args: args{
				limit: 1,
			},
			wantErr: false,
			want:    []*model.NodeRegistration{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrs := &NodeRegistrationService{
				QueryExecutor:         tt.fields.QueryExecutor,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
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

func TestNodeRegistrationService_AdmitNodes(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
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
				QueryExecutor:           &nrsMockQueryExecutorSuccess{},
				AccountBalanceQuery:     query.NewAccountBalanceQuery(),
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
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
				QueryExecutor:           &nrsMockQueryExecutorSuccess{},
				AccountBalanceQuery:     query.NewAccountBalanceQuery(),
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
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
				QueryExecutor:           tt.fields.QueryExecutor,
				AccountBalanceQuery:     tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				ParticipationScoreQuery: tt.fields.ParticipationScoreQuery,
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

func TestNodeRegistrationService_ExpelNodes(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		NodeAddressInfoService  NodeAddressInfoServiceInterface
		NodeAdmittanceCycle     uint32
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
				QueryExecutor:           &nrsMockQueryExecutorSuccess{},
				AccountBalanceQuery:     query.NewAccountBalanceQuery(),
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				NodeAddressInfoService:  &mockExpelNodesNodeAddressInfoSuccess{},
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
				QueryExecutor:           tt.fields.QueryExecutor,
				AccountBalanceQuery:     tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				ParticipationScoreQuery: tt.fields.ParticipationScoreQuery,
				NodeAddressInfoService:  tt.fields.NodeAddressInfoService,
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

func TestNodeRegistrationService_SelectNodesToBeExpelled(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		NodeAdmittanceCycle     uint32
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
				QueryExecutor:         &nrsMockQueryExecutorSuccess{},
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			want: []*model.NodeRegistration{
				{
					NodeID:             int64(1),
					NodePublicKey:      nrsNodePubKey1,
					AccountAddress:     nrsAddress1,
					RegistrationHeight: 10,
					LockedBalance:      100000000,
					RegistrationStatus: uint32(model.NodeRegistrationState_NodeQueued),
					Latest:             true,
					Height:             100,
				},
			},
			wantErr: false,
		},
		// {
		// 	name: "SelectNodesToBeExpelled:fail-{NoNodeRegistered}",
		// 	fields: fields{
		// 		QueryExecutor:         &nrsMockQueryExecutorFailNoNodeRegistered{},
		// 		AccountBalanceQuery:   query.NewAccountBalanceQuery(),
		// 		NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
		// 	},
		// 	wantErr: false,
		// 	want:    []*model.NodeRegistration{},
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrs := &NodeRegistrationService{
				QueryExecutor:           tt.fields.QueryExecutor,
				AccountBalanceQuery:     tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				ParticipationScoreQuery: tt.fields.ParticipationScoreQuery,
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
	executorBuildScrambleNodesSuccess struct {
		query.Executor
	}
	executorBuildScrambleNodesFail struct {
		query.Executor
	}
	nrsMockNodeAddressInfoService struct {
		NodeAddressInfoService
		getRegisteredNodesWithConsolidatedAddressesFail bool
	}
)

func (nrsMock *nrsMockNodeAddressInfoService) GetRegisteredNodesWithConsolidatedAddresses(
	height uint32,
	preferredStatus model.NodeAddressStatus) ([]*model.NodeRegistration, error) {
	if nrsMock.getRegisteredNodesWithConsolidatedAddressesFail {
		return nil, errors.New("MockedError")
	}
	return []*model.NodeRegistration{
		{
			NodeID:             0,
			NodePublicKey:      []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 76},
			AccountAddress:     "accountA",
			RegistrationHeight: 0,
			LockedBalance:      0,
			RegistrationStatus: uint32(model.NodeRegistrationState_NodeRegistered),
			Latest:             true,
			Height:             0,
		},
		{
			NodeID:             0,
			NodePublicKey:      []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 78},
			AccountAddress:     "accountB",
			RegistrationHeight: 0,
			LockedBalance:      0,
			RegistrationStatus: uint32(model.NodeRegistrationState_NodeRegistered),
			Latest:             true,
			Height:             0,
		},
	}, nil
}

func (*executorBuildScrambleNodesSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WithArgs(1).WillReturnRows(sqlmock.NewRows(
		query.NewNodeRegistrationQuery().Fields,
	).AddRow(
		0, []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 76}, "accountA", 0, 0, 0, true, 0,
	).AddRow(
		0, []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 78}, "accountB", 0, 0, 0, true, 0,
	))

	return db.Query(qStr, 1)
}

func (*executorBuildScrambleNodesFail) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("mockError:executeSelectFail")
}

type (
	mockBuildScrambledNodesNodeAddressInfoServiceSuccess struct {
		NodeAddressInfoServiceInterface
	}
)

func (*mockBuildScrambledNodesNodeAddressInfoServiceSuccess) GetAddressInfoByNodeIDWithPreferredStatus(
	nodeID int64,
	preferredStatus model.NodeAddressStatus,
) (*model.NodeAddressInfo, error) {
	return &model.NodeAddressInfo{
		BlockHeight: 10,
		NodeID:      int64(111),
		Address:     "127.0.0.1",
		Port:        3001,
		Signature:   make([]byte, 64),
		BlockHash:   make([]byte, 32),
		Status:      model.NodeAddressStatus_NodeAddressConfirmed,
	}, nil
}

func TestNodeRegistrationService_BuildScrambledNodes(t *testing.T) {
	db, mock, _ := sqlmock.New()
	type fields struct {
		QueryExecutor          query.ExecutorInterface
		NodeRegistrationQuery  query.NodeRegistrationQueryInterface
		NodeAddressInfoService NodeAddressInfoServiceInterface
		Logger                 *log.Logger
	}
	type args struct {
		block *model.Block
	}

	// test the building logic and result as well
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.ScrambledNodes
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				QueryExecutor: &executorBuildScrambleNodesSuccess{
					query.Executor{
						Db: db,
					},
				},
				NodeRegistrationQuery:  query.NewNodeRegistrationQuery(),
				NodeAddressInfoService: &mockBuildScrambledNodesNodeAddressInfoServiceSuccess{},
				Logger:                 log.New(),
			},
			args: args{
				block: &model.Block{
					Height:    1,
					BlockSeed: []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 76},
				},
			},
		},
		{
			name: "wantFail",
			fields: fields{
				QueryExecutor: &executorBuildScrambleNodesFail{
					query.Executor{
						Db: db,
					},
				},
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				Logger:                log.New(),
			},
			args: args{
				block: &model.Block{
					Height:    1,
					BlockSeed: []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 76},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrs := &NodeRegistrationService{
				QueryExecutor:          tt.fields.QueryExecutor,
				NodeRegistrationQuery:  tt.fields.NodeRegistrationQuery,
				NodeAddressInfoService: tt.fields.NodeAddressInfoService,
				ScrambledNodes:         map[uint32]*model.ScrambledNodes{},
				Logger:                 tt.fields.Logger,
			}
			errResult := nrs.BuildScrambledNodes(tt.args.block)
			if (errResult != nil) != tt.wantErr {
				t.Errorf("NodeRegistrationService.BuildScrambledNodes() error = %v, wantErr %v", errResult, tt.wantErr)
				return
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
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

func TestNodeRegistrationService_AddParticipationScore(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
		Logger                  *log.Logger
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
				QueryExecutor: &mockQueryExecutorAddParticipationScoreSuccess{
					prevScore: constant.MaxParticipationScore,
				},
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				Logger:                  log.New(),
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
				QueryExecutor: &mockQueryExecutorAddParticipationScoreSuccess{
					prevScore: 0,
				},
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				Logger:                  log.New(),
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
				QueryExecutor: &mockQueryExecutorAddParticipationScoreSuccess{
					prevScore: constant.MaxParticipationScore - 5,
				},
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				Logger:                  log.New(),
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
				QueryExecutor: &mockQueryExecutorAddParticipationScoreSuccess{
					prevScore: 5,
				},
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				Logger:                  log.New(),
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
					prevScore: constant.MaxParticipationScore - 11,
				},
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				Logger:                  log.New(),
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
					prevScore: 11,
				},
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				Logger:                  log.New(),
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
				QueryExecutor:           tt.fields.QueryExecutor,
				ParticipationScoreQuery: tt.fields.ParticipationScoreQuery,
				Logger:                  tt.fields.Logger,
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
		QueryExecutor                query.ExecutorInterface
		AccountBalanceQuery          query.AccountBalanceQueryInterface
		NodeRegistrationQuery        query.NodeRegistrationQueryInterface
		ParticipationScoreQuery      query.ParticipationScoreQueryInterface
		BlockQuery                   query.BlockQueryInterface
		NodeAdmissionTimestampQuery  query.NodeAdmissionTimestampQueryInterface
		NextNodeAdmissionStorage     storage.CacheStorageInterface
		Logger                       *log.Logger
		ScrambledNodes               map[uint32]*model.ScrambledNodes
		MemoizedLatestScrambledNodes *model.ScrambledNodes
		BlockchainStatusService      BlockchainStatusServiceInterface
		CurrentNodePublicKey         []byte
		Signature                    crypto.SignatureInterface
		NodeAddressInfoService       NodeAddressInfoServiceInterface
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
				QueryExecutor:                tt.fields.QueryExecutor,
				AccountBalanceQuery:          tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery:        tt.fields.NodeRegistrationQuery,
				ParticipationScoreQuery:      tt.fields.ParticipationScoreQuery,
				BlockQuery:                   tt.fields.BlockQuery,
				NodeAdmissionTimestampQuery:  tt.fields.NodeAdmissionTimestampQuery,
				NextNodeAdmissionStorage:     tt.fields.NextNodeAdmissionStorage,
				Logger:                       tt.fields.Logger,
				ScrambledNodes:               tt.fields.ScrambledNodes,
				MemoizedLatestScrambledNodes: tt.fields.MemoizedLatestScrambledNodes,
				BlockchainStatusService:      tt.fields.BlockchainStatusService,
				CurrentNodePublicKey:         tt.fields.CurrentNodePublicKey,
				Signature:                    tt.fields.Signature,
				NodeAddressInfoService:       tt.fields.NodeAddressInfoService,
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
		QueryExecutor                query.ExecutorInterface
		AccountBalanceQuery          query.AccountBalanceQueryInterface
		NodeRegistrationQuery        query.NodeRegistrationQueryInterface
		ParticipationScoreQuery      query.ParticipationScoreQueryInterface
		BlockQuery                   query.BlockQueryInterface
		NodeAdmissionTimestampQuery  query.NodeAdmissionTimestampQueryInterface
		NextNodeAdmissionStorage     storage.CacheStorageInterface
		Logger                       *log.Logger
		ScrambledNodes               map[uint32]*model.ScrambledNodes
		MemoizedLatestScrambledNodes *model.ScrambledNodes
		BlockchainStatusService      BlockchainStatusServiceInterface
		CurrentNodePublicKey         []byte
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
				QueryExecutor:                tt.fields.QueryExecutor,
				AccountBalanceQuery:          tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery:        tt.fields.NodeRegistrationQuery,
				ParticipationScoreQuery:      tt.fields.ParticipationScoreQuery,
				BlockQuery:                   tt.fields.BlockQuery,
				NodeAdmissionTimestampQuery:  tt.fields.NodeAdmissionTimestampQuery,
				NextNodeAdmissionStorage:     tt.fields.NextNodeAdmissionStorage,
				Logger:                       tt.fields.Logger,
				ScrambledNodes:               tt.fields.ScrambledNodes,
				MemoizedLatestScrambledNodes: tt.fields.MemoizedLatestScrambledNodes,
				BlockchainStatusService:      tt.fields.BlockchainStatusService,
				CurrentNodePublicKey:         tt.fields.CurrentNodePublicKey,
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
	mockNodeAddressInfoOutDated = &model.NodeAddressInfo{
		NodeID:      int64(1111),
		Address:     "192.168.1.2",
		Port:        uint32(8080),
		BlockHeight: mockValidateNodeAddressInfoNodeAddressInfoValid.BlockHeight - 1,
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

func TestNodeRegistrationService_ValidateNodeAddressInfo(t *testing.T) {
	type fields struct {
		QueryExecutor                query.ExecutorInterface
		AccountBalanceQuery          query.AccountBalanceQueryInterface
		NodeRegistrationQuery        query.NodeRegistrationQueryInterface
		ParticipationScoreQuery      query.ParticipationScoreQueryInterface
		BlockQuery                   query.BlockQueryInterface
		NodeAdmittanceCycle          uint32
		Logger                       *log.Logger
		ScrambledNodes               map[uint32]*model.ScrambledNodes
		MemoizedLatestScrambledNodes *model.ScrambledNodes
		BlockchainStatusService      BlockchainStatusServiceInterface
		CurrentNodePublicKey         []byte
		Signature                    crypto.SignatureInterface
		NodeAddressInfoService       NodeAddressInfoServiceInterface
	}
	type args struct {
		nodeAddressInfo *model.NodeAddressInfo
		validateNotInDb bool
	}

	nodePublicKey := []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	nodeAddressInfo := &model.NodeAddressInfo{
		NodeID:      int64(1111),
		Address:     "192.168.1.1",
		Port:        uint32(8080),
		BlockHeight: uint32(10),
		BlockHash:   mockValidateNodeAddressInfoValidBlockHash,
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		errMsg  string
	}{
		{
			name: "ValidateNodeAddressInfo:fail-{NodeIDNotFound}",
			args: args{
				nodeAddressInfo: nodeAddressInfo,
				validateNotInDb: true,
			},
			fields: fields{
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor: &validateNodeAddressInfoExecutorMock{
					nodeIDNotFound: true,
				},
				Logger: log.New(),
			},
			wantErr: true,
			errMsg:  "NodeIDNotFound",
		},
		{
			name: "ValidateNodeAddressInfo:fail-{InvalidSignature}",
			args: args{
				nodeAddressInfo: nodeAddressInfo,
				validateNotInDb: true,
			},
			fields: fields{
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor: &validateNodeAddressInfoExecutorMock{
					nodePublicKey: nodePublicKey,
				},
				Signature: &validateNodeAddressInfoSignatureMock{
					isValid: false,
				},
				NodeAddressInfoService: &mockValidateNodeAddressInfoNodeAddressInfoServiceSuccess{},
				Logger:                 log.New(),
			},
			wantErr: true,
			errMsg:  "InvalidSignature",
		},
		{
			name: "ValidateNodeAddressInfo:fail-{InvalidBlockHeight}",
			args: args{
				nodeAddressInfo: nodeAddressInfo,
				validateNotInDb: true,
			},
			fields: fields{
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				QueryExecutor: &validateNodeAddressInfoExecutorMock{
					nodePublicKey: nodePublicKey,
					blockNotFound: true,
				},
				Signature: &validateNodeAddressInfoSignatureMock{
					isValid: true,
				},
				NodeAddressInfoService: &mockValidateNodeAddressInfoNodeAddressInfoServiceSuccess{},
				Logger:                 log.New(),
			},
			wantErr: true,
			errMsg:  "InvalidBlockHeight",
		},
		{
			name: "ValidateNodeAddressInfo:fail-{InvalidBlockHash}",
			args: args{
				nodeAddressInfo: nodeAddressInfo,
				validateNotInDb: true,
			},
			fields: fields{
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				QueryExecutor: &validateNodeAddressInfoExecutorMock{
					nodePublicKey: nodePublicKey,
					blockHash:     []byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
				},
				Signature: &validateNodeAddressInfoSignatureMock{
					isValid: true,
				},
				NodeAddressInfoService: &mockValidateNodeAddressInfoNodeAddressInfoServiceSuccess{},
				Logger:                 log.New(),
			},
			wantErr: true,
			errMsg:  "InvalidBlockHash",
		},
		{
			name: "ValidateNodeAddressInfo:fail-{OutdatedNodeAddressInfo}",
			args: args{
				nodeAddressInfo: mockNodeAddressInfoOutDated,
				validateNotInDb: true,
			},
			fields: fields{
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				QueryExecutor: &validateNodeAddressInfoExecutorMock{
					nodePublicKey: nodePublicKey,
					blockHash:     mockValidateNodeAddressInfoValidBlockHash,
				},
				Signature: &validateNodeAddressInfoSignatureMock{
					isValid: true,
				},
				NodeAddressInfoService: &mockValidateNodeAddressInfoNodeAddressInfoServiceSuccess{},
				Logger:                 log.New(),
			},
			wantErr: true,
			errMsg:  "OutdatedNodeAddressInfo",
		},
		{
			name: "ValidateNodeAddressInfo:success",
			args: args{
				nodeAddressInfo: mockValidateNodeAddressInfoNodeAddressInfoValid,
				validateNotInDb: true,
			},
			fields: fields{
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				BlockQuery:            query.NewBlockQuery(&chaintype.MainChain{}),
				QueryExecutor: &validateNodeAddressInfoExecutorMock{
					nodePublicKey: nodePublicKey,
					blockHash:     mockValidateNodeAddressInfoValidBlockHash,
				},
				Signature: &validateNodeAddressInfoSignatureMock{
					isValid: true,
				},
				NodeAddressInfoService: &mockValidateNodeAddressInfoNodeAddressInfoServiceSuccess{},
				Logger:                 log.New(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrs := &NodeRegistrationService{
				QueryExecutor:                tt.fields.QueryExecutor,
				AccountBalanceQuery:          tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery:        tt.fields.NodeRegistrationQuery,
				ParticipationScoreQuery:      tt.fields.ParticipationScoreQuery,
				BlockQuery:                   tt.fields.BlockQuery,
				Logger:                       tt.fields.Logger,
				ScrambledNodes:               tt.fields.ScrambledNodes,
				MemoizedLatestScrambledNodes: tt.fields.MemoizedLatestScrambledNodes,
				BlockchainStatusService:      tt.fields.BlockchainStatusService,
				CurrentNodePublicKey:         tt.fields.CurrentNodePublicKey,
				Signature:                    tt.fields.Signature,
				NodeAddressInfoService:       tt.fields.NodeAddressInfoService,
			}

			if _, err := nrs.ValidateNodeAddressInfo(tt.args.nodeAddressInfo); err != nil {
				if tt.wantErr {
					errorMsg := err.Error()
					errCasted, ok := err.(blocker.Blocker)
					if ok {
						errorMsg = errCasted.Message
					}
					if tt.errMsg != errorMsg {
						t.Errorf("error differs from what expected. wrong test exit line. gotErr %s, wantErr %s",
							err.Error(),
							tt.errMsg)
					}
					return
				}
				t.Errorf("NodeRegistrationService.ValidateNodeAddressInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type (
	mockGenerateNodeAddressInfoMainBlockStateStorageSuccess struct {
		storage.CacheStorageInterface
	}
)

func (*mockGenerateNodeAddressInfoMainBlockStateStorageSuccess) GetItem(lastChange, item interface{}) error {
	var blockCopy, _ = item.(*model.Block)
	var mockLastGoodBlock = mockGoodBlock
	mockLastGoodBlock.Height = uint32(1000)
	*blockCopy = *mockLastGoodBlock
	return nil
}

func TestNodeRegistrationService_GenerateNodeAddressInfo(t *testing.T) {
	type fields struct {
		QueryExecutor                query.ExecutorInterface
		AccountBalanceQuery          query.AccountBalanceQueryInterface
		NodeRegistrationQuery        query.NodeRegistrationQueryInterface
		ParticipationScoreQuery      query.ParticipationScoreQueryInterface
		BlockQuery                   query.BlockQueryInterface
		NodeAdmittanceCycle          uint32
		Logger                       *log.Logger
		ScrambledNodes               map[uint32]*model.ScrambledNodes
		MemoizedLatestScrambledNodes *model.ScrambledNodes
		BlockchainStatusService      BlockchainStatusServiceInterface
		CurrentNodePublicKey         []byte
		Signature                    crypto.SignatureInterface
		NodeAddressInfoService       NodeAddressInfoServiceInterface
		MainBlockStateStorage        storage.CacheStorageInterface
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
			name: "GenerateNodeAddressInfo:success",
			args: args{
				nodeID:           int64(111),
				nodeAddress:      "127.0.0.1",
				port:             uint32(8080),
				nodeSecretPhrase: "shhhhhhh",
			},
			fields: fields{
				QueryExecutor: &nrsMockQueryExecutorSuccess{},
				BlockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
				Signature: &validateNodeAddressInfoSignatureMock{
					isValid: true,
				},
				NodeAddressInfoService: &nodeAddressInfoServiceMock{
					nodeAddressInfoBytes: make([]byte, 64),
				},
				Logger:                log.New(),
				MainBlockStateStorage: &mockGenerateNodeAddressInfoMainBlockStateStorageSuccess{},
			},
			want: &model.NodeAddressInfo{
				NodeID:      int64(111),
				Address:     "127.0.0.1",
				Port:        uint32(8080),
				BlockHeight: uint32(280),
				BlockHash:   make([]byte, 32),
				Signature:   make([]byte, 64),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrs := &NodeRegistrationService{
				QueryExecutor:                tt.fields.QueryExecutor,
				AccountBalanceQuery:          tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery:        tt.fields.NodeRegistrationQuery,
				ParticipationScoreQuery:      tt.fields.ParticipationScoreQuery,
				BlockQuery:                   tt.fields.BlockQuery,
				Logger:                       tt.fields.Logger,
				ScrambledNodes:               tt.fields.ScrambledNodes,
				MemoizedLatestScrambledNodes: tt.fields.MemoizedLatestScrambledNodes,
				BlockchainStatusService:      tt.fields.BlockchainStatusService,
				CurrentNodePublicKey:         tt.fields.CurrentNodePublicKey,
				Signature:                    tt.fields.Signature,
				NodeAddressInfoService:       tt.fields.NodeAddressInfoService,
				MainBlockStateStorage:        tt.fields.MainBlockStateStorage,
			}
			got, err := nrs.GenerateNodeAddressInfo(tt.args.nodeID, tt.args.nodeAddress, tt.args.port, tt.args.nodeSecretPhrase)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistrationService.GenerateNodeAddressInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeRegistrationService.GenerateNodeAddressInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
