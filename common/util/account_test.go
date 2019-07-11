package util

import (
	"reflect"
	"testing"
)

func TestGetAccountIDByPublicKey(t *testing.T) {
	type args struct {
		accountType int32
		publicKey   []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "GetAccountIDByPublicKey:success",
			args: args{
				accountType: 0,
				publicKey: []byte{4, 38, 103, 73, 250, 169, 63, 155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242,
					84, 174, 239, 46, 190, 78, 68, 90, 83, 142, 11},
			},
			want: []byte{7, 205, 139, 247, 101, 123, 250, 42, 95, 96, 199, 181, 108, 85, 197, 164, 168, 36, 49, 12, 251, 252,
				209, 82, 181, 112, 94, 41, 107, 240, 83, 180},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAccountIDByPublicKey(tt.args.accountType, tt.args.publicKey); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAccountIDByPublicKey() = %v, want %v", got, tt.want)
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
