// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
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
	case int32(model.AccountType_EstoniaEidAccountType):
		acc = &EstoniaEidAccountType{}
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
		zbcAccount                = &ZbcAccountType{}
		dummyAccount              = &BTCAccountType{}
		emptyAccount              = &EmptyAccountType{}
		multiSignatureAccountType = &MultiSignatureAccountType{}
	)
	return map[uint32]AccountTypeInterface{
		uint32(zbcAccount.GetTypeInt()):                zbcAccount,
		uint32(dummyAccount.GetTypeInt()):              dummyAccount,
		uint32(emptyAccount.GetTypeInt()):              dummyAccount,
		uint32(multiSignatureAccountType.GetTypeInt()): multiSignatureAccountType,
	}
}
