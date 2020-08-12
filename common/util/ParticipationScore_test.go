package util

import (
	"testing"

	"github.com/zoobc/zoobc-core/common/constant"
)

func TestCalculateParticipationScore(t *testing.T) {
	type args struct {
		linkedReceipt   uint32
		unlinkedReceipt uint32
		maxReceipt      uint32
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "wantSuccess:full-minus",
			args: args{
				linkedReceipt:   0,
				unlinkedReceipt: 0,
				maxReceipt:      10,
			},
			want:    -1000000000,
			wantErr: false,
		},
		{
			name: "wantSuccess:full-plus",
			args: args{
				linkedReceipt:   10,
				unlinkedReceipt: 0,
				maxReceipt:      10,
			},
			want:    1000000000,
			wantErr: false,
		},
		{
			name: "wantSuccess:half-linked-half-non-linked",
			args: args{
				linkedReceipt:   5,
				unlinkedReceipt: 5,
				maxReceipt:      10,
			},
			want:    250000000,
			wantErr: false,
		},
		{
			name: "wantSuccess:all-unlinked",
			args: args{
				linkedReceipt:   0,
				unlinkedReceipt: 20,
				maxReceipt:      20,
			},
			want:    -500000000,
			wantErr: false,
		},
		{
			name: "wantError",
			args: args{
				linkedReceipt:   5,
				unlinkedReceipt: 100,
				maxReceipt:      10,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "wantError:max=0",
			args: args{
				linkedReceipt:   0,
				unlinkedReceipt: 0,
				maxReceipt:      0,
			},
			want:    constant.MaxScoreChange,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateParticipationScore(
				tt.args.linkedReceipt,
				tt.args.unlinkedReceipt,
				tt.args.maxReceipt,
			)
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
