package transaction

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/zoobc/zoobc-core/common/constant"

	"github.com/zoobc/zoobc-core/common/transaction"

	"github.com/zoobc/zoobc-core/common/crypto"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
)

var txTypeMap = map[string][]byte{
	"sendMoney":            {1, 0, 0, 0},
	"registerNode":         {2, 0, 0, 0},
	"setupAccountDataset":  {3, 0, 0, 0},
	"removeAccountDataset": {3, 1, 0, 0},
}

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
				seed := "prune filth cleaver removable earthworm tricky sulfur citation hesitate stout snort guy"

				tx := getTransaction(txTypeMap[txType])
				unsignedTxBytes, _ := util.GetTransactionBytes(tx, false)
				tx.Signature = signature.Sign(
					unsignedTxBytes,
					constant.NodeSignatureTypeDefault,
					seed,
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
	switch util.ConvertBytesToUint32(txType) {
	case util.ConvertBytesToUint32(txTypeMap["sendMoney"]):
		amount := int64(10000)
		return &model.Transaction{
			Version:                 1,
			TransactionType:         util.ConvertBytesToUint32(txTypeMap["sendMoney"]),
			Timestamp:               time.Now().Unix(),
			SenderAccountAddress:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			RecipientAccountAddress: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
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
		poowMessage := []byte("HelloBlock")
		signature := (&crypto.Signature{}).SignByNode(
			poowMessage,
			"prune filth cleaver removable earthworm tricky sulfur citation hesitate stout snort guy")
		txBody := &model.NodeRegistrationTransactionBody{
			AccountAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			NodePublicKey: []byte{
				0, 14, 6, 218, 170, 54, 60, 50, 2, 66, 130, 119, 226, 235, 126, 203, 5, 12, 152, 194, 170, 146, 43, 63, 224,
				101, 127, 241, 62, 152, 187, 255,
			},
			NodeAddress:   "127.0.0.1",
			LockedBalance: 100000,
			Poown: &model.ProofOfOwnership{
				MessageBytes: poowMessage,
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
			SenderAccountAddress:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
			RecipientAccountAddress: "BCZKLvgUYZ1KKx-jtF9KoJskjVPvB9jpIjfzzI6zDW0J",
			Fee:                     1,
			TransactionBodyLength:   uint32(len(txBodyBytes)),
			TransactionBody: &model.Transaction_NodeRegistrationTransactionBody{
				NodeRegistrationTransactionBody: txBody,
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
