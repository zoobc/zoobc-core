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
	"reflect"
	"testing"
)

func TestConvertBytesToUint64(t *testing.T) {
	type args struct {
		bytes []byte
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{

		{
			name: "ConvertBytesToUint64:one",
			args: args{
				[]byte{12, 43, 54, 45, 12, 5, 2, 5},
			},
			want: 360856469999332108,
		},
		{
			name: "ConvertBytesToUint64:two",
			args: args{
				[]byte{12, 43, 54, 45, 12, 5, 2, 54},
			},
			want: 3891678577857800972,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertBytesToUint64(tt.args.bytes); got != tt.want {
				t.Errorf("ConvertBytesToUint64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertBytesToUint32(t *testing.T) {
	type args struct {
		bytes []byte
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{
			name: "ConvertBytesToUint64:one",
			args: args{
				[]byte{12, 43, 54, 45},
			},
			want: 758524684,
		},
		{
			name: "ConvertBytesToUint64:two",
			args: args{
				[]byte{54, 23, 54, 45},
			},
			want: 758519606,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertBytesToUint32(tt.args.bytes); got != tt.want {
				t.Errorf("ConvertBytesToUint32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertBytesToUint16(t *testing.T) {
	type args struct {
		bytes []byte
	}
	tests := []struct {
		name string
		args args
		want uint16
	}{
		{
			name: "ConvertBytesToUint64:one",
			args: args{
				[]byte{12, 43},
			},
			want: 11020,
		},
		{
			name: "ConvertBytesToUint64:two",
			args: args{
				[]byte{54, 23},
			},
			want: 5942,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertBytesToUint16(tt.args.bytes); got != tt.want {
				t.Errorf("ConvertBytesToUint16() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertUint64ToBytes(t *testing.T) {
	type args struct {
		number uint64
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "ConvertBytesToUint64:one",
			args: args{
				360856469999332108,
			},
			want: []byte{12, 43, 54, 45, 12, 5, 2, 5},
		},
		{
			name: "ConvertBytesToUint64:two",
			args: args{
				3891678577857800972,
			},
			want: []byte{12, 43, 54, 45, 12, 5, 2, 54},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertUint64ToBytes(tt.args.number); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertUint64ToBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertUint32ToBytes(t *testing.T) {
	type args struct {
		number uint32
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "ConvertBytesToUint64:one",
			args: args{
				758524684,
			},
			want: []byte{12, 43, 54, 45},
		},
		{
			name: "ConvertBytesToUint64:two",
			args: args{
				758519606,
			},
			want: []byte{54, 23, 54, 45},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertUint32ToBytes(tt.args.number); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertUint32ToBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertUint16ToBytes(t *testing.T) {
	type args struct {
		number uint16
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "ConvertBytesToUint64:one",
			args: args{
				11020,
			},
			want: []byte{12, 43},
		},
		{
			name: "ConvertBytesToUint64:two",
			args: args{
				5942,
			},
			want: []byte{54, 23},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertUint16ToBytes(tt.args.number); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertUint16ToBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertIntToBytes(t *testing.T) {
	type args struct {
		number int
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "ConvertIntToByte:one",
			args: args{
				5942,
			},
			want: []byte{54, 23, 0, 0},
		},
		{
			name: "ConvertIntToByte:two",
			args: args{
				11020,
			},
			want: []byte{12, 43, 0, 0},
		},
		{
			name: "ConvertIntToByte:three",
			args: args{
				758519606,
			},
			want: []byte{54, 23, 54, 45},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertIntToBytes(tt.args.number); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertIntToBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertStringToBytes(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "ConvertStringToBytes:success",
			args: args{
				str: "dummy random string here",
			},
			want: []byte{24, 0, 0, 0, 100, 117, 109, 109, 121, 32, 114, 97, 110, 100, 111, 109, 32, 115, 116, 114, 105, 110, 103, 32,
				104, 101, 114, 101},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertStringToBytes(tt.args.str); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertStringToBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
