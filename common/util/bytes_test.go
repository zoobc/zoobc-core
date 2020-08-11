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
