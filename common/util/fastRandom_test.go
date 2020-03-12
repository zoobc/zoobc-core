package util

import (
	"math/rand"
	"testing"
)

func TestGetFastRandom(t *testing.T) {
	type args struct {
		seed *rand.Rand
		max  int
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "GetFastRandom:firstFixedSeed",
			args: args{
				max:  10000,
				seed: rand.New(rand.NewSource(1)),
			},
			want: 8081,
		},
		{
			name: "GetFastRandom:secondFixedSeed",
			args: args{
				max:  10000,
				seed: rand.New(rand.NewSource(10)),
			},
			want: 3454,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetFastRandom(tt.args.seed, tt.args.max); got != tt.want {
				t.Errorf("GetFastRandom() = %v, want %v", got, tt.want)
			}
		})
	}
}
