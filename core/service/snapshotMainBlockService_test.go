package service

import (
	"github.com/zoobc/zoobc-core/common/constant"
	"testing"

	"github.com/zoobc/zoobc-core/common/chaintype"
)

type (
	mockChainType struct {
		chaintype.MainChain
		SnapshotInterval uint32
	}
)

func (mct *mockChainType) GetSnapshotInterval() uint32 {
	return mct.SnapshotInterval
}

func TestSnapshotMainBlockService_IsSnapshotHeight(t *testing.T) {
	type fields struct {
		chainType chaintype.ChainType
	}
	type args struct {
		height uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "IsSnapshotHeight_{interval_lower_than_minRollback_1}:",
			fields: fields{
				chainType: &mockChainType{
					SnapshotInterval: 10,
				},
			},
			args: args{
				height: 1,
			},
			want: false,
		},
		{
			name: "IsSnapshotHeight_{interval_lower_than_minRollback_2}:",
			fields: fields{
				chainType: &mockChainType{
					SnapshotInterval: 10,
				},
			},
			args: args{
				height: constant.MinRollbackBlocks,
			},
			want: true,
		},
		{
			name: "IsSnapshotHeight_{interval_lower_than_minRollback_3}:",
			fields: fields{
				chainType: &mockChainType{
					SnapshotInterval: 10,
				},
			},
			args: args{
				height: constant.MinRollbackBlocks + 9,
			},
			want: false,
		},
		{
			name: "IsSnapshotHeight_{interval_lower_than_minRollback_4}:",
			fields: fields{
				chainType: &mockChainType{
					SnapshotInterval: 10,
				},
			},
			args: args{
				height: constant.MinRollbackBlocks + 10,
			},
			want: true,
		},
		{
			name: "IsSnapshotHeight_{interval_lower_than_minRollback_5}:",
			fields: fields{
				chainType: &mockChainType{
					SnapshotInterval: 10,
				},
			},
			args: args{
				height: constant.MinRollbackBlocks + 20,
			},
			want: true,
		},
		{
			name: "IsSnapshotHeight_{interval_higher_than_minRollback_1}:",
			fields: fields{
				chainType: &mockChainType{
					SnapshotInterval: constant.MinRollbackBlocks + 10,
				},
			},
			args: args{
				height: constant.MinRollbackBlocks,
			},
			want: false,
		},
		{
			name: "IsSnapshotHeight_{interval_higher_than_minRollback_2}:",
			fields: fields{
				chainType: &mockChainType{
					SnapshotInterval: constant.MinRollbackBlocks + 10,
				},
			},
			args: args{
				height: constant.MinRollbackBlocks,
			},
			want: false,
		},
		{
			name: "IsSnapshotHeight_{interval_higher_than_minRollback_3}:",
			fields: fields{
				chainType: &mockChainType{
					SnapshotInterval: constant.MinRollbackBlocks + 10,
				},
			},
			args: args{
				height: constant.MinRollbackBlocks + 10,
			},
			want: true,
		},
		{
			name: "IsSnapshotHeight_{interval_higher_than_minRollback_4}:",
			fields: fields{
				chainType: &mockChainType{
					SnapshotInterval: constant.MinRollbackBlocks + 10,
				},
			},
			args: args{
				height: 2 * (constant.MinRollbackBlocks + 10),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := &SnapshotMainBlockService{
				chainType: tt.fields.chainType,
			}
			if got := ss.IsSnapshotHeight(tt.args.height); got != tt.want {
				t.Errorf("SnapshotMainBlockService.IsSnapshotHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}
