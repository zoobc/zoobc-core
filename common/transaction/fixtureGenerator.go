package transaction

import (
	"golang.org/x/crypto/sha3"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

var (
	senderAddress1 = "ZNK_TE5DFSAH_HVWOLTBQ_Y6IRKY35_JMYS25TB_3NIPF5DE_Q2IPMJMQ_2WDWZB5Q"
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
	TransactionWithEscrow = &model.Transaction{
		ID:                      670925173877174625,
		Version:                 1,
		TransactionType:         2,
		BlockID:                 0,
		Height:                  0,
		Timestamp:               1562806389280,
		SenderAccountAddress:    "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
		RecipientAccountAddress: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
		Fee:                     1,
		TransactionHash: []byte{
			59, 106, 191, 6, 145, 54, 181, 186, 75, 93, 234, 139, 131, 96, 153, 252, 40, 245, 235, 132,
			187, 45, 245, 113, 210, 87, 23, 67, 157, 117, 41, 143,
		},
		TransactionBodyLength: 8,
		TransactionBodyBytes:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
		Signature: []byte{
			0, 0, 0, 0, 4, 38, 103, 73, 250, 169, 63, 155, 106, 21, 9, 76, 77, 137, 3, 120, 21, 69, 90, 118, 242, 84, 174,
			239, 46, 190, 78, 68, 90, 83, 142, 11, 4, 38, 68, 24, 230, 247, 88, 220, 119, 124, 51, 149, 127, 214, 82, 224, 72, 239, 56,
			139, 255, 81, 229, 184, 77, 80, 80, 39, 254, 173, 28, 169,
		},
		Escrow: &model.Escrow{
			ApproverAddress: "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
			Commission:      1,
			Timeout:         100,
		},
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
	txBodyBytes = nr.GetBodyBytes()
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
	txBodyBytes = nr.GetBodyBytes()
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
	txBodyBytes = nr.GetBodyBytes()
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
	txBodyBytes = nr.GetBodyBytes()
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
	return txBody, sa.GetBodyBytes()
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
	return txBody, ra.GetBodyBytes()
}

func GetFixturesForTransactionBytes(tx *model.Transaction, sign bool) (txBytes []byte, hashed [32]byte) {
	byteValue, _ := (&Util{}).GetTransactionBytes(tx, sign)
	transactionHash := sha3.Sum256(byteValue)
	return byteValue, transactionHash
}

func GetFixturesForTransaction(
	timestamp int64,
	sender, recipient string,
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
			ApproverAddress: "",
			Commission:      0,
			Timeout:         0,
		},
	}
	if escrow {
		tx.Escrow = &model.Escrow{
			ID:              tx.GetID(),
			ApproverAddress: "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
			Commission:      1,
			Timeout:         100,
		}
	}

	return &tx
}

func GetFixturesForSignedMempoolTransaction(
	id, timestamp int64,
	sender, recipient string,
	escrow bool,
) *model.MempoolTransaction {
	transactionUtil := &Util{}
	tx := GetFixturesForTransaction(timestamp, sender, recipient, escrow)
	txBytes, _ := transactionUtil.GetTransactionBytes(tx, false)
	txBytesHash := sha3.Sum256(txBytes)
	signature, _ := (&crypto.Signature{}).Sign(txBytesHash[:], model.SignatureType_DefaultSignature,
		"concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved")
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
	return txBody, sa.GetBodyBytes()
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
	return txBody, sa.GetBodyBytes()
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
	return txBody, sa.GetBodyBytes()
}

func GetFixtureForSpecificTransaction(
	id, timestamp int64,
	sender, recipient string,
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
		Escrow: &model.Escrow{
			ApproverAddress: "",
			Commission:      0,
			Timeout:         0,
			Instruction:     "",
		},
	}

	if escrow {
		tx.Escrow = &model.Escrow{
			ApproverAddress: "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
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
			model.SignatureType_DefaultSignature,
			"concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved",
		)
		transactionBytes, _ = transactionUtil.GetTransactionBytes(tx, true)
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
	return txBody, sa.GetBodyBytes()
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
		model.SignatureType_DefaultSignature,
		seed,
	)

	tx.Body.VoterSignature = feeVoteSigned

	return tx.Body
}
