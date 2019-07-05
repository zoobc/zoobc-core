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
		BlockSize:   0,
		BlockHeight: 0,
		Blocks:      []*model.Block{},
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
				BlockSize:   uint32(len(mockData.Blocks)),
			},
			wantErr: false,
		},
	}

	chainType := chaintype.GetChainType(0)
	blockQuery := query.NewBlockQuery()
	queryStr := blockQuery.GetBlocks(chainType, 0)
	mock.ExpectQuery(queryStr).
		WillReturnRows(sqlmock.NewRows(blockQuery.Fields))

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
	blockQuery := query.NewBlockQuery()
	queryStr := blockQuery.GetBlockByID(chainType, 0)
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
	blockQuery := query.NewBlockQuery()
	queryStr := blockQuery.GetBlockByHeight(chainType, 0)
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
