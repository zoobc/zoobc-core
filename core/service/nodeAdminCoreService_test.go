package service

import (
	"bytes"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/query"
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

	type fields struct {
		QueryExecutor query.ExecutorInterface
		BlockQuery    query.BlockQueryInterface
		AccountQuery  query.AccountQueryInterface
		Signature     crypto.SignatureInterface
	}

	tests := []struct {
		name   string
		fields fields
		params *paramsStruct
		want   *wantStruct
	}{
		{
			name: "Generate Proof Of Ownership",
			fields: fields{
				QueryExecutor: &mockQueryExecutorSuccess{},
				BlockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
				AccountQuery:  nil,
				Signature:     nil,
			},
			params: &paramsStruct{
				accountType:    1,
				accountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				signature:      []byte{4, 38, 113, 185},
			},
			want: &wantStruct{
				proofOfOwnershipSign: []byte{115, 74, 30, 212, 221, 118, 106, 246, 87, 93, 149, 146, 141, 111, 100, 45, 29, 48,
					16, 212, 236, 60, 30, 50, 73, 134, 217, 91, 220, 41, 69, 7, 44, 181, 253, 159, 156, 174, 68, 206, 19, 51, 47, 211,
					90, 100, 38, 32, 178, 155, 204, 215, 194, 5, 109, 251, 106, 118, 238, 8, 24, 127, 170, 4},
				nodeMessages: []byte{1, 0, 66, 67, 90, 69, 71, 79, 98, 51, 87, 78, 120, 51, 102, 68, 79, 86,
					102, 57, 90, 83, 52, 69, 106, 118, 79, 73, 118, 95, 85, 101, 87, 52, 84, 86, 66, 81, 74, 95, 54, 116, 72, 75,
					108, 69, 28, 67, 36, 177, 18, 86, 20, 83, 73, 100, 118, 236, 245, 79, 57, 156, 69, 220, 166, 222, 128, 172,
					119, 169, 85, 168, 111, 124, 143, 109, 18, 226, 91, 149, 235, 82, 49, 204, 97, 180, 91, 82, 40, 9, 12, 15, 94,
					49, 245, 175, 150, 243, 217, 140, 133, 89, 117, 200, 193, 235, 101, 145, 8, 195, 1, 0, 0, 0},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			nodeAdminService := &NodeAdminService{
				QueryExecutor: tt.fields.QueryExecutor,
				BlockQuery:    tt.fields.BlockQuery,
				AccountQuery:  tt.fields.AccountQuery,
				Signature:     tt.fields.Signature,
			}
			res1, res2 := nodeAdminService.GenerateProofOfOwnership(tt.params.accountType, tt.params.accountAddress, tt.params.signature)

			if !bytes.Equal(res1, tt.want.nodeMessages) {
				t.Errorf("GetGenerateProofOfOwnership() \ngot = %v, \nwant = %v", res1, tt.want.nodeMessages)
				return
			}
			if !bytes.Equal(res2, tt.want.proofOfOwnershipSign) {
				t.Errorf("GetGenerateProofOfOwnership() \ngot = %v, \nwant = %v", res2, tt.want.proofOfOwnershipSign)
				return
			}
		})
	}
}

func TestValidateProofOfOwnership(t *testing.T) {

	type paramsStruct struct {
		nodeMessages []byte
		signature    []byte
		publicKey    []byte
	}

	type wantStruct struct {
		err error
	}

	type fields struct {
		QueryExecutor query.ExecutorInterface
		BlockQuery    query.BlockQueryInterface
		AccountQuery  query.AccountQueryInterface
		Signature     crypto.SignatureInterface
	}

	tests := []struct {
		name   string
		fields fields
		params *paramsStruct
		want   *wantStruct
	}{
		{
			name: "Validate Proof Of Ownership",
			fields: fields{
				QueryExecutor: &mockQueryExecutorSuccess{},
				BlockQuery:    query.NewBlockQuery(&chaintype.MainChain{}),
				AccountQuery:  nil,
				Signature:     nil,
			},
			params: &paramsStruct{
				nodeMessages: []byte{1, 0, 66, 67, 90, 69, 71, 79, 98, 51, 87, 78, 120, 51, 102, 68, 79, 86,
					102, 57, 90, 83, 52, 69, 106, 118, 79, 73, 118, 95, 85, 101, 87, 52, 84, 86, 66, 81, 74, 95, 54, 116, 72, 75,
					108, 69, 28, 67, 36, 177, 18, 86, 20, 83, 73, 100, 118, 236, 245, 79, 57, 156, 69, 220, 166, 222, 128, 172,
					119, 169, 85, 168, 111, 124, 143, 109, 18, 226, 91, 149, 235, 82, 49, 204, 97, 180, 91, 82, 40, 9, 12, 15, 94,
					49, 245, 175, 150, 243, 217, 140, 133, 89, 117, 200, 193, 235, 101, 145, 8, 195, 1, 0, 0, 0},
				signature: []byte{115, 74, 30, 212, 221, 118, 106, 246, 87, 93, 149, 146, 141, 111, 100, 45, 29, 48,
					16, 212, 236, 60, 30, 50, 73, 134, 217, 91, 220, 41, 69, 7, 44, 181, 253, 159, 156, 174, 68, 206, 19, 51, 47, 211,
					90, 100, 38, 32, 178, 155, 204, 215, 194, 5, 109, 251, 106, 118, 238, 8, 24, 127, 170, 4},
				publicKey: []byte{4, 38, 103, 73, 250, 169, 63, 155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242,
					84, 174, 239, 46, 190, 78, 68, 90, 83, 142, 11},
			},
			want: &wantStruct{
				err: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			nodeAdminService := &NodeAdminService{
				QueryExecutor: tt.fields.QueryExecutor,
				BlockQuery:    tt.fields.BlockQuery,
				AccountQuery:  tt.fields.AccountQuery,
				Signature:     tt.fields.Signature,
			}
			res := nodeAdminService.ValidateProofOfOwnership(tt.params.nodeMessages, tt.params.signature, tt.params.publicKey)

			if res != tt.want.err {
				t.Errorf("Validate proof of ownership \ngot = %v, \nwant = %v", res, tt.want.err)
				return
			}

		})
	}
}
