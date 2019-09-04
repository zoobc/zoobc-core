package service

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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
		NodeAddress:        "10.10.10.10",
		LockedBalance:      100000000,
		Queued:             true,
		Latest:             true,
		Height:             100,
	}
	nrsRegisteredNode1 = &model.NodeRegistration{
		NodeID:             int64(1),
		NodePublicKey:      nrsNodePubKey1,
		AccountAddress:     nrsAddress1,
		RegistrationHeight: 10,
		NodeAddress:        "10.10.10.10",
		LockedBalance:      100000000,
		Queued:             false,
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
		BlocksmithAddress:    "BCZ",
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

func (*nrsMockQueryExecutorFailNodeRegistryListener) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("MockedError")
}

func (*nrsMockQueryExecutorFailNoNodeRegistered) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance," +
		" queued, latest, height FROM node_registry WHERE locked_balance > 0 AND latest=1 ORDER BY locked_balance DESC LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"queued",
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

func (*nrsMockQueryExecutorSuccess) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT id, node_public_key, account_address, registration_height, node_address, locked_balance," +
		" queued, latest, height FROM node_registry WHERE locked_balance > 0 AND latest=1 ORDER BY locked_balance DESC LIMIT 1":
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"node_address",
			"locked_balance",
			"queued",
			"latest",
			"height",
		},
		).AddRow(1, nrsNodePubKey1, nrsAddress1, 10, "10.10.10.10", 100000000, true, true, 100))
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
		QueryExecutor         query.ExecutorInterface
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
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
				QueryExecutor:         &nrsMockQueryExecutorSuccess{},
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
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
				QueryExecutor:         &nrsMockQueryExecutorSuccess{},
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			args: args{
				nodeRegistrations: []*model.NodeRegistration{},
				height:            200,
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
			if err := nrs.AdmitNodes(tt.args.nodeRegistrations, tt.args.height); (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistrationService.AdmitNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeRegistrationService_KickOutNode(t *testing.T) {
	type fields struct {
		QueryExecutor         query.ExecutorInterface
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
	}
	type args struct {
		nodeRegistration *model.NodeRegistration
		height           uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "KickOutNode:success",
			fields: fields{
				QueryExecutor:         &nrsMockQueryExecutorSuccess{},
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			},
			args: args{
				nodeRegistration: nrsRegisteredNode1,
				height:           300,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrs := &NodeRegistrationService{
				QueryExecutor:         tt.fields.QueryExecutor,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
			}
			if err := nrs.KickOutNode(tt.args.nodeRegistration, tt.args.height); (err != nil) != tt.wantErr {
				t.Errorf("NodeRegistrationService.KickOutNode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeRegistrationService_NodeRegistryListener(t *testing.T) {
	type (
		fields struct {
			QueryExecutor         query.ExecutorInterface
			AccountBalanceQuery   query.AccountBalanceQueryInterface
			NodeRegistrationQuery query.NodeRegistrationQueryInterface
			NodeAdmittanceCycle   uint32
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
				QueryExecutor:         &nrsMockQueryExecutorSuccess{},
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				NodeAdmittanceCycle:   blockAdmittanceHeight1,
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
				QueryExecutor:         &nrsMockQueryExecutorFailNodeRegistryListener{},
				AccountBalanceQuery:   query.NewAccountBalanceQuery(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				NodeAdmittanceCycle:   blockAdmittanceHeight1 + 1,
			},
			args: args{
				block: nrsBlock1,
			},
			want: observer.Listener{
				OnNotify: func(data interface{}, args interface{}) {

				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nrs := &NodeRegistrationService{
				QueryExecutor:         tt.fields.QueryExecutor,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				NodeAdmittanceCycle:   tt.fields.NodeAdmittanceCycle,
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
