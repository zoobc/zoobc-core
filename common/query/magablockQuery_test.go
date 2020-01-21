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
		FullFileHash:     make([]byte, 64), // sha3-512
		SpineBlockHeight: 1,
		MegablockHeight:  720,
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
			want: "INSERT INTO megablock (full_file_hash,megablock_payload_length,megablock_payload_hash,spine_block_height," +
				"megablock_height,chain_type," +
				"megablock_type) VALUES(? , ?, ?, ?, ?, ?, ?)",
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

func TestMegablockQuery_Rollback(t *testing.T) {
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
			name: "RollBack:success",
			fields: fields{
				Fields:    NewMegablockQuery().Fields,
				TableName: NewMegablockQuery().TableName,
			},
			args: args{
				spineBlockHeight: 1,
			},
			want: append(want, append([]interface{}{"DELETE FROM megablock WHERE spine_block_height > ?"}, uint32(1))),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mbl := &MegablockQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if gotMultiQueries := mbl.Rollback(tt.args.spineBlockHeight); !reflect.DeepEqual(gotMultiQueries, tt.want) {
				t.Errorf("MegablockQuery.Rollback() = %v, want %v", gotMultiQueries, tt.want)
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
			want: "SELECT full_file_hash, megablock_payload_length, megablock_payload_hash, spine_block_height, " +
				"megablock_height, chain_type, " +
				"megablock_type FROM megablock WHERE chain_type = 0 AND megablock_type = 0 ORDER BY spine_block_height DESC" +
				" LIMIT 1",
		}}
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

func TestMegablockQuery_GetMegablocksBySpineBlockHeightAndChaintypeAndMegablockType(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		height uint32
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
			name: "GetMegablocksBySpineBlockHeightAndChaintypeAndMegablockType:mainchain-snapshots",
			fields: fields{
				Fields:    NewMegablockQuery().Fields,
				TableName: NewMegablockQuery().TableName,
			},
			args: args{
				height: 1,
				ct:     &chaintype.MainChain{},
				mbType: model.MegablockType_Snapshot,
			},
			want: "SELECT full_file_hash, megablock_payload_length, megablock_payload_hash, spine_block_height, " +
				"megablock_height, chain_type, " +
				"megablock_type FROM megablock WHERE spine_block_height = 1 AND chain_type = 0 AND megablock_type = 0 LIMIT 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mbl := &MegablockQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if got := mbl.GetMegablocksBySpineBlockHeightAndChaintypeAndMegablockType(
				tt.args.height,
				tt.args.ct,
				tt.args.mbType,
			); got != tt.want {
				t.Errorf("MegablockQuery.GetMegablocksBySpineBlockHeightAndChaintypeAndMegablockType() "+
					"= %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMegablockQuery_GetMegablocksBySpineBlockHeight(t *testing.T) {
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
			name: "GetMegablocksBySpineBlockHeight:success",
			fields: fields{
				Fields:    NewMegablockQuery().Fields,
				TableName: NewMegablockQuery().TableName,
			},
			args: args{
				height: 1,
			},
			wantStr: "SELECT full_file_hash, megablock_payload_length, megablock_payload_hash, spine_block_height, " +
				"megablock_height, chain_type, " +
				"megablock_type FROM megablock WHERE spine_block_height = 1 ORDER BY megablock_type, chain_type, id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mbl := &MegablockQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if gotStr := mbl.GetMegablocksBySpineBlockHeight(tt.args.height); gotStr != tt.wantStr {
				t.Errorf("MegablockQuery.GetMegablocksBySpineBlockHeight() = %v, want %v", gotStr, tt.wantStr)
			}
		})
	}
}
