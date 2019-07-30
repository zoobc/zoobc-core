package transaction

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/zoobc/zoobc-core/common/crypto"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
)

func GenerateTransactionBytes(logger *logrus.Logger,
	signature crypto.SignatureInterface) *cobra.Command {
	var txCmd = &cobra.Command{
		Use:   "tx",
		Short: "tx command used to generate transaction.",
		Long: `tx command generate signed transaction bytes in form of hex or []bytes
		`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if args[0] == "generate" {
				amount := int64(10000)
				seed := "prune filth cleaver removable earthworm tricky sulfur citation hesitate stout snort guy"

				tx := &model.Transaction{
					Version:                 1,
					TransactionType:         util.ConvertBytesToUint32([]byte{1, 0, 0, 0}),
					Timestamp:               time.Now().Unix(),
					SenderAccountType:       0,
					SenderAccountAddress:    "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
					RecipientAccountType:    0,
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
				unsignedTxBytes, _ := util.GetTransactionBytes(tx, false)
				tx.Signature = signature.Sign(
					unsignedTxBytes,
					tx.SenderAccountType,
					tx.SenderAccountAddress,
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
	return txCmd
}
