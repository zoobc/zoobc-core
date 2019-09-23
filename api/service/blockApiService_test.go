// service package serve as service layer for our api
// business logic on fetching data, processing information will be processed in this package.
package service

import (
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	coreService "github.com/zoobc/zoobc-core/core/service"
)

// ResetBlockService resets the singleton back to nil, used in test case teardown
func ResetBlockService() {
	blockServiceInstance = nil
}

var (
	basNodePubKey1 = []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
		45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135}
)

type (
	mockBlockService struct {
		coreService.BlockService
	}
	mockQueryExecutorBlockByIDFail struct {
		query.Executor
	}

	mockQueryExecutorBlockByIDNotFound struct {
		query.Executor
	}

	mockQueryExecutorGetBlocksSuccess struct {
		query.Executor
	}

	mockQueryExecutorGetBlocksFail struct {
		query.Executor
	}
)

func (*mockBlockService) GetBlockExtendedInfo(block *model.Block) (*model.BlockExtendedInfo, error) {
	return &model.BlockExtendedInfo{}, nil
}

func (*mockQueryExecutorBlockByIDFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("mockError:executeSelectFail")
}

func (*mockQueryExecutorBlockByIDNotFound) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{
		"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
		"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithPublicKey", "TotalAmount", "TotalFee", "TotalCoinBase",
		"Version"}))
	rows, _ := db.Query(qe)
	return rows, nil
}

func (*mockQueryExecutorGetBlocksSuccess) ExecuteSelect(qe string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{
		"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
		"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithPublicKey", "TotalAmount", "TotalFee", "TotalCoinBase",
		"Version"}).AddRow(1, []byte{}, 1, 10000, []byte{}, []byte{}, "", 1, 2, []byte{}, []byte{}, 0, 0, 0, 1).AddRow(1,
		[]byte{}, 2, 10000, []byte{}, []byte{}, "", 1, 2, []byte{}, []byte{}, 0, 0, 0, 1))
	rows, _ := db.Query(qe)
	return rows, nil
}

func (*mockQueryExecutorGetBlocksFail) ExecuteSelect(query string, tx bool, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("mockError:executeSelectFail")
}

type (
	mockQueryGetBlockByIDSuccess struct {
		query.Executor
	}
)

func (*mockQueryGetBlockByIDSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	blockQ := query.NewBlockQuery(&chaintype.MainChain{})

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(
		sqlmock.NewRows(blockQ.Fields).AddRow(
			1,
			[]byte{1},
			1,
			10000,
			[]byte{2},
			[]byte{3},
			"cumulative",
			1,
			1,
			[]byte{4},
			basNodePubKey1,
			1,
			1,
			1,
			1,
		),
	)
	return db.Query(qStr)
}
func TestBlockService_GetBlockByID(t *testing.T) {
	type fields struct {
		Query             query.ExecutorInterface
		BlockCoreServices map[int32]coreService.BlockServiceInterface
	}
	type args struct {
		chainType chaintype.ChainType
		id        int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.BlockExtendedInfo
		wantErr bool
	}{
		{
			name: "GetBlockByID:success",
			fields: fields{
				Query: &mockQueryGetBlockByIDSuccess{},
				BlockCoreServices: map[int32]coreService.BlockServiceInterface{
					0: &mockBlockService{},
				},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				id:        1,
			},
			wantErr: false,
			want:    &model.BlockExtendedInfo{},
		},
		{
			name: "GetBlockByID:fail-{ExecuteSelectFail}",
			fields: fields{
				Query: &mockQueryExecutorBlockByIDFail{},
				BlockCoreServices: map[int32]coreService.BlockServiceInterface{
					0: &mockBlockService{},
				},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				id:        1,
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "GetBlockByID:fail-{Block.ID notfound}",
			fields: fields{
				Query: &mockQueryExecutorBlockByIDNotFound{},
				BlockCoreServices: map[int32]coreService.BlockServiceInterface{
					0: &mockBlockService{},
				},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				id:        1,
			},
			wantErr: true,
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Query:             tt.fields.Query,
				BlockCoreServices: tt.fields.BlockCoreServices,
			}
			got, err := bs.GetBlockByID(tt.args.chainType, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockService.GetBlockByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockService.GetBlockByID() got = \n%v, want = \n%v", got, tt.want)
			}
		})
	}
}

type (
	mockQueryGetBlockByHeightSuccess struct {
		query.Executor
	}
)

func (*mockQueryGetBlockByHeightSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	blockQ := query.NewBlockQuery(&chaintype.MainChain{})

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(
		sqlmock.NewRows(blockQ.Fields).AddRow(
			1,
			[]byte{1},
			1,
			10000,
			[]byte{2},
			[]byte{3},
			"cumulative",
			1,
			1,
			[]byte{4},
			basNodePubKey1,
			1,
			1,
			1,
			1,
		),
	)
	return db.Query(qStr)
}

func TestBlockService_GetBlockByHeight(t *testing.T) {
	type fields struct {
		Query             query.ExecutorInterface
		BlockCoreServices map[int32]coreService.BlockServiceInterface
	}
	type args struct {
		chainType chaintype.ChainType
		height    uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.BlockExtendedInfo
		wantErr bool
	}{
		{
			name: "GetBlockByHeight:success",
			fields: fields{
				Query: &mockQueryGetBlockByHeightSuccess{},
				BlockCoreServices: map[int32]coreService.BlockServiceInterface{
					0: &mockBlockService{},
				},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				height:    1,
			},
			wantErr: false,
			want:    &model.BlockExtendedInfo{},
		},
		{
			name: "GetBlockByHeight:fail-{ExecuteSelectFail}",
			fields: fields{
				Query: &mockQueryExecutorBlockByIDFail{},
				BlockCoreServices: map[int32]coreService.BlockServiceInterface{
					0: &mockBlockService{},
				},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				height:    1,
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "GetBlockByHeight:fail-{Block.ID notfound}",
			fields: fields{
				Query: &mockQueryExecutorBlockByIDNotFound{},
				BlockCoreServices: map[int32]coreService.BlockServiceInterface{
					0: &mockBlockService{},
				},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				height:    1,
			},
			wantErr: true,
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Query:             tt.fields.Query,
				BlockCoreServices: tt.fields.BlockCoreServices,
			}
			got, err := bs.GetBlockByHeight(tt.args.chainType, tt.args.height)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockService.GetBlockByHeight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockService.GetBlockByHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockQueryGetBlocksSuccess struct {
		query.Executor
	}
)

func (*mockQueryGetBlocksSuccess) ExecuteSelect(qStr string, tx bool, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	blockQ := query.NewBlockQuery(&chaintype.MainChain{})

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(
		sqlmock.NewRows(blockQ.Fields).AddRow(
			1,
			[]byte{1},
			1,
			10000,
			[]byte{2},
			[]byte{3},
			"cumulative",
			1,
			1,
			[]byte{4},
			basNodePubKey1,
			1,
			1,
			1,
			1,
		),
	)
	return db.Query(qStr)
}

func TestBlockService_GetBlocks(t *testing.T) {
	type fields struct {
		Query             query.ExecutorInterface
		BlockCoreServices map[int32]coreService.BlockServiceInterface
	}
	type args struct {
		chainType chaintype.ChainType
		blockSize uint32
		height    uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.GetBlocksResponse
		wantErr bool
	}{
		{
			name: "GetBlocks:success",
			fields: fields{
				Query: &mockQueryGetBlocksSuccess{},
				BlockCoreServices: map[int32]coreService.BlockServiceInterface{
					0: &mockBlockService{},
				},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				blockSize: 2,
				height:    1,
			},
			want: &model.GetBlocksResponse{
				Height: 1,
				Count:  1,
				Blocks: []*model.BlockExtendedInfo{{}},
			},
			wantErr: false,
		},
		{
			name: "GetBlocks:success",
			fields: fields{
				Query: &mockQueryExecutorGetBlocksFail{},
				BlockCoreServices: map[int32]coreService.BlockServiceInterface{
					0: &mockBlockService{},
				},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				blockSize: 2,
				height:    1,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs := &BlockService{
				Query:             tt.fields.Query,
				BlockCoreServices: tt.fields.BlockCoreServices,
			}
			got, err := bs.GetBlocks(tt.args.chainType, tt.args.blockSize, tt.args.height)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockService.GetBlocks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockService.GetBlocks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewBlockService(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error while opening database connection")
	}
	defer db.Close()

	type args struct {
		queryExecutor     query.ExecutorInterface
		blockCoreServices map[int32]coreService.BlockServiceInterface
	}
	tests := []struct {
		name string
		args args
		want *BlockService
	}{
		{
			name: "NewBlockService:InitiateBlockServiceInstance",
			args: args{
				queryExecutor: query.NewQueryExecutor(db),
				blockCoreServices: map[int32]coreService.BlockServiceInterface{
					0: &coreService.BlockService{},
				},
			},
			want: &BlockService{
				Query: query.NewQueryExecutor(db),
				BlockCoreServices: map[int32]coreService.BlockServiceInterface{
					0: &coreService.BlockService{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlockService(tt.args.queryExecutor, tt.args.blockCoreServices); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlockService() = %v, want %v", got, tt.want)
			}
		})
	}
}
