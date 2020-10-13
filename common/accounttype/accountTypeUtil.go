package accounttype

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/zoobc/lib/address"
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
	accTypeIntBytes := bufferBytes.Next(int(constant.AccountAddressTypeLength))
	if len(accTypeIntBytes) < int(constant.AccountAddressTypeLength) {
		return nil, errors.New("InvalidAccountFormat")
	}
	accTypeInt := int32(binary.LittleEndian.Uint32(accTypeIntBytes))
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
	accPubKeyLength := int(acc.GetAccountPublicKeyLength())
	accPubKey = bufferBytes.Next(accPubKeyLength)
	if len(accPubKey) < accPubKeyLength {
		return nil, errors.New("EndOfBufferReached")
	}
	acc.SetAccountPublicKey(accPubKey)
	return acc, nil
}

// ParseEncodedAccountToAccountAddress parse an encoded account type into a full account address ([]byte)
// Note: we must know the account type first to do it
func ParseEncodedAccountToAccountAddress(accTypeInt int32, encodedAccountAddress string) ([]byte, error) {
	var (
		accPubKey []byte
		err       error
	)
	switch accTypeInt {
	case int32(model.AccountType_ZbcAccountType):
		accPubKey = make([]byte, 32)
		err = address.DecodeZbcID(encodedAccountAddress, accPubKey)
		if err != nil {
			return nil, err
		}
	case int32(model.AccountType_BTCAccountType):
		// TODO: not implemented yet!
		return nil, errors.New("Parsing encoded BTC accounts is not implemented yet")
	}
	accType, err := NewAccountType(int32(model.AccountType_ZbcAccountType), accPubKey)
	if err != nil {
		return nil, err
	}
	return accType.GetAccountAddress()
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

// ParseEncodedAccountToAccountAddressHex parse an encoded account type into a full account address (hex encoded)
// Note: we must know the account type first to do it
func ParseEncodedAccountToAccountAddressHex(accTypeInt int32, encodedAccountAddress string) (string, error) {
	accountAddress, err := ParseEncodedAccountToAccountAddress(accTypeInt, encodedAccountAddress)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(accountAddress), nil
}

// IsZbcAccount validates whether an account type is a default account (ZBC)
func IsZbcAccount(at AccountType) bool {
	_, ok := at.(*ZbcAccountType)
	return ok
}
