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
	"github.com/zoobc/zoobc-core/common/model"
)

// AccountTypeInterface interface define the different behavior of each address
type (
	AccountTypeInterface interface {
		// SetAccountPublicKey set/update account public key
		SetAccountPublicKey(accountPublicKey []byte)
		// GetAccountAddress return the full (raw) account address in bytes
		GetAccountAddress() ([]byte, error)
		// GetTypeInt return the value of the account address type in int
		GetTypeInt() int32
		// GetAccountPublicKey return an account address in bytes
		GetAccountPublicKey() []byte
		// GetAccountPrefix return the value of current account address table prefix in the database
		GetAccountPrefix() string
		// GetName return the name of the account address type
		GetName() string
		// GetAccountPublicKeyLength return the length of this account address type (for parsing tx and message bytes that embed an address)
		GetAccountPublicKeyLength() uint32
		// GetEncodedAddress return a string encoded/formatted account address
		GetEncodedAddress() (string, error)
		// GetAccountPublicKeyString return a string encoded account public key
		GetAccountPublicKeyString() (string, error)
		GetAccountPrivateKey() ([]byte, error)
		// IsEqual checks if two account have same type and pub key
		IsEqual(acc AccountTypeInterface) bool
		// GetSignatureType return the signature type number for this account type
		GetSignatureType() model.SignatureType
		// GetSignatureLength return the signature length for this account type
		GetSignatureLength() uint32

		// Sign accept a payload to be signed with an account seed then return the signature byte based on the
		Sign(payload []byte, seed string, optionalParams ...interface{}) ([]byte, error)
		VerifySignature(payload, signature, accountAddress []byte) error
		GenerateAccountFromSeed(seed string, optionalParams ...interface{}) error
	}
)
