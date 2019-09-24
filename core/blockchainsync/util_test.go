package blockchainsync

import (
	"testing"

	"github.com/zoobc/zoobc-core/common/constant"
)

func TestUtil_getMinRollbackHeight(t *testing.T) {
	type fields struct {
		Height uint32
	}
	tests := []struct {
		name   string
		fields fields
		want   uint32
	}{
		// TODO: Add test cases.
		{
			name: "GetMinRollbackHeight Successful",
			fields: fields{
				Height: constant.MinRollbackBlocks - 1,
			},
			want: 0,
		},
		{
			name: "GetMinROllbackHeight Failed",
			fields: fields{
				Height: constant.MinRollbackBlocks + 1,
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getMinRollbackHeight(tt.fields.Height)
			if got != tt.want {
				t.Errorf("Service.getMinRollbackHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}
