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
package transaction

import (
	"github.com/zoobc/zoobc-core/common/crypto"
)

var (
	txTypeMap = map[string][]byte{
		"sendMoney":              {1, 0, 0, 0},
		"registerNode":           {2, 0, 0, 0},
		"updateNodeRegistration": {2, 1, 0, 0},
		"removeNodeRegistration": {2, 2, 0, 0},
		"claimNodeRegistration":  {2, 3, 0, 0},
		"setupAccountDataset":    {3, 0, 0, 0},
		"removeAccountDataset":   {3, 1, 0, 0},
		"approvalEscrow":         {4, 0, 0, 0},
		"multiSignature":         {5, 0, 0, 0},
		"liquidPayment":          {6, 0, 0, 0},
		"liquidPaymentStop":      {6, 1, 0, 0},
		"feeVoteCommit":          {7, 0, 0, 0},
		"feeVoteReveal":          {7, 1, 0, 0},
		"atomic":                 {8, 0, 0, 0},
	}
	signature = &crypto.Signature{}

	// Basic transaction data
	message                    string
	outputType                 string
	version                    uint32
	timestamp                  int64
	senderSeed                 string
	recipientAccountAddressHex string
	fee                        int64
	post                       bool
	postHost                   string
	senderAddressHex           string
	sign                       bool

	// Send money transaction
	sendAmount int64

	// node registration transaction
	nodeSeed                   string
	nodeOwnerAccountAddressHex string
	lockedBalance              int64
	proofOfOwnershipHex        string
	databasePath               string
	databaseName               string

	// dataset transaction
	property string
	value    string
	// escrowable
	escrow               bool
	esApproverAddressHex string
	esCommission         int64
	esTimeout            uint64
	esInstruction        string

	// escrowApproval
	approval      bool
	transactionID int64

	// multiSignature
	unsignedTxHex     string
	addressSignatures map[string]string
	txHash            string
	addressesHex      []string
	nonce             int64
	minSignature      uint32

	// fee vote
	recentBlockHeight uint32
	feeVote           int64
	dbPath, dBName    string
	// liquidPayment
	completeMinutes uint64

	// atomic transaction
	inners uint32
)
