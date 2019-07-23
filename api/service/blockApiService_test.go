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
	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

// ResetBlockService resets the singleton back to nil, used in test case teardown
func ResetBlockService() {
	blockServiceInstance = nil
}

type (
	mockQueryExecutorBlockByIDSuccess struct {
		query.Executor
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

func (*mockQueryExecutorBlockByIDSuccess) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{
		"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
		"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithID", "TotalAmount", "TotalFee", "TotalCoinBase",
		"Version"}).AddRow(1, []byte{}, 1, 10000, []byte{}, []byte{}, "", 1, 2, []byte{}, []byte{}, 0, 0, 0, 1))
	rows, _ := db.Query(qe)
	return rows, nil
}

func (*mockQueryExecutorBlockByIDFail) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("mockError:executeSelectFail")
}

func (*mockQueryExecutorBlockByIDNotFound) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{
		"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
		"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithID", "TotalAmount", "TotalFee", "TotalCoinBase",
		"Version"}))
	rows, _ := db.Query(qe)
	return rows, nil
}

func (*mockQueryExecutorGetBlocksSuccess) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{
		"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
		"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithID", "TotalAmount", "TotalFee", "TotalCoinBase",
		"Version"}).AddRow(1, []byte{}, 1, 10000, []byte{}, []byte{}, "", 1, 2, []byte{}, []byte{}, 0, 0, 0, 1).AddRow(1,
		[]byte{}, 2, 10000, []byte{}, []byte{}, "", 1, 2, []byte{}, []byte{}, 0, 0, 0, 1))
	rows, _ := db.Query(qe)
	return rows, nil
}

func (*mockQueryExecutorGetBlocksFail) ExecuteSelect(qe string, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("mockError:executeSelectFail")
}

func TestNewBlockervice(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error while opening database connection")
	}
	defer db.Close()

	tests := []struct {
		name string
		want *BlockService
	}{
		{
			name: "NewBlockService:InitiateBlockServiceInstance",
			want: &BlockService{Query: query.NewQueryExecutor(db)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlockService(query.NewQueryExecutor(db)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlockService() = %v, want %v", got, tt.want)
			}
			defer ResetBlockService()
		})
	}
}

func TestBlockService_GetBlockByID(t *testing.T) {
	type fields struct {
		Query query.ExecutorInterface
	}
	type args struct {
		chainType contract.ChainType
		id        int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Block
		wantErr bool
	}{
		{
			name: "GetBlockByID:success",
			fields: fields{
				Query: &mockQueryExecutorBlockByIDSuccess{},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				id:        1,
			},
			wantErr: false,
			want: &model.Block{
				ID:                   1,
				PreviousBlockHash:    []byte{},
				Height:               1,
				Timestamp:            10000,
				BlockSeed:            []byte{},
				BlockSignature:       []byte{},
				CumulativeDifficulty: "",
				SmithScale:           1,
				PayloadLength:        2,
				PayloadHash:          []byte{},
				BlocksmithID:         []byte{},
				TotalAmount:          0,
				TotalFee:             0,
				TotalCoinBase:        0,
				Version:              1,
			},
		},
		{
			name: "GetBlockByID:fail-{ExecuteSelectFail}",
			fields: fields{
				Query: &mockQueryExecutorBlockByIDFail{},
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
				Query: tt.fields.Query,
			}
			got, err := bs.GetBlockByID(tt.args.chainType, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockService.GetBlockByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockService.GetBlockByID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockService_GetBlockByHeight(t *testing.T) {
	type fields struct {
		Query query.ExecutorInterface
	}
	type args struct {
		chainType contract.ChainType
		height    uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Block
		wantErr bool
	}{
		{
			name: "GetBlockByHeight:success",
			fields: fields{
				Query: &mockQueryExecutorBlockByIDSuccess{},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				height:    1,
			},
			wantErr: false,
			want: &model.Block{
				ID:                   1,
				PreviousBlockHash:    []byte{},
				Height:               1,
				Timestamp:            10000,
				BlockSeed:            []byte{},
				BlockSignature:       []byte{},
				CumulativeDifficulty: "",
				SmithScale:           1,
				PayloadLength:        2,
				PayloadHash:          []byte{},
				BlocksmithID:         []byte{},
				TotalAmount:          0,
				TotalFee:             0,
				TotalCoinBase:        0,
				Version:              1,
			},
		},
		{
			name: "GetBlockByHeight:fail-{ExecuteSelectFail}",
			fields: fields{
				Query: &mockQueryExecutorBlockByIDFail{},
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
				Query: tt.fields.Query,
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

func TestBlockService_GetBlocks(t *testing.T) {
	type fields struct {
		Query query.ExecutorInterface
	}
	type args struct {
		chainType contract.ChainType
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
				Query: &mockQueryExecutorGetBlocksSuccess{},
			},
			args: args{
				chainType: &chaintype.MainChain{},
				blockSize: 2,
				height:    1,
			},
			want: &model.GetBlocksResponse{
				Height: 1,
				Count:  2,
				Blocks: []*model.Block{
					{
						ID:                   1,
						PreviousBlockHash:    []byte{},
						Height:               1,
						Timestamp:            10000,
						BlockSeed:            []byte{},
						BlockSignature:       []byte{},
						CumulativeDifficulty: "",
						SmithScale:           1,
						PayloadLength:        2,
						PayloadHash:          []byte{},
						BlocksmithID:         []byte{},
						TotalAmount:          0,
						TotalFee:             0,
						TotalCoinBase:        0,
						Version:              1,
					},
					{
						ID:                   1,
						PreviousBlockHash:    []byte{},
						Height:               2,
						Timestamp:            10000,
						BlockSeed:            []byte{},
						BlockSignature:       []byte{},
						CumulativeDifficulty: "",
						SmithScale:           1,
						PayloadLength:        2,
						PayloadHash:          []byte{},
						BlocksmithID:         []byte{},
						TotalAmount:          0,
						TotalFee:             0,
						TotalCoinBase:        0,
						Version:              1,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "GetBlocks:success",
			fields: fields{
				Query: &mockQueryExecutorGetBlocksFail{},
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
				Query: tt.fields.Query,
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
