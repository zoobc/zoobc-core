package query

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
)

func TestMegablockQuery_GetMegablocksByBlockHeight(t *testing.T) {
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
			name: "GetMegablocksByBlockHeight:mainchain",
			fields: fields{
				Fields:    NewMegablockQuery().Fields,
				TableName: NewMegablockQuery().TableName,
			},
			args: args{
				height: 1,
				ct:     &chaintype.MainChain{},
			},
			wantStr: "SELECT full_snapshot_hash, spine_block_height, main_block_height FROM megablock WHERE main_block_height = 1",
		},
		{
			name: "GetMegablocksByBlockHeight:mainchain",
			fields: fields{
				Fields:    NewMegablockQuery().Fields,
				TableName: NewMegablockQuery().TableName,
			},
			args: args{
				height: 1,
				ct:     &chaintype.SpineChain{},
			},
			wantStr: "SELECT full_snapshot_hash, spine_block_height, " +
				"main_block_height FROM megablock WHERE spine_block_height = 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mbl := &MegablockQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
			}
			if gotStr := mbl.GetMegablocksByBlockHeight(tt.args.height, tt.args.ct); gotStr != tt.wantStr {
				t.Errorf("MegablockQuery.GetMegablocksByBlockHeight() = %v, want %v", gotStr, tt.wantStr)
			}
		})
	}
}

func TestMegablockQuery_InsertMegablock(t *testing.T) {
	type fields struct {
		Fields    []string
		TableName string
	}
	type args struct {
		megablock *model.Megablock
	}

	mb1 := &model.Megablock{
		FullSnapshotHash: make([]byte, 64), // sha3-512
		SpineBlockHeight: 1,
		MainBlockHeight:  720,
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
			want: "INSERT INTO megablock (full_snapshot_hash,spine_block_height," +
				"main_block_height) VALUES(? , ?, ?)",
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
		want1 [][]interface{}
	)
	want1 = append(want1, append([]interface{}{"DELETE FROM megablock WHERE spine_block_height > ?"}, uint32(1)))
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
			want: want1,
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
