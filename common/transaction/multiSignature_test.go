package transaction

import (
	"errors"
	"math/rand"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/crypto"

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
					SignatureInfo:            nil,
				},
				NormalFee: nil,
			},
			want: constant.MultisigFieldLength + constant.MultiSigInfoMinSignature + constant.MultiSigInfoNonce +
				constant.MultiSigNumberOfAddress + constant.MultiSigUnsignedTxBytesLength + constant.MultisigFieldLength,
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
					SignatureInfo:            nil,
				},
				NormalFee: nil,
			},
			want: constant.MultisigFieldLength + constant.MultiSigInfoMinSignature + constant.MultiSigInfoNonce +
				constant.MultiSigNumberOfAddress + constant.MultiSigUnsignedTxBytesLength + constant.MultisigFieldLength +
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
					SignatureInfo: &model.SignatureInfo{
						TransactionHash: make([]byte, constant.MultiSigTransactionHash),
						Signatures: map[string][]byte{
							"A": make([]byte, 64),
						},
					},
				},
				NormalFee: nil,
			},
			want: constant.MultisigFieldLength + constant.MultiSigInfoMinSignature + constant.MultiSigInfoNonce +
				constant.MultiSigNumberOfAddress + constant.MultiSigAddressLength + uint32(len([]byte("A"))) +
				constant.MultiSigUnsignedTxBytesLength + constant.MultisigFieldLength + constant.MultiSigTransactionHash +
				constant.MultiSigNumberOfSignatures + constant.MultiSigAddressLength + uint32(len([]byte("A"))) +
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
					SignatureInfo: &model.SignatureInfo{
						TransactionHash: make([]byte, constant.MultiSigTransactionHash),
						Signatures: map[string][]byte{
							"A": make([]byte, 64),
						},
					},
				},
				NormalFee: nil,
			},
			want: constant.MultisigFieldLength + constant.MultiSigInfoMinSignature + constant.MultiSigInfoNonce +
				constant.MultiSigNumberOfAddress + constant.MultiSigAddressLength + uint32(len([]byte("A"))) +
				constant.MultisigFieldLength + constant.MultiSigTransactionHash + constant.MultiSigNumberOfSignatures +
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
		SignatureInfo: &model.SignatureInfo{
			TransactionHash: make([]byte, constant.MultiSigTransactionHash),
			Signatures: map[string][]byte{
				"A": make([]byte, 64),
			},
		},
	}
	mockMultipleSignatureBodyBytes = []byte{
		1, 0, 0, 0, 2, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 1, 0, 0, 0, 65, 1, 0, 0, 0, 66, 1, 0, 0, 0,
		67, 120, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 65, 64, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
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
			want: make([]byte, 12),
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

type (
	// MultiSignatureTransactionValidate mocks
	validateTransactionUtilParseFail struct {
		UtilInterface
	}
	validateTransactionUtilParseSuccess struct {
		UtilInterface
	}
	validateTypeSwitcherGetTxTypeFail struct {
		TypeSwitcher
	}
	validateTypeSwitcheGetTxTypeSuccessInnerValidateFail struct {
		TypeSwitcher
	}
	validateTypeSwitcheGetTxTypeSuccessInnerValidateSuccess struct {
		TypeSwitcher
	}
	validateMockSendmoneyValidateFail struct {
		TypeAction
	}
	validateMockSendmoneyValidateSuccess struct {
		TypeAction
	}
	validateSignatureValidateFail struct {
		crypto.Signature
	}
	validateSignatureValidateSuccess struct {
		crypto.Signature
	}
	// MultiSignatureTransactionValidate mocks
)

func (*validateSignatureValidateFail) VerifySignature(payload, signature []byte, accountAddress string) bool {
	return false
}

func (*validateSignatureValidateSuccess) VerifySignature(payload, signature []byte, accountAddress string) bool {
	return true
}

func (*validateMockSendmoneyValidateFail) Validate(bool) error {
	return errors.New("mockedError")
}

func (*validateMockSendmoneyValidateSuccess) Validate(bool) error {
	return nil
}

func (*validateTransactionUtilParseFail) ParseTransactionBytes([]byte, bool) (*model.Transaction, error) {
	return nil, errors.New("mockedError")
}
func (*validateTransactionUtilParseSuccess) ParseTransactionBytes([]byte, bool) (*model.Transaction, error) {
	return &model.Transaction{}, nil
}
func (*validateTypeSwitcherGetTxTypeFail) GetTransactionType(*model.Transaction) (TypeAction, error) {
	return nil, errors.New("mockedError")
}
func (*validateTypeSwitcheGetTxTypeSuccessInnerValidateSuccess) GetTransactionType(*model.Transaction) (TypeAction, error) {
	return &validateMockSendmoneyValidateSuccess{}, nil
}
func (*validateTypeSwitcheGetTxTypeSuccessInnerValidateFail) GetTransactionType(*model.Transaction) (TypeAction, error) {
	return &validateMockSendmoneyValidateFail{}, nil
}
func TestMultiSignatureTransaction_Validate(t *testing.T) {
	type fields struct {
		Body            *model.MultiSignatureTransactionBody
		NormalFee       fee.FeeModelInterface
		TransactionUtil UtilInterface
		TypeSwitcher    TypeActionSwitcher
		Signature       crypto.SignatureInterface
	}
	type args struct {
		dbTx bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "MultisignatureTransaction_Validate-Fail-NothingProvided",
			fields: fields{
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo:       nil,
					UnsignedTransactionBytes: nil,
					SignatureInfo:            nil,
				},
				NormalFee:       nil,
				TransactionUtil: nil,
				TypeSwitcher:    &TypeSwitcher{},
				Signature:       nil,
			},
			args: args{
				dbTx: false,
			},
			wantErr: true,
		},
		{
			name: "MultisignatureTransaction_Validate-Fail-MultisigInfoExist-Addresses<2",
			fields: fields{
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo: &model.MultiSignatureInfo{
						MinimumSignatures: 0,
						Nonce:             0,
						Addresses:         make([]string, 1),
						MultisigAddress:   "",
						BlockHeight:       0,
					},
					UnsignedTransactionBytes: nil,
					SignatureInfo:            nil,
				},
				NormalFee:       nil,
				TransactionUtil: nil,
				TypeSwitcher:    nil,
				Signature:       nil,
			},
			args: args{
				dbTx: false,
			},
			wantErr: true,
		},
		{
			name: "MultisignatureTransaction_Validate-Fail-MultisigInfoExist-MinimumSignature<1",
			fields: fields{
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo: &model.MultiSignatureInfo{
						MinimumSignatures: 0,
						Nonce:             0,
						Addresses:         make([]string, 2),
						MultisigAddress:   "",
						BlockHeight:       0,
					},
					UnsignedTransactionBytes: nil,
					SignatureInfo:            nil,
				},
				NormalFee:       nil,
				TransactionUtil: nil,
				TypeSwitcher:    nil,
				Signature:       nil,
			},
			args: args{
				dbTx: false,
			},
			wantErr: true,
		},
		{
			name: "MultisignatureTransaction_Validate-Success-MultisigInfoExist",
			fields: fields{
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo: &model.MultiSignatureInfo{
						MinimumSignatures: 1,
						Nonce:             0,
						Addresses:         make([]string, 2),
						MultisigAddress:   "",
						BlockHeight:       0,
					},
					UnsignedTransactionBytes: nil,
					SignatureInfo:            nil,
				},
				NormalFee:       nil,
				TransactionUtil: nil,
				TypeSwitcher:    nil,
				Signature:       nil,
			},
			args: args{
				dbTx: false,
			},
			wantErr: false,
		},
		{
			name: "MultisignatureTransaction_Validate-Success-TransactionBytesExist-FailParse",
			fields: fields{
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo:       nil,
					UnsignedTransactionBytes: make([]byte, 100),
					SignatureInfo:            nil,
				},
				NormalFee:       nil,
				TransactionUtil: &validateTransactionUtilParseFail{},
				TypeSwitcher:    nil,
				Signature:       nil,
			},
			args: args{
				dbTx: false,
			},
			wantErr: true,
		},
		{
			name: "MultisignatureTransaction_Validate-Success-TransactionBytesExist-FailGetType",
			fields: fields{
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo:       nil,
					UnsignedTransactionBytes: make([]byte, 100),
					SignatureInfo:            nil,
				},
				NormalFee:       nil,
				TransactionUtil: &validateTransactionUtilParseSuccess{},
				TypeSwitcher:    &validateTypeSwitcherGetTxTypeFail{},
				Signature:       nil,
			},
			args: args{
				dbTx: false,
			},
			wantErr: true,
		},
		{
			name: "MultisignatureTransaction_Validate-Success-TransactionBytesExist-InnerValidateFail",
			fields: fields{
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo:       nil,
					UnsignedTransactionBytes: make([]byte, 100),
					SignatureInfo:            nil,
				},
				NormalFee:       nil,
				TransactionUtil: &validateTransactionUtilParseSuccess{},
				TypeSwitcher:    &validateTypeSwitcheGetTxTypeSuccessInnerValidateFail{},
				Signature:       nil,
			},
			args: args{
				dbTx: false,
			},
			wantErr: true,
		},
		{
			name: "MultisignatureTransaction_Validate-Success-TransactionBytesExist-InnerValidateSuccess",
			fields: fields{
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo:       nil,
					UnsignedTransactionBytes: make([]byte, 100),
					SignatureInfo:            nil,
				},
				NormalFee:       nil,
				TransactionUtil: &validateTransactionUtilParseSuccess{},
				TypeSwitcher:    &validateTypeSwitcheGetTxTypeSuccessInnerValidateSuccess{},
				Signature:       nil,
			},
			args: args{
				dbTx: false,
			},
			wantErr: false,
		},
		{
			name: "MultisignatureTransaction_Validate-Success-SignatureInfoExist-TransactionhHashNil",
			fields: fields{
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo:       nil,
					UnsignedTransactionBytes: nil,
					SignatureInfo: &model.SignatureInfo{
						TransactionHash: nil,
						Signatures:      nil,
					},
				},
				NormalFee:       nil,
				TransactionUtil: nil,
				TypeSwitcher:    nil,
				Signature:       nil,
			},
			args: args{
				dbTx: false,
			},
			wantErr: true,
		},
		{
			name: "MultisignatureTransaction_Validate-Success-SignatureInfoExist-NoSignature",
			fields: fields{
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo:       nil,
					UnsignedTransactionBytes: nil,
					SignatureInfo: &model.SignatureInfo{
						TransactionHash: make([]byte, 32),
						Signatures:      make(map[string][]byte),
					},
				},
				NormalFee:       nil,
				TransactionUtil: nil,
				TypeSwitcher:    nil,
				Signature:       nil,
			},
			args: args{
				dbTx: false,
			},
			wantErr: true,
		},
		{
			name: "MultisignatureTransaction_Validate-Success-SignatureInfoExist-WrongSignature",
			fields: fields{
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo:       nil,
					UnsignedTransactionBytes: nil,
					SignatureInfo: &model.SignatureInfo{
						TransactionHash: make([]byte, 32),
						Signatures: map[string][]byte{
							"A": []byte{1, 2, 3},
						},
					},
				},
				NormalFee:       nil,
				TransactionUtil: nil,
				TypeSwitcher:    nil,
				Signature:       &validateSignatureValidateFail{},
			},
			args: args{
				dbTx: false,
			},
			wantErr: true,
		},
		{
			name: "MultisignatureTransaction_Validate-Success-SignatureInfoExist-NilSignature",
			fields: fields{
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo:       nil,
					UnsignedTransactionBytes: nil,
					SignatureInfo: &model.SignatureInfo{
						TransactionHash: make([]byte, 32),
						Signatures: map[string][]byte{
							"A": nil,
						},
					},
				},
				NormalFee:       nil,
				TransactionUtil: nil,
				TypeSwitcher:    nil,
				Signature:       &validateSignatureValidateSuccess{},
			},
			args: args{
				dbTx: false,
			},
			wantErr: true,
		},
		{
			name: "MultisignatureTransaction_Validate-Success-SignatureInfoExist-ValidSignature",
			fields: fields{
				Body: &model.MultiSignatureTransactionBody{
					MultiSignatureInfo:       nil,
					UnsignedTransactionBytes: nil,
					SignatureInfo: &model.SignatureInfo{
						TransactionHash: make([]byte, 32),
						Signatures: map[string][]byte{
							"A": []byte{1, 2, 3},
						},
					},
				},
				NormalFee:       nil,
				TransactionUtil: nil,
				TypeSwitcher:    nil,
				Signature:       &validateSignatureValidateSuccess{},
			},
			args: args{
				dbTx: false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &MultiSignatureTransaction{
				Body:            tt.fields.Body,
				NormalFee:       tt.fields.NormalFee,
				TransactionUtil: tt.fields.TransactionUtil,
				TypeSwitcher:    tt.fields.TypeSwitcher,
				Signature:       tt.fields.Signature,
			}
			if err := tx.Validate(tt.args.dbTx); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
