package query

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

var (
	sc1 = &model.SnapshotChunk{
		SpineBlockHeight:  1,
		PreviousChunkHash: nil,
		ChunkHash: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
			1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		ChunkIndex: 0,
	}
)

func TestSnapshotChunk_InsertSnapshotChunk(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		snapshotChunk *model.SnapshotChunk
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantStr  string
		wantArgs []interface{}
	}{
		{
			name: "InsertSnapshotChunk:success",
			fields: fields{
				Fields:    NewSnapshotChunkQuery().Fields,
				TableName: NewSnapshotChunkQuery().TableName,
			},
			args: args{
				snapshotChunk: sc1,
			},
			wantStr: "INSERT INTO snapshot_chunk (chunk_hash,chunk_index,previous_chunk_hash,spine_block_height) VALUES(? , ?, " +
				"?, ?)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scl := &SnapshotChunkQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, _ := scl.InsertSnapshotChunk(tt.args.snapshotChunk)
			if gotStr != tt.wantStr {
				t.Errorf("SnapshotChunkQuery.InsertSnapshotChunk() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
		})
	}
}

func TestSnapshotChunkQuery_GetSnapshotChunksByBlockHeight(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		height uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantStr string
	}{
		{
			name: "GetSnapshotChunksByBlockHeight:success",
			fields: fields{
				Fields:    NewSnapshotChunkQuery().Fields,
				TableName: NewSnapshotChunkQuery().TableName,
			},
			args: args{
				height: 1,
			},
			wantStr: "SELECT chunk_hash, chunk_index, previous_chunk_hash, " +
				"spine_block_height FROM snapshot_chunk WHERE spine_block_height = 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scl := &SnapshotChunkQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if gotStr := scl.GetSnapshotChunksByBlockHeight(tt.args.height); gotStr != tt.wantStr {
				t.Errorf("SnapshotChunkQuery.GetSnapshotChunksByBlockHeight() = %v, want %v", gotStr, tt.wantStr)
			}
		})
	}
}

func TestSnapshotChunkQuery_GetSnapshotChunkByChunkHash(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		chunkHash []byte
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantStr  string
		wantArgs []interface{}
	}{
		{
			name: "GetSnapshotChunksByBlockHeight:success",
			fields: fields{
				Fields:    NewSnapshotChunkQuery().Fields,
				TableName: NewSnapshotChunkQuery().TableName,
			},
			args: args{
				chunkHash: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
					1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			},
			wantStr: "SELECT chunk_hash, chunk_index, previous_chunk_hash, " +
				"spine_block_height FROM snapshot_chunk WHERE chunk_hash = ?",
			wantArgs: []interface{}{[]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scl := &SnapshotChunkQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := scl.GetSnapshotChunkByChunkHash(tt.args.chunkHash)
			if gotStr != tt.wantStr {
				t.Errorf("SnapshotChunkQuery.GetSnapshotChunkByChunkHash() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("SnapshotChunkQuery.GetSnapshotChunkByChunkHash() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestSnapshotChunkQuery_GetLastSnapshotChunk(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "GetLastSnapshotChunk:success",
			fields: fields{
				Fields:    NewSnapshotChunkQuery().Fields,
				TableName: NewSnapshotChunkQuery().TableName,
			},
			want: "SELECT chunk_hash, chunk_index, previous_chunk_hash, " +
				"spine_block_height FROM snapshot_chunk ORDER BY spine_block_height, chunk_index DESC LIMIT 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scl := &SnapshotChunkQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := scl.GetLastSnapshotChunk(); got != tt.want {
				t.Errorf("SnapshotChunkQuery.GetLastSnapshotChunk() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSnapshotChunkQuery_Rollback(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		spineBlockHeight uint32
	}
	var (
		want [][]interface{}
	)

	tests := []struct {
		name   string
		fields fields
		args   args
		want   [][]interface{}
	}{
		{
			name: "GetSnapshotChunksByBlockHeight:success",
			fields: fields{
				Fields:    NewSnapshotChunkQuery().Fields,
				TableName: NewSnapshotChunkQuery().TableName,
			},
			args: args{
				spineBlockHeight: 1,
			},
			want: append(want, append([]interface{}{"DELETE FROM snapshot_chunk WHERE spine_block_height > ?"}, uint32(1))),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scl := &SnapshotChunkQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := scl.Rollback(tt.args.spineBlockHeight); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SnapshotChunkQuery.Rollback() = %v, want %v", got, tt.want)
			}
		})
	}
}
