package accounttype

import (
	"bytes"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"reflect"
	"testing"
)

func TestZbcAccountType_GenerateAccountFromSeed(t *testing.T) {
	var (
		seed         = "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved"
		pubKeySlip10 = []byte{149, 1, 110, 5, 224, 150, 132, 85, 59, 205, 45, 168, 107, 143, 209, 215, 181, 221, 109, 23, 39, 95, 248, 147, 114,
			91, 115, 75, 51, 31, 148, 108}
		pubKey = []byte{4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184, 77,
			80, 80, 39, 254, 173, 28, 169}
	)
	type fields struct {
		privateKey      []byte
		publicKey       []byte
		fullAddress     []byte
		publicKeyString string
		encodedAddress  string
	}
	type args struct {
		seed           string
		optionalParams []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    []byte
	}{
		{
			name: "GenerateAccountFromSeed:success-{ed25519-slip10}",
			args: args{
				seed:           seed,
				optionalParams: []interface{}{true},
			},
			want: pubKeySlip10,
		},
		{
			name: "GenerateAccountFromSeed:success-{ed25519}",
			args: args{
				seed: seed,
			},
			want: pubKey,
		},
		{
			name: "GenerateAccountFromSeed:fail-{ed25519-wrongOptionalParam}",
			args: args{
				seed:           seed,
				optionalParams: []interface{}{"invalidParam"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &ZbcAccountType{
				privateKey:      tt.fields.privateKey,
				publicKey:       tt.fields.publicKey,
				fullAddress:     tt.fields.fullAddress,
				publicKeyString: tt.fields.publicKeyString,
				encodedAddress:  tt.fields.encodedAddress,
			}
			if err := acc.GenerateAccountFromSeed(tt.args.seed, tt.args.optionalParams...); (err != nil) != tt.wantErr {
				t.Errorf("GenerateAccountFromSeed() error = %v, wantErr %v", err, tt.wantErr)
			} else if !bytes.Equal(acc.GetAccountPublicKey(), tt.want) {
				t.Errorf("GenerateAccountFromSeed() error = %v, want %v", tt.want, acc.GetAccountPublicKey())
			}
		})
	}
}

func TestZbcAccountType_GetAccountAddress(t *testing.T) {
	var (
		pubKey = []byte{4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184, 77,
			80, 80, 39, 254, 173, 28, 169}
		fullAddr = []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229,
			184, 77, 80, 80, 39, 254, 173, 28, 169}
	)

	type fields struct {
		privateKey      []byte
		publicKey       []byte
		fullAddress     []byte
		publicKeyString string
		encodedAddress  string
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "GetAccountAddress:success",
			fields: fields{
				publicKey: pubKey,
			},
			want: fullAddr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &ZbcAccountType{
				privateKey:      tt.fields.privateKey,
				publicKey:       tt.fields.publicKey,
				fullAddress:     tt.fields.fullAddress,
				publicKeyString: tt.fields.publicKeyString,
				encodedAddress:  tt.fields.encodedAddress,
			}
			got, err := acc.GetAccountAddress()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAccountAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAccountAddress() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZbcAccountType_GetAccountPrefix(t *testing.T) {
	type fields struct {
		privateKey      []byte
		publicKey       []byte
		fullAddress     []byte
		publicKeyString string
		encodedAddress  string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "GetAccountPrefix:success",
			want: constant.PrefixZoobcDefaultAccount,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &ZbcAccountType{
				privateKey:      tt.fields.privateKey,
				publicKey:       tt.fields.publicKey,
				fullAddress:     tt.fields.fullAddress,
				publicKeyString: tt.fields.publicKeyString,
				encodedAddress:  tt.fields.encodedAddress,
			}
			if got := acc.GetAccountPrefix(); got != tt.want {
				t.Errorf("GetAccountPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZbcAccountType_GetAccountPrivateKey(t *testing.T) {
	type fields struct {
		privateKey      []byte
		publicKey       []byte
		fullAddress     []byte
		publicKeyString string
		encodedAddress  string
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name:    "GetAccountPrivateKey:fail-{accountNotGenerated}",
			wantErr: true,
		},
		{
			name: "GetAccountPrivateKey:success",
			fields: fields{
				privateKey: []byte{1, 2, 3},
			},
			want: []byte{1, 2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &ZbcAccountType{
				privateKey:      tt.fields.privateKey,
				publicKey:       tt.fields.publicKey,
				fullAddress:     tt.fields.fullAddress,
				publicKeyString: tt.fields.publicKeyString,
				encodedAddress:  tt.fields.encodedAddress,
			}
			got, err := acc.GetAccountPrivateKey()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAccountPrivateKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAccountPrivateKey() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZbcAccountType_GetAccountPublicKey(t *testing.T) {
	type fields struct {
		privateKey      []byte
		publicKey       []byte
		fullAddress     []byte
		publicKeyString string
		encodedAddress  string
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		{
			name: "GetAccountPublicKey:success",
			fields: fields{
				publicKey: []byte{1, 2, 3},
			},
			want: []byte{1, 2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &ZbcAccountType{
				privateKey:      tt.fields.privateKey,
				publicKey:       tt.fields.publicKey,
				fullAddress:     tt.fields.fullAddress,
				publicKeyString: tt.fields.publicKeyString,
				encodedAddress:  tt.fields.encodedAddress,
			}
			if got := acc.GetAccountPublicKey(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAccountPublicKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZbcAccountType_GetAccountPublicKeyLength(t *testing.T) {
	type fields struct {
		privateKey      []byte
		publicKey       []byte
		fullAddress     []byte
		publicKeyString string
		encodedAddress  string
	}
	tests := []struct {
		name   string
		fields fields
		want   uint32
	}{
		{
			name: "GetAccountPublicKeyLength:success",
			want: 32,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &ZbcAccountType{
				privateKey:      tt.fields.privateKey,
				publicKey:       tt.fields.publicKey,
				fullAddress:     tt.fields.fullAddress,
				publicKeyString: tt.fields.publicKeyString,
				encodedAddress:  tt.fields.encodedAddress,
			}
			if got := acc.GetAccountPublicKeyLength(); got != tt.want {
				t.Errorf("GetAccountPublicKeyLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZbcAccountType_GetAccountPublicKeyString(t *testing.T) {
	var (
		pubKey = []byte{4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184, 77,
			80, 80, 39, 254, 173, 28, 169}
		accountPubKeyStr = "ZNK_AQTEIGHG_65MNY534_GOKX7VSS_4BEO6OEL_75I6LOCN_KBICP7VN_DSUUXGSS"
	)
	type fields struct {
		privateKey      []byte
		publicKey       []byte
		fullAddress     []byte
		publicKeyString string
		encodedAddress  string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "GetAccountPublicKeyString:success-{cached}",
			fields: fields{
				publicKeyString: "aaa",
			},
			want: "aaa",
		},
		{
			name:    "GetAccountPublicKeyString:fail-{EmptyAccountPublicKey}",
			wantErr: true,
		},
		{
			name: "GetAccountPublicKeyString:success",
			fields: fields{
				publicKey: pubKey,
			},
			want: accountPubKeyStr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &ZbcAccountType{
				privateKey:      tt.fields.privateKey,
				publicKey:       tt.fields.publicKey,
				fullAddress:     tt.fields.fullAddress,
				publicKeyString: tt.fields.publicKeyString,
				encodedAddress:  tt.fields.encodedAddress,
			}
			got, err := acc.GetAccountPublicKeyString()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAccountPublicKeyString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetAccountPublicKeyString() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZbcAccountType_GetEncodedAddress(t *testing.T) {
	var (
		pubKey = []byte{4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184, 77,
			80, 80, 39, 254, 173, 28, 169}
		encodedAddr = "ZBC_AQTEIGHG_65MNY534_GOKX7VSS_4BEO6OEL_75I6LOCN_KBICP7VN_DSUWBLM7"
	)
	type fields struct {
		privateKey      []byte
		publicKey       []byte
		fullAddress     []byte
		publicKeyString string
		encodedAddress  string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name:    "GetEncodedAddress:fail-{EmptyAccountPublicKey}",
			wantErr: true,
		},
		{
			name: "GetEncodedAddress:success",
			fields: fields{
				publicKey: pubKey,
			},
			want: encodedAddr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &ZbcAccountType{
				privateKey:      tt.fields.privateKey,
				publicKey:       tt.fields.publicKey,
				fullAddress:     tt.fields.fullAddress,
				publicKeyString: tt.fields.publicKeyString,
				encodedAddress:  tt.fields.encodedAddress,
			}
			got, err := acc.GetEncodedAddress()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEncodedAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetEncodedAddress() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZbcAccountType_GetName(t *testing.T) {
	type fields struct {
		privateKey      []byte
		publicKey       []byte
		fullAddress     []byte
		publicKeyString string
		encodedAddress  string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "GetName:success",
			want: "ZooBC",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &ZbcAccountType{
				privateKey:      tt.fields.privateKey,
				publicKey:       tt.fields.publicKey,
				fullAddress:     tt.fields.fullAddress,
				publicKeyString: tt.fields.publicKeyString,
				encodedAddress:  tt.fields.encodedAddress,
			}
			if got := acc.GetName(); got != tt.want {
				t.Errorf("GetName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZbcAccountType_GetSignatureLength(t *testing.T) {
	type fields struct {
		privateKey      []byte
		publicKey       []byte
		fullAddress     []byte
		publicKeyString string
		encodedAddress  string
	}
	tests := []struct {
		name   string
		fields fields
		want   uint32
	}{
		{
			name: "GetSignatureLength:success",
			want: constant.ZBCSignatureLength,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &ZbcAccountType{
				privateKey:      tt.fields.privateKey,
				publicKey:       tt.fields.publicKey,
				fullAddress:     tt.fields.fullAddress,
				publicKeyString: tt.fields.publicKeyString,
				encodedAddress:  tt.fields.encodedAddress,
			}
			if got := acc.GetSignatureLength(); got != tt.want {
				t.Errorf("GetSignatureLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZbcAccountType_GetSignatureType(t *testing.T) {
	type fields struct {
		privateKey      []byte
		publicKey       []byte
		fullAddress     []byte
		publicKeyString string
		encodedAddress  string
	}
	tests := []struct {
		name   string
		fields fields
		want   model.SignatureType
	}{
		{
			name: "GetSignatureType:success",
			want: model.SignatureType_DefaultSignature,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &ZbcAccountType{
				privateKey:      tt.fields.privateKey,
				publicKey:       tt.fields.publicKey,
				fullAddress:     tt.fields.fullAddress,
				publicKeyString: tt.fields.publicKeyString,
				encodedAddress:  tt.fields.encodedAddress,
			}
			if got := acc.GetSignatureType(); got != tt.want {
				t.Errorf("GetSignatureType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZbcAccountType_GetTypeInt(t *testing.T) {
	type fields struct {
		privateKey      []byte
		publicKey       []byte
		fullAddress     []byte
		publicKeyString string
		encodedAddress  string
	}
	tests := []struct {
		name   string
		fields fields
		want   int32
	}{
		{
			name: "GetTypeInt:success",
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &ZbcAccountType{
				privateKey:      tt.fields.privateKey,
				publicKey:       tt.fields.publicKey,
				fullAddress:     tt.fields.fullAddress,
				publicKeyString: tt.fields.publicKeyString,
				encodedAddress:  tt.fields.encodedAddress,
			}
			if got := acc.GetTypeInt(); got != tt.want {
				t.Errorf("GetTypeInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZbcAccountType_IsEqual(t *testing.T) {
	var (
		pubKey = []byte{4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81, 229, 184, 77,
			80, 80, 39, 254, 173, 28, 169}
		accType = &ZbcAccountType{
			publicKey: pubKey,
		}
	)
	type fields struct {
		privateKey      []byte
		publicKey       []byte
		fullAddress     []byte
		publicKeyString string
		encodedAddress  string
	}
	type args struct {
		acc2 AccountTypeInterface
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "IsEqual:success-{true}",
			args: args{
				acc2: accType,
			},
			fields: fields{
				publicKey: pubKey,
			},
			want: true,
		}, {
			name: "IsEqual:success-{false}",
			args: args{
				acc2: accType,
			},
			fields: fields{
				publicKey: append([]byte{1, 2, 3}, pubKey...),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &ZbcAccountType{
				privateKey:      tt.fields.privateKey,
				publicKey:       tt.fields.publicKey,
				fullAddress:     tt.fields.fullAddress,
				publicKeyString: tt.fields.publicKeyString,
				encodedAddress:  tt.fields.encodedAddress,
			}
			if got := acc.IsEqual(tt.args.acc2); got != tt.want {
				t.Errorf("IsEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZbcAccountType_Sign(t *testing.T) {
	var (
		seed      = "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved"
		payload   = []byte{1, 2, 3}
		signature = []byte{8, 55, 103, 141, 152, 224, 207, 186, 41, 134, 223, 127, 182, 49, 56, 186, 161, 2, 181, 82, 114, 5, 103, 167, 15,
			213, 246, 183, 25, 175, 115, 235, 21, 103, 173, 111, 111, 117, 3, 114, 117, 241, 203, 205, 148, 114, 161, 39, 210, 124,
			29, 86, 51, 154, 213, 34, 132, 76, 100, 186, 151, 31, 132, 15}
	)
	type fields struct {
		privateKey      []byte
		publicKey       []byte
		fullAddress     []byte
		publicKeyString string
		encodedAddress  string
	}
	type args struct {
		payload        []byte
		seed           string
		optionalParams []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "Sign:success",
			args: args{
				seed:    seed,
				payload: payload,
			},
			want: signature,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &ZbcAccountType{
				privateKey:      tt.fields.privateKey,
				publicKey:       tt.fields.publicKey,
				fullAddress:     tt.fields.fullAddress,
				publicKeyString: tt.fields.publicKeyString,
				encodedAddress:  tt.fields.encodedAddress,
			}
			got, err := acc.Sign(tt.args.payload, tt.args.seed, tt.args.optionalParams...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Sign() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestZbcAccountType_VerifySignature(t *testing.T) {
	var (
		accountAddress = []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255,
			81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169}
		payload   = []byte{1, 2, 3}
		signature = []byte{8, 55, 103, 141, 152, 224, 207, 186, 41, 134, 223, 127, 182, 49, 56, 186, 161, 2, 181, 82, 114, 5, 103, 167, 15,
			213, 246, 183, 25, 175, 115, 235, 21, 103, 173, 111, 111, 117, 3, 114, 117, 241, 203, 205, 148, 114, 161, 39, 210, 124,
			29, 86, 51, 154, 213, 34, 132, 76, 100, 186, 151, 31, 132, 15}
	)
	type fields struct {
		privateKey      []byte
		publicKey       []byte
		fullAddress     []byte
		publicKeyString string
		encodedAddress  string
	}
	type args struct {
		payload        []byte
		signature      []byte
		accountAddress []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "VerifySignature:success",
			args: args{
				signature:      signature,
				payload:        payload,
				accountAddress: accountAddress,
			},
		},
		{
			name: "VerifySignature:fail",
			args: args{
				signature:      append(signature, []byte{1, 2, 3}...),
				payload:        payload,
				accountAddress: accountAddress,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &ZbcAccountType{
				privateKey:      tt.fields.privateKey,
				publicKey:       tt.fields.publicKey,
				fullAddress:     tt.fields.fullAddress,
				publicKeyString: tt.fields.publicKeyString,
				encodedAddress:  tt.fields.encodedAddress,
			}
			if err := acc.VerifySignature(tt.args.payload, tt.args.signature, tt.args.accountAddress); (err != nil) != tt.wantErr {
				t.Errorf("VerifySignature() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
