package service

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
	commonUtil "github.com/zoobc/zoobc-core/common/util"
)

type (
	spyNodeAdminCoreServiceHelper struct {
		NodeAdminService
	}
	blockServiceMocked struct {
		BlockService
	}
)

func (*blockServiceMocked) GetLastBlock() (*model.Block, error) {
	return new(model.Block), nil
}

func (*blockServiceMocked) GetBlockByHeight(height uint32) (*model.Block, error) {
	return new(model.Block), nil
}

func TestNodeAdminService_GenerateProofOfOwnership(t *testing.T) {
	if err := commonUtil.LoadConfig("./resource", "config", "toml"); err != nil {
		panic(err)
	}
	type fields struct {
		Helpers      NodeAdminServiceHelpersInterface
		BlockService BlockServiceInterface
	}
	type args struct {
		accountAddress string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.ProofOfOwnership
		wantErr bool
	}{
		{
			name: "GenerateProofOfOwnership:Success",
			fields: fields{
				Helpers:      &spyNodeAdminCoreServiceHelper{},
				BlockService: &blockServiceMocked{},
			},
			args: args{
				accountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
			},
			want: &model.ProofOfOwnership{
				MessageBytes: []byte{8, 1, 18, 44, 66, 67, 90, 69, 71, 79, 98, 51, 87, 78, 120, 51, 102, 68, 79, 86, 102, 57, 90, 83, 52, 69,
					106, 118, 79, 73, 118, 95, 85, 101, 87, 52, 84, 86, 66, 81, 74, 95, 54, 116, 72, 75, 108, 69, 26, 64, 166, 159, 115, 204,
					162, 58, 154, 197, 200, 181, 103, 220, 24, 90, 117, 110, 151, 201, 130, 22, 79, 226, 88, 89, 224, 209, 220, 193, 71, 92, 128,
					166, 21, 178, 18, 58, 241, 245, 249, 76, 17, 227, 233, 64, 44, 58, 197, 88, 245, 0, 25, 157, 149, 182, 211, 227, 1, 117, 133,
					134, 40, 29, 205, 38},
				Signature: []byte{156, 114, 29, 63, 218, 45, 128, 0, 15, 148, 102, 248, 215, 237, 93, 241, 87, 188, 65, 94, 74,
					181, 85, 195, 131, 214, 109, 192, 81, 171, 210, 24, 14, 200, 53, 58, 193, 24, 252, 225, 149, 135, 223, 66, 122, 125, 147,
					213, 223, 105, 100, 83, 102, 46, 106, 144, 116, 58, 228, 191, 53, 225, 215, 15},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nas := &NodeAdminService{
				Helpers:      tt.fields.Helpers,
				BlockService: tt.fields.BlockService,
			}
			got, err := nas.GenerateProofOfOwnership(tt.args.accountAddress)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeAdminService.GenerateProofOfOwnership() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeAdminService.GenerateProofOfOwnership() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeAdminService_ValidateProofOfOwnership(t *testing.T) {
	type fields struct {
		Helpers      NodeAdminServiceHelpersInterface
		BlockService BlockServiceInterface
	}
	type args struct {
		poown         *model.ProofOfOwnership
		nodePublicKey []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ValidateProofOfOwnership:Success",
			fields: fields{
				Helpers:      &spyNodeAdminCoreServiceHelper{},
				BlockService: &blockServiceMocked{},
			},
			args: args{
				poown: &model.ProofOfOwnership{
					MessageBytes: []byte{8, 1, 18, 44, 66, 67, 90, 69, 71, 79, 98, 51, 87, 78, 120, 51, 102, 68, 79, 86, 102, 57, 90, 83, 52, 69,
						106, 118, 79, 73, 118, 95, 85, 101, 87, 52, 84, 86, 66, 81, 74, 95, 54, 116, 72, 75, 108, 69, 26, 64, 166, 159, 115, 204,
						162, 58, 154, 197, 200, 181, 103, 220, 24, 90, 117, 110, 151, 201, 130, 22, 79, 226, 88, 89, 224, 209, 220, 193, 71, 92, 128,
						166, 21, 178, 18, 58, 241, 245, 249, 76, 17, 227, 233, 64, 44, 58, 197, 88, 245, 0, 25, 157, 149, 182, 211, 227, 1, 117, 133,
						134, 40, 29, 205, 38},
					Signature: []byte{156, 114, 29, 63, 218, 45, 128, 0, 15, 148, 102, 248, 215, 237, 93, 241, 87, 188, 65, 94, 74,
						181, 85, 195, 131, 214, 109, 192, 81, 171, 210, 24, 14, 200, 53, 58, 193, 24, 252, 225, 149, 135, 223, 66, 122, 125, 147,
						213, 223, 105, 100, 83, 102, 46, 106, 144, 116, 58, 228, 191, 53, 225, 215, 15},
				},
				nodePublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49, 45, 118, 97, 219, 80, 242,
					244, 100, 134, 144, 246, 37, 144, 213, 135},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nas := &NodeAdminService{
				Helpers:      tt.fields.Helpers,
				BlockService: tt.fields.BlockService,
			}
			if err := nas.ValidateProofOfOwnership(tt.args.poown, tt.args.nodePublicKey); (err != nil) != tt.wantErr {
				t.Errorf("NodeAdminService.ValidateProofOfOwnership() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
