package crypto

import "testing"

func TestPseudoRandomGenerator(t *testing.T) {
	type args struct {
		id     uint64
		offset uint64
		algo   int
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		{
			name: "TestPseudoRandomGenerator:success-{PseudoRandomSha3256-NoOffset}",
			args: args{
				offset: 0,
				id:     3014845244095079110,
				algo:   PseudoRandomSha3256,
			},
			want: 18041622792886434681,
		},
		{
			name: "TestPseudoRandomGenerator:success-{PseudoRandomSha3256-NoOffset-2}",
			args: args{
				offset: 0,
				id:     1941309198183084506,
				algo:   PseudoRandomSha3256,
			},
			want: 3953548740169852696,
		},
		{
			name: "TestPseudoRandomGenerator:success-{PseudoRandomSha3256-Offset}",
			args: args{
				offset: 132553774296354339,
				id:     3014845244095079110,
				algo:   PseudoRandomSha3256,
			},
			want: 14251496166035092223,
		},
		{
			name: "TestPseudoRandomGenerator:success-{PseudoRandomXoroshiro128-NoOffset}",
			args: args{
				offset: 0,
				id:     3014845244095079110,
				algo:   PseudoRandomXoroshiro128,
			},
			want: 17061115035045365337,
		},
		{
			name: "TestPseudoRandomGenerator:success-{PseudoRandomXoroshiro128-NoOffset-2}",
			args: args{
				offset: 0,
				id:     1941309198183084506,
				algo:   PseudoRandomXoroshiro128,
			},
			want: 2623596903506267843,
		},
		{
			name: "TestPseudoRandomGenerator:success-{PseudoRandomXoroshiro128-Offset}",
			args: args{
				offset: 132553774296354339,
				id:     3014845244095079110,
				algo:   PseudoRandomXoroshiro128,
			},
			want: 8913237946701621685,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PseudoRandomGenerator(tt.args.id, tt.args.offset, tt.args.algo); got != tt.want {
				t.Errorf("PseudoRandomGenerator() = %v, want %v", got, tt.want)
			}
		})
	}
}
