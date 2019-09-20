package crypto

import (
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/constant"
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
		payload       []byte
		signatureType uint32
		seed          string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "Sign:valid",
			args: args{
				payload:       []byte{12, 43, 65, 65, 12, 123, 43, 12, 1, 24, 5, 5, 12, 54},
				signatureType: constant.SignatureTypeDefault,
				seed:          "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved",
			},
			want: []byte{0, 0, 0, 0, 42, 62, 47, 200, 180, 101, 85, 204, 179, 147, 143, 68, 30, 111, 6, 94, 81, 248, 219, 43, 90, 6, 167,
				45, 132, 96, 130, 0, 153, 244, 159, 137, 159, 113, 78, 9, 164, 154, 213, 255, 17, 206, 153, 156, 176, 206, 33,
				103, 72, 182, 228, 148, 234, 15, 176, 243, 50, 221, 106, 152, 53, 54, 173, 15},
		},
		{
			name: "Sign:valid-{default type}",
			args: args{
				payload:       []byte{12, 43, 65, 65, 12, 123, 43, 12, 1, 24, 5, 5, 12, 54},
				signatureType: 1011,
				seed:          "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved",
			},
			want: []byte{243, 3, 0, 0, 42, 62, 47, 200, 180, 101, 85, 204, 179, 147, 143, 68, 30, 111, 6, 94, 81, 248, 219, 43, 90, 6, 167,
				45, 132, 96, 130, 0, 153, 244, 159, 137, 159, 113, 78, 9, 164, 154, 213, 255, 17, 206, 153, 156, 176, 206, 33,
				103, 72, 182, 228, 148, 234, 15, 176, 243, 50, 221, 106, 152, 53, 54, 173, 15},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Signature{}
			got := s.Sign(tt.args.payload, tt.args.signatureType, tt.args.seed)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Signature.Sign() = %v, want %v", got, tt.want)
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
			want: []byte{0, 0, 0, 0, 42, 62, 47, 200, 180, 101, 85, 204, 179, 147, 143, 68, 30, 111, 6, 94, 81, 248, 219, 43, 90, 6, 167,
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
		accountAddress string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "VerifySignature:success",
			args: args{
				payload:        []byte{12, 43, 65, 65, 12, 123, 43, 12, 1, 24, 5, 5, 12, 54},
				accountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				signature: []byte{0, 0, 0, 0, 42, 62, 47, 200, 180, 101, 85, 204, 179, 147, 143, 68, 30, 111, 6, 94, 81, 248, 219, 43, 90, 6, 167,
					45, 132, 96, 130, 0, 153, 244, 159, 137, 159, 113, 78, 9, 164, 154, 213, 255, 17, 206, 153, 156, 176, 206, 33,
					103, 72, 182, 228, 148, 234, 15, 176, 243, 50, 221, 106, 152, 53, 54, 173, 15},
			},
			want: true,
		},
		{
			name: "VerifySignature:success-{default}",
			args: args{
				payload:        []byte{12, 43, 65, 65, 12, 123, 43, 12, 1, 24, 5, 5, 12, 54},
				accountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
				signature: []byte{255, 255, 0, 255, 42, 62, 47, 200, 180, 101, 85, 204, 179, 147, 143, 68, 30, 111, 6, 94, 81, 248, 219, 43, 90, 6, 167,
					45, 132, 96, 130, 0, 153, 244, 159, 137, 159, 113, 78, 9, 164, 154, 213, 255, 17, 206, 153, 156, 176, 206, 33,
					103, 72, 182, 228, 148, 234, 15, 176, 243, 50, 221, 106, 152, 53, 54, 173, 15},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Signature{}
			if got := s.VerifySignature(tt.args.payload, tt.args.signature, tt.args.accountAddress); got != tt.want {
				t.Errorf("Signature.VerifySignature() = %v, want %v", got, tt.want)
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
				signature: []byte{0, 0, 0, 0, 42, 62, 47, 200, 180, 101, 85, 204, 179, 147, 143, 68, 30, 111, 6, 94, 81, 248, 219, 43, 90, 6, 167,
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
