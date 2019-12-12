package util

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/constant"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

var (
	mockBlockData = model.Block{
		Version:              uint32(1),
		PreviousBlockHash:    []byte{1, 2, 4, 5, 67, 89, 86, 3, 6, 22},
		BlockSeed:            []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 76},
		Timestamp:            int64(15875592),
		TotalAmount:          int64(0),
		TotalFee:             int64(0),
		TotalCoinBase:        int64(0),
		Transactions:         []*model.Transaction{},
		PayloadHash:          []byte{},
		CumulativeDifficulty: "355353517378119",
		SmithScale:           int64(48985),
		BlocksmithPublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
			45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
	}
)

func TestGetBlockByte(t *testing.T) {
	type args struct {
		block *model.Block
		sign  bool
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "GetBlockByte:one",
			args: args{
				block: &model.Block{
					Version:           1,
					PreviousBlockHash: []byte{1, 2, 4, 5, 67, 89, 86, 3, 6, 22},
					BlockSeed:         []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 76},
					BlocksmithPublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
						45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
					Timestamp:            15875592,
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Transactions:         []*model.Transaction{},
					PayloadHash:          []byte{},
					CumulativeDifficulty: "355353517378119",
					SmithScale:           48985,
				},
				sign: false,
			},
			want: []byte{1, 0, 0, 0, 8, 62, 242, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21,
				99, 125, 75, 49, 45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135, 2, 65, 76, 32, 76,
				12, 12, 34, 65, 76, 1, 2, 4, 5, 67, 89, 86, 3, 6, 22},
			wantErr: false,
		},
		{
			name: "GetBlockByte:withSignature",
			args: args{
				block: &model.Block{
					Version:           1,
					PreviousBlockHash: []byte{1, 2, 4, 5, 67, 89, 86, 3, 6, 22},
					BlockSeed:         []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 76},
					BlocksmithPublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
						45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
					Timestamp:            15875592,
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Transactions:         []*model.Transaction{},
					PayloadHash:          []byte{},
					CumulativeDifficulty: "355353517378119",
					BlockSignature:       []byte{1, 3, 4, 54, 65, 76, 3, 3, 54, 12, 5, 64, 23, 12, 21},
					SmithScale:           48985,
				},
				sign: true,
			},
			want: []byte{1, 0, 0, 0, 8, 62, 242, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199,
				145, 21, 99, 125, 75, 49, 45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135, 2, 65,
				76, 32, 76, 12, 12, 34, 65, 76, 1, 2, 4, 5, 67, 89, 86, 3, 6, 22, 1, 3, 4, 54, 65, 76, 3, 3, 54,
				12, 5, 64, 23, 12, 21},
			wantErr: false,
		},
		{
			name: "GetBlockByte:error-{sign true without signature}",
			args: args{
				block: &model.Block{
					Version:           1,
					PreviousBlockHash: []byte{1, 2, 4, 5, 67, 89, 86, 3, 6, 22},
					BlockSeed:         []byte{2, 65, 76, 32, 76, 12, 12, 34, 65, 76},
					BlocksmithPublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
						45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
					Timestamp:            15875592,
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Transactions:         []*model.Transaction{},
					PayloadHash:          []byte{},
					CumulativeDifficulty: "355353517378119",
					SmithScale:           48985,
				},
				sign: true,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBlockByte(tt.args.block, tt.args.sign)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBlockByte() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBlockByte() = %v, want %v", got, tt.want)
			}
		})
	}
}

type (
	mockExecutorGetLastBlockSuccess struct {
		query.Executor
	}
	mockExecutorGetLastBlockNoRow struct {
		query.Executor
	}
	mockExecutorGetLastBlockFail struct {
		query.Executor
	}
)

func (*mockExecutorGetLastBlockFail) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mockRows := mock.NewRows([]string{"fake"})
	mockRows.AddRow("1")
	mock.ExpectQuery(qStr).WillReturnRows(mockRows)
	return db.QueryRow(qStr), nil
}

func (*mockExecutorGetLastBlockNoRow) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mockRows := mock.NewRows(query.NewBlockQuery(chaintype.GetChainType(0)).Fields)
	mock.ExpectQuery(qStr).WillReturnRows(mockRows)
	return db.QueryRow(qStr), nil
}

func (*mockExecutorGetLastBlockSuccess) ExecuteSelectRow(qStr string, tx bool, args ...interface{}) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	mockRows := mock.NewRows(query.NewBlockQuery(chaintype.GetChainType(0)).Fields)
	mockRows.AddRow(
		mockBlockData.GetID(),
		mockBlockData.GetBlockHash(),
		mockBlockData.GetPreviousBlockHash(),
		mockBlockData.GetHeight(),
		mockBlockData.GetTimestamp(),
		mockBlockData.GetBlockSeed(),
		mockBlockData.GetBlockSignature(),
		mockBlockData.GetCumulativeDifficulty(),
		mockBlockData.GetSmithScale(),
		mockBlockData.GetPayloadLength(),
		mockBlockData.GetPayloadHash(),
		mockBlockData.GetBlocksmithPublicKey(),
		mockBlockData.GetTotalAmount(),
		mockBlockData.GetTotalFee(),
		mockBlockData.GetTotalCoinBase(),
		mockBlockData.GetVersion(),
	)
	mock.ExpectQuery("").WillReturnRows(mockRows)
	return db.QueryRow(""), nil
}

func TestGetLastBlock(t *testing.T) {
	type args struct {
		queryExecutor query.ExecutorInterface
		blockQuery    query.BlockQueryInterface
	}
	tests := []struct {
		name    string
		args    args
		want    *model.Block
		wantErr bool
	}{
		{
			name: "wantFail:Fail",
			args: args{
				queryExecutor: &mockExecutorGetLastBlockFail{},
				blockQuery:    query.NewBlockQuery(chaintype.GetChainType(0)),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:NoRow",
			args: args{
				queryExecutor: &mockExecutorGetLastBlockNoRow{},
				blockQuery:    query.NewBlockQuery(chaintype.GetChainType(0)),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantSuccess",
			args: args{
				queryExecutor: &mockExecutorGetLastBlockSuccess{},
				blockQuery:    query.NewBlockQuery(chaintype.GetChainType(0)),
			},
			want:    &mockBlockData,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetLastBlock(tt.args.queryExecutor, tt.args.blockQuery)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLastBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}

func TestUtil_getMinRollbackHeight(t *testing.T) {
	type fields struct {
		Height uint32
	}
	tests := []struct {
		name   string
		fields fields
		want   uint32
	}{
		// TODO: Add test cases.
		{
			name: "GetMinRollbackHeight Successful",
			fields: fields{
				Height: constant.MinRollbackBlocks - 1,
			},
			want: 0,
		},
		{
			name: "GetMinROllbackHeight Failed",
			fields: fields{
				Height: constant.MinRollbackBlocks + 1,
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetMinRollbackHeight(tt.fields.Height)
			if got != tt.want {
				t.Errorf("Service.getMinRollbackHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}
