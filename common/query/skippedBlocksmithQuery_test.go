package query

import (
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
)

func TestSkippedBlocksmithQuery_SelectDataForSnapshot(t *testing.T) {
	qry := NewSkippedBlocksmithQuery()
	type fields struct {
		Fields    []string
		TableName string
		ChainType chaintype.ChainType
	}
	type args struct {
		fromHeight uint32
		toHeight   uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "SelectDataForSnapshot",
			fields: fields{
				Fields:    qry.Fields,
				TableName: qry.TableName,
				ChainType: &chaintype.MainChain{},
			},
			args: args{
				fromHeight: 0,
				toHeight:   10,
			},
			want: "SELECT blocksmith_public_key, pop_change, block_height, " +
				"blocksmith_index FROM skipped_blocksmith WHERE block_height >= 0 AND block_height <= 10 ORDER BY block_height",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sbq := &SkippedBlocksmithQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
				ChainType: tt.fields.ChainType,
			}
			if got := sbq.SelectDataForSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("SkippedBlocksmithQuery.SelectDataForSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSkippedBlocksmithQuery_TrimDataBeforeSnapshot(t *testing.T) {
	qry := NewSkippedBlocksmithQuery()
	type fields struct {
		Fields    []string
		TableName string
		ChainType chaintype.ChainType
	}
	type args struct {
		fromHeight uint32
		toHeight   uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "TrimDataBeforeSnapshot",
			fields: fields{
				Fields:    qry.Fields,
				TableName: qry.TableName,
				ChainType: &chaintype.MainChain{},
			},
			args: args{
				fromHeight: 0,
				toHeight:   10,
			},
			want: "DELETE FROM skipped_blocksmith WHERE block_height >= 0 AND block_height <= 10",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sbq := &SkippedBlocksmithQuery{
				Fields:    tt.fields.Fields,
				TableName: tt.fields.TableName,
				ChainType: tt.fields.ChainType,
			}
			if got := sbq.TrimDataBeforeSnapshot(tt.args.fromHeight, tt.args.toHeight); got != tt.want {
				t.Errorf("SkippedBlocksmithQuery.TrimDataBeforeSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}
