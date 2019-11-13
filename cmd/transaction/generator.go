package transaction

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
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

func GenerateTxRegisterNode(
	tx *model.Transaction,
	nodeOwnerAccountAddress, nodeSeed, recipientAccountAddress, nodeAddress string,
	lockedBalance int64,
	sqliteDB *sql.DB,
) *model.Transaction {
	lastBlock, err := util.GetLastBlock(query.NewQueryExecutor(sqliteDB), query.NewBlockQuery(chaintype.GetChainType(0)))
	if err != nil {
		panic(err)
	}
	poowMessage := &model.ProofOfOwnershipMessage{
		AccountAddress: nodeOwnerAccountAddress,
		BlockHash:      lastBlock.BlockHash,
		BlockHeight:    lastBlock.Height,
	}

	nodePubKey := util.GetPublicKeyFromSeed(nodeSeed)
	poownMessageBytes := util.GetProofOfOwnershipMessageBytes(poowMessage)
	signature := (&crypto.Signature{}).SignByNode(poownMessageBytes, nodeSeed)
	txBody := &model.NodeRegistrationTransactionBody{
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

func GenerateTxUpdateNode(
	tx *model.Transaction,
	nodeOwnerAccountAddress, nodeSeed, nodeAddress string,
	lockedBalance int64,
	sqliteDB *sql.DB,
) *model.Transaction {
	lastBlock, err := util.GetLastBlock(query.NewQueryExecutor(sqliteDB), query.NewBlockQuery(chaintype.GetChainType(0)))
	if err != nil {
		panic(err)
	}
	poowMessage := &model.ProofOfOwnershipMessage{
		AccountAddress: nodeOwnerAccountAddress,
		BlockHash:      lastBlock.BlockHash,
		BlockHeight:    lastBlock.Height,
	}

	nodePubKey := util.GetPublicKeyFromSeed(nodeSeed)
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

func GenerateTxClaimNode(
	tx *model.Transaction,
	nodeOwnerAccountAddress, nodeSeed, recipientAccountAddress string,
	sqliteDB *sql.DB,
) *model.Transaction {
	lastBlock, err := util.GetLastBlock(query.NewQueryExecutor(sqliteDB), query.NewBlockQuery(chaintype.GetChainType(0)))
	if err != nil {
		panic(err)
	}
	poowMessage := &model.ProofOfOwnershipMessage{
		AccountAddress: nodeOwnerAccountAddress,
		BlockHash:      lastBlock.BlockHash,
		BlockHeight:    lastBlock.Height,
	}

	nodePubKey := util.GetPublicKeyFromSeed(nodeSeed)
	poownMessageBytes := util.GetProofOfOwnershipMessageBytes(poowMessage)
	signature := (&crypto.Signature{}).SignByNode(
		poownMessageBytes,
		nodeSeed)
	txBody := &model.ClaimNodeRegistrationTransactionBody{
		NodePublicKey: nodePubKey,
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

func GenerateTxSetupAccountDataset(
	tx *model.Transaction,
	senderAccountAddress, recipientAccountAddress, property, value string,
	activeTime uint64,
) *model.Transaction {
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

func GenerateTxRemoveAccountDataset(
	tx *model.Transaction,
	senderAccountAddress, recipientAccountAddress, property, value string,
) *model.Transaction {
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

/*
Basic Func
*/
func GenerateBasicTransaction(senderSeed string,
	version uint32,
	timestamp, fee int64,
	recipientAccountAddress string,
) *model.Transaction {
	senderAccountAddress := util.GetAddressFromSeed(senderSeed)
	if timestamp <= 0 {
		timestamp = time.Now().Unix()
	}
	return &model.Transaction{
		Version:                 version,
		Timestamp:               timestamp,
		SenderAccountAddress:    senderAccountAddress,
		RecipientAccountAddress: recipientAccountAddress,
		Fee:                     fee,
	}
}

func PrintTx(signedTxBytes []byte, outputType string) {
	var resultStr string
	switch outputType {
	case "hex":
		resultStr = hex.EncodeToString(signedTxBytes)
	default:
		var byteStrArr []string
		for _, bt := range signedTxBytes {
			byteStrArr = append(byteStrArr, fmt.Sprintf("%v", bt))
		}
		resultStr = strings.Join(byteStrArr, ", ")
	}
	fmt.Println(resultStr)
}

func GenerateSignedTxBytes(tx *model.Transaction, senderSeed string) []byte {
	unsignedTxBytes, _ := transaction.GetTransactionBytes(tx, false)
	tx.Signature = signature.Sign(
		unsignedTxBytes,
		constant.SignatureTypeDefault,
		senderSeed,
	)
	signedTxBytes, _ := transaction.GetTransactionBytes(tx, true)
	return signedTxBytes
}
