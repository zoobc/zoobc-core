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
package crypto

import (
	"bytes"
	"github.com/zoobc/zoobc-core/common/accounttype"
	"github.com/zoobc/zoobc-core/common/signaturetype"
	"golang.org/x/crypto/sha3"

	"github.com/zoobc/zed25519/zed"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	// SignatureInterface represent interface of signature
	SignatureInterface interface {
		Sign(payload []byte, accountType model.AccountType, seed string, optionalParams ...interface{}) ([]byte, error)
		SignByNode(payload []byte, nodeSeed string) []byte
		VerifySignature(payload, signature, accountAddress []byte) error
		VerifyNodeSignature(payload, signature []byte, nodePublicKey []byte) bool
		GenerateAccountFromSeed(accountType accounttype.AccountTypeInterface, seed string, optionalParams ...interface{}) (
			privateKey, publicKey []byte,
			publicKeyString, encodedAddress string,
			fullAccountAddress []byte,
			err error,
		)
		GenerateBlockSeed(payload []byte, nodeSeed string) []byte
	}

	// Signature object handle signing and verifying different signature
	Signature struct {
	}
)

// NewSignature create new instance of signature object
func NewSignature() *Signature {
	return &Signature{}
}

// Sign accept account ID and payload to be signed then return the signature byte based on the
// signature method associated with account.Type
func (*Signature) Sign(
	payload []byte,
	accountTypeEnum model.AccountType,
	seed string,
	optionalParams ...interface{},
) ([]byte, error) {
	accountType, err := accounttype.NewAccountType(int32(accountTypeEnum), nil)
	if err != nil {
		return nil, err
	}
	return accountType.Sign(payload, seed, optionalParams...)
}

// SignByNode special method for signing block only, there will be no multiple signature options
func (*Signature) SignByNode(payload []byte, nodeSeed string) []byte {
	var (
		buffer           = bytes.NewBuffer([]byte{})
		ed25519Signature = signaturetype.NewEd25519Signature()
		nodePrivateKey   = ed25519Signature.GetPrivateKeyFromSeed(nodeSeed)
		signature        = ed25519Signature.Sign(nodePrivateKey, payload)
	)
	buffer.Write(signature)
	return buffer.Bytes()
}

// VerifySignature accept payload (before without signature), signature and the account id
// then verify the signature + public key against the payload based on the
func (*Signature) VerifySignature(payload, signature, accountAddress []byte) error {
	var (
		accountTypeInt = int32(util.ConvertBytesToUint32(accountAddress[:4]))
	)
	accountType, err := accounttype.NewAccountType(accountTypeInt, accountAddress[4:])
	if err != nil {
		return err
	}
	return accountType.VerifySignature(payload, signature, accountAddress)
}

// VerifyNodeSignature Verify a signature of a block or message signed with a node private key
// Note: this function is a wrapper around the ed25519 algorithm
func (*Signature) VerifyNodeSignature(payload, signature, nodePublicKey []byte) bool {
	var result = signaturetype.NewEd25519Signature().Verify(nodePublicKey, payload, signature)
	return result
}

// GenerateAccountFromSeed to generate account based on provided seed
func (*Signature) GenerateAccountFromSeed(accountType accounttype.AccountTypeInterface, seed string, optionalParams ...interface{}) (
	privateKey, publicKey []byte,
	publicKeyString, encodedAddress string,
	fullAccountAddress []byte,
	err error,
) {
	err = accountType.GenerateAccountFromSeed(seed, optionalParams...)
	if err != nil {
		return nil, nil, "", "", nil, err
	}
	privateKey, err = accountType.GetAccountPrivateKey()
	if err != nil {
		return nil, nil, "", "", nil, err
	}
	publicKey = accountType.GetAccountPublicKey()
	publicKeyString, err = accountType.GetAccountPublicKeyString()
	if err != nil {
		return nil, nil, "", "", nil, err
	}
	encodedAddress, err = accountType.GetEncodedAddress()
	if err != nil {
		return nil, nil, "", "", nil, err
	}
	fullAccountAddress, err = accountType.GetAccountAddress()
	if err != nil {
		return nil, nil, "", "", nil, err
	}
	return
}

// GenerateBlockSeed special method for generating block seed using zed
func (*Signature) GenerateBlockSeed(payload []byte, nodeSeed string) []byte {
	var (
		buffer       = bytes.NewBuffer([]byte{})
		seedBuffer   = []byte(nodeSeed)
		seedHash     = sha3.Sum256(seedBuffer)
		seedByte     = seedHash[:]
		zedSecret    = zed.SecretFromSeed(seedByte)
		zedSignature = zedSecret.Sign(payload)
	)
	buffer.Write(zedSignature[:])
	return buffer.Bytes()
}
