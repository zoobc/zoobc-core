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
