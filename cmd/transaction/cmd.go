package transaction

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	TXGeneratorCommands struct {
		DB *sql.DB
	}
	RunCommand func(ccmd *cobra.Command, args []string)
)

var (
	txGeneratorCommandsInstance *TXGeneratorCommands
	txCmd                       = &cobra.Command{
		Use:   "transaction",
		Short: "transaction command used to generate transaction.",
	}
	sendMoneyCmd = &cobra.Command{
		Use:   "send-money",
		Short: "send-money command used to generate \"send money\" transaction",
	}
	registerNodeCmd = &cobra.Command{
		Use:   "register-node",
		Short: "send-money command used to generate \"send money\" transaction",
	}
	updateNodeCmd = &cobra.Command{
		Use:   "update-node",
		Short: "update-node command used to generate \"update node\" transaction",
	}
	removeNodeCmd = &cobra.Command{
		Use:   "remove-node",
		Short: "remove-node command used to generate \"remove node\" transaction",
	}
	claimNodeCmd = &cobra.Command{
		Use:   "claim-node",
		Short: "claim-node command used to generate \"claim node\" transaction",
	}
	setupAccountDatasetCmd = &cobra.Command{
		Use:   "set-account-dataset",
		Short: "set-account-dataset command used to generate \"set account dataset\" transaction",
	}
	removeAccountDatasetCmd = &cobra.Command{
		Use:   "remove-account-dataset",
		Short: "remove-account-dataset command used to generate \"remove account dataset\" transaction",
	}
)

func init() {
	/*
		TXCommandRoot
	*/
	txCmd.PersistentFlags().StringVar(&outputType, "output", "bytes", "defines the type of the output to be generated [\"bytes\", \"hex\"]")
	txCmd.PersistentFlags().Uint32Var(&version, "version", 1, "defines version of the transaction")
	txCmd.PersistentFlags().Int64Var(&timestamp, "timestamp", time.Now().Unix(), "defines timestamp of the transaction")
	txCmd.PersistentFlags().StringVar(&senderSeed, "sender-seed", "",
		"defines the sender seed that's used to sign transaction and whose public key will be used in the"+
			"`Sender Account Address` field of the transaction")
	txCmd.PersistentFlags().StringVar(&recipientAccountAddress, "recipient", "", "defines the recipient intended for the transaction")
	txCmd.PersistentFlags().Int64Var(&fee, "fee", 1, "defines the fee of the transaction")

	/*
		SendMoney Command
	*/
	sendMoneyCmd.Flags().Int64Var(&sendAmount, "amount", 0, "Amount of money we want to send")

	/*
		RegisterNode Command
	*/
	registerNodeCmd.Flags().StringVar(&nodeOwnerAccountAddress, "node-owner-account-address", "", "Account address of the owner of the node")
	registerNodeCmd.Flags().StringVar(&nodeSeed, "node-seed", "", "Private key of the node")
	registerNodeCmd.Flags().StringVar(&nodeAddress, "node-address", "", "(ip) Address of the node")
	registerNodeCmd.Flags().Int64Var(&lockedBalance, "locked-balance", 0, "Amount of money wanted to be locked")

	/*
		UpdateNode Command
	*/
	updateNodeCmd.Flags().StringVar(&nodeOwnerAccountAddress, "node-owner-account-address", "", "Account address of the owner of the node")
	updateNodeCmd.Flags().StringVar(&nodeSeed, "node-seed", "", "Private key of the node")
	updateNodeCmd.Flags().StringVar(&nodeAddress, "node-address", "", "(ip) Address of the node")
	updateNodeCmd.Flags().Int64Var(&lockedBalance, "locked-balance", 0, "Amount of money wanted to be locked")

	/*
		RemoveNode Command
	*/
	removeNodeCmd.Flags().StringVar(&nodeSeed, "node-seed", "", "Private key of the node")

	/*
		ClaimNode Command
	*/
	claimNodeCmd.Flags().StringVar(&nodeOwnerAccountAddress, "node-owner-account-address", "", "Account address of the owner of the node")
	claimNodeCmd.Flags().StringVar(&nodeSeed, "node-seed", "", "Private key of the node")

	/*
		SetupAccountDataset Command
	*/
	setupAccountDatasetCmd.Flags().StringVar(&property, "property", "", "Property of dataset wanted to be set")
	setupAccountDatasetCmd.Flags().StringVar(&value, "value", "", "Value of dataset wanted to be set")
	// 2592000 = 30 days
	setupAccountDatasetCmd.Flags().Uint64Var(&activeTime, "active-time", 2592000, "Active Time of dataset wanted to be set")

	/*
		RemoveAccountDataset Command
	*/
	removeAccountDatasetCmd.Flags().StringVar(&property, "property", "", "Property of dataset wanted to be removed")
	removeAccountDatasetCmd.Flags().StringVar(&value, "value", "", "Value of dataset wanted to be removed")
}

// Commands set TXGeneratorCommandsInstance that will used by whole commands
func Commands(sqliteDB *sql.DB) *cobra.Command {
	if txGeneratorCommandsInstance == nil {
		txGeneratorCommandsInstance = &TXGeneratorCommands{DB: sqliteDB}
	}

	sendMoneyCmd.Run = txGeneratorCommandsInstance.SendMoneyProcess()
	txCmd.AddCommand(sendMoneyCmd)
	registerNodeCmd.Run = txGeneratorCommandsInstance.RegisterNodeProcess()
	txCmd.AddCommand(registerNodeCmd)
	updateNodeCmd.Run = txGeneratorCommandsInstance.UpdateNodeProcess()
	txCmd.AddCommand(updateNodeCmd)
	removeNodeCmd.Run = txGeneratorCommandsInstance.RemoveAccountDatasetProcess()
	txCmd.AddCommand(removeNodeCmd)
	claimNodeCmd.Run = txGeneratorCommandsInstance.ClaimNodeProcess()
	txCmd.AddCommand(claimNodeCmd)
	setupAccountDatasetCmd.Run = txGeneratorCommandsInstance.SetupAccountDatasetProcess()
	txCmd.AddCommand(setupAccountDatasetCmd)
	removeAccountDatasetCmd.Run = txGeneratorCommandsInstance.RemoveAccountDatasetProcess()
	txCmd.AddCommand(removeAccountDatasetCmd)
	return txCmd
}

// SendMoneyProcess for generate TX SendMoney type
func (*TXGeneratorCommands) SendMoneyProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		tx := GenerateBasicTransaction(senderSeed)
		tx = GenerateTxSendMoney(tx, sendAmount)
		result := PrintTx(GenerateSignedTxBytes(tx, senderSeed), outputType)
		fmt.Println(result)
	}
}

func (txg *TXGeneratorCommands) RegisterNodeProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		tx := GenerateBasicTransaction(senderSeed)
		tx = GenerateTxRegisterNode(
			tx,
			nodeOwnerAccountAddress,
			nodeSeed,
			recipientAccountAddress,
			nodeAddress,
			lockedBalance,
			txg.DB,
		)
		result := PrintTx(GenerateSignedTxBytes(tx, senderSeed), outputType)
		fmt.Println(result)
	}
}

func (txg *TXGeneratorCommands) UpdateNodeProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		tx := GenerateBasicTransaction(senderSeed)
		tx = GenerateTxUpdateNode(
			tx,
			nodeOwnerAccountAddress,
			nodeSeed,
			nodeAddress,
			lockedBalance,
			txg.DB,
		)
		result := PrintTx(GenerateSignedTxBytes(tx, senderSeed), outputType)
		fmt.Println(result)
	}
}

func (*TXGeneratorCommands) RemoveNodeProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		tx := GenerateBasicTransaction(senderSeed)
		tx = GenerateTxRemoveNode(tx, nodeSeed)
		result := PrintTx(GenerateSignedTxBytes(tx, senderSeed), outputType)
		fmt.Println(result)
	}
}

func (txg *TXGeneratorCommands) ClaimNodeProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		tx := GenerateBasicTransaction(senderSeed)
		tx = GenerateTxClaimNode(
			tx,
			nodeOwnerAccountAddress,
			nodeSeed,
			recipientAccountAddress,
			txg.DB,
		)
		result := PrintTx(GenerateSignedTxBytes(tx, senderSeed), outputType)
		fmt.Println(result)
	}
}
func (*TXGeneratorCommands) SetupAccountDatasetProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		senderAccountAddress := util.GetAddressFromSeed(senderSeed)
		tx := GenerateBasicTransaction(senderSeed)
		tx = GenerateTxSetupAccountDataset(tx, senderAccountAddress, recipientAccountAddress, property, value, activeTime)
		result := PrintTx(GenerateSignedTxBytes(tx, senderSeed), outputType)
		fmt.Println(result)
	}
}

func (*TXGeneratorCommands) RemoveAccountDatasetProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		senderAccountAddress := util.GetAddressFromSeed(senderSeed)
		tx := GenerateBasicTransaction(senderSeed)
		tx = GenerateTxRemoveAccountDataset(tx, senderAccountAddress, recipientAccountAddress, property, value)
		result := PrintTx(GenerateSignedTxBytes(tx, senderSeed), outputType)
		fmt.Println(result)
	}
}
