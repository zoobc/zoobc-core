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
	"encoding/hex"

	"github.com/zoobc/zoobc-core/common/accounttype"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/signaturetype"
	"github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/sha3"
)

var (
	// ZBC_D2EDT53U_5VSQXGQD_COZMETMY_FUVV23NQ_UPLXTR7F_6LKVWNNF_J2SPLUDQ
	senderAddress1 = []byte{0, 0, 0, 0, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56, 139, 255, 81,
		229, 184, 77, 80, 80, 39, 254, 173, 28, 169}
	senderAddress1PassPhrase = "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved"
	// ZBC_BZP2BUBM_NIFDFNSM_BP7J2K5H_RXSPH5OT_2WTPVIUU_KLH6I3DZ_TTD6XEHE
	senderAddress2 = []byte{0, 0, 0, 0, 14, 95, 160, 208, 44, 106, 10, 50, 182, 76, 11, 254, 157, 43, 167, 141, 228, 243, 245, 211, 213,
		166, 250, 162, 148, 82, 207, 228, 108, 121, 156, 199}
	// ZBC_GRIZVZTE_RPDU4OVB_OUF64H22_F4KDXJBQ_3UIWOXPI_SQE2ILRV_WK6BK6G4
	senderAddress3 = []byte{0, 0, 0, 0, 52, 81, 154, 230, 100, 139, 199, 78, 58, 161, 117, 11, 238, 31, 90, 47, 20, 59, 164, 48, 221, 17,
		103, 93, 232, 148, 9, 164, 46, 53, 178, 188}
	// ZBC_HHBIMCR5_7GTKH3SE_HVM2QPDI_DQIR4OLD_3NU5UQDT_BY2HHOS6_DBBEJSLT
	senderAddress4 = []byte{0, 0, 0, 0, 57, 194, 134, 10, 61, 249, 166, 163, 238, 68, 61, 89, 168, 60, 104, 28, 17, 30, 57, 99, 219, 105,
		218, 64, 115, 14, 52, 115, 186, 94, 24, 66}
	// ZNK_IGXGYIX2_Q67MFEYO_7TVQRL7X_NKEVRI4H_OIKR5NXK_FKMFMPZT_G4ZZZ3TE
	recipientAddress1 = []byte{0, 0, 0, 0, 65, 174, 108, 34, 250, 135, 190, 194, 147, 14, 252, 235, 8, 175, 247, 106, 137, 88, 163, 135,
		114, 21, 30, 182, 234, 42, 152, 86, 63, 51, 55, 51}
	// ZBC_EFA2GBTM_UJLAQGZ7_VGCV63HY_CHDBDXLZ_YNIMK67W_QJG7MJMB_3VKFLLYQ
	approverAddress1 = []byte{0, 0, 0, 0, 33, 65, 163, 6, 108, 162, 86, 8, 27, 63, 169, 133, 95, 108, 248, 17, 198, 17, 221, 121, 195, 80,
		197, 123, 246, 130, 77, 246, 37, 129, 221, 84}
	// var senderSeed1 = "prune filth cleaver removable earthworm tricky sulfur citation hesitate stout snort guy"
	nodeSeed1   = "sprinkled sneak species pork outpost thrift unwind cheesy vexingly dizzy neurology neatness"
	nodePubKey1 = []byte{153, 58, 50, 200, 7, 61, 108, 229, 204, 48, 199, 145, 21, 99, 125, 75, 49,
		45, 118, 97, 219, 80, 242, 244, 100, 134, 144, 246, 37, 144, 213, 135}
	nodePubKey2 = []byte{41, 235, 184, 214, 70, 23, 153, 89, 104, 41, 250, 248, 51, 7, 69, 89, 234,
		181, 100, 163, 45, 69, 152, 70, 52, 201, 147, 70, 6, 242, 52, 220}
	block1 = &model.Block{
		ID:                   0,
		PreviousBlockHash:    []byte{},
		Height:               1,
		Timestamp:            1562806389280,
		BlockSeed:            []byte{},
		BlockSignature:       []byte{},
		CumulativeDifficulty: string(100000000),
		PayloadLength:        0,
		PayloadHash:          []byte{},
		BlocksmithPublicKey:  nodePubKey1,
		TotalAmount:          100000000,
		TotalFee:             10000000,
		TotalCoinBase:        1,
		Version:              0,
	}
)

func GetFixturesForNoderegistration(nodeRegistrationQuery query.NodeRegistrationQueryInterface) (
	poownMessage *model.ProofOfOwnershipMessage,
	poown *model.ProofOfOwnership,
	txBody *model.NodeRegistrationTransactionBody,
	txBodyBytes []byte,
) {
	blockHash, _ := util.GetBlockHash(block1, &chaintype.MainChain{})
	poownMessage = &model.ProofOfOwnershipMessage{
		AccountAddress: senderAddress1,
		BlockHash:      blockHash,
		BlockHeight:    0,
	}
	poownMessageBytes := util.GetProofOfOwnershipMessageBytes(poownMessage)
	poownSignature := crypto.NewSignature().SignByNode(poownMessageBytes, nodeSeed1)
	poown = &model.ProofOfOwnership{
		MessageBytes: poownMessageBytes,
		Signature:    poownSignature,
	}
	txBody = &model.NodeRegistrationTransactionBody{
		NodePublicKey:  nodePubKey1,
		AccountAddress: senderAddress1,
		LockedBalance:  10000000000,
		Poown:          poown,
	}
	nr := NodeRegistration{
		Body:                  txBody,
		NodeRegistrationQuery: nodeRegistrationQuery,
	}
	txBodyBytes, _ = nr.GetBodyBytes()
	return poownMessage, poown, txBody, txBodyBytes
}

func GetFixturesForUpdateNoderegistration(nodeRegistrationQuery query.NodeRegistrationQueryInterface) (
	poownMessage *model.ProofOfOwnershipMessage,
	poown *model.ProofOfOwnership,
	txBody *model.UpdateNodeRegistrationTransactionBody,
	txBodyBytes []byte,
) {
	blockHash, _ := util.GetBlockHash(block1, &chaintype.MainChain{})

	poownMessage = &model.ProofOfOwnershipMessage{
		AccountAddress: senderAddress1,
		BlockHash:      blockHash,
		BlockHeight:    0,
	}
	poownMessageBytes := util.GetProofOfOwnershipMessageBytes(poownMessage)
	poownSignature := crypto.NewSignature().SignByNode(poownMessageBytes, nodeSeed1)
	poown = &model.ProofOfOwnership{
		MessageBytes: poownMessageBytes,
		Signature:    poownSignature,
	}
	txBody = &model.UpdateNodeRegistrationTransactionBody{
		NodePublicKey: nodePubKey1,
		LockedBalance: 10000000000,
		Poown:         poown,
	}
	nr := UpdateNodeRegistration{
		Body:                  txBody,
		NodeRegistrationQuery: nodeRegistrationQuery,
	}
	txBodyBytes, _ = nr.GetBodyBytes()
	return poownMessage, poown, txBody, txBodyBytes
}

func GetFixturesForClaimNoderegistration() (
	poown *model.ProofOfOwnership,
	txBody *model.ClaimNodeRegistrationTransactionBody,
	txBodyBytes []byte,
) {

	blockHash, _ := util.GetBlockHash(block1, &chaintype.MainChain{})
	poownMessage := &model.ProofOfOwnershipMessage{
		AccountAddress: senderAddress1,
		BlockHash:      blockHash,
		BlockHeight:    0,
	}
	poownMessageBytes := util.GetProofOfOwnershipMessageBytes(poownMessage)
	poownSignature := crypto.NewSignature().SignByNode(poownMessageBytes, nodeSeed1)
	poown = &model.ProofOfOwnership{
		MessageBytes: poownMessageBytes,
		Signature:    poownSignature,
	}
	txBody = &model.ClaimNodeRegistrationTransactionBody{
		NodePublicKey: nodePubKey1,
		Poown:         poown,
	}
	nr := ClaimNodeRegistration{
		Body: txBody,
	}
	txBodyBytes, _ = nr.GetBodyBytes()
	return
}

func GetFixturesForRemoveNoderegistration() (
	txBody *model.RemoveNodeRegistrationTransactionBody,
	txBodyBytes []byte,
) {

	txBody = &model.RemoveNodeRegistrationTransactionBody{
		NodePublicKey: nodePubKey1,
	}
	nr := RemoveNodeRegistration{
		Body: txBody,
	}
	txBodyBytes, _ = nr.GetBodyBytes()
	return txBody, txBodyBytes
}

func GetFixturesForSetupAccountDataset() (
	txBody *model.SetupAccountDatasetTransactionBody,
	txBodyBytes []byte,
) {
	txBody = &model.SetupAccountDatasetTransactionBody{
		Property: "Admin",
		Value:    "Welcome",
	}

	sa := SetupAccountDataset{
		Body: txBody,
	}
	txBodyBytes, _ = sa.GetBodyBytes()
	return txBody, txBodyBytes
}

func GetFixturesForRemoveAccountDataset() (
	txBody *model.RemoveAccountDatasetTransactionBody,
	txBodyBytes []byte,
) {
	txBody = &model.RemoveAccountDatasetTransactionBody{
		Property: "Admin",
		Value:    "Good bye",
	}

	ra := RemoveAccountDataset{
		Body: txBody,
	}
	txBodyBytes, _ = ra.GetBodyBytes()
	return txBody, txBodyBytes
}

func GetFixturesForTransactionBytes(tx *model.Transaction, sign bool) (txBytes []byte, hashed [32]byte) {
	byteValue, _ := (&Util{}).GetTransactionBytes(tx, sign)
	transactionHash := sha3.Sum256(byteValue)
	return byteValue, transactionHash
}

func GetFixturesForTransaction(
	timestamp int64,
	sender, recipient []byte,
	escrow bool,
) *model.Transaction {

	tx := model.Transaction{
		Version:                 1,
		ID:                      2774809487,
		BlockID:                 1,
		Height:                  1,
		SenderAccountAddress:    sender,
		RecipientAccountAddress: recipient,
		TransactionType:         0,
		Fee:                     1,
		Timestamp:               timestamp,
		TransactionHash:         make([]byte, 32),
		TransactionBodyLength:   0,
		TransactionBodyBytes:    make([]byte, 0),
		TransactionBody:         nil,
		Signature:               make([]byte, 64),
		Escrow: &model.Escrow{
			ApproverAddress: nil,
			Commission:      0,
			Timeout:         0,
		},
	}
	if escrow {
		tx.Escrow = &model.Escrow{
			ID:              tx.GetID(),
			ApproverAddress: approverAddress1,
			Commission:      1,
			Timeout:         100,
		}
	}

	return &tx
}

func GetFixturesForSignedMempoolTransaction(
	id, timestamp int64,
	sender, recipient []byte,
	escrow bool,
) *model.MempoolTransaction {
	transactionUtil := &Util{}
	tx := GetFixturesForTransaction(timestamp, sender, recipient, escrow)
	txBytes, _ := transactionUtil.GetTransactionBytes(tx, false)
	txBytesHash := sha3.Sum256(txBytes)
	signature, _ := (&crypto.Signature{}).Sign(txBytesHash[:], model.AccountType_ZbcAccountType,
		senderAddress1PassPhrase)
	tx.Signature = signature
	txBytes, _ = transactionUtil.GetTransactionBytes(tx, true)
	return &model.MempoolTransaction{
		ID:                      id,
		BlockHeight:             0,
		FeePerByte:              1,
		ArrivalTimestamp:        timestamp,
		TransactionBytes:        txBytes,
		SenderAccountAddress:    sender,
		RecipientAccountAddress: recipient,
	}
}
func GetFixturesForApprovalEscrowTransaction() (
	txBody *model.ApprovalEscrowTransactionBody,
	txBodyBytes []byte,
) {
	txBody = &model.ApprovalEscrowTransactionBody{
		Approval:      model.EscrowApproval_Approve,
		TransactionID: 100,
	}

	sa := ApprovalEscrowTransaction{
		Body: txBody,
	}
	txBodyBytes, _ = sa.GetBodyBytes()
	return txBody, txBodyBytes
}

func GetFixturesForLiquidPaymentTransaction() (
	txBody *model.LiquidPaymentTransactionBody,
	txBodyBytes []byte,
) {
	txBody = &model.LiquidPaymentTransactionBody{
		Amount:          100,
		CompleteMinutes: 200,
	}

	sa := LiquidPaymentTransaction{
		Body: txBody,
	}
	txBodyBytes, _ = sa.GetBodyBytes()
	return txBody, txBodyBytes

}

func GetFixturesForLiquidPaymentStopTransaction() (
	txBody *model.LiquidPaymentStopTransactionBody,
	txBodyBytes []byte,
) {
	txBody = &model.LiquidPaymentStopTransactionBody{
		TransactionID: 123,
	}

	sa := LiquidPaymentStopTransaction{
		Body: txBody,
	}
	txBodyBytes, _ = sa.GetBodyBytes()
	return txBody, txBodyBytes
}

func GetFixtureForSpecificTransaction(
	id, timestamp int64,
	sender, recipient []byte,
	bodyLength uint32,
	transactionType model.TransactionType,
	transactionBody model.TransactionBodyInterface,
	escrow, sign bool,
) (tx *model.Transaction, txBytes []byte) {
	var (
		transactionBytes []byte
	)

	tx = &model.Transaction{
		Version:                 1,
		ID:                      id,
		SenderAccountAddress:    sender,
		RecipientAccountAddress: recipient,
		TransactionType:         uint32(transactionType),
		Fee:                     1,
		Timestamp:               timestamp,
		TransactionBodyLength:   bodyLength,
		TransactionBodyBytes:    make([]byte, bodyLength),
		TransactionIndex:        0,
		TransactionBody:         transactionBody,
		Signature:               nil,
	}

	if escrow {
		tx.Escrow = &model.Escrow{
			ApproverAddress: approverAddress1,
			Commission:      1,
			Timeout:         100,
			Instruction:     "",
		}
	}

	var transactionUtil = &Util{}
	transactionBytes, _ = transactionUtil.GetTransactionBytes(tx, false)
	if sign {
		tx.Signature, _ = (&crypto.Signature{}).Sign(
			transactionBytes,
			model.AccountType_ZbcAccountType,
			senderAddress1PassPhrase,
		)
		transactionBytes, _ = transactionUtil.GetTransactionBytes(tx, true)
		hashed := sha3.Sum256(transactionBytes)
		tx.TransactionHash = hashed[:]

	} else {
		hashed := sha3.Sum256(transactionBytes)
		tx.TransactionHash = hashed[:]
	}

	tx.TransactionBody = nil
	return tx, transactionBytes
}

func GetFixturesForBlock(height uint32, id int64) *model.Block {
	return &model.Block{
		ID:                   id,
		BlockHash:            []byte{},
		PreviousBlockHash:    []byte{},
		Height:               height,
		Timestamp:            10000,
		BlockSeed:            []byte{},
		BlockSignature:       []byte{3},
		CumulativeDifficulty: "1",
		PayloadLength:        1,
		PayloadHash:          []byte{},
		BlocksmithPublicKey:  []byte{},
		TotalAmount:          1000,
		TotalFee:             0,
		TotalCoinBase:        1,
		Version:              0,
	}
}

func GetFixtureForFeeVoteCommitTransaction(
	feeVoteInfo *model.FeeVoteInfo,
	seed string,
) (txBody *model.FeeVoteCommitTransactionBody, txBodyBytes []byte) {
	revealBody := GetFixtureForFeeVoteRevealTransaction(feeVoteInfo, seed)
	digest := sha3.New256()
	_, _ = digest.Write((&FeeVoteRevealTransaction{
		Body: revealBody,
	}).GetFeeVoteInfoBytes())

	txBody = &model.FeeVoteCommitTransactionBody{
		VoteHash: digest.Sum([]byte{}),
	}

	sa := FeeVoteCommitTransaction{
		Body: txBody,
	}
	txBodyBytes, _ = sa.GetBodyBytes()
	return txBody, txBodyBytes
}

func GetFixtureForFeeVoteRevealTransaction(
	voteInfo *model.FeeVoteInfo,
	seed string,
) (body *model.FeeVoteRevealTransactionBody) {
	tx := &FeeVoteRevealTransaction{
		Body: &model.FeeVoteRevealTransactionBody{
			FeeVoteInfo: voteInfo,
		},
	}

	feeVoteSigned, _ := (&crypto.Signature{}).Sign(
		tx.GetFeeVoteInfoBytes(),
		model.AccountType_ZbcAccountType,
		seed,
	)

	tx.Body.VoterSignature = feeVoteSigned

	return tx.Body
}

func GetFixtureForAtomicTransaction(
	unsignedBytes map[string][][]byte,
) (body *model.AtomicTransactionBody, txBodyBytes []byte) {

	var (
		uBytes     = make([][]byte, 0)
		signatures = make(map[string][]byte)
	)
	for seed, txsBytes := range unsignedBytes {
		for _, txBytes := range txsBytes {
			var (
				accType accounttype.AccountTypeInterface
				address []byte
				err     error
			)

			uBytes = append(uBytes, txBytes)
			signature, _ := (&crypto.Signature{}).Sign(
				txBytes,
				model.AccountType_ZbcAccountType,
				seed,
			)
			privateKey, _ := signaturetype.NewEd25519Signature().GetPrivateKeyFromSeedUseSlip10(seed)
			pubKey, _ := signaturetype.NewEd25519Signature().GetPublicKeyFromPrivateKeyUseSlip10(privateKey)
			accType, err = accounttype.NewAccountType((&accounttype.ZbcAccountType{}).GetTypeInt(), pubKey)
			if err != nil {
				panic(err)
			}
			address, err = accType.GetAccountAddress()
			if err != nil {
				panic(err)
			}
			signatures[hex.EncodeToString(address)] = signature
		}
	}

	tx := &AtomicTransaction{
		Body: &model.AtomicTransactionBody{
			UnsignedTransactionBytes: uBytes,
			Signatures:               signatures,
		},
	}
	txBodyBytes, _ = tx.GetBodyBytes()
	return tx.Body, txBodyBytes
}
