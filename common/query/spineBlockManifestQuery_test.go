package query

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
)

func TestSpineBlockManifestQuery_InsertSpineBlockManifest(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		spineBlockManifest *model.SpineBlockManifest
	}

	mb1 := &model.SpineBlockManifest{
		FullFileHash:             make([]byte, 64), // sha3-512
		SpineBlockManifestHeight: 720,
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "InsertSpineBlockManifest:success",
			fields: fields{
				Fields:    NewSpineBlockManifestQuery().Fields,
				TableName: NewSpineBlockManifestQuery().TableName,
			},
			args: args{
				spineBlockManifest: mb1,
			},
			want: "INSERT OR REPLACE INTO spine_block_manifest (id,full_file_hash,file_chunk_hashes,manifest_reference_height," +
				"chain_type,manifest_type," +
				"manifest_timestamp) VALUES(? , ?, ?, ?, ?, ?, ?)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mbl := &SpineBlockManifestQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got, _ := mbl.InsertSpineBlockManifest(tt.args.spineBlockManifest); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SpineBlockManifestQuery.InsertSpineBlockManifest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpineBlockManifestQuery_GetLastSpineBlockManifest(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		ct     chaintype.ChainType
		mbType model.SpineBlockManifestType
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "GetLastSpineBlockManifest:success",
			fields: fields{
				Fields:    NewSpineBlockManifestQuery().Fields,
				TableName: NewSpineBlockManifestQuery().TableName,
			},
			args: args{
				ct:     &chaintype.MainChain{},
				mbType: model.SpineBlockManifestType_Snapshot,
			},
			want: "SELECT id, full_file_hash, file_chunk_hashes, manifest_reference_height, chain_type, manifest_type, " +
				"manifest_timestamp FROM spine_block_manifest WHERE chain_type = 0 AND manifest_type = 0 ORDER BY manifest_reference_height" +
				" DESC LIMIT 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mbl := &SpineBlockManifestQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := mbl.GetLastSpineBlockManifest(tt.args.ct, tt.args.mbType); got != tt.want {
				t.Errorf("SpineBlockManifestQuery.GetLastSpineBlockManifest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpineBlockManifestQuery_GetSpineBlockManifestsInTimeInterval(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		fromTimestamp int64
		toTimestamp   int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "GetSpineBlockManifestTimeInterval:success",
			fields: fields{
				Fields:    NewSpineBlockManifestQuery().Fields,
				TableName: NewSpineBlockManifestQuery().TableName,
			},
			args: args{
				fromTimestamp: 10,
				toTimestamp:   20,
			},
			want: "SELECT id, full_file_hash, file_chunk_hashes, manifest_reference_height, chain_type, manifest_type, " +
				"manifest_timestamp FROM spine_block_manifest WHERE manifest_timestamp > 10 AND manifest_timestamp <= 20 ORDER" +
				" BY manifest_type, chain_type, manifest_reference_height",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mbl := &SpineBlockManifestQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := mbl.GetSpineBlockManifestTimeInterval(tt.args.fromTimestamp, tt.args.toTimestamp); got != tt.want {
				t.Errorf("SpineBlockManifestQuery.GetSpineBlockManifestTimeInterval() = %v, want %v", got, tt.want)
			}
		})
	}
}
