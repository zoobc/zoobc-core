// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
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
