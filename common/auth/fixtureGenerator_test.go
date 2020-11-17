package auth

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

func TestGetFixturesProofOfOwnershipValidation(t *testing.T) {
	type args struct {
		height uint32
		hash   []byte
		block  *model.Block
	}
	tests := []struct {
		name      string
		args      args
		wantPoown *model.ProofOfOwnership
		wantErr   bool
	}{
		{
			name: "GetFixturesProofOfOwnershipValidation:Success",
			args: args{
				height: 0,
				hash:   nil,
				block: &model.Block{
					ID:                   0,
					BlockHash:            nil,
					PreviousBlockHash:    nil,
					Height:               0,
					Timestamp:            0,
					BlockSeed:            nil,
					BlockSignature:       nil,
					CumulativeDifficulty: "",
					BlocksmithPublicKey:  nil,
					TotalAmount:          0,
					TotalFee:             0,
					TotalCoinBase:        0,
					Version:              0,
					PayloadLength:        0,
					PayloadHash:          nil,
					MerkleRoot:           nil,
					MerkleTree:           nil,
					ReferenceBlockHeight: 0,
					Transactions:         nil,
					PublishedReceipts:    nil,
					SpinePublicKeys:      nil,
					SpineBlockManifests:  nil,
					TransactionIDs:       nil,
				},
			},
			wantPoown: &model.ProofOfOwnership{
				MessageBytes: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
					0, 0, 0, 0, 0, 0, 0, 0, 167, 255, 198, 248, 191, 30, 215, 102, 81, 193, 71, 86, 160, 97, 214, 98,
					245, 128, 255, 77, 228, 59, 73, 250, 130, 216, 10, 75, 128, 248, 67, 74, 0, 0, 0, 0},
				Signature: []byte{101, 230, 53, 25, 61, 205, 146, 135, 222, 231, 68, 146, 143, 55, 215, 32, 48, 34,
					170, 203, 33, 211, 96, 205, 232, 105, 143, 246, 17, 52, 126, 32, 139, 117, 114, 210, 58, 21, 162,
					90, 231, 200, 239, 160, 247, 50, 228, 40, 30, 120, 138, 162, 115, 207, 161, 28, 36, 167, 171, 169,
					35, 48, 239, 4},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotPoown := GetFixturesProofOfOwnershipValidation(tt.args.height, tt.args.hash, tt.args.block); !reflect.DeepEqual(gotPoown, tt.wantPoown) {
				t.Errorf("GetFixturesProofOfOwnershipValidation() = %v, want %v", gotPoown, tt.wantPoown)
			}
		})
	}
}
