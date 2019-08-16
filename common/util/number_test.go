package util

import "testing"

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
