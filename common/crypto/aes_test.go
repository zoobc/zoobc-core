package crypto

import (
	"reflect"
	"testing"
)

func TestDecryptString(t *testing.T) {
	type args struct {
		passphrase            string
		encryptedBase64String string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantStr string
		wantErr bool
	}{
		{
			name: "OpenSSLDecrypt:success",
			args: args{
				passphrase:            "ultra-strong-password",
				encryptedBase64String: "U2FsdGVkX1+3xAKFIuJuOYLvFnf50slLil9ZVXqSYdF7F8e4Ty8572n0X+rMziq5",
			},
			want: []byte{84, 104, 105, 115, 32, 115, 101, 110, 116, 101, 110, 99, 101, 32, 105, 115, 32, 115, 117, 112, 101, 114, 32,
				115, 101, 99, 114, 101, 116},
			wantStr: "This sentence is super secret",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := OpenSSLDecrypt(tt.args.passphrase, tt.args.encryptedBase64String)
			if (err != nil) != tt.wantErr {
				t.Errorf("OpenSSLDecrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OpenSSLDecrypt() got = %v, want %v", got, tt.want)
			}
			if string(got) != tt.wantStr {
				t.Errorf("OpenSSLDecrypt() got = %v, want %v", string(got), tt.wantStr)
			}
		})
	}
}
