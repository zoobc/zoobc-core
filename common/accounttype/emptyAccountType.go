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
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/model"
)

// EmptyAccountType the default account type
type EmptyAccountType struct {
	accountPublicKey []byte
}

func (acc *EmptyAccountType) SetAccountPublicKey(accountPublicKey []byte) {
	if accountPublicKey == nil {
		acc.accountPublicKey = make([]byte, 0)
	}
	// could be a zero-padded pub key
	acc.accountPublicKey = accountPublicKey
}

func (acc *EmptyAccountType) GetAccountAddress() ([]byte, error) {
	buff := bytes.NewBuffer([]byte{})
	tmpBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmpBuf, uint32(acc.GetTypeInt()))
	buff.Write(tmpBuf)
	buff.Write(acc.GetAccountPublicKey())
	return buff.Bytes(), nil
}

func (acc *EmptyAccountType) GetTypeInt() int32 {
	return int32(model.AccountType_EmptyAccountType)
}

func (acc *EmptyAccountType) GetAccountPublicKey() []byte {
	if acc.accountPublicKey == nil {
		return make([]byte, 0)
	}
	return acc.accountPublicKey
}

func (acc *EmptyAccountType) GetAccountPrefix() string {
	return ""
}

func (acc *EmptyAccountType) GetName() string {
	return "Empty"
}

func (acc *EmptyAccountType) GetAccountPublicKeyLength() uint32 {
	return 0
}

func (acc *EmptyAccountType) IsEqual(acc2 AccountTypeInterface) bool {
	return bytes.Equal(acc.GetAccountPublicKey(), acc2.GetAccountPublicKey()) && acc.GetTypeInt() == acc2.GetTypeInt()
}

func (acc *EmptyAccountType) GetSignatureType() model.SignatureType {
	return model.SignatureType_DefaultSignature
}

func (acc *EmptyAccountType) GetSignatureLength() uint32 {
	return 0
}

func (acc *EmptyAccountType) GetEncodedAddress() (string, error) {
	return "", nil
}

func (acc *EmptyAccountType) Sign(payload []byte, seed string, optionalParams ...interface{}) ([]byte, error) {
	return nil, blocker.NewBlocker(
		blocker.ValidationErr,
		"EmptyAccountTypeCannotSign",
	)
}

func (acc *EmptyAccountType) VerifySignature(payload, signature, accountAddress []byte) error {
	return blocker.NewBlocker(
		blocker.ValidationErr,
		"EmptyAccountTypeCannotSign",
	)
}

func (acc *EmptyAccountType) GetAccountPublicKeyString() (string, error) {
	return "", nil
}

func (acc *EmptyAccountType) GetAccountPrivateKey() ([]byte, error) {
	return []byte{}, nil
}

func (acc *EmptyAccountType) GenerateAccountFromSeed(seed string, optionalParams ...interface{}) error {
	return blocker.NewBlocker(
		blocker.ValidationErr,
		"EmptyAccountTypeCannotGenerateAccountFromSeed",
	)
}
