package transaction

import (
	"github.com/zoobc/zoobc-core/common/transaction"

	"github.com/zoobc/zoobc-core/common/crypto"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
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
	}
)

func GenerateTxSendMoney(tx *model.Transaction, sendAmount int64) *model.Transaction {
	tx.TransactionType = util.ConvertBytesToUint32(txTypeMap["sendMoney"])
	tx.TransactionBody = &model.Transaction_SendMoneyTransactionBody{
		SendMoneyTransactionBody: &model.SendMoneyTransactionBody{
			Amount: sendAmount,
		},
	}
	tx.TransactionBodyBytes = util.ConvertUint64ToBytes(uint64(sendAmount))
	tx.TransactionBodyLength = 8
	return tx
}

func GenerateTxRegisterNode(tx *model.Transaction,
	nodeOwnerAccountAddress, nodeSeed, recipientAccountAddress, nodeAddress string,
	lockedBalance int64) *model.Transaction {
	nodePubKey := util.GetPublicKeyFromSeed(nodeSeed)
	poowMessage := GenerateMockPoowMessage(nodeOwnerAccountAddress)
	poownMessageBytes := util.GetProofOfOwnershipMessageBytes(poowMessage)
	signature := (&crypto.Signature{}).SignByNode(
		poownMessageBytes,
		nodeSeed)
	txBody := &model.NodeRegistrationTransactionBody{
		AccountAddress: recipientAccountAddress,
		NodePublicKey:  nodePubKey,
		NodeAddress: &model.NodeAddress{
			Address: nodeAddress,
		},
		LockedBalance: lockedBalance,
		Poown: &model.ProofOfOwnership{
			MessageBytes: poownMessageBytes,
			Signature:    signature,
		},
	}
	txBodyBytes := (&transaction.NodeRegistration{
		Body:                  txBody,
		NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
	}).GetBodyBytes()

	tx.TransactionType = util.ConvertBytesToUint32(txTypeMap["registerNode"])
	tx.TransactionBody = &model.Transaction_NodeRegistrationTransactionBody{
		NodeRegistrationTransactionBody: txBody,
	}
	tx.TransactionBodyBytes = txBodyBytes
	tx.TransactionBodyLength = uint32(len(txBodyBytes))

	return tx
}

func GenerateTxUpdateNode(tx *model.Transaction, nodeOwnerAccountAddress, nodeSeed, nodeAddress string,
	lockedBalance int64) *model.Transaction {
	nodePubKey := util.GetPublicKeyFromSeed(nodeSeed)
	poowMessage := GenerateMockPoowMessage(nodeOwnerAccountAddress)
	poownMessageBytes := util.GetProofOfOwnershipMessageBytes(poowMessage)
	signature := (&crypto.Signature{}).SignByNode(
		poownMessageBytes,
		nodeSeed)
	txBody := &model.UpdateNodeRegistrationTransactionBody{
		NodePublicKey: nodePubKey,
		NodeAddress: &model.NodeAddress{
			Address: nodeAddress,
		},
		LockedBalance: lockedBalance,
		Poown: &model.ProofOfOwnership{
			MessageBytes: poownMessageBytes,
			Signature:    signature,
		},
	}
	txBodyBytes := (&transaction.UpdateNodeRegistration{
		Body:                  txBody,
		NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
	}).GetBodyBytes()

	tx.TransactionBodyLength = uint32(len(txBodyBytes))
	tx.TransactionType = util.ConvertBytesToUint32(txTypeMap["updateNodeRegistration"])
	tx.TransactionBody = &model.Transaction_UpdateNodeRegistrationTransactionBody{
		UpdateNodeRegistrationTransactionBody: txBody,
	}
	tx.TransactionBodyBytes = txBodyBytes
	return tx
}

func GenerateTxRemoveNode(tx *model.Transaction, nodeSeed string) *model.Transaction {
	nodePubKey := util.GetPublicKeyFromSeed(nodeSeed)
	txBody := &model.RemoveNodeRegistrationTransactionBody{
		NodePublicKey: nodePubKey,
	}
	txBodyBytes := (&transaction.RemoveNodeRegistration{
		Body:                  txBody,
		NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
	}).GetBodyBytes()

	tx.TransactionType = util.ConvertBytesToUint32(txTypeMap["removeNodeRegistration"])
	tx.TransactionBody = &model.Transaction_RemoveNodeRegistrationTransactionBody{
		RemoveNodeRegistrationTransactionBody: txBody,
	}
	tx.TransactionBodyBytes = txBodyBytes
	tx.TransactionBodyLength = uint32(len(txBodyBytes))
	return tx
}

func GenerateTxClaimNode(tx *model.Transaction, nodeOwnerAccountAddress, nodeSeed, recipientAccountAddress string) *model.Transaction {
	nodePubKey := util.GetPublicKeyFromSeed(nodeSeed)
	poowMessage := GenerateMockPoowMessage(nodeOwnerAccountAddress)
	poownMessageBytes := util.GetProofOfOwnershipMessageBytes(poowMessage)
	signature := (&crypto.Signature{}).SignByNode(
		poownMessageBytes,
		nodeSeed)
	txBody := &model.ClaimNodeRegistrationTransactionBody{
		AccountAddress: recipientAccountAddress,
		NodePublicKey:  nodePubKey,
		Poown: &model.ProofOfOwnership{
			MessageBytes: poownMessageBytes,
			Signature:    signature,
		},
	}
	txBodyBytes := (&transaction.ClaimNodeRegistration{
		Body: txBody,
	}).GetBodyBytes()

	tx.TransactionType = util.ConvertBytesToUint32(txTypeMap["claimNodeRegistration"])
	tx.TransactionBody = &model.Transaction_ClaimNodeRegistrationTransactionBody{
		ClaimNodeRegistrationTransactionBody: txBody,
	}
	tx.TransactionBodyBytes = txBodyBytes
	tx.TransactionBodyLength = uint32(len(txBodyBytes))
	return tx
}

func GenerateTxSetupAccountDataset(tx *model.Transaction,
	senderAccountAddress, recipientAccountAddress, property, value string,
	activeTime uint64) *model.Transaction {
	txBody := &model.SetupAccountDatasetTransactionBody{
		SetterAccountAddress:    senderAccountAddress,
		RecipientAccountAddress: recipientAccountAddress,
		Property:                property,
		Value:                   value,
		MuchTime:                activeTime,
	}
	txBodyBytes := (&transaction.SetupAccountDataset{
		Body: txBody,
	}).GetBodyBytes()

	tx.TransactionType = util.ConvertBytesToUint32(txTypeMap["setupAccountDataset"])
	tx.TransactionBody = &model.Transaction_SetupAccountDatasetTransactionBody{
		SetupAccountDatasetTransactionBody: txBody,
	}
	tx.TransactionBodyBytes = txBodyBytes
	tx.TransactionBodyLength = uint32(len(txBodyBytes))
	return tx
}

func GenerateTxRemoveAccountDataset(tx *model.Transaction,
	senderAccountAddress, recipientAccountAddress, property, value string) *model.Transaction {
	txBody := &model.RemoveAccountDatasetTransactionBody{
		SetterAccountAddress:    senderAccountAddress,
		RecipientAccountAddress: recipientAccountAddress,
		Property:                property,
		Value:                   value,
	}
	txBodyBytes := (&transaction.RemoveAccountDataset{
		Body: txBody,
	}).GetBodyBytes()

	tx.TransactionType = util.ConvertBytesToUint32(txTypeMap["removeAccountDataset"])
	tx.TransactionBody = &model.Transaction_RemoveAccountDatasetTransactionBody{
		RemoveAccountDatasetTransactionBody: txBody,
	}
	tx.TransactionBodyBytes = txBodyBytes
	tx.TransactionBodyLength = uint32(len(txBodyBytes))
	return tx
}
