package crypto

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/constant"
)

var (
	ed25519MockAddress   = "ZBC_AQTHDOKQ_2USUORFR_WB7PCOQD_ECAQDHCB_Y5X7DAVQ_OQ7SH2HR_2KWPQ7UC"
	ed25519MockSeed      = "compile fernlike laptop scouring bobsled tremble probably immunity babble elsewhere throwing thrill"
	ed25519MockPublicKey = []byte{4, 38, 113, 185, 80, 213, 37, 71, 68, 177, 176, 126, 241, 58, 3, 32, 129, 1, 156, 65, 199, 111,
		241, 130, 176, 116, 63, 35, 232, 241, 210, 172}
)

func TestNewEd25519Signature(t *testing.T) {
	tests := []struct {
		name string
		want *Ed25519Signature
	}{
		{
			name: "wantSuccess",
			want: &Ed25519Signature{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewEd25519Signature(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewEd25519Signature() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEd25519Signature_GetPublicKeyFromSeed(t *testing.T) {
	type args struct {
		seed string
	}
	tests := []struct {
		name string
		es   *Ed25519Signature
		args args
		want []byte
	}{
		{
			name: "wantSuccess",
			es:   &Ed25519Signature{},
			args: args{
				seed: ed25519MockSeed,
			},
			want: ed25519MockPublicKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			es := &Ed25519Signature{}
			if got := es.GetPublicKeyFromSeed(tt.args.seed); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ed25519Signature.GetPublicKeyFromSeed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEd25519Signature_GetAddressFromSeed(t *testing.T) {
	type args struct {
		seed string
	}
	tests := []struct {
		name string
		es   *Ed25519Signature
		args args
		want string
	}{
		{
			name: "wantSuccess",
			es:   &Ed25519Signature{},
			args: args{
				seed: ed25519MockSeed,
			},
			want: ed25519MockAddress,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			es := &Ed25519Signature{}
			if got := es.GetAddressFromSeed(constant.PrefixZoobcDefaultAccount, tt.args.seed); got != tt.want {
				t.Errorf("Ed25519Signature.GetAddressFromSeed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEd25519Signature_GetAddressFromPublicKey(t *testing.T) {
	type args struct {
		publicKey []byte
	}
	tests := []struct {
		name    string
		es      *Ed25519Signature
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "wantSuccess",
			args: args{
				publicKey: []byte{4, 38, 103, 73, 250, 169, 63, 155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242,
					84, 174, 239, 46, 190, 78, 68, 90, 83, 142, 11},
			},
			want:    "ZBC_AQTGOSP2_VE7ZW2QV_BFGE3CID_PAKUKWTW_6JKK53ZO_XZHEIWST_RYFXR672",
			wantErr: false,
		},
		{
			name: "want:fail-{public key length < 32}",
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
			es := &Ed25519Signature{}
			got, err := es.GetAddressFromPublicKey(constant.PrefixZoobcDefaultAccount, tt.args.publicKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ed25519Signature.GetAddressFromPublicKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Ed25519Signature.GetAddressFromPublicKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEd25519Signature_GetPublicKeyFromAddress(t *testing.T) {
	type args struct {
		address string
	}
	tests := []struct {
		name    string
		es      *Ed25519Signature
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "wantSuccess",
			args: args{
				address: "ZBC_AQTGOSP2_VE7ZW2QV_BFGE3CID_PAKUKWTW_6JKK53ZO_XZHEIWST_RYFXR672",
			},
			want: []byte{4, 38, 103, 73, 250, 169, 63, 155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242,
				84, 174, 239, 46, 190, 78, 68, 90, 83, 142, 11},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			es := &Ed25519Signature{}
			got, err := es.GetPublicKeyFromEncodedAddress(tt.args.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ed25519Signature.GetPublicKeyFromEncodedAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ed25519Signature.GetPublicKeyFromEncodedAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEd25519Signature_GetPublicKeyFromPrivateKey(t *testing.T) {
	var mockPrivateKey = []byte{188, 149, 6, 52, 103, 250, 141, 133, 84, 93, 225, 77, 118, 252, 111, 71, 115, 7, 109, 188, 229, 212,
		31, 2, 189, 96, 77, 11, 89, 44, 125, 12, 183, 227, 106, 207, 27, 80, 168, 160, 252, 110, 172, 177, 36, 171,
		249, 50, 21, 23, 18, 168, 96, 125, 32, 72, 124, 33, 96, 51, 231, 163, 156, 224}
	type args struct {
		privateKey []byte
	}
	tests := []struct {
		name    string
		e       *Ed25519Signature
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "wantSuccess",
			args: args{
				privateKey: mockPrivateKey,
			},
			want:    mockPrivateKey[32:],
			wantErr: false,
		},
		{
			name: "wantFailed",
			args: args{
				privateKey: []byte{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Ed25519Signature{}
			got, err := e.GetPublicKeyFromPrivateKey(tt.args.privateKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ed25519Signature.GetPublicKeyFromPrivateKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ed25519Signature.GetPublicKeyFromPrivateKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEd25519Signature_GetPrivateKeyFromSeed(t *testing.T) {
	type args struct {
		seed string
	}
	tests := []struct {
		name string
		e    *Ed25519Signature
		args args
		want []byte
	}{
		{
			name: "wantSuccess",
			args: args{
				seed: ed25519MockSeed,
			},
			want: []byte{64, 146, 182, 11, 22, 166, 212, 34, 188, 208, 94, 27, 124, 93, 182, 102, 136, 68, 246,
				28, 225, 249, 230, 196, 189, 79, 108, 178, 194, 39, 16, 203, 4, 38, 113, 185, 80, 213, 37, 71,
				68, 177, 176, 126, 241, 58, 3, 32, 129, 1, 156, 65, 199, 111, 241, 130, 176, 116, 63, 35, 232,
				241, 210, 172,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Ed25519Signature{}
			if got := e.GetPrivateKeyFromSeed(tt.args.seed); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ed25519Signature.GetPrivateKeyFromSeed() = %v\nwant %v", got, tt.want)
			}
		})
	}
}

func TestEd25519Signature_GetPrivateKeyFromSeedUseSlip10(t *testing.T) {
	type args struct {
		seed string
	}
	tests := []struct {
		name    string
		e       *Ed25519Signature
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "wantSuccess",
			args: args{
				seed: ed25519MockSeed,
			},
			want: []byte{28, 54, 122, 202, 22, 213, 226, 171, 212, 4, 201, 23, 18, 83, 234, 116,
				168, 202, 38, 81, 62, 84, 121, 59, 175, 165, 81, 161, 131, 173, 227, 20},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Ed25519Signature{}
			got, err := e.GetPrivateKeyFromSeedUseSlip10(tt.args.seed)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ed25519Signature.GetPrivateKeyFromSeedUseSlip10() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ed25519Signature.GetPrivateKeyFromSeedUseSlip10() = %v, want %v", got, tt.want)
			}
		})
	}
}
