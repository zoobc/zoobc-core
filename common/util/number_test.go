package util

import (
	"testing"
)

func TestMinUint32(t *testing.T) {
	type args struct {
		number1 uint32
		number2 uint32
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{
			name: "TestMinUint32: first number is smaller",
			args: args{
				number1: 1,
				number2: 2,
			},
			want: 1,
		},
		{
			name: "TestMinUint32: second number is smaller",
			args: args{
				number1: 2,
				number2: 1,
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MinUint32(tt.args.number1, tt.args.number2); got != tt.want {
				t.Errorf("TestMinUint32() = %v want %v", got, tt.want)
			}
		})
	}
}

func TestMaxUint32(t *testing.T) {
	type args struct {
		number1 uint32
		number2 uint32
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{
			name: "TestMaxUint32: first number is larger",
			args: args{
				number1: 2,
				number2: 1,
			},
			want: 2,
		},
		{
			name: "TestMaxUint32: second number is larger",
			args: args{
				number1: 1,
				number2: 2,
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaxUint32(tt.args.number1, tt.args.number2); got != tt.want {
				t.Errorf("TestMaxUint32() = %v want %v", got, tt.want)
			}
		})
	}
}

func TestGetNextStep(t *testing.T) {
	type args struct {
		curStep  int64
		interval int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "GetNextSnapshotHeight:success-{height_same_as_nextStep}",
			args: args{
				curStep:  74057,
				interval: 74057,
			},
			want: int64(74057),
		},
		{
			name: "GetNextSnapshotHeight:success-{height_lower_than_nextStep}",
			args: args{
				curStep:  1000,
				interval: 74057,
			},
			want: int64(74057),
		},
		{
			name: "GetNextSnapshotHeight:success-{height_higher_than_nextStep}",
			args: args{
				curStep:  84057,
				interval: 74057,
			},
			want: int64(148114),
		},
		{
			name: "GetNextSnapshotHeight:success-{height_more_than_double_nextStep}",
			args: args{
				curStep:  148115,
				interval: 74057,
			},
			want: int64(222171),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNextStep(tt.args.curStep, tt.args.interval); got != tt.want {
				t.Errorf("GetNextStep() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMinInt64(t *testing.T) {
	type args struct {
		x int64
		y int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "want-x-smaller",
			args: args{
				x: 1,
				y: 2,
			},
			want: 1,
		},
		{
			name: "want-y-smaller",
			args: args{
				x: 2,
				y: 1,
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MinInt64(tt.args.x, tt.args.y); got != tt.want {
				t.Errorf("MinInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMaxInt64(t *testing.T) {
	type args struct {
		x int64
		y int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "want-x-larger",
			args: args{
				x: 2,
				y: 1,
			},
			want: 2,
		},
		{
			name: "want-y-larger",
			args: args{
				x: 1,
				y: 2,
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaxInt64(tt.args.x, tt.args.y); got != tt.want {
				t.Errorf("MaxInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}
