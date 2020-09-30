package accounttype

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/zoobc/zoobc-core/common/constant"
)

// GetAccountType returns the appropriate AccountType object based on the account type index
func NewAccountType(accTypeInt uint32, accPubKey []byte) AccountType {
	var (
		acc AccountType
	)
	switch accTypeInt {
	case 0:
		acc = &ZbcAccountType{}
	case 1:
		acc = &DummyAccountType{}
	default:
		return nil
	}
	acc.SetAccountPublicKey(accPubKey)
	return acc
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
		acc = &DummyAccountType{}
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
		dummyAccount = &DummyAccountType{}
	)
	return map[uint32]AccountType{
		zbcAccount.GetTypeInt():   zbcAccount,
		dummyAccount.GetTypeInt(): dummyAccount,
	}
}

// IsZbcAccount validates whether an account type is a default account (ZBC)
func IsZbcAccount(at AccountType) bool {
	_, ok := at.(*ZbcAccountType)
	return ok
}
