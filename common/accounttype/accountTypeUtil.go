package accounttype

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
)

// NewAccountType returns the appropriate AccountType object based on the account account type nul and account public key
func NewAccountType(accTypeInt int32, accPubKey []byte) (AccountType, error) {
	var (
		acc AccountType
	)
	switch accTypeInt {
	case int32(model.AccountType_ZbcAccountType):
		acc = &ZbcAccountType{}
	case int32(model.AccountType_BTCAccountType):
		acc = &BTCAccountType{}
	case int32(model.AccountType_EmptyAccountType):
		acc = &EmptyAccountType{}
	default:
		return nil, errors.New("InvalidAccountType")
	}
	acc.SetAccountPublicKey(accPubKey)
	return acc, nil
}

// NewAccountTypeFromAccount returns the appropriate AccountType object based on the account full address (account type + account public key)
func NewAccountTypeFromAccount(accountAddress []byte) (AccountType, error) {
	buff := bytes.NewBuffer(accountAddress)
	return ParseBytesToAccountType(buff)
}

// ParseBytesToAccountType parse an AccountAddress from a bytes.Buffer and returns the appropriate AccountType object
func ParseBytesToAccountType(bufferBytes *bytes.Buffer) (AccountType, error) {
	var (
		accPubKey []byte
		acc       AccountType
	)
	accTypeIntBytes := bufferBytes.Next(int(constant.AccountAddressType))
	if len(accTypeIntBytes) < int(constant.AccountAddressType) {
		return nil, errors.New("InvalidAccountFormat")
	}
	accTypeInt := binary.LittleEndian.Uint32(accTypeIntBytes)
	switch accTypeInt {
	case 0:
		acc = &ZbcAccountType{}
	case 1:
		acc = &BTCAccountType{}
	default:
		return nil, errors.New("InvalidAccountType")
	}
	accPubKeyLength := int(acc.GetAccountPublicKeyLength())
	accPubKey = bufferBytes.Next(accPubKeyLength)
	if len(accPubKey) < accPubKeyLength {
		return nil, errors.New("EndOfBufferReached")
	}
	acc.SetAccountPublicKey(accPubKey)
	return acc, nil
}

// GetAccountTypes returns all AccountType (useful for loops)
func GetAccountTypes() map[uint32]AccountType {
	var (
		zbcAccount   = &ZbcAccountType{}
		dummyAccount = &BTCAccountType{}
	)
	return map[uint32]AccountType{
		uint32(zbcAccount.GetTypeInt()):   zbcAccount,
		uint32(dummyAccount.GetTypeInt()): dummyAccount,
	}
}

// IsZbcAccount validates whether an account type is a default account (ZBC)
func IsZbcAccount(at AccountType) bool {
	_, ok := at.(*ZbcAccountType)
	return ok
}
