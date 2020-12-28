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
package util

import (
	"bytes"
	"github.com/zoobc/zoobc-core/common/accounttype"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
)

// GetProofOfOwnershipSize returns size in bytes of a proof of ownership message
func GetProofOfOwnershipSize(accountAddressType accounttype.AccountTypeInterface, withSignature bool) uint32 {
	var (
		accountAddressSize = constant.AccountAddressTypeLength + accountAddressType.GetAccountPublicKeyLength()
	)
	message := accountAddressSize + constant.BlockHash + constant.Height
	if withSignature {
		return message + constant.NodeSignature
	}
	return message
}

// GetProofOfOwnershipBytes serialize ProofOfOwnership struct into bytes
func GetProofOfOwnershipBytes(poown *model.ProofOfOwnership) []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(poown.MessageBytes)
	buffer.Write(poown.Signature)
	return buffer.Bytes()
}

// ParseProofOfOwnershipBytes parse a byte array into a ProofOfOwnership struct (message + signature)
// poownBytes if true returns size of message + signature
func ParseProofOfOwnershipBytes(poownBytes []byte) (*model.ProofOfOwnership, error) {
	// copy poown bytes and parse first bytes as accountAddress to get the address size
	var tmpPoonBytes = make([]byte, len(poownBytes))
	copy(tmpPoonBytes, poownBytes)
	tmpBuffer := bytes.NewBuffer(tmpPoonBytes)
	accType, err := accounttype.ParseBytesToAccountType(tmpBuffer)
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(poownBytes)
	poownMessageBytes, err := ReadTransactionBytes(buffer, int(GetProofOfOwnershipSize(accType, false)))
	if err != nil {
		return nil, err
	}
	signature, err := ReadTransactionBytes(buffer, int(constant.NodeSignature))
	if err != nil {
		return nil, err
	}
	return &model.ProofOfOwnership{
		MessageBytes: poownMessageBytes,
		Signature:    signature,
	}, nil
}

// GetProofOfOwnershipMessageBytes serialize ProofOfOwnershipMessage struct into bytes
func GetProofOfOwnershipMessageBytes(poownMessage *model.ProofOfOwnershipMessage) []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(poownMessage.AccountAddress)
	buffer.Write(poownMessage.BlockHash)
	buffer.Write(ConvertUint32ToBytes(poownMessage.BlockHeight))
	return buffer.Bytes()
}

// ParseProofOfOwnershipMessageBytes parse a byte array into a ProofOfOwnershipMessage struct (only the message, no signature)
func ParseProofOfOwnershipMessageBytes(poownMessageBytes []byte) (*model.ProofOfOwnershipMessage, error) {
	buffer := bytes.NewBuffer(poownMessageBytes)
	account, err := accounttype.ParseBytesToAccountType(buffer)
	if err != nil {
		return nil, err
	}
	blockHash, err := ReadTransactionBytes(buffer, int(constant.BlockHash))
	if err != nil {
		return nil, err
	}
	heightBytes, err := ReadTransactionBytes(buffer, int(constant.Height))
	if err != nil {
		return nil, err
	}
	height := ConvertBytesToUint32(heightBytes)
	accountAddress, err := account.GetAccountAddress()
	if err != nil {
		return nil, err
	}
	return &model.ProofOfOwnershipMessage{
		AccountAddress: accountAddress,
		BlockHash:      blockHash,
		BlockHeight:    height,
	}, nil
}
