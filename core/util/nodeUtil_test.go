package util

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
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

func TestNodeKeyConfig_ParseKeysFile(t *testing.T) {
	type fields struct {
		filePath string
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
				filePath: "/IDontExist",
			},
			wantErr: true,
			errText: "AppErr: NodeKeysFileNotExist",
		},
		{
			name: "ParseKeysFile:fail{InvalidNodeKeysFile}",
			fields: fields{
				filePath: "./testdata/node_keys_invalid.json",
			},
			wantErr: true,
			errText: "AppErr: InvalidNodeKeysFile",
		},
		{
			name: "ParseKeysFile:success",
			fields: fields{
				filePath: "./testdata/node_keys.json",
			},
			wantErr: false,
			want:    nodeUtilfixtureNodeKeysJSON,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nk := &NodeKeyConfig{
				filePath: tt.fields.filePath,
			}
			got, err := nk.ParseKeysFile()
			if err != nil {
				if !tt.wantErr {
					t.Errorf("NodeKeyConfig.ParseKeysFile() error = %v, wantErr %v", err, tt.wantErr)
				}
				if err.Error() != tt.errText {
					t.Errorf("NodeKeyConfig.ParseKeysFile() error text = %s, wantErr text %s", err.Error(), tt.errText)
				}
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeKeyConfig.ParseKeysFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeKeyConfig_GetLastNodeKey(t *testing.T) {
	type fields struct {
		filePath string
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
			nk := &NodeKeyConfig{
				filePath: tt.fields.filePath,
			}
			if got := nk.GetLastNodeKey(tt.args.nodeKeys); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeKeyConfig.GetLastNodeKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodeKeyConfig_GenerateNodeKey(t *testing.T) {
	// add tmp file for test with previous keys
	tmpFilePath := "testdata/node_keys_tmp"
	tmpFilePath2 := "testdata/node_keys2_tmp"
	file, _ := json.MarshalIndent(nodeUtilfixtureNodeKeysJSON, "", " ")
	_ = ioutil.WriteFile(tmpFilePath, file, 0644)

	type args struct {
		seed string
	}
	type fields struct {
		filePath string
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
				filePath: tmpFilePath,
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
				filePath: tmpFilePath2,
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
			nk := &NodeKeyConfig{
				filePath: tt.fields.filePath,
			}
			got, err := nk.GenerateNodeKey(tt.args.seed)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodeKeyConfig.GenerateNodeKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeKeyConfig.GenerateNodeKey() = %v, want %v", got, tt.want)
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
