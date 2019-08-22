package service

import (
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/model"
	commonUtil "github.com/zoobc/zoobc-core/common/util"
)

type (
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
				BlockService: &blockServiceMocked{},
			},
			args: args{
				accountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
			},
			want: &model.ProofOfOwnership{
				MessageBytes: []byte{66, 67, 90, 69, 71, 79, 98, 51, 87, 78, 120, 51, 102, 68, 79, 86,
					102, 57, 90, 83, 52, 69, 106, 118, 79, 73, 118, 95, 85, 101, 87, 52, 84, 86, 66, 81,
					74, 95, 54, 116, 72, 75, 108, 69, 166, 159, 115, 204, 162, 58, 154, 197, 200, 181,
					103, 220, 24, 90, 117, 110, 151, 201, 130, 22, 79, 226, 88, 89, 224, 209, 220, 193,
					71, 92, 128, 166, 21, 178, 18, 58, 241, 245, 249, 76, 17, 227, 233, 64, 44, 58, 197,
					88, 245, 0, 25, 157, 149, 182, 211, 227, 1, 117, 133, 134, 40, 29, 205, 38, 0, 0, 0, 0},
				Signature: []byte{0, 0, 0, 0, 240, 15, 67, 88, 228, 24, 201, 104, 160, 102, 6, 249, 82, 197, 58,
					181, 142, 56, 129, 159, 102, 104, 119, 208, 63, 199, 57, 32, 142, 174, 210, 60, 30,
					243, 9, 24, 69, 24, 185, 61, 117, 30, 228, 212, 189, 50, 49, 105, 225, 73, 239, 184,
					113, 147, 225, 158, 4, 245, 221, 216, 217, 164, 219, 7},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nas := &NodeAdminService{
				BlockService: tt.fields.BlockService,
			}
			got, err := nas.GenerateProofOfOwnership(tt.args.accountAddress)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeAdminService.GenerateProofOfOwnership() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				log.Printf("%v", got.MessageBytes)
				log.Printf("%v", got.Signature)
				t.Errorf("NodeAdminService.GenerateProofOfOwnership() get %v\n, want %v\n", got, tt.want)
			}
		})
	}
}
