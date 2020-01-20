package query

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
)

var (
	sc1 = &model.FileChunk{
		SpineBlockHeight:  1,
		PreviousChunkHash: nil,
		ChunkHash: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
			1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		ChunkIndex: 0,
	}
)

func TestFileChunk_InsertFileChunk(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		snapshotChunk *model.FileChunk
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantStr  string
		wantArgs []interface{}
	}{
		{
			name: "InsertFileChunk:success",
			fields: fields{
				Fields:    NewFileChunkQuery().Fields,
				TableName: NewFileChunkQuery().TableName,
			},
			args: args{
				snapshotChunk: sc1,
			},
			wantStr: "INSERT INTO file_chunk (chunk_hash,megablock_id,chunk_index,previous_chunk_hash,spine_block_height," +
				"chain_type) VALUES(? , ?, ?, ?, ?, ?)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scl := &FileChunkQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, _ := scl.InsertFileChunk(tt.args.snapshotChunk)
			if gotStr != tt.wantStr {
				t.Errorf("FileChunkQuery.InsertFileChunk() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
		})
	}
}

func TestFileChunkQuery_GetFileChunksByBlockHeight(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		height uint32
		ct     chaintype.ChainType
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantStr string
	}{
		{
			name: "GetFileChunksByBlockHeight:success",
			fields: fields{
				Fields:    NewFileChunkQuery().Fields,
				TableName: NewFileChunkQuery().TableName,
			},
			args: args{
				height: 1,
				ct:     &chaintype.MainChain{},
			},
			wantStr: "SELECT chunk_hash, megablock_id, chunk_index, previous_chunk_hash, spine_block_height, " +
				"chain_type FROM file_chunk WHERE spine_block_height = 1 AND chain_type = 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scl := &FileChunkQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if gotStr := scl.GetFileChunksByBlockHeight(tt.args.height, tt.args.ct); gotStr != tt.wantStr {
				t.Errorf("FileChunkQuery.GetFileChunksByBlockHeight() = %v, want %v", gotStr, tt.wantStr)
			}
		})
	}
}

func TestFileChunkQuery_GetFileChunkByChunkHash(t *testing.T) {
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
			name: "GetFileChunksByBlockHeight:success",
			fields: fields{
				Fields:    NewFileChunkQuery().Fields,
				TableName: NewFileChunkQuery().TableName,
			},
			args: args{
				chunkHash: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
					1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			},
			wantStr: "SELECT chunk_hash, megablock_id, chunk_index, previous_chunk_hash, spine_block_height, " +
				"chain_type FROM file_chunk WHERE chunk_hash = ?",
			wantArgs: []interface{}{[]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
				1, 1}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scl := &FileChunkQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			gotStr, gotArgs := scl.GetFileChunkByChunkHash(tt.args.chunkHash)
			if gotStr != tt.wantStr {
				t.Errorf("FileChunkQuery.GetFileChunkByChunkHash() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("FileChunkQuery.GetFileChunkByChunkHash() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestFileChunkQuery_GetLastFileChunk(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		ct chaintype.ChainType
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "GetLastFileChunk:success",
			fields: fields{
				Fields:    NewFileChunkQuery().Fields,
				TableName: NewFileChunkQuery().TableName,
			},
			args: args{
				ct: &chaintype.MainChain{},
			},
			want: "SELECT chunk_hash, megablock_id, chunk_index, previous_chunk_hash, spine_block_height, " +
				"chain_type FROM file_chunk WHERE chain_type = 0 ORDER BY spine_block_height, chunk_index DESC LIMIT 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scl := &FileChunkQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := scl.GetLastFileChunk(tt.args.ct); got != tt.want {
				t.Errorf("FileChunkQuery.GetLastFileChunk() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileChunkQuery_Rollback(t *testing.T) {
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
			name: "GetFileChunksByBlockHeight:success",
			fields: fields{
				Fields:    NewFileChunkQuery().Fields,
				TableName: NewFileChunkQuery().TableName,
			},
			args: args{
				spineBlockHeight: 1,
			},
			want: append(want, append([]interface{}{"DELETE FROM file_chunk WHERE spine_block_height > ?"}, uint32(1))),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scl := &FileChunkQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := scl.Rollback(tt.args.spineBlockHeight); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FileChunkQuery.Rollback() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileChunkQuery_GetFileChunksByMegablockID(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		megablockID int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantStr string
	}{
		{
			name: "GetFileChunksByMegablockID:success",
			fields: fields{
				Fields:    NewFileChunkQuery().Fields,
				TableName: NewFileChunkQuery().TableName,
			},
			args: args{
				megablockID: int64(1),
			},
			wantStr: "SELECT chunk_hash, megablock_id, chunk_index, previous_chunk_hash, spine_block_height, " +
				"chain_type FROM file_chunk WHERE megablock_id = 1",
		},	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scl := &FileChunkQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if gotStr := scl.GetFileChunksByMegablockID(tt.args.megablockID); gotStr != tt.wantStr {
				t.Errorf("FileChunkQuery.GetFileChunksByMegablockID() = %v, want %v", gotStr, tt.wantStr)
			}
		})
	}
}
