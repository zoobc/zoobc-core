package util

import "testing"

func TestGetDayOfMonthUTC(t *testing.T) {
	type args struct {
		timestamp int64
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "wantSuccess",
			args: args{
				timestamp: 1,
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDayOfMonthUTC(tt.args.timestamp); got != tt.want {
				t.Errorf("GetDayOfMonthUTC() = %v, want %v", got, tt.want)
			}
		})
	}
}
