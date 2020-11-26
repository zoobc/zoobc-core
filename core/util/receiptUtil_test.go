package util

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/util"
)

var (
	receiptUtilInstance = &ReceiptUtil{}
	mockReceipt1        = &model.Receipt{
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
	mockGenerateBatchReceiptBlock := &storage.BlockCacheObject{
		ID:        mockBlock.ID,
		Height:    mockBlock.Height,
		BlockHash: mockReceipt1.ReferenceBlockHash,
	}
	type args struct {
		ct                 chaintype.ChainType
		referenceBlock     *storage.BlockCacheObject
		senderPublicKey    []byte
		recipientPublicKey []byte
		datumHash          []byte
		datumType          uint32
	}
	tests := []struct {
		name    string
		args    args
		want    *model.Receipt
		wantErr bool
	}{
		{
			name: "GenerateReceipt : success",
			args: args{
				ct:                 &chaintype.MainChain{},
				referenceBlock:     mockGenerateBatchReceiptBlock,
				senderPublicKey:    mockReceipt1.SenderPublicKey,
				recipientPublicKey: mockReceipt1.RecipientPublicKey,
				datumHash:          mockReceipt1.DatumHash,
				datumType:          mockReceipt1.DatumType,
			},
			want:    mockReceipt1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := receiptUtilInstance.GenerateReceipt(
				tt.args.ct, tt.args.referenceBlock, tt.args.senderPublicKey, tt.args.recipientPublicKey,
				tt.args.datumHash, nil, tt.args.datumType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateReceipt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateReceipt() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetUnsignedReceiptBytes(t *testing.T) {
	mockReceipt1.DatumHash, _ = util.GetBlockHash(mockBlock, &chaintype.MainChain{})
	type args struct {
		receipt *model.Receipt
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "GetUnsignedReceiptBytes:success",
			args: args{receipt: mockReceipt1},
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
			if got := receiptUtilInstance.GetUnsignedReceiptBytes(tt.args.receipt); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUnsignedReceiptBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
