package transaction

import (
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/zoobc/zoobc-core/common/constant"

	"github.com/zoobc/zoobc-core/common/transaction"

	"github.com/zoobc/zoobc-core/common/crypto"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/common/model"
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
	// Core node test account in genesis block
	senderAccountSeed = constant.MainchainGenesisFundReceivers[0].AccountSeed
)

func GenerateTransactionBytes(logger *logrus.Logger,
	signature crypto.SignatureInterface) *cobra.Command {
	var (
		txType string
	)
	var txCmd = &cobra.Command{
		Use:   "tx",
		Short: "tx command used to generate transaction.",
		Long: `tx command generate signed transaction bytes in form of hex or []bytes
		`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if args[0] == "generate" {
				tx := getTransaction(txTypeMap[txType])
				unsignedTxBytes, _ := util.GetTransactionBytes(tx, false)
				tx.Signature = signature.Sign(
					unsignedTxBytes,
					constant.SignatureTypeDefault,
					senderAccountSeed,
				)
				signedTxBytes, _ := util.GetTransactionBytes(tx, true)
				var signedTxByteString string
				for _, b := range signedTxBytes {
					signedTxByteString += fmt.Sprintf("%v, ", b)
				}
				logger.Printf("tx-bytes:byte = %v", signedTxByteString)
				logger.Printf("tx-bytes:hex = %s", hex.EncodeToString(signedTxBytes))
			} else {
				logger.Error("unknown command")
			}
		},
	}
	txCmd.Flags().StringVarP(&txType, "type", "t", "sendMoney", "number of account to generate")
	return txCmd
}

func getTransaction(txType []byte) *model.Transaction {
	nodeSeed := "sprinkled sneak species pork outpost thrift unwind cheesy vexingly dizzy neurology neatness"
	nodePubKey := util.GetPublicKeyFromSeed(nodeSeed)
	senderAccountAddress := util.GetAddressFromSeed(senderAccountSeed)
	log.Printf("%s", senderAccountAddress)
	recipientAccountSeed := "witch collapse practice feed shame open despair creek road again ice least"
	recipientAccountAddress := util.GetAddressFromSeed(recipientAccountSeed)
	account2Seed := "rocket inflict forest able frozen expose mail exit early permit dial retire"
	account2Address := util.GetAddressFromSeed(account2Seed)
	switch util.ConvertBytesToUint32(txType) {
	case util.ConvertBytesToUint32(txTypeMap["sendMoney"]):
		amount := 50 * constant.OneZBC
		return &model.Transaction{
			Version:                 1,
			TransactionType:         util.ConvertBytesToUint32(txTypeMap["sendMoney"]),
			Timestamp:               time.Now().Unix(),
			SenderAccountAddress:    senderAccountAddress,
			RecipientAccountAddress: recipientAccountAddress,
			Fee:                     1,
			TransactionBodyLength:   8,
			TransactionBody: &model.Transaction_SendMoneyTransactionBody{
				SendMoneyTransactionBody: &model.SendMoneyTransactionBody{
					Amount: amount,
				},
			},
			TransactionBodyBytes: util.ConvertUint64ToBytes(uint64(amount)),
		}
	case util.ConvertBytesToUint32(txTypeMap["registerNode"]):
		poowMessage := &model.ProofOfOwnershipMessage{
			AccountAddress: recipientAccountAddress,
			BlockHash: []byte{209, 64, 140, 231, 150, 96, 104, 137, 202, 190, 83, 202, 22, 67, 222,
				38, 48, 40, 213, 202, 144, 30, 73, 184, 186, 188, 240, 209, 252, 222, 132, 36},
			BlockHeight: 1,
		}
		poownMessageBytes := util.GetProofOfOwnershipMessageBytes(poowMessage)
		signature := (&crypto.Signature{}).SignByNode(
			poownMessageBytes,
			nodeSeed)
		txBody := &model.NodeRegistrationTransactionBody{
			AccountAddress: recipientAccountAddress,
			NodePublicKey:  nodePubKey,
			NodeAddress: &model.NodeAddress{
				Address: "127.0.0.1",
			},
			LockedBalance: 10 * constant.OneZBC,
			Poown: &model.ProofOfOwnership{
				MessageBytes: poownMessageBytes,
				Signature:    signature,
			},
		}
		txBodyBytes := (&transaction.NodeRegistration{
			Body: txBody,
		}).GetBodyBytes()
		return &model.Transaction{
			Version:                 1,
			TransactionType:         util.ConvertBytesToUint32(txTypeMap["registerNode"]),
			Timestamp:               time.Now().Unix(),
			SenderAccountAddress:    senderAccountAddress,
			RecipientAccountAddress: senderAccountAddress,
			Fee:                     1,
			TransactionBodyLength:   uint32(len(txBodyBytes)),
			TransactionBody: &model.Transaction_NodeRegistrationTransactionBody{
				NodeRegistrationTransactionBody: txBody,
			},
			TransactionBodyBytes: txBodyBytes,
		}
	case util.ConvertBytesToUint32(txTypeMap["updateNodeRegistration"]):
		poowMessage := &model.ProofOfOwnershipMessage{
			AccountAddress: recipientAccountAddress,
			BlockHash: []byte{209, 64, 140, 231, 150, 96, 104, 137, 202, 190, 83, 202, 22, 67, 222,
				38, 48, 40, 213, 202, 144, 30, 73, 184, 186, 188, 240, 209, 252, 222, 132, 36},
			BlockHeight: 1,
		}
		poownMessageBytes := util.GetProofOfOwnershipMessageBytes(poowMessage)
		signature := (&crypto.Signature{}).SignByNode(
			poownMessageBytes,
			nodeSeed)
		txBody := &model.UpdateNodeRegistrationTransactionBody{
			NodePublicKey: nodePubKey,
			NodeAddress: &model.NodeAddress{
				Address: "127.0.0.1",
			},
			LockedBalance: 10050000000000,
			Poown: &model.ProofOfOwnership{
				MessageBytes: poownMessageBytes,
				Signature:    signature,
			},
		}
		txBodyBytes := (&transaction.UpdateNodeRegistration{
			Body: txBody,
		}).GetBodyBytes()
		return &model.Transaction{
			Version:                 1,
			TransactionType:         util.ConvertBytesToUint32(txTypeMap["updateNodeRegistration"]),
			Timestamp:               time.Now().Unix(),
			SenderAccountAddress:    senderAccountAddress,
			RecipientAccountAddress: senderAccountAddress,
			Fee:                     1,
			TransactionBodyLength:   uint32(len(txBodyBytes)),
			TransactionBody: &model.Transaction_UpdateNodeRegistrationTransactionBody{
				UpdateNodeRegistrationTransactionBody: txBody,
			},
			TransactionBodyBytes: txBodyBytes,
		}
	case util.ConvertBytesToUint32(txTypeMap["claimNodeRegistration"]):
		poowMessage := &model.ProofOfOwnershipMessage{
			AccountAddress: recipientAccountAddress,
			BlockHash: []byte{209, 64, 140, 231, 150, 96, 104, 137, 202, 190, 83, 202, 22, 67, 222,
				38, 48, 40, 213, 202, 144, 30, 73, 184, 186, 188, 240, 209, 252, 222, 132, 36},
			BlockHeight: 1,
		}
		poownMessageBytes := util.GetProofOfOwnershipMessageBytes(poowMessage)
		signature := (&crypto.Signature{}).SignByNode(
			poownMessageBytes,
			nodeSeed)
		txBody := &model.ClaimNodeRegistrationTransactionBody{
			AccountAddress: account2Address,
			NodePublicKey:  nodePubKey,
			Poown: &model.ProofOfOwnership{
				MessageBytes: poownMessageBytes,
				Signature:    signature,
			},
		}
		txBodyBytes := (&transaction.ClaimNodeRegistration{
			Body: txBody,
		}).GetBodyBytes()
		return &model.Transaction{
			Version:                 1,
			TransactionType:         util.ConvertBytesToUint32(txTypeMap["claimNodeRegistration"]),
			Timestamp:               time.Now().Unix(),
			SenderAccountAddress:    senderAccountAddress,
			RecipientAccountAddress: senderAccountAddress,
			Fee:                     1,
			TransactionBodyLength:   uint32(len(txBodyBytes)),
			TransactionBody: &model.Transaction_ClaimNodeRegistrationTransactionBody{
				ClaimNodeRegistrationTransactionBody: txBody,
			},
			TransactionBodyBytes: txBodyBytes,
		}
	case util.ConvertBytesToUint32(txTypeMap["removeNodeRegistration"]):
		txBody := &model.RemoveNodeRegistrationTransactionBody{
			NodePublicKey: nodePubKey,
		}
		txBodyBytes := (&transaction.RemoveNodeRegistration{
			Body: txBody,
		}).GetBodyBytes()
		return &model.Transaction{
			Version:                 1,
			TransactionType:         util.ConvertBytesToUint32(txTypeMap["removeNodeRegistration"]),
			Timestamp:               time.Now().Unix(),
			SenderAccountAddress:    senderAccountAddress,
			RecipientAccountAddress: senderAccountAddress,
			Fee:                     1,
			TransactionBodyLength:   uint32(len(txBodyBytes)),
			TransactionBody: &model.Transaction_RemoveNodeRegistrationTransactionBody{
				RemoveNodeRegistrationTransactionBody: txBody,
			},
			TransactionBodyBytes: txBodyBytes,
		}
	case util.ConvertBytesToUint32(txTypeMap["setupAccountDataset"]):
		txBody := &model.SetupAccountDatasetTransactionBody{
			SetterAccountAddress:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			RecipientAccountAddress: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
			Property:                "Member",
			Value:                   "Welcome to the jungle",
			MuchTime:                2592000, // 30 days in second
		}
		txBodyBytes := (&transaction.SetupAccountDataset{
			Body: txBody,
		}).GetBodyBytes()
		return &model.Transaction{
			Version:                 1,
			TransactionType:         util.ConvertBytesToUint32(txTypeMap["setupAccountDataset"]),
			Timestamp:               time.Now().Unix(),
			SenderAccountAddress:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			RecipientAccountAddress: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
			Fee:                     1,
			TransactionBodyLength:   uint32(len(txBodyBytes)),
			TransactionBody: &model.Transaction_SetupAccountDatasetTransactionBody{
				SetupAccountDatasetTransactionBody: txBody,
			},
			TransactionBodyBytes: txBodyBytes,
		}
	case util.ConvertBytesToUint32(txTypeMap["removeAccountDataset"]):
		txBody := &model.RemoveAccountDatasetTransactionBody{
			SetterAccountAddress:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			RecipientAccountAddress: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
			Property:                "Member",
			Value:                   "Good bye",
		}
		txBodyBytes := (&transaction.RemoveAccountDataset{
			Body: txBody,
		}).GetBodyBytes()
		return &model.Transaction{
			Version:                 1,
			TransactionType:         util.ConvertBytesToUint32(txTypeMap["removeAccountDataset"]),
			Timestamp:               time.Now().Unix(),
			SenderAccountAddress:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			RecipientAccountAddress: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
			Fee:                     1,
			TransactionBodyLength:   uint32(len(txBodyBytes)),
			TransactionBody: &model.Transaction_RemoveAccountDatasetTransactionBody{
				RemoveAccountDatasetTransactionBody: txBody,
			},
			TransactionBodyBytes: txBodyBytes,
		}
	}
	return nil
}
