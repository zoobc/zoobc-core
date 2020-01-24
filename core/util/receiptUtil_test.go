package util

import (
	"database/sql"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dgraph-io/badger"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

var (
	receiptUtilInstance = &ReceiptUtil{}
	mockReceipt         = &model.BatchReceipt{
		SenderPublicKey:      []byte{1, 2, 3},
		RecipientPublicKey:   []byte{3, 2, 1},
		DatumType:            constant.ReceiptDatumTypeBlock,
		DatumHash:            []byte{1, 2, 3},
		ReferenceBlockHeight: 0,
		ReferenceBlockHash: []byte{
			167, 255, 198, 248, 191, 30, 215, 102, 81, 193, 71, 86, 160, 97, 214, 98, 245, 128, 255, 77, 228, 59, 73,
			250, 130, 216, 10, 75, 128, 248, 67, 74,
		},
		RMRLinked:          nil,
		RecipientSignature: nil,
	}
	mockBlock = &model.Block{
		ID:                   0,
		PreviousBlockHash:    nil,
		Height:               0,
		Timestamp:            0,
		BlockSeed:            nil,
		BlockSignature:       nil,
		CumulativeDifficulty: "",
		BlocksmithPublicKey:  []byte{},
		TotalAmount:          0,
		TotalFee:             0,
		TotalCoinBase:        0,
		Version:              0,
		PayloadLength:        0,
		PayloadHash:          nil,
		Transactions:         nil,
	}
)

type (
	mockGenerateBatchReceiptWithReminderKVExecutorSuccess struct {
		kvdb.KVExecutor
	}
	mockGenerateBatchReceiptWithReminderKVExecutorFailOtherError struct {
		kvdb.KVExecutor
	}
	mockGenerateBatchReceiptWithReminderQueryExecutorSuccess struct {
		query.Executor
	}
)

func (*mockGenerateBatchReceiptWithReminderKVExecutorSuccess) Insert(key string, value []byte, expiry int) error {
	return nil
}
func (*mockGenerateBatchReceiptWithReminderKVExecutorFailOtherError) Insert(key string, value []byte, expiry int) error {
	return badger.ErrInvalidKey
}

func (*mockGenerateBatchReceiptWithReminderQueryExecutorSuccess) ExecuteSelectRow(
	qStr string,
	tx bool, args ...interface{},
) (*sql.Row, error) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(qStr)).WillReturnRows(sqlmock.NewRows([]string{
		"ID", "Tree", "Timestamp",
	}))
	row := db.QueryRow(qStr)
	return row, nil
}

func TestGetNumberOfMaxReceipts(t *testing.T) {
	type args struct {
		numberOfSortedBlocksmiths int
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{
			name: "TotalBlocksmiths < PriorityConstant",
			args: args{numberOfSortedBlocksmiths: constant.PriorityStrategyMaxPriorityPeers - 1},
			want: constant.PriorityStrategyMaxPriorityPeers - 2,
		},
		{
			name: "TotalBlocksmiths > PriorityConstant",
			args: args{numberOfSortedBlocksmiths: constant.PriorityStrategyMaxPriorityPeers + 2},
			want: constant.PriorityStrategyMaxPriorityPeers,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := receiptUtilInstance.GetNumberOfMaxReceipts(tt.args.numberOfSortedBlocksmiths); got != tt.want {
				t.Errorf("GetNumberOfMaxReceipts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateBatchReceipt(t *testing.T) {
	mockReceipt.DatumHash, _ = util.GetBlockHash(mockBlock, &chaintype.MainChain{})
	type args struct {
		ct                 chaintype.ChainType
		referenceBlock     *model.Block
		senderPublicKey    []byte
		recipientPublicKey []byte
		datumHash          []byte
		datumType          uint32
	}
	tests := []struct {
		name    string
		args    args
		want    *model.BatchReceipt
		wantErr bool
	}{
		{
			name: "GenerateReceipt : success",
			args: args{
				ct:                 &chaintype.MainChain{},
				referenceBlock:     mockBlock,
				senderPublicKey:    mockReceipt.SenderPublicKey,
				recipientPublicKey: mockReceipt.RecipientPublicKey,
				datumHash:          mockReceipt.DatumHash,
				datumType:          mockReceipt.DatumType,
			},
			want:    mockReceipt,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := receiptUtilInstance.GenerateBatchReceipt(
				tt.args.ct, tt.args.referenceBlock, tt.args.senderPublicKey, tt.args.recipientPublicKey,
				tt.args.datumHash, nil, tt.args.datumType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateBatchReceipt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateBatchReceipt() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetUnsignedReceiptBytes(t *testing.T) {
	mockReceipt.DatumHash, _ = util.GetBlockHash(mockBlock, &chaintype.MainChain{})
	type args struct {
		receipt *model.BatchReceipt
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "GetUnsignedReceiptBytes:success",
			args: args{receipt: mockReceipt},
			want: []byte{
				1, 2, 3, 3, 2, 1, 0, 0, 0, 0, 167, 255, 198, 248, 191, 30, 215, 102, 81, 193, 71, 86, 160, 97, 214, 98,
				245, 128, 255, 77, 228, 59, 73, 250, 130, 216, 10, 75, 128, 248, 67, 74, 1, 0, 0, 0, 167, 255, 198, 248,
				191, 30, 215, 102, 81, 193, 71, 86, 160, 97, 214, 98, 245, 128, 255, 77, 228, 59, 73, 250, 130, 216, 10,
				75, 128, 248, 67, 74,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := receiptUtilInstance.GetUnsignedBatchReceiptBytes(tt.args.receipt); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUnsignedReceiptBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetReceiptKey(t *testing.T) {
	type args struct {
		dataHash        []byte
		senderPublicKey []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "GetReceiptKey:success",
			args: args{
				dataHash:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
				senderPublicKey: []byte{8, 7, 6, 5, 4, 3, 2, 1},
			},
			want: []byte{
				2, 160, 111, 100, 237, 108, 67, 150, 246, 57, 185, 79, 214, 244, 182, 125, 4, 110, 77, 16, 211, 215,
				53, 174, 50, 113, 46, 46, 80, 149, 21, 150,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := receiptUtilInstance.GetReceiptKey(tt.args.dataHash, tt.args.senderPublicKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetReceiptKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetReceiptKey() got = %v, want %v", got, tt.want)
			}
		})
	}
}
