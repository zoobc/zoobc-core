// package main

// import (
// 	"bytes"
// 	"fmt"

// 	"github.com/zoobc/zoobc-core/common/accounttype"
// 	"github.com/zoobc/zoobc-core/common/blocker"
// 	"github.com/zoobc/zoobc-core/common/constant"
// 	"github.com/zoobc/zoobc-core/common/model"
// 	"github.com/zoobc/zoobc-core/common/util"
// )

// func main() {
// 	_, err := test()
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// }

// func test() (*model.Transaction, error) {
// 	var (
// 		chunkedBytes []byte
// 		transaction  model.Transaction
// 		buffer       *bytes.Buffer
// 		escrow       model.Escrow
// 		err          error
// 	)

// 	transactionBytes := []byte{1, 0, 0, 0,
// 		1, 161, 175, 52, 96,
// 		0, 0, 0, 0,
// 		0, 0, 0, 0, 47, 113, 76, 14, 227, 177, 70, 153, 2, 189, 86, 241, 253, 68, 76, 50, 224, 253, 234, 232, 128, 235, 119, 24, 195, 79, 196, 105, 74, 74, 237, 85,
// 		0, 0, 0, 0, 47, 133, 28, 189, 198, 23, 52, 15, 179, 127, 180, 87, 162, 84, 53, 80, 58, 245, 4, 200, 144, 77, 92, 157, 135, 170, 11, 159, 137, 144, 171, 28,
// 		32, 214, 19, 0, 0, 0, 0, 0,
// 		8, 0, 0, 0, 0, 202, 154, 59, 0, 0, 0, 0,
// 		2, 0, 0, 0, 3, 0, 0, 0, 111, 107, 101,
// 		36, 94, 186, 89, 23, 14, 26, 232, 233, 163, 71, 103, 84, 80, 42, 86, 214, 196, 13, 244, 195, 194, 64, 23, 54, 250, 142, 3, 198, 86, 128, 180, 234, 49, 87, 0, 182, 5, 231, 43, 203, 125, 20, 82, 246, 91, 12, 254, 149, 207, 238, 50, 238, 25, 134, 65, 146, 153, 139, 232, 190, 83, 54, 5}
// 	buffer = bytes.NewBuffer(transactionBytes)

// 	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.TransactionType))
// 	if err != nil {
// 		return nil, err
// 	}
// 	transaction.TransactionType = util.ConvertBytesToUint32(chunkedBytes)

// 	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.TransactionVersion))
// 	if err != nil {
// 		return nil, err
// 	}
// 	transaction.Version = uint32(chunkedBytes[0])

// 	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.Timestamp))
// 	if err != nil {
// 		return nil, err
// 	}
// 	transaction.Timestamp = int64(util.ConvertBytesToUint64(chunkedBytes))

// 	senderAccType, err := accounttype.ParseBytesToAccountType(buffer)
// 	if err != nil {
// 		return nil, err
// 	}
// 	senderAddress, err := senderAccType.GetAccountAddress()
// 	if err != nil {
// 		return nil, err
// 	}
// 	transaction.SenderAccountAddress = senderAddress

// 	recipientAccType, err := accounttype.ParseBytesToAccountType(buffer)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if recipientAccType.GetTypeInt() != int32(model.AccountType_EmptyAccountType) {
// 		transaction.RecipientAccountAddress, err = recipientAccType.GetAccountAddress()
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.Fee))
// 	if err != nil {
// 		return nil, err
// 	}
// 	transaction.Fee = int64(util.ConvertBytesToUint64(chunkedBytes))

// 	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.TransactionBodyLength))
// 	if err != nil {
// 		return nil, err
// 	}
// 	transaction.TransactionBodyLength = util.ConvertBytesToUint32(chunkedBytes)
// 	transaction.TransactionBodyBytes, err = util.ReadTransactionBytes(buffer, int(transaction.TransactionBodyLength))
// 	if err != nil {
// 		return nil, err
// 	}
// 	/***
// 	Escrow part
// 	1. ApproverAddress
// 	2. Commission
// 	3. Timeout
// 	4. Instruction
// 	*/
// 	approverAccType, err := accounttype.ParseBytesToAccountType(buffer)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// if approver account is empty (== empty account type), then skip the escrow part
// 	if approverAccType.GetTypeInt() != int32(model.AccountType_EmptyAccountType) {
// 		escrow.ApproverAddress, err = approverAccType.GetAccountAddress()
// 		if err != nil {
// 			return nil, err
// 		}

// 		chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.EscrowCommissionLength))
// 		if err != nil {
// 			return nil, err
// 		}
// 		escrow.Commission = int64(util.ConvertBytesToUint64(chunkedBytes))

// 		chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.EscrowTimeoutLength))
// 		if err != nil {
// 			return nil, err
// 		}
// 		escrow.Timeout = int64(util.ConvertBytesToUint64(chunkedBytes))

// 		chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.EscrowInstructionLength))
// 		if err != nil {
// 			return nil, err
// 		}
// 		instructionLength := int(util.ConvertBytesToUint32(chunkedBytes))
// 		instruction, err := util.ReadTransactionBytes(buffer, instructionLength)
// 		if err != nil {
// 			return nil, err
// 		}
// 		escrow.Instruction = string(instruction)

// 		transaction.Escrow = &escrow
// 	}

// 	chunkedBytes, err = util.ReadTransactionBytes(buffer, int(constant.TxMessageBytesLength))
// 	if err != nil {
// 		return nil, err
// 	}
// 	messageLength := int(util.ConvertBytesToUint32(chunkedBytes))
// 	fmt.Println("messageLength", messageLength)
// 	if messageLength > 0 {
// 		messageBytes, err := util.ReadTransactionBytes(buffer, messageLength)
// 		if err != nil {
// 			return nil, err
// 		}
// 		transaction.Message = messageBytes
// 	}
// 	fmt.Println("transaction.Fee", transaction.Fee)
// 	fmt.Println("transaction.Amount", transaction.TransactionBodyBytes)
// 	fmt.Println("transaction.Message", transaction.Message)

// 	signatureLength := senderAccType.GetSignatureLength()
// 	transaction.Signature, err = util.ReadTransactionBytes(buffer, int(signatureLength))
// 	if err != nil {
// 		return nil, blocker.NewBlocker(
// 			blocker.ParserErr,
// 			"no transaction signature",
// 		)
// 	}
// 	return nil, nil
// }

package main

import (
	"encoding/hex"
)

func main() {
	// var publickey = make([]byte, 32)
	// a := &signaturetype.Ed25519Signature{}
	// publickey, _ = a.GetPublicKeyFromEncodedAddress("ZNK_GX3GA5PJ_6VWWLVQR_UJE3YT4G_SZZGBKQA_4TWVOVMM_VFDLZSWR_ECQDCMZA")
	// address.DecodeZbcID("ZNK_GX3GA5PJ_6VWWLVQR_UJE3YT4G_SZZGBKQA_4TWVOVMM_VFDLZSWR_ECQDCMZA", publickey)
	println(hex.EncodeToString([]byte{224, 87, 180, 194, 102, 108, 253, 212, 24, 59, 218, 203, 2, 224, 127, 160, 17, 14, 222, 224, 45, 201, 63, 23, 148, 90, 95, 103, 181, 176, 188, 160}))
}
