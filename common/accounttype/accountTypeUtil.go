package accounttype

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/zoobc/lib/address"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
)

// NewAccountType returns the appropriate AccountTypeInterface object based on the account account type nul and account public key
func NewAccountType(accTypeInt int32, accPubKey []byte) (AccountTypeInterface, error) {
	var (
		acc AccountTypeInterface
	)
	switch accTypeInt {
	case int32(model.AccountType_ZbcAccountType):
		acc = &ZbcAccountType{}
	case int32(model.AccountType_BTCAccountType):
		acc = &BTCAccountType{}
	case int32(model.AccountType_EmptyAccountType):
		acc = &EmptyAccountType{}
	case int32(model.AccountType_MultiSignatureAccountType):
		acc = &MultiSignatureAccountType{}
	default:
		return nil, errors.New("InvalidAccountType")
	}
	acc.SetAccountPublicKey(accPubKey)
	return acc, nil
}

// NewAccountTypeFromAccount returns the appropriate AccountTypeInterface object based on the account full address (account type + account public key)
func NewAccountTypeFromAccount(accountAddress []byte) (AccountTypeInterface, error) {
	buff := bytes.NewBuffer(accountAddress)
	return ParseBytesToAccountType(buff)
}

// ParseBytesToAccountType parse an AccountAddress from a bytes.Buffer and returns the appropriate AccountTypeInterface object
func ParseBytesToAccountType(buffer *bytes.Buffer) (AccountTypeInterface, error) {
	var (
		accPubKey []byte
		acc       AccountTypeInterface
	)
	accTypeIntBytes := buffer.Next(int(constant.AccountAddressTypeLength))
	if len(accTypeIntBytes) < int(constant.AccountAddressTypeLength) {
		return nil, errors.New("InvalidAccountFormat")
	}
	accTypeInt := int32(binary.LittleEndian.Uint32(accTypeIntBytes))
	acc, err := NewAccountType(accTypeInt, []byte{})
	if err != nil {
		return nil, err
	}
	accPubKeyLength := int(acc.GetAccountPublicKeyLength())
	accPubKey = buffer.Next(accPubKeyLength)
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
		accType   AccountTypeInterface
	)
	switch accTypeInt {
	case int32(model.AccountType_ZbcAccountType):
		accPubKey = make([]byte, 32)
		err = address.DecodeZbcID(encodedAccountAddress, accPubKey)
		if err != nil {
			return nil, err
		}
		accType, err = NewAccountType(int32(model.AccountType_ZbcAccountType), accPubKey)
		if err != nil {
			return nil, err
		}
	case int32(model.AccountType_BTCAccountType):
		// TODO: not implemented yet!
		return nil, errors.New("parsing encoded BTC accounts is not implemented yet")
	case int32(model.AccountType_MultiSignatureAccountType):
		accPubKey = make([]byte, 32)
		err = address.DecodeZbcID(encodedAccountAddress, accPubKey)
		if err != nil {
			return nil, err
		}
		accType, err = NewAccountType(int32(model.AccountType_MultiSignatureAccountType), accPubKey)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("InvalidAccountType")
	}
	return accType.GetAccountAddress()
}

// GetAccountTypes returns all AccountTypeInterface (useful for loops)
func GetAccountTypes() map[uint32]AccountTypeInterface {
	var (
		zbcAccount            = &ZbcAccountType{}
		dummyAccount          = &BTCAccountType{}
		emptyAccount          = &EmptyAccountType{}
		multiSignatureAccount = &MultiSignatureAccountType{}
	)
	return map[uint32]AccountTypeInterface{
		uint32(zbcAccount.GetTypeInt()):            zbcAccount,
		uint32(dummyAccount.GetTypeInt()):          dummyAccount,
		uint32(emptyAccount.GetTypeInt()):          dummyAccount,
		uint32(multiSignatureAccount.GetTypeInt()): multiSignatureAccount,
	}
}
