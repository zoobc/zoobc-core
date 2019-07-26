package service

import (
	"bytes"
	"testing"
)

func TestGenerateProofOfOwnership(t *testing.T) {

	type paramsStruct struct {
		accountType    uint32
		accountAddress string
		signature      []byte
	}

	type wantStruct struct {
		proofOfOwnershipSign []byte
		nodeMessages         []byte
	}

	nodeAdminService := &NodeAdminService{}

	tests := []struct {
		name   string
		params *paramsStruct
		want   *wantStruct
	}{
		{
			name: "Generate Proof Of Ownership",
			params: &paramsStruct{
				accountType:    1,
				accountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				signature:      []byte{4, 38, 113, 185},
			},
			want: &wantStruct{
				proofOfOwnershipSign: []byte{4, 38, 113, 185},
				nodeMessages:         []byte{4, 38, 113, 185},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res1, _ := nodeAdminService.GenerateProofOfOwnership(tt.params.accountType, tt.params.accountAddress, tt.params.signature)
			if !bytes.Equal(res1, tt.want.nodeMessages) {
				t.Errorf("GetGenerateProofOfOwnership() \ngot = %v, \nwant = %v", res1, tt.want.nodeMessages)
				return
			}
		})
	}
}
