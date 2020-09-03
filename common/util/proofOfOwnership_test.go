package util

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
)

func TestParseProofOfOwnershipBytes(t *testing.T) {
	poown := &model.ProofOfOwnership{
		MessageBytes: make([]byte, GetProofOfOwnershipSize(false)),
		Signature:    make([]byte, constant.NodeSignature),
	}
	poownBytes := GetProofOfOwnershipBytes(poown)
	type args struct {
		poownBytes []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *model.ProofOfOwnership
		wantErr bool
	}{
		{
			name: "ParseProofOfOwnershipBytes - fail (empty bytes)",
			args: args{
				poownBytes: []byte{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ParseProofOfOwnershipBytes - fail (wrong poown size)",
			args: args{
				poownBytes: poownBytes[:10],
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ParseProofOfOwnershipBytes - fail (no signature / wrong signature size)",
			args: args{
				poownBytes: poownBytes[:GetProofOfOwnershipSize(false)],
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "ParseProofOfOwnershipBytes - success",
			args:    args{poownBytes: poownBytes},
			want:    poown,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseProofOfOwnershipBytes(tt.args.poownBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseProofOfOwnershipBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseProofOfOwnershipBytes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetProofOfOwnershipSize(t *testing.T) {
	t.Run("WithAndWithoutSignature-Gap", func(t *testing.T) {
		withSig := GetProofOfOwnershipSize(true)
		withoutSig := GetProofOfOwnershipSize(false)
		if withSig-withoutSig != constant.NodeSignature {
			t.Errorf("GetPoownSize with and without signature should have %d difference",
				constant.NodeSignature)
		}
	})
}

func TestParseProofOfOwnershipMessageBytes(t *testing.T) {
	poownMessage := &model.ProofOfOwnershipMessage{
		AccountAddress: "ZBC_AQTEIGHG_65MNY534_GOKX7VSS_4BEO6OEL_75I6LOCN_KBICP7VN_DSUWBLM7",
		BlockHash:      make([]byte, constant.BlockHash),
		BlockHeight:    0,
	}
	poownMessageBytes := GetProofOfOwnershipMessageBytes(poownMessage)
	type args struct {
		poownMessageBytes []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *model.ProofOfOwnershipMessage
		wantErr bool
	}{
		{
			name:    "ParseProofOfOwnershipMessageBytes:fail - no bytes",
			args:    args{poownMessageBytes: []byte{}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "ParseProofOfOwnershipMessageBytes:fail - wrong account address",
			args:    args{poownMessageBytes: poownMessageBytes[:10]},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "ParseProofOfOwnershipMessageBytes:fail - no block hash",
			args:    args{poownMessageBytes: poownMessageBytes[:len([]byte(poownMessage.AccountAddress))]},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ParseProofOfOwnershipMessageBytes:fail - no block height",
			args: args{
				poownMessageBytes: poownMessageBytes[:(len([]byte(poownMessage.AccountAddress)) +
					int(constant.BlockHash))],
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ParseProofOfOwnershipMessageBytes:fail - success",
			args: args{
				poownMessageBytes: poownMessageBytes,
			},
			want:    poownMessage,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseProofOfOwnershipMessageBytes(tt.args.poownMessageBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseProofOfOwnershipMessageBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseProofOfOwnershipMessageBytes() got = %v, want %v", got, tt.want)
			}
		})
	}
}
