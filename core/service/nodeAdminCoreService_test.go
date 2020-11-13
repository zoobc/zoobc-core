package service

import (
	"encoding/json"
	"github.com/zoobc/zoobc-core/common/crypto"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	blockServiceMocked struct {
		BlockService
		height uint32
	}
)

var (
	nodeUtilfixtureNodeKeysJSON = []*model.NodeKey{
		{
			PublicKey: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
				45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
			Seed: "sprinkled sneak species pork outpost thrift unwind cheesy vexingly dizzy neurology neatness",
		},
		{
			ID: 1,
			PublicKey: []byte{0, 14, 6, 218, 170, 54, 60, 50, 2, 66, 130, 119, 226, 235, 126, 203, 5, 12,
				152, 194, 170, 146, 43, 63, 224, 101, 127, 241, 62, 152, 187, 255},
			Seed: "demanding unlined hazard neuter condone anime asleep ascent capitol sitter marathon armband",
		},
		{
			ID: 2,
			PublicKey: []byte{140, 115, 35, 51, 159, 22, 234, 192, 38, 104, 96, 24, 80, 70, 86, 211,
				123, 72, 52, 221, 97, 121, 59, 151, 158, 90, 167, 17, 110, 253, 122, 158},
			Seed: "street roast immovable escalator stinger nervy provider debug flavoring hubcap creature remix",
		},
	}
	nodeAdminSrvAccountAddress1 = []byte{0, 0, 0, 0, 30, 136, 57, 247, 116, 237, 101, 11, 154, 3, 19, 178, 194, 77, 152, 45, 43, 93, 109,
		176, 163, 215, 121, 199, 229, 242, 213, 91, 53, 165, 78, 164}
)

func (bsMock *blockServiceMocked) GetLastBlock() (*model.Block, error) {
	return &model.Block{
		Height: bsMock.height,
	}, nil
}

func (*blockServiceMocked) GetBlockByHeight(height uint32) (*model.Block, error) {
	return &model.Block{
		Height: height,
	}, nil
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
				ID: 2,
				PublicKey: []byte{140, 115, 35, 51, 159, 22, 234, 192, 38, 104, 96, 24, 80, 70, 86, 211,
					123, 72, 52, 221, 97, 121, 59, 151, 158, 90, 167, 17, 110, 253, 122, 158},
				Seed: "street roast immovable escalator stinger nervy provider debug flavoring hubcap creature remix",
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
		accountAddress []byte
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
				BlockService: &blockServiceMocked{
					height: 1,
					BlockService: BlockService{
						Chaintype: &chaintype.MainChain{},
					},
				},
				FilePath: "testdata/node_keys.json",
			},
			args: args{
				accountAddress: nodeAdminSrvAccountAddress1,
			},
			want: &model.ProofOfOwnership{
				MessageBytes: []byte{0, 0, 0, 0, 30, 136, 57, 247, 116, 237, 101, 11, 154, 3, 19, 178, 194, 77, 152, 45, 43, 93, 109, 176,
					163, 215, 121, 199, 229, 242, 213, 91, 53, 165, 78, 164, 167, 255, 198, 248, 191, 30, 215, 102, 81, 193, 71, 86, 160,
					97, 214, 98, 245, 128, 255, 77, 228, 59, 73, 250, 130, 216, 10, 75, 128, 248, 67, 74, 1, 0, 0, 0,
				},
				Signature: []byte{121, 232, 54, 161, 27, 34, 149, 127, 212, 132, 26, 220, 111, 19, 167, 176, 235, 203, 94, 215, 12, 193,
					71, 96, 187, 97, 119, 249, 99, 41, 5, 211, 147, 190, 184, 43, 32, 252, 50, 56, 104, 67, 113, 144, 137, 63, 245, 151,
					172, 30, 57, 198, 184, 15, 182, 229, 99, 173, 239, 8, 190, 108, 163, 6},
			},
			wantErr: false,
		},
		{
			name: "GenerateProofOfOwnership:Success-{safeBlockHeight}",
			fields: fields{
				BlockService: &blockServiceMocked{
					height: 11,
					BlockService: BlockService{
						Chaintype: &chaintype.MainChain{},
					},
				},
				FilePath: "testdata/node_keys.json",
			},
			args: args{
				accountAddress: nodeAdminSrvAccountAddress1,
			},
			want: &model.ProofOfOwnership{
				MessageBytes: []byte{0, 0, 0, 0, 30, 136, 57, 247, 116, 237, 101, 11, 154, 3, 19, 178, 194, 77, 152, 45, 43, 93, 109, 176,
					163, 215, 121, 199, 229, 242, 213, 91, 53, 165, 78, 164, 167, 255, 198, 248, 191, 30, 215, 102, 81, 193, 71, 86, 160,
					97, 214, 98, 245, 128, 255, 77, 228, 59, 73, 250, 130, 216, 10, 75, 128, 248, 67, 74, 1, 0, 0, 0,
				},
				Signature: []byte{121, 232, 54, 161, 27, 34, 149, 127, 212, 132, 26, 220, 111, 19, 167, 176, 235, 203, 94, 215, 12, 193,
					71, 96, 187, 97, 119, 249, 99, 41, 5, 211, 147, 190, 184, 43, 32, 252, 50, 56, 104, 67, 113, 144, 137, 63, 245, 151,
					172, 30, 57, 198, 184, 15, 182, 229, 99, 173, 239, 8, 190, 108, 163, 6},
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
