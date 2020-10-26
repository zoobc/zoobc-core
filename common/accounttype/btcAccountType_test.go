package accounttype

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"reflect"
	"testing"
)

func TestBTCAccountType_GenerateAccountFromSeed(t *testing.T) {
	var (
		seed = "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved"
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
		want    string
	}{
		{
			name: "GenerateAccountFromSeed:success-{defaultParams}",
			args: args{
				seed: seed,
			},
			want: "12Ea6WAMZhFnfM5kjyfrfykqVWFcaWorQ8",
		},
		{
			name: "GenerateAccountFromSeed:success-{withOptionalParams}",
			args: args{
				seed: seed,
				optionalParams: []interface{}{
					model.PrivateKeyBytesLength_PrivateKey256Bits,
					model.BitcoinPublicKeyFormat_PublicKeyFormatUncompressed,
				},
			},
			want: "1FjUuYPZHz3D9kvEj21uiwE3JwYNemQqcv",
		},
		{
			name: "GenerateAccountFromSeed:success-{invalidOptionalParams}",
			args: args{
				seed: seed,
				optionalParams: []interface{}{
					model.BitcoinPublicKeyFormat_PublicKeyFormatUncompressed,
					model.PrivateKeyBytesLength_PrivateKey256Bits,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &BTCAccountType{
				privateKey:      tt.fields.privateKey,
				publicKey:       tt.fields.publicKey,
				fullAddress:     tt.fields.fullAddress,
				publicKeyString: tt.fields.publicKeyString,
				encodedAddress:  tt.fields.encodedAddress,
			}
			if err := acc.GenerateAccountFromSeed(tt.args.seed, tt.args.optionalParams...); (err != nil) != tt.wantErr {
				t.Errorf("GenerateAccountFromSeed() error = %v, wantErr %v", err, tt.wantErr)
			} else if acc.encodedAddress != tt.want {
				t.Errorf("GenerateAccountFromSeed() error = %v, want %v", tt.want, acc.encodedAddress)
			}
		})
	}
}

func TestBTCAccountType_GetAccountAddress(t *testing.T) {
	var (
		pubKey = []byte{3, 82, 247, 192, 243, 36, 207, 71, 90, 3, 103, 220, 47, 115, 64, 15, 13, 59, 186, 231, 45, 42, 149, 73, 12, 5,
			166, 141, 205, 177, 156, 77, 122}
		fullAccountAddress = []byte{1, 0, 0, 0, 3, 82, 247, 192, 243, 36, 207, 71, 90, 3, 103, 220, 47, 115, 64, 15, 13, 59, 186, 231, 45,
			42, 149, 73, 12, 5, 166, 141, 205, 177, 156, 77, 122}
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
			name: "GetAccountAddress:success-{cached}",
			fields: fields{
				fullAddress: []byte{1, 2, 3},
			},
			want: []byte{1, 2, 3},
		},
		{
			name:    "GetAccountAddress:fail-{emptyPublicKey}",
			wantErr: true,
		},
		{
			name: "GetAccountAddress:success-{calculated}",
			fields: fields{
				publicKey: pubKey,
			},
			want: fullAccountAddress,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &BTCAccountType{
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

func TestBTCAccountType_GetAccountPrefix(t *testing.T) {
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
			want: "BTC",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &BTCAccountType{
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

func TestBTCAccountType_GetAccountPrivateKey(t *testing.T) {
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
			name:    "GetAccountPrivateKey:fail-{AccountNotGenerated}",
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
			acc := &BTCAccountType{
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

func TestBTCAccountType_GetAccountPublicKey(t *testing.T) {
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
			name: "GetAccountPublicKey",
			fields: fields{
				publicKey: []byte{1, 2, 3},
			},
			want: []byte{1, 2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &BTCAccountType{
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

func TestBTCAccountType_GetAccountPublicKeyLength(t *testing.T) {
	var (
		pubKey = []byte{3, 82, 247, 192, 243, 36, 207, 71, 90, 3, 103, 220, 47, 115, 64, 15, 13, 59, 186, 231, 45, 42, 149, 73, 12, 5,
			166, 141, 205, 177, 156, 77, 122}
	)
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
			name: "GetAccountPublicKeyLength:success-{calculated}",
			fields: fields{
				publicKey: pubKey,
			},
			want: uint32(len(pubKey)),
		},
		{
			name: "GetAccountPublicKeyLength:success-{default}",
			want: btcec.PubKeyBytesLenCompressed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &BTCAccountType{
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

func TestBTCAccountType_GetAccountPublicKeyString(t *testing.T) {
	var (
		pubKey = []byte{3, 82, 247, 192, 243, 36, 207, 71, 90, 3, 103, 220, 47, 115, 64, 15, 13, 59, 186, 231, 45, 42, 149, 73, 12, 5,
			166, 141, 205, 177, 156, 77, 122}
		pubKeyUncompressed = []byte{4, 23, 242, 149, 86, 46, 8, 97, 27, 163, 27, 125, 99, 98, 54, 154, 151, 74, 39, 53, 43, 252, 34, 129,
			236, 215, 63, 6, 199, 95, 12, 118, 142, 226, 203, 118, 61, 85, 225, 52, 154, 214, 231, 224, 25, 90, 120, 22, 132, 146, 6, 90,
			51, 118, 24, 70, 45, 33, 170, 152, 86, 60, 15, 193, 164}
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
				publicKeyString: "testStr",
			},
			want: "testStr",
		},
		{
			name:    "GetAccountPublicKeyString:fail-{EmptyAccountPublicKey}",
			wantErr: true,
		},
		{
			name: "GetAccountPublicKeyString:success-{calculated}",
			fields: fields{
				publicKey: pubKey,
			},
			want: "0352f7c0f324cf475a0367dc2f73400f0d3bbae72d2a95490c05a68dcdb19c4d7a",
		},
		{
			name: "GetAccountPublicKeyString:success-{calculated-pubKeyUncompressed}",
			fields: fields{
				publicKey: pubKeyUncompressed,
			},
			want: "0417f295562e08611ba31b7d6362369a974a27352bfc2281ecd73f06c75" +
				"f0c768ee2cb763d55e1349ad6e7e0195a78168492065a337618462d21aa98563c0fc1a4",
		},
		{
			name: "GetAccountPublicKeyString:fail-{invalidPublicKey}",
			fields: fields{
				publicKey: append([]byte{10, 2, 3, 4}, pubKey...),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &BTCAccountType{
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

func TestBTCAccountType_GetEncodedAddress(t *testing.T) {
	var (
		pubKey = []byte{3, 82, 247, 192, 243, 36, 207, 71, 90, 3, 103, 220, 47, 115, 64, 15, 13, 59, 186, 231, 45, 42, 149, 73, 12, 5,
			166, 141, 205, 177, 156, 77, 122}
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
			want: "12Ea6WAMZhFnfM5kjyfrfykqVWFcaWorQ8",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &BTCAccountType{
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

func TestBTCAccountType_GetName(t *testing.T) {
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
			name: "GetName",
			want: "BTCAccount",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &BTCAccountType{
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

func TestBTCAccountType_GetSignatureLength(t *testing.T) {
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
			want: constant.BTCECDSASignatureLength,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &BTCAccountType{
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

func TestBTCAccountType_GetSignatureType(t *testing.T) {
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
			want: model.SignatureType_BitcoinSignature,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &BTCAccountType{
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

func TestBTCAccountType_GetTypeInt(t *testing.T) {
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
			name: "GetTypeInt",
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &BTCAccountType{
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

func TestBTCAccountType_IsEqual(t *testing.T) {
	var (
		pubKey = []byte{3, 82, 247, 192, 243, 36, 207, 71, 90, 3, 103, 220, 47, 115, 64, 15, 13, 59, 186, 231, 45, 42, 149, 73, 12, 5,
			166, 141, 205, 177, 156, 77, 122}
		accType = &BTCAccountType{
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
			acc := &BTCAccountType{
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

func TestBTCAccountType_Sign(t *testing.T) {
	var (
		seed      = "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved"
		payload   = []byte{1, 2, 3}
		signature = []byte{33, 0, 3, 82, 247, 192, 243, 36, 207, 71, 90, 3, 103, 220, 47, 115, 64, 15, 13, 59, 186, 231, 45, 42, 149, 73,
			12, 5, 166, 141, 205, 177, 156, 77, 122, 48, 69, 2, 33, 0, 184, 15, 123, 55, 191, 208, 195, 227, 186, 140, 100, 21, 170, 80,
			65, 31, 156, 163, 120, 194, 142, 121, 103, 105, 146, 3, 242, 162, 86, 169, 141, 211, 2, 32, 16, 140, 220, 132, 123, 128, 89,
			152, 29, 29, 91, 239, 70, 201, 42, 173, 210, 30, 218, 205, 224, 93, 135, 46, 221, 182, 148, 15, 252, 76, 119, 145}
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
			acc := &BTCAccountType{
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

func TestBTCAccountType_VerifySignature(t *testing.T) {
	var (
		accountAddress = []byte{1, 0, 0, 0, 3, 82, 247, 192, 243, 36, 207, 71, 90, 3, 103, 220, 47, 115, 64, 15, 13, 59, 186, 231, 45, 42,
			149, 73, 12, 5, 166, 141, 205, 177, 156, 77, 122}
		payload   = []byte{1, 2, 3}
		signature = []byte{33, 0, 3, 82, 247, 192, 243, 36, 207, 71, 90, 3, 103, 220, 47, 115, 64, 15, 13, 59, 186, 231, 45, 42, 149, 73,
			12, 5, 166, 141, 205, 177, 156, 77, 122, 48, 69, 2, 33, 0, 184, 15, 123, 55, 191, 208, 195, 227, 186, 140, 100, 21, 170, 80,
			65, 31, 156, 163, 120, 194, 142, 121, 103, 105, 146, 3, 242, 162, 86, 169, 141, 211, 2, 32, 16, 140, 220, 132, 123, 128, 89,
			152, 29, 29, 91, 239, 70, 201, 42, 173, 210, 30, 218, 205, 224, 93, 135, 46, 221, 182, 148, 15, 252, 76, 119, 145}
	)
	type fields struct {
		privateKey      []byte
		publicKey       []byte
		fullAddress     []byte
		publicKeyString string
		encodedAddress  string
	}
	type args struct {
		payload            []byte
		signature          []byte
		fullAccountAddress []byte
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
				fullAccountAddress: accountAddress,
				payload:            payload,
				signature:          signature,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &BTCAccountType{
				privateKey:      tt.fields.privateKey,
				publicKey:       tt.fields.publicKey,
				fullAddress:     tt.fields.fullAddress,
				publicKeyString: tt.fields.publicKeyString,
				encodedAddress:  tt.fields.encodedAddress,
			}
			if err := acc.VerifySignature(tt.args.payload, tt.args.signature, tt.args.fullAccountAddress); (err != nil) != tt.wantErr {
				t.Errorf("VerifySignature() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
