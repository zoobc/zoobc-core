package accounttype

import (
	"bytes"
	"encoding/hex"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
	"golang.org/x/crypto/sha3"
)

func TestEstoniaEidAccountType_SetAccountPublicKey(t *testing.T) {
	type fields struct {
		publicKey   []byte
		fullAddress []byte
	}
	type args struct {
		accountPublicKey []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "success",
			args: args{
				accountPublicKey: []byte{1, 2, 3},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &EstoniaEidAccountType{
				publicKey:   tt.fields.publicKey,
				fullAddress: tt.fields.fullAddress,
			}
			acc.SetAccountPublicKey(tt.args.accountPublicKey)
			if !bytes.Equal(acc.GetAccountPublicKey(), tt.args.accountPublicKey) {
				t.Error("SetAccountPublicKey() is incorrect")
			}
		})
	}
}

func TestEstoniaEidAccountType_GetAccountAddress(t *testing.T) {
	type fields struct {
		publicKey   []byte
		fullAddress []byte
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name:    "fail",
			want:    nil,
			wantErr: true,
		},
		{
			name: "success",
			fields: fields{
				fullAddress: []byte{1, 2, 3},
			},
			want:    []byte{1, 2, 3},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &EstoniaEidAccountType{
				publicKey:   tt.fields.publicKey,
				fullAddress: tt.fields.fullAddress,
			}
			got, err := acc.GetAccountAddress()
			if (err != nil) != tt.wantErr {
				t.Errorf("EstoniaEidAccountType.GetAccountAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EstoniaEidAccountType.GetAccountAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEstoniaEidAccountType_GetTypeInt(t *testing.T) {
	type fields struct {
		publicKey   []byte
		fullAddress []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   int32
	}{
		{
			name: "success",
			want: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &EstoniaEidAccountType{
				publicKey:   tt.fields.publicKey,
				fullAddress: tt.fields.fullAddress,
			}
			if got := acc.GetTypeInt(); got != tt.want {
				t.Errorf("EstoniaEidAccountType.GetTypeInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEstoniaEidAccountType_GetAccountPublicKey(t *testing.T) {
	type fields struct {
		publicKey   []byte
		fullAddress []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		{
			name: "success",
			fields: fields{
				publicKey: []byte{1, 2, 3},
			},
			want: []byte{1, 2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &EstoniaEidAccountType{
				publicKey:   tt.fields.publicKey,
				fullAddress: tt.fields.fullAddress,
			}
			if got := acc.GetAccountPublicKey(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EstoniaEidAccountType.GetAccountPublicKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEstoniaEidAccountType_GetAccountPrefix(t *testing.T) {
	type fields struct {
		publicKey   []byte
		fullAddress []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "success",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &EstoniaEidAccountType{
				publicKey:   tt.fields.publicKey,
				fullAddress: tt.fields.fullAddress,
			}
			if got := acc.GetAccountPrefix(); got != tt.want {
				t.Errorf("EstoniaEidAccountType.GetAccountPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEstoniaEidAccountType_GetName(t *testing.T) {
	type fields struct {
		publicKey   []byte
		fullAddress []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "success",
			want: "EstoniaEid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &EstoniaEidAccountType{
				publicKey:   tt.fields.publicKey,
				fullAddress: tt.fields.fullAddress,
			}
			if got := acc.GetName(); got != tt.want {
				t.Errorf("EstoniaEidAccountType.GetName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEstoniaEidAccountType_GetAccountPublicKeyLength(t *testing.T) {
	type fields struct {
		publicKey   []byte
		fullAddress []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   uint32
	}{
		{
			name: "success",
			want: 97,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &EstoniaEidAccountType{
				publicKey:   tt.fields.publicKey,
				fullAddress: tt.fields.fullAddress,
			}
			if got := acc.GetAccountPublicKeyLength(); got != tt.want {
				t.Errorf("EstoniaEidAccountType.GetAccountPublicKeyLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEstoniaEidAccountType_IsEqual(t *testing.T) {
	type fields struct {
		publicKey   []byte
		fullAddress []byte
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
			name: "success: public key different",
			fields: fields{
				publicKey: []byte{1, 2, 3},
			},
			args: args{
				acc2: &EstoniaEidAccountType{
					publicKey: []byte{1, 2, 4},
				},
			},
			want: false,
		},
		{
			name: "success: public key equal",
			fields: fields{
				publicKey: []byte{1, 2, 3},
			},
			args: args{
				acc2: &EstoniaEidAccountType{
					publicKey: []byte{1, 2, 3},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &EstoniaEidAccountType{
				publicKey:   tt.fields.publicKey,
				fullAddress: tt.fields.fullAddress,
			}
			if got := acc.IsEqual(tt.args.acc2); got != tt.want {
				t.Errorf("EstoniaEidAccountType.IsEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEstoniaEidAccountType_GetSignatureType(t *testing.T) {
	type fields struct {
		publicKey   []byte
		fullAddress []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   model.SignatureType
	}{
		{
			name: "success",
			want: model.SignatureType_EstoniaEidSignature,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &EstoniaEidAccountType{
				publicKey:   tt.fields.publicKey,
				fullAddress: tt.fields.fullAddress,
			}
			if got := acc.GetSignatureType(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EstoniaEidAccountType.GetSignatureType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEstoniaEidAccountType_GetSignatureLength(t *testing.T) {
	type fields struct {
		publicKey   []byte
		fullAddress []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   uint32
	}{
		{
			name: "success",
			want: uint32(96),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &EstoniaEidAccountType{
				publicKey:   tt.fields.publicKey,
				fullAddress: tt.fields.fullAddress,
			}
			if got := acc.GetSignatureLength(); got != tt.want {
				t.Errorf("EstoniaEidAccountType.GetSignatureLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEstoniaEidAccountType_GetEncodedAddress(t *testing.T) {
	type fields struct {
		publicKey   []byte
		fullAddress []byte
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "fail: no public key",
			fields: fields{
				publicKey: nil,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "success",
			fields: fields{
				publicKey: []byte{1, 2, 3, 4},
			},
			want:    "01020304",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &EstoniaEidAccountType{
				publicKey:   tt.fields.publicKey,
				fullAddress: tt.fields.fullAddress,
			}
			got, err := acc.GetEncodedAddress()
			if (err != nil) != tt.wantErr {
				t.Errorf("EstoniaEidAccountType.GetEncodedAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("EstoniaEidAccountType.GetEncodedAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEstoniaEidAccountType_DecodePublicKeyFromAddress(t *testing.T) {
	type fields struct {
		publicKey   []byte
		fullAddress []byte
	}
	type args struct {
		address string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				address: "010203",
			},
			want:    []byte{1, 2, 3},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &EstoniaEidAccountType{
				publicKey:   tt.fields.publicKey,
				fullAddress: tt.fields.fullAddress,
			}
			got, err := acc.DecodePublicKeyFromAddress(tt.args.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("EstoniaEidAccountType.DecodePublicKeyFromAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EstoniaEidAccountType.DecodePublicKeyFromAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEstoniaEidAccountType_GenerateAccountFromSeed(t *testing.T) {
	type fields struct {
		publicKey   []byte
		fullAddress []byte
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
	}{
		{
			name:    "fail: no implementation",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &EstoniaEidAccountType{
				publicKey:   tt.fields.publicKey,
				fullAddress: tt.fields.fullAddress,
			}
			if err := acc.GenerateAccountFromSeed(tt.args.seed, tt.args.optionalParams...); (err != nil) != tt.wantErr {
				t.Errorf("EstoniaEidAccountType.GenerateAccountFromSeed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEstoniaEidAccountType_GetAccountPublicKeyString(t *testing.T) {
	type fields struct {
		publicKey   []byte
		fullAddress []byte
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				publicKey: []byte{1, 2, 3, 4},
			},
			want:    "01020304",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &EstoniaEidAccountType{
				publicKey:   tt.fields.publicKey,
				fullAddress: tt.fields.fullAddress,
			}
			got, err := acc.GetAccountPublicKeyString()
			if (err != nil) != tt.wantErr {
				t.Errorf("EstoniaEidAccountType.GetAccountPublicKeyString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("EstoniaEidAccountType.GetAccountPublicKeyString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEstoniaEidAccountType_GetAccountPrivateKey(t *testing.T) {
	type fields struct {
		publicKey   []byte
		fullAddress []byte
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name:    "success",
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &EstoniaEidAccountType{
				publicKey:   tt.fields.publicKey,
				fullAddress: tt.fields.fullAddress,
			}
			got, err := acc.GetAccountPrivateKey()
			if (err != nil) != tt.wantErr {
				t.Errorf("EstoniaEidAccountType.GetAccountPrivateKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EstoniaEidAccountType.GetAccountPrivateKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEstoniaEidAccountType_Sign(t *testing.T) {
	type fields struct {
		publicKey   []byte
		fullAddress []byte
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
			name:    "success",
			want:    []byte{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &EstoniaEidAccountType{
				publicKey:   tt.fields.publicKey,
				fullAddress: tt.fields.fullAddress,
			}
			got, err := acc.Sign(tt.args.payload, tt.args.seed, tt.args.optionalParams...)
			if (err != nil) != tt.wantErr {
				t.Errorf("EstoniaEidAccountType.Sign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EstoniaEidAccountType.Sign() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEstoniaEidAccountType_VerifySignature(t *testing.T) {
	messageRaw := "testing verification signature"
	messsageByte := []byte(messageRaw)
	digest := sha3.New256()
	_, _ = digest.Write(messsageByte)
	messageHash := digest.Sum([]byte{})

	mockPublicKeyBytes, _ := hex.DecodeString(
		"04acf16bf960e6797993a7fd08ad4464fde0b7eefe543d119552c4d1e786dd851903afe925ac1414cefaac741c520" +
			"0fa92c5f37a30a87430fc59bb543ff768a3cbc934548774b5645b2c3209b2a928c1cb7b52c2bb973690dddf7c348585907b27",
	)
	mockSignature, _ := hex.DecodeString(
		"00B87DCEC8616E0BC01D84A903B77E4BD70F7812378DDD90EF2F7253B011E49E49E8F4544E5F470227FE406E26B4A104BED" +
			"E622BC94689381C07E651CA53C1EB55160B6DB55B4E075C1289BE791CB485751E135B010DC52FE146CAB4ED31FF3F",
	)
	type fields struct {
		publicKey   []byte
		fullAddress []byte
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
			name: "success",
			fields: fields{
				publicKey: mockPublicKeyBytes,
			},
			args: args{
				payload:   messageHash,
				signature: mockSignature,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &EstoniaEidAccountType{
				publicKey:   tt.fields.publicKey,
				fullAddress: tt.fields.fullAddress,
			}
			if err := acc.VerifySignature(tt.args.payload, tt.args.signature, tt.args.accountAddress); (err != nil) != tt.wantErr {
				t.Errorf("EstoniaEidAccountType.VerifySignature() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
