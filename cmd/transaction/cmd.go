package transaction

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
)

var (
	mockPoowBlockHash = []byte{209, 64, 140, 231, 150, 96, 104, 137, 202, 190, 83, 202, 22, 67, 222,
		38, 48, 40, 213, 202, 144, 30, 73, 184, 186, 188, 240, 209, 252, 222, 132, 36}
	signature = &crypto.Signature{}

	// Basic transactions data
	outputType              string
	version                 uint32
	timestamp               int64
	senderSeed              string
	recipientAccountAddress string
	fee                     int64

	// Send money transaction
	sendAmount int64

	// node registration transactions
	nodeSeed                string
	nodeOwnerAccountAddress string
	nodeAddress             string
	lockedBalance           int64

	// dataset transactions
	property   string
	value      string
	activeTime uint64

	txCmd = &cobra.Command{
		Use:   "transaction",
		Short: "transaction command used to generate transaction.",
	}

	txSendMoneyCmd = &cobra.Command{
		Use:   "send-money",
		Short: "send-money command used to generate \"send money\" transaction",
		Run: func(ccmd *cobra.Command, args []string) {
			tx := GenerateBasicTransaction(senderSeed)
			tx = GenerateTxSendMoney(tx, sendAmount)
			result := PrintTx(GenerateSignedTxBytes(tx, senderSeed), outputType)
			fmt.Println(result)
		},
	}

	registerNodeCmd = &cobra.Command{
		Use:   "register-node",
		Short: "send-money command used to generate \"send money\" transaction",
		Run: func(ccmd *cobra.Command, args []string) {
			tx := GenerateBasicTransaction(senderSeed)
			tx = GenerateTxRegisterNode(tx, nodeOwnerAccountAddress, nodeSeed, recipientAccountAddress, nodeAddress, lockedBalance)
			result := PrintTx(GenerateSignedTxBytes(tx, senderSeed), outputType)
			fmt.Println(result)
		},
	}

	updateNodeCmd = &cobra.Command{
		Use:   "update-node",
		Short: "update-node command used to generate \"update node\" transaction",
		Run: func(ccmd *cobra.Command, args []string) {
			tx := GenerateBasicTransaction(senderSeed)
			tx = GenerateTxUpdateNode(tx, nodeOwnerAccountAddress, nodeSeed, nodeAddress, lockedBalance)
			result := PrintTx(GenerateSignedTxBytes(tx, senderSeed), outputType)
			fmt.Println(result)
		},
	}

	removeNodeCmd = &cobra.Command{
		Use:   "remove-node",
		Short: "remove-node command used to generate \"remove node\" transaction",
		Run: func(ccmd *cobra.Command, args []string) {
			tx := GenerateBasicTransaction(senderSeed)
			tx = GenerateTxRemoveNode(tx, nodeSeed)
			result := PrintTx(GenerateSignedTxBytes(tx, senderSeed), outputType)
			fmt.Println(result)
		},
	}

	claimNodeCmd = &cobra.Command{
		Use:   "claim-node",
		Short: "claim-node command used to generate \"claim node\" transaction",
		Run: func(ccmd *cobra.Command, args []string) {
			tx := GenerateBasicTransaction(senderSeed)
			tx = GenerateTxClaimNode(tx, nodeOwnerAccountAddress, nodeSeed, recipientAccountAddress)
			result := PrintTx(GenerateSignedTxBytes(tx, senderSeed), outputType)
			fmt.Println(result)
		},
	}

	setupAccountDatasetCmd = &cobra.Command{
		Use:   "set-account-dataset",
		Short: "set-account-dataset command used to generate \"set account dataset\" transaction",
		Run: func(ccmd *cobra.Command, args []string) {
			senderAccountAddress := util.GetAddressFromSeed(senderSeed)
			tx := GenerateBasicTransaction(senderSeed)
			tx = GenerateTxSetupAccountDataset(tx, senderAccountAddress, recipientAccountAddress, property, value, activeTime)
			result := PrintTx(GenerateSignedTxBytes(tx, senderSeed), outputType)
			fmt.Println(result)
		},
	}

	removeAccountDatasetCmd = &cobra.Command{
		Use:   "remove-account-dataset",
		Short: "remove-account-dataset command used to generate \"remove account dataset\" transaction",
		Run: func(ccmd *cobra.Command, args []string) {
			senderAccountAddress := util.GetAddressFromSeed(senderSeed)
			tx := GenerateBasicTransaction(senderSeed)
			tx = GenerateTxRemoveAccountDataset(tx, senderAccountAddress, recipientAccountAddress, property, value)
			result := PrintTx(GenerateSignedTxBytes(tx, senderSeed), outputType)
			fmt.Println(result)
		},
	}
)

func init() {
	txCmd.PersistentFlags().StringVar(&outputType, "output", "bytes", "defines the type of the output to be generated [\"bytes\", \"hex\"]")
	txCmd.PersistentFlags().Uint32Var(&version, "version", 1, "defines version of the transaction")
	txCmd.PersistentFlags().Int64Var(&timestamp, "timestamp", time.Now().Unix(), "defines timestamp of the transaction")
	txCmd.PersistentFlags().StringVar(&senderSeed, "sender-seed", "",
		"defines the sender seed that's used to sign transaction and whose public key will be used in the"+
			"`Sender Account Address` field of the transaction")
	txCmd.PersistentFlags().StringVar(&recipientAccountAddress, "recipient", "", "defines the recipient intended for the transaction")
	txCmd.PersistentFlags().Int64Var(&fee, "fee", 1, "defines the fee of the transaction")

	txSendMoneyCmd.Flags().Int64Var(&sendAmount, "amount", 0, "Amount of money we want to send")
	txCmd.AddCommand(txSendMoneyCmd)

	registerNodeCmd.Flags().StringVar(&nodeOwnerAccountAddress, "node-owner-account-address", "", "Account address of the owner of the node")
	registerNodeCmd.Flags().StringVar(&nodeSeed, "node-seed", "", "Private key of the node")
	registerNodeCmd.Flags().StringVar(&nodeAddress, "node-address", "", "(ip) Address of the node")
	registerNodeCmd.Flags().Int64Var(&lockedBalance, "locked-balance", 0, "Amount of money wanted to be locked")
	txCmd.AddCommand(registerNodeCmd)

	updateNodeCmd.Flags().StringVar(&nodeOwnerAccountAddress, "node-owner-account-address", "", "Account address of the owner of the node")
	updateNodeCmd.Flags().StringVar(&nodeSeed, "node-seed", "", "Private key of the node")
	updateNodeCmd.Flags().StringVar(&nodeAddress, "node-address", "", "(ip) Address of the node")
	updateNodeCmd.Flags().Int64Var(&lockedBalance, "locked-balance", 0, "Amount of money wanted to be locked")
	txCmd.AddCommand(updateNodeCmd)

	removeNodeCmd.Flags().StringVar(&nodeSeed, "node-seed", "", "Private key of the node")
	txCmd.AddCommand(removeNodeCmd)

	claimNodeCmd.Flags().StringVar(&nodeOwnerAccountAddress, "node-owner-account-address", "", "Account address of the owner of the node")
	claimNodeCmd.Flags().StringVar(&nodeSeed, "node-seed", "", "Private key of the node")
	txCmd.AddCommand(claimNodeCmd)

	setupAccountDatasetCmd.Flags().StringVar(&property, "property", "", "Property of dataset wanted to be set")
	setupAccountDatasetCmd.Flags().StringVar(&value, "value", "", "Value of dataset wanted to be set")
	// 2592000 = 30 days
	setupAccountDatasetCmd.Flags().Uint64Var(&activeTime, "active-time", 2592000, "Active Time of dataset wanted to be set")
	txCmd.AddCommand(setupAccountDatasetCmd)

	removeAccountDatasetCmd.Flags().StringVar(&property, "property", "", "Property of dataset wanted to be removed")
	removeAccountDatasetCmd.Flags().StringVar(&value, "value", "", "Value of dataset wanted to be removed")
	txCmd.AddCommand(removeAccountDatasetCmd)
}

func Commands() *cobra.Command {
	return txCmd
}

func GenerateBasicTransaction(senderSeed string) *model.Transaction {
	senderAccountAddress := util.GetAddressFromSeed(senderSeed)
	return &model.Transaction{
		Version:                 version,
		Timestamp:               timestamp,
		SenderAccountAddress:    senderAccountAddress,
		RecipientAccountAddress: recipientAccountAddress,
		Fee:                     fee,
	}
}

func PrintTx(signedTxBytes []byte, outputType string) string {
	switch outputType {
	case "hex":
		return hex.EncodeToString(signedTxBytes)
	default:
		var byteStrArr []string
		for _, bt := range signedTxBytes {
			byteStrArr = append(byteStrArr, fmt.Sprintf("%v", bt))
		}
		return strings.Join(byteStrArr, ", ")
	}
}

func GenerateSignedTxBytes(tx *model.Transaction, senderSeed string) []byte {
	unsignedTxBytes, _ := util.GetTransactionBytes(tx, false)
	tx.Signature = signature.Sign(
		unsignedTxBytes,
		constant.SignatureTypeDefault,
		senderSeed,
	)
	signedTxBytes, _ := util.GetTransactionBytes(tx, true)
	return signedTxBytes
}

func GenerateMockPoowMessage(ownerAccountAddress string) *model.ProofOfOwnershipMessage {
	return &model.ProofOfOwnershipMessage{
		AccountAddress: ownerAccountAddress,
		BlockHash:      mockPoowBlockHash,
		BlockHeight:    1,
	}
}
