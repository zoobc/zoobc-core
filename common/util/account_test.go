package util

import (
	"math"
	"reflect"
	"testing"
)

func TestGetAccountIDByPublicKey(t *testing.T) {
	type args struct {
		accountType uint32
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
			want: []byte{61, 173, 177, 191, 183, 169, 194, 0, 147, 155, 147, 2, 103, 251, 133, 203, 243, 56, 197, 6, 238, 74,
				226, 190, 77, 169, 58, 218, 234, 86, 88, 130},
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

func TestGetPrivateKeyFromSeed(t *testing.T) {
	type args struct {
		seed string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "GetPrivateKeyFromSeed:success",
			args: args{
				seed: "first elegant bird flavor life run february tissue grant begin farm surprise",
			},
			want: []byte{246, 188, 143, 79, 238, 206, 8, 182, 67, 60, 246, 31, 13, 81, 144, 22, 11, 79, 129, 224, 242, 91, 106, 213, 28, 34,
				100, 32, 202, 210, 159, 195, 182, 133, 198, 136, 42, 100, 135, 101, 197, 113, 115, 127, 171, 250, 37, 12, 39, 182, 243, 203,
				196, 207, 176, 95, 108, 116, 117, 192, 61, 215, 173, 206},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetPrivateKeyFromSeed(tt.args.seed)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPrivateKeyFromSeed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPrivateKeyFromSeed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAddressFromPublicKey(t *testing.T) {
	type args struct {
		publicKey []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "GetAddressFromPublicKey:success",
			args: args{
				publicKey: []byte{4, 38, 103, 73, 250, 169, 63, 155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242,
					84, 174, 239, 46, 190, 78, 68, 90, 83, 142, 11},
			},
			want:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			wantErr: false,
		},
		{
			name: "GetAddressFromPublicKey:fail-{public key length < 32}",
			args: args{
				publicKey: []byte{4, 38, 103, 73, 250, 169, 63, 155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242,
					84, 174, 239, 46, 190, 78, 68, 90, 83},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAddressFromPublicKey(tt.args.publicKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAddressFromPublicKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetAddressFromPublicKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPublicKeyFromAddress(t *testing.T) {
	type args struct {
		address string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "GetPublicKeyFromAddress:success",
			args: args{
				address: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			},
			want: []byte{4, 38, 103, 73, 250, 169, 63, 155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242,
				84, 174, 239, 46, 190, 78, 68, 90, 83, 142, 11},
			wantErr: false,
		},
		{
			name: "GetPublicKeyFromAddress:fail-{decode error, wrong address format/length}",
			args: args{
				address: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgt",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "GetPublicKeyFromAddress:fail-{checksum error, wrong address format}",
			args: args{
				address: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtM",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetPublicKeyFromAddress(tt.args.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPublicKeyFromAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPublicKeyFromAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateAccountIDFromAddress(t *testing.T) {
	type args struct {
		accountType uint32
		address     string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "CreateAccountIDFromAddress:success",
			args: args{
				accountType: 0,
				address:     "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			},
			want: []byte{136, 106, 141, 253, 44, 34, 145, 81, 166, 229, 33, 209, 150, 188, 204, 28, 239, 33, 152, 158, 4, 187,
				13, 109, 173, 223, 169, 232, 50, 200, 169, 25},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateAccountIDFromAddress(tt.args.accountType, tt.args.address); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateAccountIDFromAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateAccountAddress(t *testing.T) {
	type args struct {
		accType uint32
		address string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "TestValidateAccountAddress:Default",
			args: args{
				accType: math.MaxUint32,
				address: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
			},
			wantErr: true,
		},
		{
			name: "TestValidateAccountAddress:0:success",
			args: args{
				accType: 0,
				address: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
			},
			wantErr: false,
		},
		{
			name: "TestValidateAccountAddress:0:wantErr",
			args: args{
				accType: 0,
				address: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateAccountAddress(tt.args.accType, tt.args.address); (err != nil) != tt.wantErr {
				t.Errorf("ValidateAccountAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
