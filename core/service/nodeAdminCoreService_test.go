package service

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	blockServiceMocked struct {
		BlockService
	}
)

var (
	nodeUtilfixtureNodeKeysJSON = []*model.NodeKey{
		{
			ID:        0,
			PublicKey: "993a32c8073d6ce5cc30c79115637d4b312d7661db50f2f4648690f62590d587",
			Seed:      "sprinkled sneak species pork outpost thrift unwind cheesy vexingly dizzy neurology neatness",
		},
		{
			ID:        1,
			PublicKey: "000e06daaa363c3202428277e2eb7ecb050c98c2aa922b3fe0657ff13e98bbff",
			Seed:      "demanding unlined hazard neuter condone anime asleep ascent capitol sitter marathon armband",
		},
		{
			ID:        2,
			PublicKey: "8c7323339f16eac026686018504656d37b4834dd61793b979e5aa7116efd7a9e",
			Seed:      "street roast immovable escalator stinger nervy provider debug flavoring hubcap creature remix",
		},
	}
)

func (*blockServiceMocked) GetLastBlock() (*model.Block, error) {
	return new(model.Block), nil
}

func (*blockServiceMocked) GetBlockByHeight(height uint32) (*model.Block, error) {
	return new(model.Block), nil
}

func TestNodeAdminService_GenerateNodeKey(t *testing.T) {
	// add tmp file for test with previous keys
	tmpFilePath := "testdata/node_keys_tmp"
	tmpFilePath2 := "testdata/node_keys2_tmp"
	file, _ := json.MarshalIndent(nodeUtilfixtureNodeKeysJSON, "", " ")
	_ = ioutil.WriteFile(tmpFilePath, file, 0644)

	type fields struct {
		QueryExecutor query.ExecutorInterface
		BlockQuery    query.BlockQueryInterface
		Signature     crypto.SignatureInterface
		BlockService  BlockServiceInterface
		FilePath      string
	}
	type args struct {
		seed string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "GenerateNodeKey:success-{append to previous keys}",
			fields: fields{
				FilePath: tmpFilePath,
			},
			args: args{
				seed: "street roast immovable escalator stinger nervy provider debug flavoring hubcap creature remix",
			},
			wantErr: false,
			want: []byte{140, 115, 35, 51, 159, 22, 234, 192, 38, 104, 96, 24, 80, 70, 86, 211,
				123, 72, 52, 221, 97, 121, 59, 151, 158, 90, 167, 17, 110, 253, 122, 158},
		},
		{
			name: "GenerateNodeKey:success-{first node key}",
			fields: fields{
				FilePath: tmpFilePath2,
			},
			args: args{
				seed: "street roast immovable escalator stinger nervy provider debug flavoring hubcap creature remix",
			},
			wantErr: false,
			want: []byte{140, 115, 35, 51, 159, 22, 234, 192, 38, 104, 96, 24, 80, 70, 86, 211,
				123, 72, 52, 221, 97, 121, 59, 151, 158, 90, 167, 17, 110, 253, 122, 158},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nas := &NodeAdminService{
				QueryExecutor: tt.fields.QueryExecutor,
				BlockQuery:    tt.fields.BlockQuery,
				Signature:     tt.fields.Signature,
				BlockService:  tt.fields.BlockService,
				FilePath:      tt.fields.FilePath,
			}
			got, err := nas.GenerateNodeKey(tt.args.seed)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeAdminService.GenerateNodeKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeAdminService.GenerateNodeKey() = %v, want %v", got, tt.want)
			}
		})
	}

	// extra checks
	file, _ = ioutil.ReadFile(tmpFilePath)
	data := make([]*model.NodeKey, 0)
	_ = json.Unmarshal(file, &data)
	if len(data) != len(nodeUtilfixtureNodeKeysJSON)+1 {
		t.Errorf("NodeKeyConfig.GenerateNodeKey() data appended incorrectly to node keys file %s", tmpFilePath)
	}
	os.Remove(tmpFilePath)
	file, _ = ioutil.ReadFile(tmpFilePath2)
	data = make([]*model.NodeKey, 0)
	_ = json.Unmarshal(file, &data)
	if len(data) != 1 {
		t.Errorf("NodeKeyConfig.GenerateNodeKey() data appended incorrectly to node keys file %s", tmpFilePath2)
	}
	os.Remove(tmpFilePath2)

}

func TestNodeAdminService_GetLastNodeKey(t *testing.T) {
	type fields struct {
		QueryExecutor query.ExecutorInterface
		BlockQuery    query.BlockQueryInterface
		Signature     crypto.SignatureInterface
		BlockService  BlockServiceInterface
		FilePath      string
	}
	type args struct {
		nodeKeys []*model.NodeKey
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *model.NodeKey
	}{
		{
			name: "GetLastNodeKey:success",
			args: args{
				nodeKeys: nodeUtilfixtureNodeKeysJSON,
			},
			want: &model.NodeKey{
				ID:        2,
				PublicKey: "8c7323339f16eac026686018504656d37b4834dd61793b979e5aa7116efd7a9e",
				Seed:      "street roast immovable escalator stinger nervy provider debug flavoring hubcap creature remix",
			},
		},
		{
			name: "GetLastNodeKey:success-{return nil when node_keys file don't exist}",
			args: args{
				nodeKeys: nil,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NodeAdminService{
				QueryExecutor: tt.fields.QueryExecutor,
				BlockQuery:    tt.fields.BlockQuery,
				Signature:     tt.fields.Signature,
				BlockService:  tt.fields.BlockService,
				FilePath:      tt.fields.FilePath,
			}
			if got := n.GetLastNodeKey(tt.args.nodeKeys); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeAdminService.GetLastNodeKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeAdminService_ParseKeysFile(t *testing.T) {
	type fields struct {
		QueryExecutor query.ExecutorInterface
		BlockQuery    query.BlockQueryInterface
		Signature     crypto.SignatureInterface
		BlockService  BlockServiceInterface
		FilePath      string
	}
	tests := []struct {
		name    string
		fields  fields
		want    []*model.NodeKey
		wantErr bool
		errText string
	}{
		{
			name: "ParseKeysFile:fail{NodeKeysFileNotExist}",
			fields: fields{
				FilePath: "/IDontExist",
			},
			wantErr: true,
			errText: "AppErr: NodeKeysFileNotExist",
		},
		{
			name: "ParseKeysFile:fail{InvalidNodeKeysFile}",
			fields: fields{
				FilePath: "./testdata/node_keys_invalid.json",
			},
			wantErr: true,
			errText: "AppErr: InvalidNodeKeysFile",
		},
		{
			name: "ParseKeysFile:success",
			fields: fields{
				FilePath: "./testdata/node_keys.json",
			},
			wantErr: false,
			want:    nodeUtilfixtureNodeKeysJSON,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nas := &NodeAdminService{
				QueryExecutor: tt.fields.QueryExecutor,
				BlockQuery:    tt.fields.BlockQuery,
				Signature:     tt.fields.Signature,
				BlockService:  tt.fields.BlockService,
				FilePath:      tt.fields.FilePath,
			}
			got, err := nas.ParseKeysFile()
			if err != nil {
				if !tt.wantErr {
					t.Errorf("NodeAdminService.ParseKeysFile() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if err.Error() != tt.errText {
					t.Errorf("NodeAdminService.ParseKeysFile() error text = %s, wantErr text %s", err.Error(), tt.errText)
					return
				}
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeAdminService.ParseKeysFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeAdminService_GenerateProofOfOwnership(t *testing.T) {
	type fields struct {
		QueryExecutor query.ExecutorInterface
		BlockQuery    query.BlockQueryInterface
		Signature     crypto.SignatureInterface
		BlockService  BlockServiceInterface
		FilePath      string
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
				FilePath:     "testdata/node_keys.json",
			},
			args: args{
				accountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
			},
			want: &model.ProofOfOwnership{
				MessageBytes: []byte{66, 67, 90, 69, 71, 79, 98, 51, 87, 78, 120, 51, 102, 68, 79, 86, 102, 57,
					90, 83, 52, 69, 106, 118, 79, 73, 118, 95, 85, 101, 87, 52, 84, 86, 66, 81, 74, 95, 54,
					116, 72, 75, 108, 69, 166, 159, 115, 204, 162, 58, 154, 197, 200, 181, 103, 220, 24, 90,
					117, 110, 151, 201, 130, 22, 79, 226, 88, 89, 224, 209, 220, 193, 71, 92, 128, 166, 21, 178,
					18, 58, 241, 245, 249, 76, 17, 227, 233, 64, 44, 58, 197, 88, 245, 0, 25, 157, 149, 182,
					211, 227, 1, 117, 133, 134, 40, 29, 205, 38, 0, 0, 0, 0},
				Signature: []byte{0, 0, 0, 0, 116, 65, 167, 11, 140, 236, 182, 117, 244, 235, 119, 139, 107,
					55, 122, 98, 177, 92, 107, 224, 210, 54, 43, 102, 234, 173, 149, 115, 40, 73, 222, 67,
					215, 244, 76, 225, 218, 137, 183, 246, 220, 7, 239, 204, 10, 196, 105, 140, 231, 127,
					1, 225, 142, 20, 154, 21, 178, 233, 52, 165, 56, 239, 64, 6},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nas := &NodeAdminService{
				QueryExecutor: tt.fields.QueryExecutor,
				BlockQuery:    tt.fields.BlockQuery,
				Signature:     tt.fields.Signature,
				BlockService:  tt.fields.BlockService,
				FilePath:      tt.fields.FilePath,
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
