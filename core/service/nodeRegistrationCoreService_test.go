package service

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	log "github.com/sirupsen/logrus"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
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
		NodeAddress: &model.NodeAddress{
			Address: "10.10.10.10",
		},
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
		NodeAddress: &model.NodeAddress{
			Address: "10.10.10.10",
		},
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
	case "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, registration_status, " +
		"latest, height FROM node_registry WHERE locked_balance > 0 AND registration_status = 1 AND latest=1 " +
		"ORDER BY locked_balance DESC LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		},
		))
	case "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, registration_status, " +
		"latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		},
		))
	case "SELECT A.id, A.node_public_key, A.account_address, A.registration_height, A.node_address, A.locked_balance, " +
		"A.registration_status, A.latest, A.height FROM node_registry as A INNER JOIN participation_score as B ON A.id = B.node_id " +
		"WHERE B.score = 0 AND A.latest=1 AND A.registration_status=0 AND B.latest=1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
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

func (*nrsMockQueryExecutorSuccess) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, registration_status, " +
		"latest, height FROM node_registry WHERE locked_balance > 0 AND registration_status = 1 AND latest=1 " +
		"ORDER BY locked_balance DESC LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		},
		).AddRow(1, nrsNodePubKey1, nrsAddress1, 10, "10.10.10.10", 100000000, uint32(model.NodeRegistrationState_NodeQueued), true, 100))
	case "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, registration_status, " +
		"latest, height FROM node_registry WHERE node_public_key = ? AND latest=1 ORDER BY height DESC LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		},
		).AddRow(1, nrsNodePubKey1, nrsAddress1, 10, "10.10.10.10", 100000000, uint32(model.NodeRegistrationState_NodeQueued), true, 100))
	case "SELECT A.id, A.node_public_key, A.account_address, A.registration_height, A.node_address, A.locked_balance, " +
		"A.registration_status, A.latest, A.height FROM node_registry as A INNER JOIN participation_score as B ON A.id = B.node_id " +
		"WHERE B.score <= 0 AND A.latest=1 AND A.registration_status=0 AND B.latest=1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		},
		).AddRow(1, nrsNodePubKey1, nrsAddress1, 10, "10.10.10.10", 100000000, uint32(model.NodeRegistrationState_NodeQueued), true, 100))
	case "SELECT nr.id AS nodeID, nr.node_public_key AS node_public_key, ps.score AS participation_score " +
		"FROM node_registry AS nr INNER JOIN participation_score AS ps ON nr.id = ps.node_id " +
		"WHERE nr.registration_status = 0 AND nr.latest = 1 AND ps.score > 0 AND ps.latest = 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"participation_score",
		},
		).AddRow(1, nrsNodePubKey1, 8000))
	case "SELECT id, node_public_key, account_address, registration_height, node_address, " +
		"locked_balance, registration_status, latest, height " +
		"FROM node_registry where registration_status = 0 AND (id,height) in " +
		"(SELECT id,MAX(height) FROM node_registry WHERE height <= 1 GROUP BY id) " +
		"ORDER BY height DESC":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
		},
		).AddRow(1, nrsNodePubKey1, nrsAddress1, 10, "10.10.10.10", 100000000,
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

func TestNodeRegistrationService_ExpelNodes(t *testing.T) {
	type fields struct {
		QueryExecutor           query.ExecutorInterface
		AccountBalanceQuery     query.AccountBalanceQueryInterface
		NodeRegistrationQuery   query.NodeRegistrationQueryInterface
		ParticipationScoreQuery query.ParticipationScoreQueryInterface
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
				NodeAddress: &model.NodeAddress{
					Address: "10.10.10.10",
				},
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
					NodeAddress: &model.NodeAddress{
						Address: "10.10.10.10",
					},
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

func TestNodeRegistrationService_GetNodeRegistryAtHeight(t *testing.T) {
	type fields struct {
		QueryExecutor         query.ExecutorInterface
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
	}
	type args struct {
		height uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.NodeRegistration
		wantErr bool
	}{
		{
			name: "GetNodeRegistryAtHeight:success",
			fields: fields{
				QueryExecutor:         &nrsMockQueryExecutorSuccess{},
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			args: args{
				height: 1,
			},
			want: []*model.NodeRegistration{
				{
					NodeID:             int64(1),
					NodePublicKey:      nrsNodePubKey1,
					AccountAddress:     nrsAddress1,
					RegistrationHeight: 10,
					NodeAddress: &model.NodeAddress{
						Address: "10.10.10.10",
					},
					LockedBalance:      100000000,
					RegistrationStatus: uint32(model.NodeRegistrationState_NodeRegistered),
					Latest:             true,
					Height:             200,
				},
			},
			wantErr: false,
		},
		{
			name: "GetNodeRegistryAtHeight:fail-{NoNodeRegistered}",
			fields: fields{
				QueryExecutor:         &nrsMockQueryExecutorFailNoNodeRegistered{},
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			args: args{
				height: 1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrs := &NodeRegistrationService{
				QueryExecutor:         tt.fields.QueryExecutor,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
			}
			got, err := nrs.GetNodeRegistryAtHeight(tt.args.height)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistrationService.GetNodeRegistryAtHeight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeRegistrationService.GetNodeRegistryAtHeight() = %v, want %v", got, tt.want)
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
)

func (*executorBuildScrambleNodesSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WithArgs(1).WillReturnRows(sqlmock.NewRows(
		query.NewNodeRegistrationQuery().Fields,
	).AddRow(
		0, []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 76}, "accountA", 0, "127.0.0.1:3000", 0, 0, true, 0,
	).AddRow(
		0, []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 78}, "accountB", 0, "127.0.0.1:3001", 0, 0, true, 0,
	))

	return db.Query(qStr, 1)
}

func (*executorBuildScrambleNodesFail) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("mockError:executeSelectFail")
}

func TestNodeRegistrationService_BuildScrambledNodes(t *testing.T) {
	db, mock, _ := sqlmock.New()
	type fields struct {
		QueryExecutor         query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		Logger                *log.Logger
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
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				Logger:                log.New(),
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
				QueryExecutor:         tt.fields.QueryExecutor,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				ScrambledNodes:        map[uint32]*model.ScrambledNodes{},
				Logger:                tt.fields.Logger,
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
	mockQueryExecutorGetNextNodeAdmissionTimestampFail struct {
		query.Executor
	}
	mockQueryExecutorGetNextNodeAdmissionTimestampSuccess struct {
		query.Executor
	}
	mockNodeAdmissionTimestampQueryGetNextNodeAdmissionTimestampFail struct {
		query.NodeAdmissionTimestampQuery
	}
	mockNodeAdmissionTimestampQueryGetNextNodeAdmissionTimestampSuccess struct {
		query.NodeAdmissionTimestampQuery
	}
)

func (*mockQueryExecutorGetNextNodeAdmissionTimestampFail) ExecuteSelectRow(
	query string, tx bool, args ...interface{},
) (*sql.Row, error) {
	return nil, errors.New("mockedError")
}
func (*mockQueryExecutorGetNextNodeAdmissionTimestampSuccess) ExecuteSelectRow(
	query string, tx bool, args ...interface{},
) (*sql.Row, error) {
	return nil, nil
}

func (*mockNodeAdmissionTimestampQueryGetNextNodeAdmissionTimestampFail) Scan(
	nextNodeAdmission *model.NodeAdmissionTimestamp, row *sql.Row,
) error {
	return errors.New("mockedError")
}
func (*mockNodeAdmissionTimestampQueryGetNextNodeAdmissionTimestampSuccess) Scan(
	nextNodeAdmission *model.NodeAdmissionTimestamp, row *sql.Row,
) error {
	return nil
}

func TestNodeRegistrationService_GetNextNodeAdmissionTimestamp(t *testing.T) {
	var (
		mockNodeAdmissionTimesatamp = model.NodeAdmissionTimestamp{
			Timestamp:   1,
			BlockHeight: 1,
		}
	)
	type fields struct {
		QueryExecutor                query.ExecutorInterface
		AccountBalanceQuery          query.AccountBalanceQueryInterface
		NodeRegistrationQuery        query.NodeRegistrationQueryInterface
		ParticipationScoreQuery      query.ParticipationScoreQueryInterface
		BlockQuery                   query.BlockQueryInterface
		NodeAdmissionTimestampQuery  query.NodeAdmissionTimestampQueryInterface
		NextNodeAdmission            *model.NodeAdmissionTimestamp
		Logger                       *log.Logger
		ScrambledNodes               map[uint32]*model.ScrambledNodes
		ScrambledNodesLock           sync.RWMutex
		MemoizedLatestScrambledNodes *model.ScrambledNodes
		BlockchainStatusService      BlockchainStatusServiceInterface
		CurrentNodePublicKey         []byte
	}
	type args struct {
		blockHeight uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "wantSuccess:NextNodeAdmissionNotNil",
			fields: fields{
				NextNodeAdmission: &mockNodeAdmissionTimesatamp,
			},
			args: args{
				blockHeight: mockNodeAdmissionTimesatamp.BlockHeight + 1,
			},
			want:    mockNodeAdmissionTimesatamp.Timestamp,
			wantErr: false,
		},
		{
			name: "wantFail:ExecuteSelectRow",
			fields: fields{
				QueryExecutor:               &mockQueryExecutorGetNextNodeAdmissionTimestampFail{},
				NodeAdmissionTimestampQuery: query.NewNodeAdmissionTimestampQuery(),
			},
			args: args{
				blockHeight: 1,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "wantFail:ScanQuery",
			fields: fields{
				QueryExecutor:               &mockQueryExecutorGetNextNodeAdmissionTimestampSuccess{},
				NodeAdmissionTimestampQuery: &mockNodeAdmissionTimestampQueryGetNextNodeAdmissionTimestampFail{},
			},
			args: args{
				blockHeight: 1,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "wantSuccess_2",
			fields: fields{
				QueryExecutor:               &mockQueryExecutorGetNextNodeAdmissionTimestampSuccess{},
				NodeAdmissionTimestampQuery: &mockNodeAdmissionTimestampQueryGetNextNodeAdmissionTimestampSuccess{},
			},
			args: args{
				blockHeight: 1,
			},
			want:    0,
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
				NextNodeAdmission:            tt.fields.NextNodeAdmission,
				Logger:                       tt.fields.Logger,
				ScrambledNodes:               tt.fields.ScrambledNodes,
				ScrambledNodesLock:           tt.fields.ScrambledNodesLock,
				MemoizedLatestScrambledNodes: tt.fields.MemoizedLatestScrambledNodes,
				BlockchainStatusService:      tt.fields.BlockchainStatusService,
				CurrentNodePublicKey:         tt.fields.CurrentNodePublicKey,
			}
			got, err := nrs.GetNextNodeAdmissionTimestamp(tt.args.blockHeight)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistrationService.GetNextNodeAdmissionTimestamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
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
	mockNodeAdmissionTimestampQueryInsertNextNodeAdmissionTimestampFail struct {
		query.NodeAdmissionTimestampQuery
	}
	mockNodeAdmissionTimestampQueryInsertNextNodeAdmissionTimestampSuccess struct {
		query.NodeAdmissionTimestampQuery
	}
)

func (*mockQueryExecutorInsertNextNodeAdmissionTimestampFail) ExecuteSelect(
	query string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	return nil, errors.New("mockedErr")
}

func (*mockQueryExecutorInsertNextNodeAdmissionTimestampSuccess) ExecuteSelect(
	query string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	return nil, nil
}

func (*mockQueryExecutorInsertNextNodeAdmissionTimestampSuccess) ExecuteTransactions(
	queries [][]interface{},
) error {
	return nil
}
func (*mockQueryExecutorInsertNextNodeAdmissionTimestampFailExecuteTransactions) ExecuteSelect(
	query string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	return nil, nil
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
		NextNodeAdmission            *model.NodeAdmissionTimestamp
		Logger                       *log.Logger
		ScrambledNodes               map[uint32]*model.ScrambledNodes
		ScrambledNodesLock           sync.RWMutex
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
				NextNodeAdmission:            tt.fields.NextNodeAdmission,
				Logger:                       tt.fields.Logger,
				ScrambledNodes:               tt.fields.ScrambledNodes,
				ScrambledNodesLock:           tt.fields.ScrambledNodesLock,
				MemoizedLatestScrambledNodes: tt.fields.MemoizedLatestScrambledNodes,
				BlockchainStatusService:      tt.fields.BlockchainStatusService,
				CurrentNodePublicKey:         tt.fields.CurrentNodePublicKey,
			}
			if err := nrs.InsertNextNodeAdmissionTimestamp(tt.args.lastAdmissionTimestamp, tt.args.blockHeight, tt.args.dbTx); (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistrationService.InsertNextNodeAdmissionTimestamp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
