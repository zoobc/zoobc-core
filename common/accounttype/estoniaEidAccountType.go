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
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
)

// EstoniaEidAccountType the default account type
type EstoniaEidAccountType struct {
	publicKey, fullAddress []byte
}

func (acc *EstoniaEidAccountType) SetAccountPublicKey(accountPublicKey []byte) {
	if accountPublicKey == nil {
		acc.publicKey = make([]byte, 0)
	}
	acc.publicKey = accountPublicKey
}

func (acc *EstoniaEidAccountType) GetAccountAddress() ([]byte, error) {
	if acc.fullAddress != nil {
		return acc.fullAddress, nil
	}
	if acc.GetAccountPublicKey() == nil {
		return nil, errors.New("AccountAddressPublicKeyEmpty")
	}
	buff := bytes.NewBuffer([]byte{})
	tmpBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmpBuf, uint32(acc.GetTypeInt()))
	buff.Write(tmpBuf)
	buff.Write(acc.GetAccountPublicKey())
	acc.fullAddress = buff.Bytes()
	return acc.fullAddress, nil
}

func (acc *EstoniaEidAccountType) GetTypeInt() int32 {
	return int32(model.AccountType_EstoniaEidAccountType)
}

func (acc *EstoniaEidAccountType) GetAccountPublicKey() []byte {
	return acc.publicKey
}

func (acc *EstoniaEidAccountType) GetAccountPrefix() string {
	return ""
}

func (acc *EstoniaEidAccountType) GetName() string {
	return "EstoniaEid"
}

func (acc *EstoniaEidAccountType) GetAccountPublicKeyLength() uint32 {
	return 97
}

func (acc *EstoniaEidAccountType) IsEqual(acc2 AccountTypeInterface) bool {
	return bytes.Equal(acc.GetAccountPublicKey(), acc2.GetAccountPublicKey()) && acc.GetTypeInt() == acc2.GetTypeInt()
}

func (acc *EstoniaEidAccountType) GetSignatureType() model.SignatureType {
	return model.SignatureType_EstoniaEidSignature
}

func (acc *EstoniaEidAccountType) GetSignatureLength() uint32 {
	return constant.EstoniaEidSignatureLength
}

func (acc *EstoniaEidAccountType) GetEncodedAddress() (string, error) {
	if acc.GetAccountPublicKey() == nil || bytes.Equal(acc.GetAccountPublicKey(), []byte{}) {
		return "", errors.New("EmptyAccountPublicKey")
	}
	return hex.EncodeToString(acc.GetAccountPublicKey()), nil
}

func (acc *EstoniaEidAccountType) DecodePublicKeyFromAddress(address string) ([]byte, error) {
	return hex.DecodeString(address)
}

func (acc *EstoniaEidAccountType) GenerateAccountFromSeed(seed string, optionalParams ...interface{}) error {
	return errors.New("NoImplementation")
}

func (acc *EstoniaEidAccountType) GetAccountPublicKeyString() (string, error) {
	return acc.GetEncodedAddress()
}

// GetAccountPrivateKey: we don't store neither generate private key of this account type as the private key resides in the eID
func (acc *EstoniaEidAccountType) GetAccountPrivateKey() ([]byte, error) {
	return nil, nil
}

// Sign: this account type and signature comes only from Estonia eID and we don't have private key of the accounts
// so we don't sign sign for this type of account.
func (acc *EstoniaEidAccountType) Sign(payload []byte, seed string, optionalParams ...interface{}) ([]byte, error) {
	return []byte{}, nil
}

func (acc *EstoniaEidAccountType) VerifySignature(payload, signature, accountAddress []byte) error {
	publicKey := acc.loadPublicKeyFromDer(acc.GetAccountPublicKey())
	r, s, _ := acc.decodeSignatureNIST384RS(signature)
	if !ecdsa.Verify(&publicKey, payload, r, s) {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"InvalidSignature",
		)
	}
	return nil
}

// decodeSignatureNIST384RS: decode signature to R and S component
// source: https://github.com/warner/python-ecdsa/blob/master/src/ecdsa/util.py (sigdecode_string)
func (acc *EstoniaEidAccountType) decodeSignatureNIST384RS(signature []byte) (r, s *big.Int, err error) {
	// curveOrder "39402006196394479212279040100143613805079739270465446667946905279627659399113263569398956308152294913554433653942643"
	curveOrderLen := 48
	if len(signature) != curveOrderLen*2 {
		return nil, nil, fmt.Errorf("error signature length: %d", len(signature))
	}
	rBytes := signature[:48]
	sBytes := signature[48:]
	r = new(big.Int).SetBytes(rBytes)
	s = new(big.Int).SetBytes(sBytes)
	return r, s, nil
}

// loadPublicKeyFromDer loads public key from DER byte data
func (acc *EstoniaEidAccountType) loadPublicKeyFromDer(publicKeyBytes []byte) (publicKey ecdsa.PublicKey) {
	curve := elliptic.P384()
	publicKey.Curve = curve
	publicKey.X, publicKey.Y = elliptic.Unmarshal(curve, publicKeyBytes)
	return publicKey
}
