package query

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
)

func TestMegablockQuery_InsertMegablock(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		megablock *model.Megablock
	}

	mb1 := &model.Megablock{
		FullFileHash:    make([]byte, 64), // sha3-512
		MegablockHeight: 720,
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "InsertMegablock:success",
			fields: fields{
				Fields:    NewMegablockQuery().Fields,
				TableName: NewMegablockQuery().TableName,
			},
			args: args{
				megablock: mb1,
			},
			want: "INSERT INTO megablock (id,full_file_hash,file_chunk_hashes,megablock_height,chain_type,megablock_type," +
				"expiration_timestamp) VALUES(? , ?, ?, ?, ?, ?, ?)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mbl := &MegablockQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got, _ := mbl.InsertMegablock(tt.args.megablock); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MegablockQuery.InsertMegablock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMegablockQuery_GetLastMegablock(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		ct     chaintype.ChainType
		mbType model.MegablockType
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "GetLastMegablock:success",
			fields: fields{
				Fields:    NewMegablockQuery().Fields,
				TableName: NewMegablockQuery().TableName,
			},
			args: args{
				ct:     &chaintype.MainChain{},
				mbType: model.MegablockType_Snapshot,
			},
			want: "SELECT id, full_file_hash, file_chunk_hashes, megablock_height, chain_type, megablock_type, " +
				"expiration_timestamp FROM megablock WHERE chain_type = 0 AND megablock_type = 0 ORDER BY megablock_height" +
				" DESC LIMIT 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mbl := &MegablockQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := mbl.GetLastMegablock(tt.args.ct, tt.args.mbType); got != tt.want {
				t.Errorf("MegablockQuery.GetLastMegablock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMegablockQuery_GetMegablocksInTimeInterval(t *testing.T) {
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
			name: "GetMegablocksInTimeInterval:success",
			fields: fields{
				Fields:    NewMegablockQuery().Fields,
				TableName: NewMegablockQuery().TableName,
			},
			args: args{
				fromTimestamp: 10,
				toTimestamp:   20,
			},
			want: "SELECT id, full_file_hash, file_chunk_hashes, megablock_height, chain_type, megablock_type, " +
				"expiration_timestamp FROM megablock WHERE expiration_timestamp > 10 AND expiration_timestamp <= 20 ORDER" +
				" BY megablock_type, chain_type, megablock_height",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mbl := &MegablockQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := mbl.GetMegablocksInTimeInterval(tt.args.fromTimestamp, tt.args.toTimestamp); got != tt.want {
				t.Errorf("MegablockQuery.GetMegablocksInTimeInterval() = %v, want %v", got, tt.want)
			}
		})
	}
}
