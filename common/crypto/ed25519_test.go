package crypto

import (
	"reflect"
	"testing"
)

var (
	ed25519MockAddress   = "BCZxuVDVJUdEsbB-8ToDIIEBnEHHb_GCsHQ_I-jx0qw7"
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
			if got := es.GetAddressFromSeed(tt.args.seed); got != tt.want {
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
			want:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
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
			got, err := es.GetAddressFromPublicKey(tt.args.publicKey)
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
				address: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			},
			want: []byte{4, 38, 103, 73, 250, 169, 63, 155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242,
				84, 174, 239, 46, 190, 78, 68, 90, 83, 142, 11},
			wantErr: false,
		},
		{
			name: "wantFail-{decode error, wrong address format/length}",
			args: args{
				address: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgt",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wantFail:fail-{checksum error, wrong address format}",
			args: args{
				address: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtM",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			es := &Ed25519Signature{}
			got, err := es.GetPublicKeyFromAddress(tt.args.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ed25519Signature.GetPublicKeyFromAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ed25519Signature.GetPublicKeyFromAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
