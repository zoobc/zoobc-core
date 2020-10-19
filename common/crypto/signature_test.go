package crypto

import (
	"bytes"
	"github.com/zoobc/zoobc-core/common/accounttype"
	"github.com/zoobc/zoobc-core/common/blocker"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

func TestNewSignature(t *testing.T) {
	tests := []struct {
		name string
		want *Signature
	}{
		{
			name: "NewSignature:success",
			want: &Signature{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSignature(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSignature() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSignature_Sign(t *testing.T) {
	type args struct {
		payload     []byte
		accountType model.AccountType
		seed        string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "Sign:valid-DefaultSignature",
			args: args{
				payload:     []byte{12, 43, 65, 65, 12, 123, 43, 12, 1, 24, 5, 5, 12, 54},
				accountType: model.AccountType_ZbcAccountType,
				seed:        "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved",
			},
			want: []byte{42, 62, 47, 200, 180, 101, 85, 204, 179, 147, 143, 68, 30, 111, 6, 94, 81, 248, 219, 43, 90, 6, 167,
				45, 132, 96, 130, 0, 153, 244, 159, 137, 159, 113, 78, 9, 164, 154, 213, 255, 17, 206, 153, 156, 176, 206, 33,
				103, 72, 182, 228, 148, 234, 15, 176, 243, 50, 221, 106, 152, 53, 54, 173, 15},
			wantErr: false,
		},
		{
			name: "Sign:valid-BitcoinSignature",
			args: args{
				payload:     []byte{12, 43, 65, 65, 12, 123, 43, 12, 1, 24, 5, 5, 12, 54},
				accountType: model.AccountType_BTCAccountType,
				seed:        "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved",
			},
			want: []byte{33, 0, 3, 82, 247, 192, 243, 36, 207, 71, 90, 3, 103, 220, 47, 115, 64, 15, 13, 59, 186, 231, 45,
				42, 149, 73, 12, 5, 166, 141, 205, 177, 156, 77, 122, 48, 68, 2, 32, 90, 19, 249, 248, 141, 2, 142, 176, 69, 131, 63, 122,
				227, 255, 114, 26, 116, 34, 23, 167, 245, 190, 121, 156, 49, 171, 110, 198, 76, 191, 126, 236, 2, 32, 9, 133, 158, 144,
				106, 172, 10, 8, 201, 172, 22, 1, 23, 102, 80, 158, 55, 191, 144, 127, 111, 186, 226, 211, 3, 203, 131, 93, 140, 126, 90,
				133},
			wantErr: false,
		},
		{
			name: "Sign:invalid-signature-type}",
			args: args{
				payload:     []byte{12, 43, 65, 65, 12, 123, 43, 12, 1, 24, 5, 5, 12, 54},
				accountType: 1011,
				seed:        "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Signature{}
			got, err := s.Sign(tt.args.payload, tt.args.accountType, tt.args.seed)
			if (err != nil) != tt.wantErr {
				t.Errorf("Signature.Sign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Signature.Sign() = \n%v, want  \n%v", got, tt.want)
			}
		})
	}
}

func TestSignature_SignBlock(t *testing.T) {
	type args struct {
		payload  []byte
		nodeSeed string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "SignByNode:success",
			args: args{
				payload:  []byte{12, 43, 65, 65, 12, 123, 43, 12, 1, 24, 5, 5, 12, 54},
				nodeSeed: "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved",
			},
			want: []byte{42, 62, 47, 200, 180, 101, 85, 204, 179, 147, 143, 68, 30, 111, 6, 94, 81, 248, 219, 43, 90, 6, 167,
				45, 132, 96, 130, 0, 153, 244, 159, 137, 159, 113, 78, 9, 164, 154, 213, 255, 17, 206, 153, 156, 176, 206, 33,
				103, 72, 182, 228, 148, 234, 15, 176, 243, 50, 221, 106, 152, 53, 54, 173, 15},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Signature{}
			if got := s.SignByNode(tt.args.payload, tt.args.nodeSeed); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Signature.SignByNode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSignature_VerifySignature(t *testing.T) {
	type args struct {
		payload        []byte
		signature      []byte
		accountAddress []byte
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "VerifySignature:success-{ed25519-signature}",
			args: args{
				payload: []byte{12, 43, 65, 65, 12, 123, 43, 12, 1, 24, 5, 5, 12, 54},
				accountAddress: []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56,
					139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
				signature: []byte{42, 62, 47, 200, 180, 101, 85, 204, 179, 147, 143, 68, 30, 111, 6, 94, 81, 248, 219, 43, 90, 6, 167,
					45, 132, 96, 130, 0, 153, 244, 159, 137, 159, 113, 78, 9, 164, 154, 213, 255, 17, 206, 153, 156, 176, 206, 33,
					103, 72, 182, 228, 148, 234, 15, 176, 243, 50, 221, 106, 152, 53, 54, 173, 15},
			},
			want: nil,
		},
		{
			name: "VerifySignature:success-{bitcoin-signature}",
			args: args{
				payload: []byte{12, 43, 65, 65, 12, 123, 43, 12, 1, 24, 5, 5, 12, 54},
				accountAddress: []byte{1, 0, 0, 0, 3, 82, 247, 192, 243, 36, 207, 71, 90, 3, 103, 220, 47, 115, 64, 15, 13, 59, 186, 231,
					45, 42, 149, 73, 12, 5, 166, 141, 205, 177, 156, 77, 122},
				signature: []byte{33, 0, 3, 82, 247, 192, 243, 36, 207, 71, 90, 3, 103, 220, 47, 115, 64, 15, 13, 59, 186, 231, 45,
					42, 149, 73, 12, 5, 166, 141, 205, 177, 156, 77, 122, 48, 68, 2, 32, 90, 19, 249, 248, 141, 2, 142, 176, 69, 131, 63, 122,
					227, 255, 114, 26, 116, 34, 23, 167, 245, 190, 121, 156, 49, 171, 110, 198, 76, 191, 126, 236, 2, 32, 9, 133, 158, 144,
					106, 172, 10, 8, 201, 172, 22, 1, 23, 102, 80, 158, 55, 191, 144, 127, 111, 186, 226, 211, 3, 203, 131, 93, 140, 126, 90,
					133},
			},
			want: nil,
		},
		{
			name: "VerifySignature:failed-{invalid-signature}",
			args: args{
				payload: []byte{12, 43, 65, 65, 12, 123, 43, 12, 1, 24, 5, 5, 12, 54},
				accountAddress: []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56,
					139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
				signature: []byte{255, 255, 0, 255, 42, 62, 47, 200, 180, 101, 85, 204, 179, 147, 143, 68, 30, 111, 6, 94, 81, 248, 219, 43, 90, 6, 167,
					45, 132, 96, 130, 0, 153, 244, 159, 137, 159, 113, 78, 9, 164, 154, 213, 255, 17, 206, 153, 156, 176, 206, 33,
					103, 72, 182, 228, 148, 234, 15, 176, 243, 50, 221, 106, 152, 53, 54, 173, 15},
			},
			want: blocker.NewBlocker(
				blocker.ValidationErr,
				"InvalidSignature",
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Signature{}
			if got := s.VerifySignature(tt.args.payload, tt.args.signature, tt.args.accountAddress); got != tt.want {
				t.Errorf("Signature.VerifySignature() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

func TestSignature_VerifyNodeSignature(t *testing.T) {
	type args struct {
		payload       []byte
		signature     []byte
		nodePublicKey []byte
	}
	tests := []struct {
		name string
		s    *Signature
		args args
		want bool
	}{
		{
			name: "VerifyNodeSignature:success",
			args: args{
				payload: []byte{12, 43, 65, 65, 12, 123, 43, 12, 1, 24, 5, 5, 12, 54},
				signature: []byte{42, 62, 47, 200, 180, 101, 85, 204, 179, 147, 143, 68, 30, 111, 6, 94, 81, 248, 219, 43, 90, 6, 167,
					45, 132, 96, 130, 0, 153, 244, 159, 137, 159, 113, 78, 9, 164, 154, 213, 255, 17, 206, 153, 156, 176, 206, 33,
					103, 72, 182, 228, 148, 234, 15, 176, 243, 50, 221, 106, 152, 53, 54, 173, 15},
				nodePublicKey: []byte{4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
					81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Signature{}
			if got := s.VerifyNodeSignature(tt.args.payload, tt.args.signature, tt.args.nodePublicKey); got != tt.want {
				t.Errorf("Signature.VerifyNodeSignature() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSignature_GenerateAccountFromSeed(t *testing.T) {
	type args struct {
		accountType accounttype.AccountTypeInterface
		seed        string
	}
	tests := []struct {
		name                string
		s                   *Signature
		args                args
		wantPrivateKey      []byte
		wantPublicKey       []byte
		wantPublicKeyString string
		wantEncodedAddress  string
		wantFullAddress     []byte
		wantErr             bool
	}{
		{
			name: "GenerateAccountFromSeed:success-{DefaultSignature}",
			args: args{
				accountType: &accounttype.ZbcAccountType{},
				seed:        "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved",
			},
			wantPrivateKey: []byte{215, 143, 134, 166, 8, 238, 10, 130, 59, 25, 200, 58, 125, 85, 55, 94, 206, 50, 194, 93, 71,
				172, 247, 140, 12, 13, 53, 119, 251, 233, 244, 212, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214,
				82, 224, 72, 239, 56, 139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
			wantPublicKey: []byte{4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56,
				139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
			wantPublicKeyString: "ZNK_AQTEIGHG_65MNY534_GOKX7VSS_4BEO6OEL_75I6LOCN_KBICP7VN_DSUUXGSS",
			wantEncodedAddress:  "ZBC_AQTEIGHG_65MNY534_GOKX7VSS_4BEO6OEL_75I6LOCN_KBICP7VN_DSUWBLM7",
			wantFullAddress: []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56,
				139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169},
			wantErr: false,
		},
		{
			name: "GenerateAccountFromSeed:success-{BitcoinSignature}",
			args: args{
				accountType: &accounttype.BTCAccountType{},
				seed:        "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved",
			},
			wantPrivateKey: []byte{215, 143, 134, 166, 8, 238, 10, 130, 59, 25, 200, 58, 125, 85, 55, 94, 206, 50, 194, 93,
				71, 172, 247, 140, 12, 13, 53, 119, 251, 233, 244, 212},
			wantPublicKey: []byte{3, 82, 247, 192, 243, 36, 207, 71, 90, 3, 103, 220, 47, 115, 64, 15, 13, 59, 186, 231,
				45, 42, 149, 73, 12, 5, 166, 141, 205, 177, 156, 77, 122},
			wantPublicKeyString: "0352f7c0f324cf475a0367dc2f73400f0d3bbae72d2a95490c05a68dcdb19c4d7a",
			wantEncodedAddress:  "12Ea6WAMZhFnfM5kjyfrfykqVWFcaWorQ8",
			wantFullAddress: []byte{1, 0, 0, 0, 3, 82, 247, 192, 243, 36, 207, 71, 90, 3, 103, 220, 47, 115, 64, 15, 13, 59, 186, 231,
				45, 42, 149, 73, 12, 5, 166, 141, 205, 177, 156, 77, 122},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Signature{}
			gotPrivateKey, gotPublicKey, gotPublickKeyString, gotAddress, gotFullAccountAddress,
				err := s.GenerateAccountFromSeed(tt.args.accountType,
				tt.args.seed)
			if (err != nil) != tt.wantErr {
				t.Errorf("Signature.GenerateAccountFromSeed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPrivateKey, tt.wantPrivateKey) {
				t.Errorf("Signature.GenerateAccountFromSeed() gotPrivateKey = %v, want %v", gotPrivateKey, tt.wantPrivateKey)
			}
			if !reflect.DeepEqual(gotPublicKey, tt.wantPublicKey) {
				t.Errorf("Signature.GenerateAccountFromSeed() gotPublicKey = %v, want %v", gotPublicKey, tt.wantPublicKey)
			}
			if gotPublickKeyString != tt.wantPublicKeyString {
				t.Errorf("Signature.GenerateAccountFromSeed() gotPublickKeyString = %v, want %v", gotPublickKeyString, tt.wantPublicKeyString)
			}
			if gotAddress != tt.wantEncodedAddress {
				t.Errorf("Signature.GenerateAccountFromSeed() gotAddress = %v, want %v", gotAddress, tt.wantEncodedAddress)
			}
			if !bytes.Equal(gotFullAccountAddress, tt.wantFullAddress) {
				t.Errorf("Signature.GenerateAccountFromSeed() gotFullAddress = %v, want %v", gotFullAccountAddress, tt.wantFullAddress)
			}
		})
	}
}

func TestSignature_GenerateBlockSeed(t *testing.T) {
	type args struct {
		payload  []byte
		nodeSeed string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "GenerateBlockSeed:success",
			args: args{
				payload:  []byte{12, 43, 65, 65, 12, 123, 43, 12, 1, 24, 5, 5, 12, 54},
				nodeSeed: "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved",
			},
			want: []byte{42, 62, 47, 200, 180, 101, 85, 204, 179, 147, 143, 68, 30, 111, 6, 94, 81, 248, 219, 43, 90, 6, 167,
				45, 132, 96, 130, 0, 153, 244, 159, 137, 159, 113, 78, 9, 164, 154, 213, 255, 17, 206, 153, 156, 176, 206, 33,
				103, 72, 182, 228, 148, 234, 15, 176, 243, 50, 221, 106, 152, 53, 54, 173, 15},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Signature{}
			if got := s.GenerateBlockSeed(tt.args.payload, tt.args.nodeSeed); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Signature.GenerateBlockSeed() = %v, want %v", got, tt.want)
			}
		})
	}
}
