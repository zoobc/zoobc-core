package accounttype

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/zoobc/zoobc-core/common/model"
)

func TestGetAccountTypes(t *testing.T) {
	var (
		zbcAccount            = &ZbcAccountType{}
		dummyAccount          = &BTCAccountType{}
		emptyAccount          = &EmptyAccountType{}
		multiSignatureAccount = &MultiSignatureAccountType{}
	)
	tests := []struct {
		name string
		want map[uint32]AccountTypeInterface
	}{
		{
			name: "TestGetAccountTypes:success",
			want: map[uint32]AccountTypeInterface{
				uint32(zbcAccount.GetTypeInt()):            zbcAccount,
				uint32(dummyAccount.GetTypeInt()):          dummyAccount,
				uint32(emptyAccount.GetTypeInt()):          dummyAccount,
				uint32(multiSignatureAccount.GetTypeInt()): multiSignatureAccount,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAccountTypes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAccountTypes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewAccountType(t *testing.T) {
	var (
		zbcAccType = &ZbcAccountType{}
	)
	zbcAccType.SetAccountPublicKey([]byte{1, 2, 3})
	type args struct {
		accTypeInt int32
		accPubKey  []byte
	}
	tests := []struct {
		name    string
		args    args
		want    AccountTypeInterface
		wantErr bool
	}{
		{
			name: "TestNewAccountType:success",
			args: args{
				accPubKey:  []byte{1, 2, 3},
				accTypeInt: 0,
			},
			want: zbcAccType,
		},
		{
			name: "TestNewAccountType:fail-{invalidAccountType}",
			args: args{
				accPubKey:  []byte{},
				accTypeInt: 99,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAccountType(tt.args.accTypeInt, tt.args.accPubKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAccountType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAccountType() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewAccountTypeFromAccount(t *testing.T) {
	var (
		accountAddressPubKey = []byte{149, 1, 110, 5, 224, 150, 132, 85, 59, 205, 45, 168, 107, 143, 209, 215, 181, 221, 109, 23,
			39, 95, 248, 147, 114, 91, 115, 75, 51, 31, 148, 108}
		accountAddress = append([]byte{0, 0, 0, 0}, accountAddressPubKey...)
		accType        = &ZbcAccountType{}
	)
	accType.SetAccountPublicKey(accountAddressPubKey)
	type args struct {
		accountAddress []byte
	}
	tests := []struct {
		name    string
		args    args
		want    AccountTypeInterface
		wantErr bool
	}{
		{
			name: "TestNewAccountTypeFromAccount:success",
			args: args{
				accountAddress: accountAddress,
			},
			want: accType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAccountTypeFromAccount(tt.args.accountAddress)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAccountTypeFromAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAccountTypeFromAccount() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseBytesToAccountType(t *testing.T) {
	var (
		accPubKey = []byte{149, 1, 110, 5, 224, 150, 132, 85, 59, 205, 45, 168, 107, 143, 209, 215, 181, 221, 109, 23, 39,
			95, 248, 147, 114, 91, 115, 75, 51, 31, 148, 108}
		buffer = []byte{0, 0, 0, 0, 149, 1, 110, 5, 224, 150, 132, 85, 59, 205, 45, 168, 107, 143, 209, 215, 181, 221, 109, 23, 39,
			95, 248, 147, 114, 91, 115, 75, 51, 31, 148, 108, 1, 2, 3, 4, 5, 6, 0, 0, 0, 0}
		bufferInvalidAccType = []byte{255, 255, 0, 0, 149, 1, 110, 5, 224, 150, 132, 85, 59, 205, 45, 168, 107, 143, 209, 215, 181, 221, 109, 23, 39,
			95, 248, 147, 114, 91, 115, 75, 51, 31, 148, 108, 1, 2, 3, 4, 5, 6, 0, 0, 0, 0}
		bufferInvalidPubKeyLength = []byte{255, 255, 0, 0, 149, 1, 110, 5, 224, 150, 132, 85, 59, 205, 45, 168, 107, 143, 209, 215, 181, 221,
			109, 23, 39,
			95, 248, 147, 114, 91}
		accTypeRes = &ZbcAccountType{}
	)
	accTypeRes.SetAccountPublicKey(accPubKey)
	type args struct {
		buffer *bytes.Buffer
	}
	tests := []struct {
		name    string
		args    args
		want    AccountTypeInterface
		wantErr bool
	}{
		{
			name: "TestParseBytesToAccountType:success",
			args: args{
				buffer: bytes.NewBuffer(buffer),
			},
			want: accTypeRes,
		},
		{
			name: "TestParseBytesToAccountType:fail-{InvalidAccountType}",
			args: args{
				buffer: bytes.NewBuffer(bufferInvalidAccType),
			},
			wantErr: true,
		},
		{
			name: "TestParseBytesToAccountType:fail-{InvalidAccountPubKey}",
			args: args{
				buffer: bytes.NewBuffer(bufferInvalidPubKeyLength),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBytesToAccountType(tt.args.buffer)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBytesToAccountType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseBytesToAccountType() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseEncodedAccountToAccountAddress(t *testing.T) {
	var (
		encAddress1  = "ZBC_SUAW4BPA_S2CFKO6N_FWUGXD6R_262523IX_E5P7RE3S_LNZUWMY7_SRWCMI2J"
		fullAddress1 = []byte{0, 0, 0, 0, 149, 1, 110, 5, 224, 150, 132, 85, 59, 205, 45, 168, 107, 143, 209, 215, 181, 221, 109, 23, 39,
			95, 248, 147, 114, 91, 115, 75, 51, 31, 148, 108}
		musigAddress      = "ZMS_ZGL7TDDQ_HFE5OPPK_DBK3Y2GX_27OIFWUW_ZJ236YW3_JMUZQPWQ_I52QKANZ"
		musigAddressBytes = []byte{3, 0, 0, 0, 201, 151, 249, 140, 112, 57, 73, 215, 61, 234, 24, 85, 188, 104, 215, 215, 220, 130,
			218, 150, 202, 117, 191, 98, 219, 75, 41, 152, 62, 208, 71, 117}
	)
	type args struct {
		accTypeInt            int32
		encodedAccountAddress string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "TestParseEncodedAccountToAccountAddress:success",
			args: args{
				encodedAccountAddress: encAddress1,
				accTypeInt:            int32(model.AccountType_ZbcAccountType),
			},
			want: fullAddress1,
		},
		{
			name: "MusigAddress:success",
			args: args{
				encodedAccountAddress: musigAddress,
				accTypeInt:            int32(model.AccountType_MultiSignatureAccountType),
			},
			want: musigAddressBytes,
		},
		{
			name: "TestParseEncodedAccountToAccountAddress:fail-{BtcNotImplemented}",
			args: args{
				encodedAccountAddress: "12Ea6WAMZhFnfM5kjyfrfykqVWFcaWorQ8",
				accTypeInt:            int32(model.AccountType_BTCAccountType),
			},
			wantErr: true,
		},
		{
			name: "TestParseEncodedAccountToAccountAddress:fail-{InvalidAccountType}",
			args: args{
				encodedAccountAddress: "12Ea6WAMZhFnfM5kjyfrfykqVWFcaWorQ8",
				accTypeInt:            99,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseEncodedAccountToAccountAddress(tt.args.accTypeInt, tt.args.encodedAccountAddress)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEncodedAccountToAccountAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseEncodedAccountToAccountAddress() got = %v, want %v", got, tt.want)
			}
		})
	}
}
