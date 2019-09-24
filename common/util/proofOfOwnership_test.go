package util

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

func TestParseProofOfOwnershipBytes(t *testing.T) {
	poown := &model.ProofOfOwnership{
		MessageBytes: make([]byte, GetProofOfOwnershipSize(false)),
		Signature:    make([]byte, 68),
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
