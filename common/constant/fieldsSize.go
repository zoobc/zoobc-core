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
package constant

var (
	AccountAddressTypeLength   uint32 = 4
	TransactionSignatureLength uint32 = 4
	// NodePublicKey TODO: this is valid for pub keys generated using Ed25519. in future we might have more implementations
	NodePublicKey uint32 = 32
	Balance       uint32 = 8
	BlockHash     uint32 = 32
	Height        uint32 = 4
	// NodeSignature node always signs using Ed25519 algorithm, that produces 64 bytes signatures
	NodeSignature         uint32 = 64
	TransactionType       uint32 = 4
	TransactionVersion    uint32 = 1
	Timestamp             uint32 = 8
	Fee                   uint32 = 8
	TransactionBodyLength uint32 = 4
	// SignatureType variables
	SignatureType   uint32 = 4
	AuthRequestType        = 4
	AuthTimestamp          = 8
	// DatasetPropertyLength is max length of string property name in dataset
	DatasetPropertyLength uint32 = 4
	// DatasetValueLength is max length of string property value in dataset
	DatasetValueLength uint32 = 4

	TxMessageBytesLength uint32 = 4

	EscrowApproverAddressLength uint32 = 4
	EscrowCommissionLength      uint32 = 8
	EscrowTimeoutLength         uint32 = 8
	EscrowApproval              uint32 = 4
	EscrowID                    uint32 = 8
	EscrowApprovalBytesLength          = EscrowApproval + EscrowID
	EscrowInstructionLength     uint32 = 4
	MultisigFieldLength         uint32 = 4
	// MultiSigFieldMissing indicate fields is missing, no need to read the bytes
	MultiSigFieldMissing uint32
	// MultiSigFieldPresent indicate fields is present, parse the byte accordingly
	MultiSigFieldPresent           uint32 = 1
	MultiSigAddressLength          uint32 = 4
	MultiSigSignatureLength        uint32 = 4
	MultiSigSignatureAddressLength uint32 = 4
	MultiSigNumberOfAddress        uint32 = 4
	MultiSigNumberOfSignatures     uint32 = 4
	MultiSigUnsignedTxBytesLength  uint32 = 4
	MultiSigInfoSize               uint32 = 4
	MultiSigInfoSignatureInfoSize  uint32 = 4
	MultiSigInfoNonce              uint32 = 8
	MultiSigInfoMinSignature       uint32 = 4
	MultiSigTransactionHash        uint32 = 32

	// FeeVote part
	FeeVote              uint32 = 8
	RecentBlockHeight    uint32 = 4
	VoterSignatureLength uint32 = 4

	// Liquid Transaction

	LiquidPaymentCompleteMinutesLength uint32 = 8
	TransactionID                      uint32 = 8
)
