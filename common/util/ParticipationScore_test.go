package util

import (
	"testing"

	"github.com/zoobc/zoobc-core/common/constant"
)

func TestCalculateParticipationScore(t *testing.T) {
	type args struct {
		linkedReceipt   uint32
		unlinkedReceipt uint32
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "wantSuccess",
			args: args{
				linkedReceipt:   constant.MaxReceipt - 13,
				unlinkedReceipt: 13,
			},
			want:    25000000,
			wantErr: false,
		},
		{
			name: "wantError",
			args: args{
				linkedReceipt:   13,
				unlinkedReceipt: constant.MaxReceipt + 1,
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateParticipationScore(tt.args.linkedReceipt, tt.args.unlinkedReceipt)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateParticipationScore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CalculateParticipationScore() = %v, want %v", got, tt.want)
			}
		})
	}
}
