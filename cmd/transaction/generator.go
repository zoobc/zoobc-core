package transaction

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	rpc_model "github.com/zoobc/zoobc-core/common/model"
	rpc_service "github.com/zoobc/zoobc-core/common/service"
	"google.golang.org/grpc"

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

	nodePubKey := crypto.NewEd25519Signature().GetPublicKeyFromSeed(nodeSeed)
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

	nodePubKey := crypto.NewEd25519Signature().GetPublicKeyFromSeed(nodeSeed)
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
	nodePubKey := crypto.NewEd25519Signature().GetPublicKeyFromSeed(nodeSeed)
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

	nodePubKey := crypto.NewEd25519Signature().GetPublicKeyFromSeed(nodeSeed)
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

func GenerateBasicTransaction(
	senderSeed string,
	senderSignatureType int32,
	version uint32,
	timestamp, fee int64,
	recipientAccountAddress string,
) *model.Transaction {
	var (
		senderAccountAddress string
	)
	if senderSeed == "" {
		senderAccountAddress = senderAddress
	} else {
		switch model.SignatureType(senderSignatureType) {
		case model.SignatureType_DefaultSignature:
			senderAccountAddress = crypto.NewEd25519Signature().GetAddressFromSeed(senderSeed)
		case model.SignatureType_BitcoinSignature:
			var (
				bitcoinSig  = crypto.NewBitcoinSignature(crypto.DefaultBitcoinNetworkParams(), crypto.DefaultBitcoinCurve())
				pubKey, err = bitcoinSig.GetPublicKeyFromSeed(
					senderSeed,
					crypto.DefaultBitcoinPublicKeyFormat(),
					crypto.DefaultBitcoinPrivateKeyLength(),
				)
			)
			if err != nil {
				panic(fmt.Sprintln(
					"GenerateBasicTransaction-BitcoinSignature-Failed GetPublicKey",
					err.Error(),
				))
			}
			senderAccountAddress, err = bitcoinSig.GetAddressFromPublicKey(pubKey)
			if err != nil {
				panic(fmt.Sprintln(
					"GenerateBasicTransaction-BitcoinSignature-Failed GetPublicKey",
					err.Error(),
				))
			}
		default:
			panic("GenerateBasicTransaction-Invalid Signature Type")
		}
	}

	if timestamp <= 0 {
		timestamp = time.Now().Unix()
	}
	return &model.Transaction{
		Version:                 version,
		Timestamp:               timestamp,
		SenderAccountAddress:    senderAccountAddress,
		RecipientAccountAddress: recipientAccountAddress,
		Fee:                     fee,
		Escrow: &model.Escrow{
			ApproverAddress: "",
			Commission:      0,
			Timeout:         0,
		},
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
	if post {
		conn, err := grpc.Dial(postHost, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %s", err)
		}
		defer conn.Close()

		c := rpc_service.NewTransactionServiceClient(conn)

		response, err := c.PostTransaction(context.Background(), &rpc_model.PostTransactionRequest{
			TransactionBytes: signedTxBytes,
		})
		if err != nil {
			fmt.Printf("post failed: %v\n", err)
		} else {
			fmt.Printf("\n\nresult: %v\n", response)
		}
	} else {
		fmt.Println(resultStr)
	}
}

func GenerateSignedTxBytes(tx *model.Transaction, senderSeed string, signatureType int32) []byte {
	var (
		transactionUtil = &transaction.Util{}
		txType          transaction.TypeAction
	)
	txType, _ = (&transaction.TypeSwitcher{}).GetTransactionType(tx)
	minimumFee, _ := txType.GetMinimumFee()
	tx.Fee += minimumFee

	unsignedTxBytes, _ := transactionUtil.GetTransactionBytes(tx, false)
	if senderSeed == "" {
		return unsignedTxBytes
	}
	tx.Signature, _ = signature.Sign(
		unsignedTxBytes,
		model.SignatureType(signatureType),
		senderSeed,
	)
	signedTxBytes, _ := transactionUtil.GetTransactionBytes(tx, true)
	return signedTxBytes
}

// GenerateEscrowApprovalTransaction set escrow approval body
func GenerateEscrowApprovalTransaction(tx *model.Transaction) *model.Transaction {

	var chosen model.EscrowApproval
	switch approval {
	case true:
		chosen = model.EscrowApproval_Approve
	default:
		chosen = model.EscrowApproval_Reject
	}

	txBody := &model.ApprovalEscrowTransactionBody{
		Approval:      chosen,
		TransactionID: transactionID,
	}
	txBodyBytes := (&transaction.ApprovalEscrowTransaction{
		Body: txBody,
	}).GetBodyBytes()

	tx.TransactionBody = txBody
	tx.TransactionBodyBytes = txBodyBytes
	tx.TransactionBodyLength = constant.EscrowApprovalBytesLength
	tx.TransactionType = util.ConvertBytesToUint32(txTypeMap["approvalEscrow"])

	return tx
}

/*
GenerateEscrowedTransaction inject escrow. Need:
		1. esApproverAddress
		2. Commission
		3. Timeout
Invalid escrow validation when those fields has not set
*/
func GenerateEscrowedTransaction(
	tx *model.Transaction,
) *model.Transaction {
	tx.Escrow = &model.Escrow{
		ApproverAddress: esApproverAddress,
		Commission:      esCommission,
		Timeout:         esTimeout,
		Instruction:     esInstruction,
	}
	return tx
}

/*
GeneratedMultiSignatureTransaction inject escrow. Need:
		1. unsignedTxHex
		2. signatures
		3. multisigInfo:
			- minSignature
			- nonce
			- addresses
Invalid escrow validation when those fields has not set
*/
func GeneratedMultiSignatureTransaction(
	tx *model.Transaction,
	minSignature uint32,
	nonce int64,
	unsignedTxHex, txHash string,
	addressSignatures map[string]string, addresses []string,
) *model.Transaction {
	var (
		signatures    = make(map[string][]byte)
		signatureInfo *model.SignatureInfo
		unsignedTx    []byte
		multiSigInfo  *model.MultiSignatureInfo
		err           error
	)
	if minSignature > 0 && len(addresses) > 0 {
		multiSigInfo = &model.MultiSignatureInfo{
			MinimumSignatures: minSignature,
			Nonce:             nonce,
			Addresses:         addresses,
		}
	}
	if unsignedTxHex != "" {
		unsignedTx, err = hex.DecodeString(unsignedTxHex)
		if err != nil {
			return nil
		}
	}
	if txHash != "" {
		transactionHash, err := hex.DecodeString(txHash)
		if err != nil {
			return nil
		}
		for k, v := range addressSignatures {
			if v == "" {
				sigType := util.ConvertUint32ToBytes(2)
				signatures[k] = sigType
			} else {
				signature, err := hex.DecodeString(v)
				if err != nil {
					return nil
				}
				signatures[k] = signature
			}
		}
		fmt.Printf("signatures: %v\n\n\n", signatures)
		signatureInfo = &model.SignatureInfo{
			TransactionHash: transactionHash,
			Signatures:      signatures,
		}
	}
	tx.TransactionType = util.ConvertBytesToUint32(txTypeMap["multiSignature"])
	txBody := &model.MultiSignatureTransactionBody{
		MultiSignatureInfo:       multiSigInfo,
		UnsignedTransactionBytes: unsignedTx,
		SignatureInfo:            signatureInfo,
	}
	tx.TransactionBodyBytes = (&transaction.MultiSignatureTransaction{
		Body: txBody,
	}).GetBodyBytes()
	fmt.Printf("length: %v\n", len(tx.TransactionBodyBytes))
	tx.TransactionBodyLength = uint32(len(tx.TransactionBodyBytes))
	return tx
}
