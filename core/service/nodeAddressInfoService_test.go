package service

import (
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"reflect"
	"regexp"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	nrcuMockQueryExecutor struct {
		query.Executor
	}
)

func (*nrcuMockQueryExecutor) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	switch qe {
	case "SELECT id, node_public_key, account_address, registration_height, t2.address || ':' || t2.port AS node_address, locked_balance, " +
		"registration_status, latest, height, t2.status as ai_status FROM node_registry " +
		"INNER JOIN node_address_info AS t2 ON id = t2.node_id WHERE registration_status = 0 AND (id,height) in " +
		"(SELECT t1.id,MAX(t1.height) FROM node_registry AS t1 WHERE t1.height <= 10 GROUP BY t1.id) GROUP BY id ORDER BY t2.status":
		mockedRows := sqlmock.NewRows([]string{
			"id",
			"node_public_key",
			"account_address",
			"registration_height",
			"t2.address || ':' || t2.port AS node_address",
			"locked_balance",
			"registration_status",
			"latest",
			"height",
			// TODO: add these fields when dropping address field from node_registry table
			// "t2.address AS ai_Address",
			// "t2.port AS ai_Port",
			"t2.status as ai_status",
		})
		mockedRows.AddRow(
			int64(111),
			[]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			uint32(0),
			"127.0.0.1:3000",
			10000000000,
			model.NodeRegistrationState_NodeRegistered,
			true,
			10,
			model.NodeAddressStatus_NodeAddressConfirmed,
		)
		mockedRows.AddRow(
			int64(111),
			[]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			uint32(0),
			"127.0.0.2:4000",
			10000000000,
			model.NodeRegistrationState_NodeRegistered,
			true,
			11,
			model.NodeAddressStatus_NodeAddressPending,
		)
		mockedRows.AddRow(
			int64(222),
			[]byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
			"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjAAA",
			uint32(0),
			"127.0.0.3:5000",
			10000000000,
			model.NodeRegistrationState_NodeRegistered,
			true,
			8,
			model.NodeAddressStatus_NodeAddressPending,
		)
		mockedRows.AddRow(
			int64(333),
			[]byte{3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3},
			"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjBBB",
			uint32(0),
			"127.0.0.4:6000",
			10000000000,
			model.NodeRegistrationState_NodeRegistered,
			true,
			18,
			model.NodeAddressStatus_NodeAddressPending,
		)
		mockedRows.AddRow(
			int64(333),
			[]byte{3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3},
			"BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjBBB",
			uint32(0),
			"127.0.0.4:7000",
			10000000000,
			model.NodeRegistrationState_NodeRegistered,
			true,
			20,
			model.NodeAddressStatus_NodeAddressConfirmed,
		)
		mock.ExpectQuery(regexp.QuoteMeta(qe)).WillReturnRows(mockedRows)
	default:
		return nil, errors.New("InvalidQuery")
	}

	rows, _ := db.Query(qe)
	return rows, nil
}

func TestNodeAddressInfoService_GetRegisteredNodesWithConsolidatedAddresses(t *testing.T) {
	type fields struct {
		QueryExecutor         query.ExecutorInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		Logger                *log.Logger
	}
	type args struct {
		height          uint32
		preferredStatus model.NodeAddressStatus
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*model.NodeRegistration
		wantErr bool
	}{
		{
			name: "GetRegisteredNodesWithConsolidatedAddresses:preferPending",
			fields: fields{
				Logger:                log.New(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor:         &nrcuMockQueryExecutor{},
			},
			args: args{
				height:          10,
				preferredStatus: model.NodeAddressStatus_NodeAddressPending,
			},
			want: []*model.NodeRegistration{
				{
					NodeID:         int64(111),
					NodePublicKey:  []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
					AccountAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
					NodeAddress: &model.NodeAddress{
						Address: "127.0.0.2",
						Port:    uint32(4000),
					},
					LockedBalance: int64(10000000000),
					Latest:        true,
					Height:        uint32(11),
					NodeAddressInfo: &model.NodeAddressInfo{
						Status: model.NodeAddressStatus_NodeAddressPending,
					},
				},
				{
					NodeID:         int64(222),
					NodePublicKey:  []byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
					AccountAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjAAA",
					NodeAddress: &model.NodeAddress{
						Address: "127.0.0.3",
						Port:    uint32(5000),
					},
					LockedBalance: int64(10000000000),
					Latest:        true,
					Height:        uint32(8),
					NodeAddressInfo: &model.NodeAddressInfo{
						Status: model.NodeAddressStatus_NodeAddressPending,
					},
				},
				{
					NodeID:         int64(333),
					NodePublicKey:  []byte{3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3},
					AccountAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjBBB",
					NodeAddress: &model.NodeAddress{
						Address: "127.0.0.4",
						Port:    uint32(6000),
					},
					LockedBalance: int64(10000000000),
					Latest:        true,
					Height:        uint32(18),
					NodeAddressInfo: &model.NodeAddressInfo{
						Status: model.NodeAddressStatus_NodeAddressPending,
					},
				},
			},
		},
		{
			name: "GetRegisteredNodesWithConsolidatedAddresses:preferConfirmed",
			fields: fields{
				Logger:                log.New(),
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor:         &nrcuMockQueryExecutor{},
			},
			args: args{
				height:          10,
				preferredStatus: model.NodeAddressStatus_NodeAddressConfirmed,
			},
			want: []*model.NodeRegistration{
				{
					NodeID:         int64(111),
					NodePublicKey:  []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
					AccountAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
					NodeAddress: &model.NodeAddress{
						Address: "127.0.0.1",
						Port:    uint32(3000),
					},
					LockedBalance: int64(10000000000),
					Latest:        true,
					Height:        uint32(10),
					NodeAddressInfo: &model.NodeAddressInfo{
						Status: model.NodeAddressStatus_NodeAddressConfirmed,
					},
				},
				{
					NodeID:         int64(222),
					NodePublicKey:  []byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
					AccountAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjAAA",
					NodeAddress: &model.NodeAddress{
						Address: "127.0.0.3",
						Port:    uint32(5000),
					},
					LockedBalance: int64(10000000000),
					Latest:        true,
					Height:        uint32(8),
					NodeAddressInfo: &model.NodeAddressInfo{
						Status: model.NodeAddressStatus_NodeAddressPending,
					},
				},
				{
					NodeID:         int64(333),
					NodePublicKey:  []byte{3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3},
					AccountAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjBBB",
					NodeAddress: &model.NodeAddress{
						Address: "127.0.0.4",
						Port:    uint32(7000),
					},
					LockedBalance: int64(10000000000),
					Latest:        true,
					Height:        uint32(20),
					NodeAddressInfo: &model.NodeAddressInfo{
						Status: model.NodeAddressStatus_NodeAddressConfirmed,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nru := &NodeAddressInfoService{
				QueryExecutor:         tt.fields.QueryExecutor,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				Logger:                tt.fields.Logger,
			}
			got, err := nru.GetRegisteredNodesWithConsolidatedAddresses(tt.args.height, tt.args.preferredStatus)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeAddressInfoService.GetRegisteredNodesWithConsolidatedAddresses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			var found bool
			for _, gotNai := range got {
				for _, wantNai := range tt.want {
					if reflect.DeepEqual(gotNai, wantNai) {
						found = true
						break
					}
				}
			}
			if !found {
				t.Errorf("NodeAddressInfoService.GetRegisteredNodesWithConsolidatedAddresses() = %v, want %v", got, tt.want)
			}

		})
	}
}
