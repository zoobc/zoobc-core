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
	"bytes"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/constant"
)

func TestReadTransactionBytes(t *testing.T) {
	type args struct {
		buf    *bytes.Buffer
		nBytes int
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "ReadTransactionBytes:wrong-bytes",
			args: args{
				buf:    bytes.NewBuffer([]byte{1, 2}),
				nBytes: 4,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadTransactionBytes(tt.args.buf, tt.args.nBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadTransactionBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadTransactionBytes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFeePerByteTransaction(t *testing.T) {
	type args struct {
		feeTransaction   int64
		transactionBytes []byte
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "wantSuccess",
			args: args{
				feeTransaction:   1,
				transactionBytes: []byte{1},
			},
			want: constant.OneFeePerByteTransaction,
		},
		{
			name: "wantSuccess:zeroLengthBytes",
			args: args{
				feeTransaction:   1,
				transactionBytes: []byte{},
			},
			want: 1 * constant.OneFeePerByteTransaction,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FeePerByteTransaction(tt.args.feeTransaction, tt.args.transactionBytes); got != tt.want {
				t.Errorf("FeePerByteTransaction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSplitByteSliceByChunkSize(t *testing.T) {
	type args struct {
		b         []byte
		chunkSize int
	}
	tests := []struct {
		name           string
		args           args
		wantSplitSlice [][]byte
	}{
		{
			name: "SplitByteSliceByChunkSize:chunkLength<sliceLength",
			args: args{
				b: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
					45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
				chunkSize: 10,
			},
			wantSplitSlice: [][]byte{
				{153, 58, 50, 200, 7, 61, 108, 229, 204, 48},
				{199, 145, 21, 99, 125, 75, 49, 45, 118, 97},
				{219, 80, 242, 244, 100, 134, 144, 246, 37, 144},
				{213, 135},
			},
		},
		{
			name: "SplitByteSliceByChunkSize:chunkLength=sliceLength",
			args: args{
				b: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
					45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
				chunkSize: 32,
			},
			wantSplitSlice: [][]byte{
				{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
					45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
			},
		},
		{
			name: "SplitByteSliceByChunkSize:chunkLength>sliceLength",
			args: args{
				b: []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
					45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
				chunkSize: 100,
			},
			wantSplitSlice: [][]byte{
				{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
					45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotSplitSlice := SplitByteSliceByChunkSize(tt.args.b, tt.args.chunkSize); !reflect.DeepEqual(gotSplitSlice, tt.wantSplitSlice) {
				t.Errorf("SplitByteSliceByChunkSize() = %v, want %v", gotSplitSlice, tt.wantSplitSlice)
			}
		})
	}
}

func TestGetChecksumByte(t *testing.T) {
	type args struct {
		bytes []byte
	}
	tests := []struct {
		name string
		args args
		want byte
	}{
		{
			name: "GetChecksumByte:success",
			args: args{
				bytes: []byte{1, 2, 3},
			},
			want: 6,
		},
		{
			name: "GetChecksumByte:zeroValue",
			args: args{
				bytes: []byte{254, 1, 1},
			},
			want: 0,
		},
		{
			name: "GetChecksumByte:overFlow",
			args: args{
				bytes: []byte{254, 1, 1, 5},
			},
			want: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetChecksumByte(tt.args.bytes); got != tt.want {
				t.Errorf("GetChecksumByte() = %v, want %v", got, tt.want)
			}
		})
	}
}
