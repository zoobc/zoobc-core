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
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"golang.org/x/crypto/sha3"
)

// ETHAccountType the default account type
type ETHAccountType struct {
	privateKey, publicKey, fullAddress []byte
}

func (acc *ETHAccountType) SetAccountPublicKey(accountPublicKey []byte) {
	if uint32(len(accountPublicKey)) > acc.GetAccountPublicKeyLength() {
		return
	}
	if accountPublicKey == nil {
		acc.publicKey = make([]byte, 0)
	}
	acc.publicKey = accountPublicKey
}

func (acc *ETHAccountType) GetAccountAddress() ([]byte, error) {
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

func (acc *ETHAccountType) GetTypeInt() int32 {
	return int32(model.AccountType_ETHAccountType)
}

func (acc *ETHAccountType) GetAccountPublicKey() []byte {
	return acc.publicKey
}

func (acc *ETHAccountType) GetAccountPrefix() string {
	return "ETH"
}

func (acc *ETHAccountType) GetName() string {
	return "ETHAccount"
}

func (acc *ETHAccountType) GetAccountPublicKeyLength() uint32 {
	return 64
}

func (acc *ETHAccountType) IsEqual(acc2 AccountTypeInterface) bool {
	return bytes.Equal(acc.GetAccountPublicKey(), acc2.GetAccountPublicKey()) && acc.GetTypeInt() == acc2.GetTypeInt()
}

func (acc *ETHAccountType) GetSignatureType() model.SignatureType {
	return model.SignatureType_EthereumSignature
}

func (acc *ETHAccountType) GetSignatureLength() uint32 {
	return constant.EthereumSignatureLength
}

func (acc *ETHAccountType) GetEncodedAddress() (string, error) {
	if acc.GetAccountPublicKey() == nil || bytes.Equal(acc.GetAccountPublicKey(), []byte{}) {
		return "", errors.New("EmptyAccountPublicKey")
	}
	hash := sha3.NewLegacyKeccak256()
	_, err := hash.Write(acc.publicKey)
	if err != nil {
		return "", err
	}
	return hexutil.Encode(hash.Sum(nil)[12:]), nil
}

func (acc *ETHAccountType) DecodePublicKeyFromAddress(address string) ([]byte, error) {
	return nil, errors.New("NoImplementation")
}

func (acc *ETHAccountType) GenerateAccountFromSeed(seed string, optionalParams ...interface{}) error {
	return errors.New("NoImplementation")
}

func (acc *ETHAccountType) GetAccountPublicKeyString() (string, error) {
	return acc.GetEncodedAddress()
}

// GetAccountPrivateKey: we don't store neither generate private key of this account type as the private key resides in the eID
func (acc *ETHAccountType) GetAccountPrivateKey() ([]byte, error) {
	return acc.privateKey, nil
}

// Sign: this account type and signature comes only from Estonia eID and we don't have private key of the accounts
// so we don't sign sign for this type of account.
func (acc *ETHAccountType) Sign(payload []byte, seed string, optionalParams ...interface{}) ([]byte, error) {
	hash := crypto.Keccak256Hash(payload)
	privateKey, err := crypto.ToECDSA(acc.privateKey)
	if err != nil {
		return nil, fmt.Errorf("InvalidPrivateKey:%s", err.Error())
	}
	return crypto.Sign(hash.Bytes(), privateKey)
}

func (acc *ETHAccountType) VerifySignature(payload, signature, accountAddress []byte) error {
	hash := crypto.Keccak256Hash(payload)

	signatureNoRecoverID := signature[:len(signature)-1] // remove recovery id

	publicKeyWithPrefix := []byte{4}
	publicKeyWithPrefix = append(publicKeyWithPrefix, acc.publicKey...)
	verified := crypto.VerifySignature(publicKeyWithPrefix, hash.Bytes(), signatureNoRecoverID)
	if !verified {
		return errors.New("InvalidSignature")
	}

	return nil
}
