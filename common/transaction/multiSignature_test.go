package transaction

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/constant"

	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
)

func TestMultiSignatureTransaction_GetSize(t *testing.T) {
	type fields struct {
		Body      *model.MultiSignatureTransactionBody
		NormalFee fee.FeeModelInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   uint32
	}{
		{
			name: "GetSize-Success-no_addresses-no_signatures-no_transactionBytes",
			fields: fields{
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo: &model.MultiSignatureInfo{
						MinimumSignatures: rand.Uint32(),
						Nonce:             rand.Int63(),
						Addresses:         nil,
					},
					UnsignedTransactionBytes: nil,
					Signatures:               nil,
				},
				NormalFee: nil,
			},
			want: constant.MultiSigInfoMinSignature + constant.MultiSigInfoNonce + constant.MultiSigNumberOfAddress +
				constant.MultiSigUnsignedTxBytesLength + constant.MultiSigNumberOfSignatures,
		},
		{
			name: "GetSize-Success-with_addresses-no_signatures-no_transactionBytes",
			fields: fields{
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo: &model.MultiSignatureInfo{
						MinimumSignatures: rand.Uint32(),
						Nonce:             rand.Int63(),
						Addresses: []string{
							"A",
						},
					},
					UnsignedTransactionBytes: nil,
					Signatures:               nil,
				},
				NormalFee: nil,
			},
			want: constant.MultiSigInfoMinSignature + constant.MultiSigInfoNonce + constant.MultiSigNumberOfAddress +
				constant.MultiSigUnsignedTxBytesLength + constant.MultiSigNumberOfSignatures +
				constant.MultiSigAddressLength + uint32(len([]byte("A"))),
		},
		{
			name: "GetSize-Success-with_addresses-with_signatures-no_transactionBytes",
			fields: fields{
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo: &model.MultiSignatureInfo{
						MinimumSignatures: rand.Uint32(),
						Nonce:             rand.Int63(),
						Addresses: []string{
							"A",
						},
					},
					UnsignedTransactionBytes: nil,
					Signatures: [][]byte{
						make([]byte, 64),
					},
				},
				NormalFee: nil,
			},
			want: constant.MultiSigInfoMinSignature + constant.MultiSigInfoNonce + constant.MultiSigNumberOfAddress +
				constant.MultiSigUnsignedTxBytesLength + constant.MultiSigNumberOfSignatures +
				constant.MultiSigAddressLength + uint32(len([]byte("A"))) +
				constant.MultiSigSignatureLength + 64,
		},
		{
			name: "GetSize-Success-with_addresses-with_signatures-with_transactionBytes",
			fields: fields{
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo: &model.MultiSignatureInfo{
						MinimumSignatures: rand.Uint32(),
						Nonce:             rand.Int63(),
						Addresses: []string{
							"A",
						},
					},
					UnsignedTransactionBytes: make([]byte, 120),
					Signatures: [][]byte{
						make([]byte, 64),
					},
				},
				NormalFee: nil,
			},
			want: constant.MultiSigInfoMinSignature + constant.MultiSigInfoNonce + constant.MultiSigNumberOfAddress +
				constant.MultiSigNumberOfSignatures +
				constant.MultiSigAddressLength + uint32(len([]byte("A"))) +
				constant.MultiSigSignatureLength + 64 +
				constant.MultiSigUnsignedTxBytesLength + 120,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &MultiSignatureTransaction{
				Body:      tt.fields.Body,
				NormalFee: tt.fields.NormalFee,
			}
			if got := tx.GetSize(); got != tt.want {
				t.Errorf("GetSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

var (
	// mock for GetBodyBytes & ParseBodyBytes
	mockMultipleSignatureBody = &model.MultiSignatureTransactionBody{
		MultiSignatureInfo: &model.MultiSignatureInfo{
			MinimumSignatures: 2,
			Nonce:             1,
			Addresses: []string{
				"A",
				"B",
				"C",
			},
		},
		UnsignedTransactionBytes: make([]byte, 120),
		Signatures: [][]byte{
			make([]byte, 64),
			make([]byte, 64),
			make([]byte, 64),
		},
	}
	mockMultipleSignatureBodyBytes = []byte{
		2, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 1, 0, 0, 0, 65, 1, 0, 0, 0, 66, 1, 0, 0, 0, 67,
		120, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0,
		64, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 64, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 64, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
	}
	// mock for GetBodyBytes & ParseBodyBytes
)

func TestMultiSignatureTransaction_GetBodyBytes(t *testing.T) {
	type fields struct {
		Body      *model.MultiSignatureTransactionBody
		NormalFee fee.FeeModelInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		{
			name: "GetBodyBytes-Success",
			fields: fields{
				Body:      nil,
				NormalFee: nil,
			},
			want: make([]byte, 24),
		},
		{
			name: "GetBodyBytes-Success-Complete",
			fields: fields{
				Body:      mockMultipleSignatureBody,
				NormalFee: nil,
			},
			want: mockMultipleSignatureBodyBytes,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &MultiSignatureTransaction{
				Body:      tt.fields.Body,
				NormalFee: tt.fields.NormalFee,
			}
			if got := tx.GetBodyBytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBodyBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultiSignatureTransaction_ParseBodyBytes(t *testing.T) {
	type fields struct {
		Body      *model.MultiSignatureTransactionBody
		NormalFee fee.FeeModelInterface
	}
	type args struct {
		txBodyBytes []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    model.TransactionBodyInterface
		wantErr bool
	}{
		{
			name: "ParseBodyBytes-success",
			fields: fields{
				Body:      nil,
				NormalFee: nil,
			},
			args: args{
				txBodyBytes: mockMultipleSignatureBodyBytes,
			},
			want:    mockMultipleSignatureBody,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &MultiSignatureTransaction{
				Body:      tt.fields.Body,
				NormalFee: tt.fields.NormalFee,
			}
			got, err := tx.ParseBodyBytes(tt.args.txBodyBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBodyBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseBodyBytes() got = %v, want %v", got, tt.want)
			}
		})
	}
}
