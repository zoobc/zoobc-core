package signaturetype

import (
	"reflect"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/zoobc/zoobc-core/common/model"
	"golang.org/x/crypto/sha3"
)

var (
	mockBitcoinSeed           = "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved"
	mockBitcoinPrivKey32Bytes = sha3.Sum256([]byte(mockBitcoinSeed))
	mockBitcoinPrivKey48Bytes = sha3.Sum384([]byte(mockBitcoinSeed))
	mockBitcoinPrivKey64Bytes = sha3.Sum512([]byte(mockBitcoinSeed))
	mockBitcoinPublicKetBytes = []byte{3, 82, 247, 192, 243, 36, 207, 71, 90, 3, 103, 220, 47, 115, 64, 15, 13,
		59, 186, 231, 45, 42, 149, 73, 12, 5, 166, 141, 205, 177, 156, 77, 122}
)

func TestNewBitcoinSignature(t *testing.T) {
	type args struct {
		netParams *chaincfg.Params
		curve     *btcec.KoblitzCurve
	}
	tests := []struct {
		name string
		args args
		want *BitcoinSignature
	}{
		{
			name: "wantSuccess",
			args: args{
				netParams: DefaultBitcoinNetworkParams(),
				curve:     DefaultBitcoinCurve(),
			},
			want: &BitcoinSignature{
				NetworkParams: DefaultBitcoinNetworkParams(),
				Curve:         DefaultBitcoinCurve(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBitcoinSignature(tt.args.netParams, tt.args.curve); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBitcoinSignature() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBitcoinSignature_GetPrivateKeyFromSeed(t *testing.T) {
	var (
		mockWantedPrivKey256Bits, _ = btcec.PrivKeyFromBytes(DefaultBitcoinCurve(), mockBitcoinPrivKey32Bytes[:])
		mockWantedPrivKey384Bits, _ = btcec.PrivKeyFromBytes(DefaultBitcoinCurve(), mockBitcoinPrivKey48Bytes[:])
		mockWantedPrivKey512Bits, _ = btcec.PrivKeyFromBytes(DefaultBitcoinCurve(), mockBitcoinPrivKey64Bytes[:])
	)
	type fields struct {
		NetworkParams *chaincfg.Params
		Curve         *btcec.KoblitzCurve
	}
	type args struct {
		seed          string
		privkeyLength model.PrivateKeyBytesLength
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *btcec.PrivateKey
		wantErr bool
	}{

		{
			name: "wantSuccess:PrivateKey256Bits",
			fields: fields{
				NetworkParams: DefaultBitcoinNetworkParams(),
				Curve:         DefaultBitcoinCurve(),
			},
			args: args{
				seed:          mockBitcoinSeed,
				privkeyLength: model.PrivateKeyBytesLength_PrivateKey256Bits,
			},
			want:    mockWantedPrivKey256Bits,
			wantErr: false,
		},
		{
			name: "wantSuccess:PrivateKey384Bits",
			fields: fields{
				NetworkParams: DefaultBitcoinNetworkParams(),
				Curve:         DefaultBitcoinCurve(),
			},
			args: args{
				seed:          mockBitcoinSeed,
				privkeyLength: model.PrivateKeyBytesLength_PrivateKey384Bits,
			},
			want:    mockWantedPrivKey384Bits,
			wantErr: false,
		},
		{
			name: "wantSuccess:PrivateKey512Bits",
			fields: fields{
				NetworkParams: DefaultBitcoinNetworkParams(),
				Curve:         DefaultBitcoinCurve(),
			},
			args: args{
				seed:          mockBitcoinSeed,
				privkeyLength: model.PrivateKeyBytesLength_PrivateKey512Bits,
			},
			want:    mockWantedPrivKey512Bits,
			wantErr: false,
		},
		{
			name:   "wantFail:InvalidPrivateKeyLength",
			fields: fields{},
			args: args{
				seed:          mockBitcoinSeed,
				privkeyLength: model.PrivateKeyBytesLength_PrivateKeyInvalid,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BitcoinSignature{
				NetworkParams: tt.fields.NetworkParams,
				Curve:         tt.fields.Curve,
			}
			got, err := b.GetPrivateKeyFromSeed(tt.args.seed, tt.args.privkeyLength)
			if (err != nil) != tt.wantErr {
				t.Errorf("BitcoinSignature.GetPrivateKeyFromSeed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BitcoinSignature.GetPrivateKeyFromSeed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBitcoinSignature_GetPublicKeyFromSeed(t *testing.T) {
	type fields struct {
		NetworkParams *chaincfg.Params
		Curve         *btcec.KoblitzCurve
	}
	type args struct {
		seed          string
		format        model.BitcoinPublicKeyFormat
		privkeyLength model.PrivateKeyBytesLength
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				NetworkParams: DefaultBitcoinNetworkParams(),
				Curve:         DefaultBitcoinCurve(),
			},
			args: args{
				seed:          mockBitcoinSeed,
				format:        model.BitcoinPublicKeyFormat_PublicKeyFormatCompressed,
				privkeyLength: model.PrivateKeyBytesLength_PrivateKey256Bits,
			},
			want:    mockBitcoinPublicKetBytes,
			wantErr: false,
		},
		{
			name: "wantFail:InvalidPrivateKey",
			fields: fields{
				NetworkParams: DefaultBitcoinNetworkParams(),
				Curve:         DefaultBitcoinCurve(),
			},
			args: args{
				seed:          mockBitcoinSeed,
				format:        model.BitcoinPublicKeyFormat_PublicKeyFormatCompressed,
				privkeyLength: model.PrivateKeyBytesLength_PrivateKeyInvalid,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:InvalidPublicKeyFormat",
			fields: fields{
				NetworkParams: DefaultBitcoinNetworkParams(),
				Curve:         DefaultBitcoinCurve(),
			},
			args: args{
				seed:          mockBitcoinSeed,
				format:        model.BitcoinPublicKeyFormat(-1),
				privkeyLength: model.PrivateKeyBytesLength_PrivateKey256Bits,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BitcoinSignature{
				NetworkParams: tt.fields.NetworkParams,
				Curve:         tt.fields.Curve,
			}
			got, err := b.GetPublicKeyFromSeed(tt.args.seed, tt.args.format, tt.args.privkeyLength)
			if (err != nil) != tt.wantErr {
				t.Errorf("BitcoinSignature.GetPublicKeyFromSeed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BitcoinSignature.GetPublicKeyFromSeed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBitcoinSignature_GetPublicKeyFromPrivateKey(t *testing.T) {
	var mockPrivateKey, mockPubKey = btcec.PrivKeyFromBytes(DefaultBitcoinCurve(), mockBitcoinPrivKey32Bytes[:])
	type fields struct {
		NetworkParams *chaincfg.Params
		Curve         *btcec.KoblitzCurve
	}
	type args struct {
		privateKey *btcec.PrivateKey
		format     model.BitcoinPublicKeyFormat
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:   "wantSuccess:CompressedFormat",
			fields: fields{},
			args: args{
				privateKey: mockPrivateKey,
				format:     model.BitcoinPublicKeyFormat_PublicKeyFormatCompressed,
			},
			want:    mockPubKey.SerializeCompressed(),
			wantErr: false,
		},
		{
			name:   "wantSuccess:UncompressedFormat",
			fields: fields{},
			args: args{
				privateKey: mockPrivateKey,
				format:     model.BitcoinPublicKeyFormat_PublicKeyFormatUncompressed,
			},
			want:    mockPubKey.SerializeUncompressed(),
			wantErr: false,
		},
		{
			name:   "wantFail:InvalidFormat",
			fields: fields{},
			args: args{
				privateKey: mockPrivateKey,
				format:     model.BitcoinPublicKeyFormat(-1),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BitcoinSignature{
				NetworkParams: tt.fields.NetworkParams,
				Curve:         tt.fields.Curve,
			}
			got, err := b.GetPublicKeyFromPrivateKey(tt.args.privateKey, tt.args.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("BitcoinSignature.GetPublicKeyFromPrivateKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BitcoinSignature.GetPublicKeyFromPrivateKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBitcoinSignature_GetPublicKeyString(t *testing.T) {
	type fields struct {
		NetworkParams *chaincfg.Params
		Curve         *btcec.KoblitzCurve
	}
	type args struct {
		publicKey []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				NetworkParams: DefaultBitcoinNetworkParams(),
				Curve:         DefaultBitcoinCurve(),
			},
			args: args{
				publicKey: mockBitcoinPublicKetBytes,
			},
			want:    "0352f7c0f324cf475a0367dc2f73400f0d3bbae72d2a95490c05a68dcdb19c4d7a",
			wantErr: false,
		},
		{
			name: "wantFailed",
			fields: fields{
				NetworkParams: DefaultBitcoinNetworkParams(),
				Curve:         DefaultBitcoinCurve(),
			},
			args: args{
				publicKey: []byte{2, 231, 191, 45, 151, 205, 3, 121, 159, 114, 31, 223, 160, 57},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BitcoinSignature{
				NetworkParams: tt.fields.NetworkParams,
				Curve:         tt.fields.Curve,
			}
			got, err := b.GetPublicKeyString(tt.args.publicKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("BitcoinSignature.GetPublicKeyString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BitcoinSignature.GetPublicKeyString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBitcoinSignature_GetAddressFromPublicKey(t *testing.T) {
	type fields struct {
		NetworkParams *chaincfg.Params
		Curve         *btcec.KoblitzCurve
	}
	type args struct {
		publicKey []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				NetworkParams: DefaultBitcoinNetworkParams(),
				Curve:         DefaultBitcoinCurve(),
			},
			args: args{
				publicKey: mockBitcoinPublicKetBytes,
			},
			want:    "12Ea6WAMZhFnfM5kjyfrfykqVWFcaWorQ8",
			wantErr: false,
		},
		{
			name: "wnatFailed",
			fields: fields{
				NetworkParams: DefaultBitcoinNetworkParams(),
				Curve:         DefaultBitcoinCurve(),
			},
			args: args{
				publicKey: []byte{},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BitcoinSignature{
				NetworkParams: tt.fields.NetworkParams,
				Curve:         tt.fields.Curve,
			}
			got, err := b.GetAddressFromPublicKey(tt.args.publicKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("BitcoinSignature.GetAddressFromPublicKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BitcoinSignature.GetAddressFromPublicKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBitcoinSignature_GetAddressBytes(t *testing.T) {
	type fields struct {
		NetworkParams *chaincfg.Params
		Curve         *btcec.KoblitzCurve
	}
	type args struct {
		encodedAddress string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				NetworkParams: DefaultBitcoinNetworkParams(),
				Curve:         DefaultBitcoinCurve(),
			},
			args: args{
				encodedAddress: "12Ea6WAMZhFnfM5kjyfrfykqVWFcaWorQ8",
			},
			want:    []byte{13, 137, 40, 212, 218, 119, 144, 80, 70, 113, 150, 129, 2, 84, 45, 144, 145, 17, 64, 134},
			wantErr: false,
		},
		{
			name: "wantFailed",
			fields: fields{
				NetworkParams: DefaultBitcoinNetworkParams(),
				Curve:         DefaultBitcoinCurve(),
			},
			args: args{
				encodedAddress: "",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BitcoinSignature{
				NetworkParams: tt.fields.NetworkParams,
				Curve:         tt.fields.Curve,
			}
			got, err := b.GetAddressBytes(tt.args.encodedAddress)
			if (err != nil) != tt.wantErr {
				t.Errorf("BitcoinSignature.GetAddressBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BitcoinSignature.GetAddressBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBitcoinSignature_GetSignatureFromBytes(t *testing.T) {
	var (
		mockPrivateKey, _ = btcec.PrivKeyFromBytes(DefaultBitcoinCurve(), mockBitcoinPrivKey32Bytes[:])
		mockSignature, _  = mockPrivateKey.Sign([]byte{12, 1, 2, 1})
	)
	type fields struct {
		NetworkParams *chaincfg.Params
		Curve         *btcec.KoblitzCurve
	}
	type args struct {
		signatureBytes []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *btcec.Signature
		wantErr bool
	}{
		{
			name: "wantSuccess",
			fields: fields{
				NetworkParams: DefaultBitcoinNetworkParams(),
				Curve:         DefaultBitcoinCurve(),
			},
			args: args{
				signatureBytes: mockSignature.Serialize(),
			},
			want:    mockSignature,
			wantErr: false,
		},
		{
			name: "wantFailed",
			fields: fields{
				NetworkParams: DefaultBitcoinNetworkParams(),
				Curve:         DefaultBitcoinCurve(),
			},
			args: args{
				signatureBytes: []byte{2, 1, 2},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BitcoinSignature{
				NetworkParams: tt.fields.NetworkParams,
				Curve:         tt.fields.Curve,
			}
			got, err := b.GetSignatureFromBytes(tt.args.signatureBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("BitcoinSignature.GetSignatureFromBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BitcoinSignature.GetSignatureFromBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
