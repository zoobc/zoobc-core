package service

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/observer"
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
		RegistrationStatus: uint32(constant.NodeQueued),
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
		RegistrationStatus: uint32(constant.NodeRegistered),
		Latest:             true,
		Height:             200,
	}
	blockAdmittanceHeight1 uint32 = 1440
	nrsBlock1                     = &model.Block{
		ID:                   0,
		Height:               blockAdmittanceHeight1,
		Version:              1,
		CumulativeDifficulty: "",
		SmithScale:           0,
		PreviousBlockHash:    []byte{},
		BlockSeed:            []byte{},
		BlocksmithPublicKey:  nrsNodePubKey1,
		Timestamp:            12345678,
		TotalAmount:          0,
		TotalFee:             0,
		TotalCoinBase:        0,
		Transactions:         []*model.Transaction{},
		PayloadHash:          []byte{},
		PayloadLength:        0,
		BlockSignature:       []byte{},
	}
	nrsBlock2 = &model.Block{
		ID:                   1000,
		Height:               blockAdmittanceHeight1,
		Version:              1,
		CumulativeDifficulty: "",
		SmithScale:           0,
		PreviousBlockHash:    []byte{},
		BlockSeed:            []byte{1, 1, 1, 1, 1, 1, 1, 1},
		BlocksmithPublicKey:  nrsNodePubKey1,
		Timestamp:            12345678,
		TotalAmount:          0,
		TotalFee:             0,
		TotalCoinBase:        0,
		Transactions:         []*model.Transaction{},
		PayloadHash:          []byte{},
		PayloadLength:        0,
		BlockSignature:       []byte{},
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
		"latest, height FROM node_registry WHERE node_public_key = ? AND latest=1":
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
		).AddRow(1, nrsNodePubKey1, nrsAddress1, 10, "10.10.10.10", 100000000, constant.NodeQueued, true, 100))
	case "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, registration_status, " +
		"latest, height FROM node_registry WHERE node_public_key = ? AND latest=1":
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
		).AddRow(1, nrsNodePubKey1, nrsAddress1, 10, "10.10.10.10", 100000000, constant.NodeQueued, true, 100))
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
		).AddRow(1, nrsNodePubKey1, nrsAddress1, 10, "10.10.10.10", 100000000, constant.NodeQueued, true, 100))
	case "SELECT nr.id AS nodeID, nr.node_public_key AS node_public_key, ps.score AS participation_score " +
		"FROM node_registry AS nr INNER JOIN participation_score AS ps ON nr.id = ps.node_id " +
		"WHERE account_address = " + constant.DeletedNodeAccountAddress +
		" AND nr.latest = 1 AND nr.registration_status = 0 AND ps.score > 0 AND ps.latest = 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"participation_score",
		},
		).AddRow(1, nrsNodePubKey1, 8000))
	case "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance, registration_status, " +
		"latest, height, max(height) AS max_height FROM node_registry where height <= 1 AND registration_status = 0 " +
		"GROUP BY id ORDER BY height DESC":
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
			"max_height",
		},
		).AddRow(1, nrsNodePubKey1, nrsAddress1, 10, "10.10.10.10", 100000000, constant.NodeRegistered, true, 200, 200))
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

func TestNodeRegistrationService_NodeRegistryListener(t *testing.T) {
	type (
		fields struct {
			QueryExecutor           query.ExecutorInterface
			AccountBalanceQuery     query.AccountBalanceQueryInterface
			NodeRegistrationQuery   query.NodeRegistrationQueryInterface
			ParticipationScoreQuery query.ParticipationScoreQueryInterface
			NodeAdmittanceCycle     uint32
		}
		args struct {
			block *model.Block
		}
	)
	tests := []struct {
		name   string
		fields fields
		args   args
		want   observer.Listener
	}{
		{
			name: "NodeRegistryListener:success",
			fields: fields{
				QueryExecutor:           &nrsMockQueryExecutorSuccess{},
				AccountBalanceQuery:     query.NewAccountBalanceQuery(),
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				NodeAdmittanceCycle:     blockAdmittanceHeight1,
			},
			args: args{
				block: nrsBlock1,
			},
			want: observer.Listener{
				OnNotify: func(data interface{}, args interface{}) {

				},
			},
		},
		{
			name: "NodeRegistryListener:success-{noAdmittanceBlock}",
			fields: fields{
				QueryExecutor:           &nrsMockQueryExecutorFailNodeRegistryListener{},
				AccountBalanceQuery:     query.NewAccountBalanceQuery(),
				NodeRegistrationQuery:   query.NewNodeRegistrationQuery(),
				ParticipationScoreQuery: query.NewParticipationScoreQuery(),
				NodeAdmittanceCycle:     blockAdmittanceHeight1 + 1,
			},
			args: args{
				block: nrsBlock1,
			},
			want: observer.Listener{
				OnNotify: func(data interface{}, args interface{}) {},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrs := &NodeRegistrationService{
				QueryExecutor:           tt.fields.QueryExecutor,
				AccountBalanceQuery:     tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				ParticipationScoreQuery: tt.fields.ParticipationScoreQuery,
				NodeAdmittanceCycle:     tt.fields.NodeAdmittanceCycle,
			}

			got := nrs.NodeRegistryListener()
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("NodeRegistrationService.NodeRegistryListener() = %v, want %v", got, tt.want)
			}
			testOnNotifyPushBlockListener(got.OnNotify, tt.args.block)
		})
	}
}

func testOnNotifyPushBlockListener(fn observer.OnNotify, block *model.Block) {
	fn(block, nil)
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
				NodeAdmittanceCycle:     tt.fields.NodeAdmittanceCycle,
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
				RegistrationStatus: uint32(constant.NodeQueued),
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
				NodeAdmittanceCycle:     tt.fields.NodeAdmittanceCycle,
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
					RegistrationStatus: uint32(constant.NodeQueued),
					Latest:             true,
					Height:             100,
				},
			},
			wantErr: false,
		},
		{
			name: "SelectNodesToBeExpelled:fail-{NoNodeRegistered}",
			fields: fields{
				QueryExecutor:         &nrsMockQueryExecutorFailNoNodeRegistered{},
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			wantErr: false,
			want:    []*model.NodeRegistration{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrs := &NodeRegistrationService{
				QueryExecutor:           tt.fields.QueryExecutor,
				AccountBalanceQuery:     tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery:   tt.fields.NodeRegistrationQuery,
				ParticipationScoreQuery: tt.fields.ParticipationScoreQuery,
				NodeAdmittanceCycle:     tt.fields.NodeAdmittanceCycle,
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
					RegistrationStatus: uint32(constant.NodeRegistered),
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
