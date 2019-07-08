// service package serve as service layer for our api
// business logic on fetching data, processing information will be processed in this package.
package service

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

// ResetBlockService resets the singleton back to nil, used in test case teardown
func ResetBlockService() {
	blockServiceInstance = nil
}

func TestNewBlockervice(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error while opening database connection")
	}
	// defer db.Close()

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

func Test_BlockService_GetBlocks(t *testing.T) {
	mockData := struct {
		BlockSize   uint32
		BlockHeight uint32
		Blocks      []*model.Block
	}{
		BlockSize:   2,
		BlockHeight: 0,
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
				Timestamp:            11000,
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
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error while opening database connection")
	}
	defer db.Close()
	instance := NewBlockService(query.NewQueryExecutor(db))
	defer ResetBlockService()
	tests := []struct {
		name    string
		bs      *BlockService
		want    *model.GetBlocksResponse
		wantErr bool
	}{
		{
			name: "GetBlocks:success",
			bs:   instance,
			want: &model.GetBlocksResponse{
				Blocks:      mockData.Blocks,
				BlockHeight: mockData.BlockHeight,
				BlockSize:   2,
			},
			wantErr: false,
		},
	}

	chainType := chaintype.GetChainType(0)
	blockQuery := query.NewBlockQuery(chainType)
	queryStr := blockQuery.GetBlocks(mockData.BlockHeight, mockData.BlockSize)

	mock.ExpectQuery(queryStr).WillReturnRows(sqlmock.NewRows([]string{"ID", "PreviousBlockHash", "Height", "Timestamp", "BlockSeed", "BlockSignature", "CumulativeDifficulty",
		"SmithScale", "PayloadLength", "PayloadHash", "BlocksmithID", "TotalAmount", "TotalFee", "TotalCoinBase", "Version",
	}).AddRow(1, []byte{}, 1, 10000, []byte{}, []byte{}, "", 1, 2, []byte{}, []byte{}, 0, 0, 0, 1).AddRow(1, []byte{}, 2, 11000, []byte{}, []byte{}, "", 1, 2, []byte{}, []byte{}, 0, 0, 0, 1))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := instance.GetBlocks(chainType, mockData.BlockSize, mockData.BlockHeight)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockService.GetBlocks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockService.GetBlocks() = %v, want %v", got, tt.want)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func Test_BlockService_GetBlockByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error while opening database connection")
	}
	defer db.Close()
	var bl model.Block
	if err != nil {
		panic(err)
	}
	instance := NewBlockService(query.NewQueryExecutor(db))
	defer ResetBlockService()

	tests := []struct {
		name    string
		bs      *BlockService
		want    *model.Block
		wantErr bool
	}{
		{
			name:    "GetBlockByID:success",
			bs:      instance,
			want:    &bl,
			wantErr: false,
		},
	}

	chainType := chaintype.GetChainType(0)
	blockQuery := query.NewBlockQuery(chainType)
	queryStr := blockQuery.GetBlockByID(0)
	mock.ExpectQuery(regexp.QuoteMeta(queryStr)).
		WillReturnRows(sqlmock.NewRows(blockQuery.Fields))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := instance.GetBlockByID(chainType, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockService.GetBlockByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockService.GetBlockByID() = %v, want %v", got, tt.want)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func Test_BlockService_GetBlockByHeight(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error while opening database connection")
	}
	defer db.Close()
	var bl model.Block
	instance := NewBlockService(query.NewQueryExecutor(db))
	defer ResetBlockService()
	tests := []struct {
		name    string
		bs      *BlockService
		want    *model.Block
		wantErr bool
	}{
		{
			name:    "GetBlockByHeight:success",
			bs:      instance,
			want:    &bl,
			wantErr: false,
		},
	}

	chainType := chaintype.GetChainType(0)
	blockQuery := query.NewBlockQuery(chainType)
	queryStr := blockQuery.GetBlockByHeight(0)
	mock.ExpectQuery(regexp.QuoteMeta(queryStr)).
		WillReturnRows(sqlmock.NewRows(blockQuery.Fields))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := instance.GetBlockByHeight(chainType, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("BlockService.GetBlockByHeight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockService.GetBlockByHeight() = %v, want %v", got, tt.want)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
