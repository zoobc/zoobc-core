package service

import (
	"bytes"
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	// GetBlocksmithAccountAddress mocks
	mockExecutorGetBlocksmithAccountAddressExecuteSelectFail struct {
		query.Executor
	}

	mockGetBlocksmithAccountAddressExecutorSuccess struct {
		query.Executor
	}
	mockGetBlocksmithAccountAddressNodeRegistrationFail struct {
		query.NodeRegistrationQuery
	}
	mockGetBlocksmithAccountAddressNodeRegistrationQueryBuildNoRows struct {
		query.NodeRegistrationQuery
	}
	mockGetBlocksmithAccountAddressNodeRegistrationQueryBuildOneRows struct {
		query.NodeRegistrationQuery
	}
	// GetBlocksmithAccountAddress mocks
	// RewardBlocksmithAccountAddresses mocks
	mockRewardBlocksmithAccountAddressesExecutorFail struct {
		query.Executor
	}
	mockRewardBlocksmithAccountAddressesExecutorSuccess struct {
		query.Executor
	}
	// RewardBlocksmithAccountAddresses mocks
)

var (
	// GetBlocksmithAccountAddress mocks
	mockGetBlocksmithAccountAddressNodeRegistry = &model.NodeRegistration{AccountAddress: []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88,
		220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169}}
)

func (*mockExecutorGetBlocksmithAccountAddressExecuteSelectFail) ExecuteSelect(
	qe string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	return nil, errors.New("mockedError")
}

func (*mockGetBlocksmithAccountAddressExecutorSuccess) ExecuteSelect(
	qe string, tx bool, args ...interface{},
) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta("MOCKQUERY")).WillReturnRows(sqlmock.NewRows([]string{
		"dummyColumn"}).AddRow(
		[]byte{1}))
	rows, _ := db.Query("MOCKQUERY")
	return rows, nil
}

func (*mockGetBlocksmithAccountAddressNodeRegistrationFail) BuildModel(
	nodeRegistrations []*model.NodeRegistration,
	rows *sql.Rows,
) ([]*model.NodeRegistration, error) {
	return nil, errors.New("mockedError")
}

func (*mockGetBlocksmithAccountAddressNodeRegistrationQueryBuildNoRows) BuildModel(
	nodeRegistrations []*model.NodeRegistration,
	rows *sql.Rows,
) ([]*model.NodeRegistration, error) {
	return []*model.NodeRegistration{}, nil
}

func (*mockGetBlocksmithAccountAddressNodeRegistrationQueryBuildOneRows) BuildModel(
	nodeRegistrations []*model.NodeRegistration,
	rows *sql.Rows,
) ([]*model.NodeRegistration, error) {
	return []*model.NodeRegistration{
		mockGetBlocksmithAccountAddressNodeRegistry,
	}, nil
}

func (*mockRewardBlocksmithAccountAddressesExecutorFail) ExecuteTransactions(queries [][]interface{}) error {
	return errors.New("mockedError")
}

func (*mockRewardBlocksmithAccountAddressesExecutorSuccess) ExecuteTransactions(queries [][]interface{}) error {
	return nil
}

func TestBlocksmithService_GetBlocksmithAccountAddress(t *testing.T) {
	type fields struct {
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
		QueryExecutor         query.ExecutorInterface
	}
	type args struct {
		block *model.Block
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "GetBlocksmithAccountAddress-ExecuteSelectFail",
			fields: fields{
				NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
				QueryExecutor:         &mockExecutorGetBlocksmithAccountAddressExecuteSelectFail{},
			},
			args: args{
				block: &model.Block{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetBlocksmithAccountAddress-BuildModelFail-IncorrectColumn",
			fields: fields{
				NodeRegistrationQuery: &mockGetBlocksmithAccountAddressNodeRegistrationFail{},
				QueryExecutor:         &mockGetBlocksmithAccountAddressExecutorSuccess{},
			},
			args: args{
				block: &model.Block{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetBlocksmithAccountAddress-VersionedNodeRegistrationNotFound",
			fields: fields{
				NodeRegistrationQuery: &mockGetBlocksmithAccountAddressNodeRegistrationQueryBuildNoRows{},
				QueryExecutor:         &mockGetBlocksmithAccountAddressExecutorSuccess{},
			},
			args: args{
				block: &model.Block{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetBlocksmithAccountAddress-VersionedNodeRegistrationNotFound",
			fields: fields{
				NodeRegistrationQuery: &mockGetBlocksmithAccountAddressNodeRegistrationQueryBuildOneRows{},
				QueryExecutor:         &mockGetBlocksmithAccountAddressExecutorSuccess{},
			},
			args: args{
				block: &model.Block{},
			},
			want:    mockGetBlocksmithAccountAddressNodeRegistry.AccountAddress,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlocksmithService{
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
				QueryExecutor:         tt.fields.QueryExecutor,
			}
			got, err := bs.GetBlocksmithAccountAddress(tt.args.block)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBlocksmithAccountAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("GetBlocksmithAccountAddress() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockService_RewardBlocksmithAccountAddresses(t *testing.T) {
	type fields struct {
		QueryExecutor         query.ExecutorInterface
		AccountLedgerQuery    query.AccountLedgerQueryInterface
		AccountBalanceQuery   query.AccountBalanceQueryInterface
		NodeRegistrationQuery query.NodeRegistrationQueryInterface
	}
	type args struct {
		blocksmithAccountAddresses [][]byte
		totalReward                int64
		timestamp                  int64
		height                     uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "RewardBlocksmithAccountAddress:NoAccountToBeRewarded",
			args: args{
				blocksmithAccountAddresses: [][]byte{},
				totalReward:                10000,
				timestamp:                  1578549075,
				height:                     1,
			},
			fields: fields{
				QueryExecutor:       nil,
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountLedgerQuery:  query.NewAccountLedgerQuery(),
			},
			wantErr: true,
		},
		{
			name: "RewardBlocksmithAccountAddress:executorFailExecuteTransactions",
			args: args{
				blocksmithAccountAddresses: [][]byte{bcsAddress1},
				totalReward:                10000,
				timestamp:                  1578549075,
				height:                     1,
			},
			fields: fields{
				QueryExecutor:       &mockRewardBlocksmithAccountAddressesExecutorFail{},
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountLedgerQuery:  query.NewAccountLedgerQuery(),
			},
			wantErr: true,
		},
		{
			name: "RewardBlocksmithAccountAddress:success",
			args: args{
				blocksmithAccountAddresses: [][]byte{bcsAddress1},
				totalReward:                10000,
				timestamp:                  1578549075,
				height:                     1,
			},
			fields: fields{
				QueryExecutor:       &mockRewardBlocksmithAccountAddressesExecutorSuccess{},
				AccountBalanceQuery: query.NewAccountBalanceQuery(),
				AccountLedgerQuery:  query.NewAccountLedgerQuery(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlocksmithService{
				QueryExecutor:         tt.fields.QueryExecutor,
				AccountLedgerQuery:    tt.fields.AccountLedgerQuery,
				AccountBalanceQuery:   tt.fields.AccountBalanceQuery,
				NodeRegistrationQuery: tt.fields.NodeRegistrationQuery,
			}
			if err := bs.RewardBlocksmithAccountAddresses(
				tt.args.blocksmithAccountAddresses,
				tt.args.totalReward,
				tt.args.timestamp,
				tt.args.height,
			); (err != nil) != tt.wantErr {
				t.Errorf("BlockService.RewardBlocksmithAccountAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewBlocksmithService(t *testing.T) {
	type args struct {
		accountBalanceQuery   query.AccountBalanceQueryInterface
		accountLedgerQuery    query.AccountLedgerQueryInterface
		nodeRegistrationQuery query.NodeRegistrationQueryInterface
		queryExecutor         query.ExecutorInterface
		chaintype             chaintype.ChainType
	}
	tests := []struct {
		name string
		args args
		want *BlocksmithService
	}{
		{
			name: "NewBlocksmithServiceSuccess",
			args: args{
				accountLedgerQuery:    nil,
				accountBalanceQuery:   nil,
				nodeRegistrationQuery: nil,
				queryExecutor:         nil,
			},
			want: &BlocksmithService{
				AccountBalanceQuery:   nil,
				AccountLedgerQuery:    nil,
				NodeRegistrationQuery: nil,
				QueryExecutor:         nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlocksmithService(tt.args.accountBalanceQuery, tt.args.accountLedgerQuery,
				tt.args.nodeRegistrationQuery, tt.args.queryExecutor, tt.args.chaintype); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlocksmithService() = %v, want %v", got, tt.want)
			}
		})
	}
}
