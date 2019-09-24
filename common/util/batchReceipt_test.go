package util

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
)

var (
	mockReceipt = &model.BatchReceipt{
		SenderPublicKey:      []byte{1, 2, 3},
		RecipientPublicKey:   []byte{3, 2, 1},
		DatumType:            constant.ReceiptDatumTypeBlock,
		DatumHash:            []byte{1, 2, 3},
		ReferenceBlockHeight: 0,
		ReferenceBlockHash: []byte{
			166, 159, 115, 204, 162, 58, 154, 197, 200, 181, 103, 220, 24, 90, 117, 110, 151, 201, 130, 22, 79, 226, 88,
			89, 224, 209, 220, 193, 71, 92, 128, 166, 21, 178, 18, 58, 241,
			245, 249, 76, 17, 227, 233, 64, 44, 58, 197, 88, 245, 0, 25, 157,
			149, 182, 211, 227, 1, 117, 133, 134, 40, 29, 205, 38,
		},
		ReceiptMerkleRoot:  nil,
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
		SmithScale:           0,
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

func TestGenerateReceipt(t *testing.T) {
	mockReceipt.DatumHash, _ = GetBlockHash(mockBlock)
	type args struct {
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
			got, err := GenerateBatchReceipt(tt.args.referenceBlock, tt.args.senderPublicKey, tt.args.recipientPublicKey,
				tt.args.datumHash, tt.args.datumType)
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
	mockReceipt.DatumHash, _ = GetBlockHash(mockBlock)
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
				1, 2, 3, 3, 2, 1, 0, 0, 0, 0, 166, 159, 115, 204, 162, 58, 154, 197, 200, 181, 103, 220, 24, 90, 117, 110, 151, 201, 130,
				22, 79, 226, 88, 89, 224, 209, 220, 193, 71, 92, 128, 166, 21, 178, 18, 58, 241, 245, 249, 76, 17, 227, 233, 64, 44, 58,
				197, 88, 245, 0, 25, 157, 149, 182, 211, 227, 1, 117, 133, 134, 40, 29, 205, 38, 1, 0, 0, 0, 166, 159, 115, 204, 162, 58,
				154, 197, 200, 181, 103, 220, 24, 90, 117, 110, 151, 201, 130, 22, 79, 226, 88, 89, 224, 209, 220, 193, 71, 92, 128, 166, 21,
				178, 18, 58, 241, 245, 249, 76, 17, 227, 233, 64, 44, 58, 197, 88, 245, 0, 25, 157, 149, 182, 211, 227, 1, 117, 133, 134, 40,
				29, 205, 38,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetUnsignedBatchReceiptBytes(tt.args.receipt); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUnsignedReceiptBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
