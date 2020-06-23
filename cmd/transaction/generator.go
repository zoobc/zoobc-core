package transaction

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"google.golang.org/grpc"

	"github.com/zoobc/zoobc-core/cmd/noderegistry"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	rpcService "github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
)

// GenerateTxSendMoney return send money transaction based on provided basic transaction & ammunt
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

/*
GenerateTxRegisterNode return register node transaction based on provided basic transaction &
others specific field for generate register node transaction
*/
func GenerateTxRegisterNode(
	tx *model.Transaction,
	nodeAddress string,
	lockedBalance int64,
	nodePubKey []byte,
	proofOfOwnerShip *model.ProofOfOwnership,
) *model.Transaction {

	txBody := &model.NodeRegistrationTransactionBody{
		NodePublicKey: nodePubKey,
		NodeAddress: &model.NodeAddress{
			Address: nodeAddress,
		},
		LockedBalance: lockedBalance,
		Poown:         proofOfOwnerShip,
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

/*
GenerateTxUpdateNode return update node transaction based on provided basic transaction &
others specific field for update register node transaction
*/
func GenerateTxUpdateNode(
	tx *model.Transaction,
	nodeAddress string,
	lockedBalance int64,
	nodePubKey []byte,
	proofOfOwnerShip *model.ProofOfOwnership,
) *model.Transaction {
	txBody := &model.UpdateNodeRegistrationTransactionBody{
		NodePublicKey: nodePubKey,
		NodeAddress: &model.NodeAddress{
			Address: nodeAddress,
		},
		LockedBalance: lockedBalance,
		Poown:         proofOfOwnerShip,
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

/*
GenerateTxRemoveNode return remove node transaction based on provided basic transaction &
others specific field for remove node transaction
*/
func GenerateTxRemoveNode(tx *model.Transaction, nodePubKey []byte) *model.Transaction {
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

/*
GenerateTxClaimNode return claim node transaction based on provided basic transaction &
others specific field for claim node transaction
*/
func GenerateTxClaimNode(
	tx *model.Transaction,
	nodePubKey []byte,
	proofOfOwnerShip *model.ProofOfOwnership,
) *model.Transaction {
	var (
		txBody = &model.ClaimNodeRegistrationTransactionBody{
			NodePublicKey: nodePubKey,
			Poown:         proofOfOwnerShip,
		}
		txBodyBytes = (&transaction.ClaimNodeRegistration{
			Body: txBody,
		}).GetBodyBytes()
	)
	tx.TransactionType = util.ConvertBytesToUint32(txTypeMap["claimNodeRegistration"])
	tx.TransactionBody = &model.Transaction_ClaimNodeRegistrationTransactionBody{
		ClaimNodeRegistrationTransactionBody: txBody,
	}
	tx.TransactionBodyBytes = txBodyBytes
	tx.TransactionBodyLength = uint32(len(txBodyBytes))
	return tx
}

// GenerateProofOfOwnership generate proof of owner ship for transaction related with node registry
func GenerateProofOfOwnership(
	dbPath, dbname, nodeOwnerAccountAddress, nodeSeed, proofOfOwnershipHex string,
) *model.ProofOfOwnership {
	if proofOfOwnershipHex != "" {
		powBytes, err := hex.DecodeString(proofOfOwnershipHex)
		if err != nil {
			panic(fmt.Sprintln("failed decode proofOfOwnershipHex, ", err.Error()))
		}
		pow, err := util.ParseProofOfOwnershipBytes(powBytes)
		if err != nil {
			panic(fmt.Sprintln("failed parse proofOfOwnership, ", err.Error()))
		}
		return pow
	}
	return noderegistry.GetProofOfOwnerShip(dbPath, dbname, nodeOwnerAccountAddress, nodeSeed)
}

/*
GenerateTxSetupAccountDataset return setup account dataset transaction based on provided basic transaction &
others specific field for setup account dataset transaction
*/
func GenerateTxSetupAccountDataset(
	tx *model.Transaction,
	property, value string,
) *model.Transaction {
	txBody := &model.SetupAccountDatasetTransactionBody{
		Property: property,
		Value:    value,
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

/*
GenerateTxRemoveAccountDataset return remove account dataset transaction based on provided basic transaction &
others specific field for remove account dataset transaction
*/
func GenerateTxRemoveAccountDataset(
	tx *model.Transaction,
	property, value string,
) *model.Transaction {
	txBody := &model.RemoveAccountDatasetTransactionBody{
		Property: property,
		Value:    value,
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

// GenerateBasicTransaction return  basic transaction based on common transaction field
func GenerateBasicTransaction(
	senderAddress, senderSeed string,
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

// PrintTx will print out the signed transaction based on provided format
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

		c := rpcService.NewTransactionServiceClient(conn)

		response, err := c.PostTransaction(context.Background(), &model.PostTransactionRequest{
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

// GenerateSignedTxBytes retrun signed transaction bytes
func GenerateSignedTxBytes(
	tx *model.Transaction,
	senderSeed string,
	signatureType int32,
	optionalSignParams ...interface{},
) []byte {
	var (
		transactionUtil = &transaction.Util{}
		txType          transaction.TypeAction
		err             error
	)
	txType, err = (&transaction.TypeSwitcher{}).GetTransactionType(tx)
	if err != nil {
		log.Fatalf("fail get transaction type: %s", err)
	}
	minimumFee, err := txType.GetMinimumFee()
	if err != nil {
		log.Fatalf("fail get minimum fee: %s", err)
	}
	tx.Fee += minimumFee

	unsignedTxBytes, _ := transactionUtil.GetTransactionBytes(tx, false)
	if senderSeed == "" {
		return unsignedTxBytes
	}
	tx.Signature, err = signature.Sign(
		unsignedTxBytes,
		model.SignatureType(signatureType),
		senderSeed,
		optionalSignParams...,
	)
	if err != nil {
		log.Fatalf("fail get sign tx: %s", err)
	}
	signedTxBytes, err := transactionUtil.GetTransactionBytes(tx, true)
	if err != nil {
		log.Fatalf("fail get get signed transactionBytes: %s", err)
	}
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

func GenerateTxRemoveNodeHDwallet(tx *model.Transaction, nodePubKey []byte) *model.Transaction {
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

/*
GenerateTxFeeVoteCommitment return fee vote commit vote transaction based on provided basic transaction &
others specific field for fee vote commit vote transaction
*/
func GenerateTxFeeVoteCommitment(
	tx *model.Transaction,
	voteHash []byte,
) *model.Transaction {
	var (
		txBody = &model.FeeVoteCommitTransactionBody{
			VoteHash: voteHash,
		}
		txBodyBytes = (&transaction.FeeVoteCommitTransaction{Body: txBody}).GetBodyBytes()
	)
	tx.TransactionType = util.ConvertBytesToUint32(txTypeMap["feeVoteCommit"])
	tx.TransactionBody = &model.Transaction_FeeVoteCommitTransactionBody{
		FeeVoteCommitTransactionBody: txBody,
	}
	tx.TransactionBodyBytes = txBodyBytes
	tx.TransactionBodyLength = uint32(len(txBodyBytes))
	return tx
}

func GenerateTxFeeVoteRevealPhase(tx *model.Transaction, voteInfo *model.FeeVoteInfo, voteInfoSigned []byte) *model.Transaction {

	var (
		txBody = &model.FeeVoteRevealTransactionBody{
			FeeVoteInfo:    voteInfo,
			VoterSignature: voteInfoSigned,
		}
		txBodyBytes = (&transaction.FeeVoteRevealTransaction{
			Body: txBody,
		}).GetBodyBytes()
	)
	tx.TransactionType = util.ConvertBytesToUint32(txTypeMap["feeVoteReveal"])
	tx.TransactionBody = &model.Transaction_FeeVoteRevealTransactionBody{
		FeeVoteRevealTransactionBody: txBody,
	}
	tx.TransactionBodyBytes = txBodyBytes
	tx.TransactionBodyLength = uint32(len(txBodyBytes))
	return tx
}

// GenerateTxLiquidPayment return liquid payment transaction based on provided basic transaction & ammunt
func GenerateTxLiquidPayment(tx *model.Transaction, sendAmount int64, completeMinutes uint64) *model.Transaction {
	txBody := &model.LiquidPaymentTransactionBody{
		Amount:          sendAmount,
		CompleteMinutes: completeMinutes,
	}
	tx.TransactionType = util.ConvertBytesToUint32(txTypeMap["liquidPayment"])
	tx.TransactionBody = &model.Transaction_LiquidPaymentTransactionBody{
		LiquidPaymentTransactionBody: txBody,
	}
	txBodyBytes := (&transaction.LiquidPaymentTransaction{
		Body: txBody,
	}).GetBodyBytes()
	tx.TransactionBodyBytes = txBodyBytes
	tx.TransactionBodyLength = uint32(len(txBodyBytes))
	return tx
}

// GenerateTxLiquidPaymentStop return liquid payment stop transaction based on provided basic transaction & ammunt
func GenerateTxLiquidPaymentStop(tx *model.Transaction, transactionID int64) *model.Transaction {
	txBody := &model.LiquidPaymentStopTransactionBody{
		TransactionID: transactionID,
	}
	tx.TransactionType = util.ConvertBytesToUint32(txTypeMap["liquidPaymentStop"])
	tx.TransactionBody = &model.Transaction_LiquidPaymentStopTransactionBody{
		LiquidPaymentStopTransactionBody: txBody,
	}
	txBodyBytes := (&transaction.LiquidPaymentStopTransaction{
		Body: txBody,
	}).GetBodyBytes()
	tx.TransactionBodyBytes = txBodyBytes
	tx.TransactionBodyLength = uint32(len(txBodyBytes))
	return tx
}
