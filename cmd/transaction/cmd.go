package transaction

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
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
	registerNodeCmdScheduler = &cobra.Command{
		Use:   "register-node",
		Short: "send-money command used to generate \"send money\" transaction used on scheduler",
	}
	registerNodeCmd = &cobra.Command{
		Use:   "register-node",
		Short: "send-money command used to generate \"send money\" transaction",
	}
	updateNodeCmdScheduler = &cobra.Command{
		Use:   "update-node",
		Short: "update-node command used to generate \"update node\" transaction on scheduler",
	}
	updateNodeCmd = &cobra.Command{
		Use:   "update-node",
		Short: "update-node command used to generate \"update node\" transaction",
	}
	removeNodeCmdScheduler = &cobra.Command{
		Use:   "remove-node",
		Short: "remove-node command used to generate \"remove node\" transaction on scheduler",
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
	escrowApprovalCmd = &cobra.Command{
		Use:   "escrow-approval",
		Short: "transaction sub command used to generate 'escrow approval' transaction",
		Long:  "transaction sub command used to generate 'escrow approval' transaction. required transaction id and approval = true:false",
	}
	multiSigCmd = &cobra.Command{
		Use:   "multi-signature",
		Short: "transaction sub command used to generate 'multi signature' transaction",
		Long: "transaction sub command used to generate 'multi signature' transaction that require multiple account to submit their signature " +
			"before it is valid to be executed",
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
	txCmd.PersistentFlags().BoolVar(&post, "post", false, "post generated bytes to [127.0.0.1:7000](default)")
	txCmd.PersistentFlags().StringVar(&postHost, "post-host", "127.0.0.1:7000", "destination of post action")
	txCmd.PersistentFlags().StringVar(&senderAddress, "sender-address", "", "transaction's sender address")
	txCmd.PersistentFlags().Int32Var(
		&senderSignatureType,
		"sender-signature-type",
		int32(model.SignatureType_DefaultSignature),
		"signature-type that provide type of signature want to use to generate the account",
	)

	/*
		SendMoney Command
	*/
	sendMoneyCmd.Flags().Int64Var(&sendAmount, "amount", 0, "Amount of money we want to send")
	sendMoneyCmd.Flags().BoolVar(&escrow, "escrow", false, "Escrowable transaction ? need approver-address if yes")
	sendMoneyCmd.Flags().StringVar(&esApproverAddress, "approver-address", "", "Escrow fields: Approver account address")
	sendMoneyCmd.Flags().Uint64Var(&esTimeout, "timeout", 0, "Escrow fields: Timeout transaction id")
	sendMoneyCmd.Flags().Int64Var(&esCommission, "commission", 0, "Escrow fields: Commission")
	sendMoneyCmd.Flags().StringVar(&esInstruction, "instruction", "", "Escrow fields: Instruction")

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
	/*
		EscrowApproval Command
	*/
	escrowApprovalCmd.Flags().Int64Var(&transactionID, "transaction-id", 0, "escrow approval body field which is int64")
	escrowApprovalCmd.Flags().BoolVar(&approval, "approval", false, "escrow approval body field which is bool")
	/*
		MultiSig Command
	*/
	multiSigCmd.Flags().StringSliceVar(&addresses, "addresses", []string{}, "list of participants "+
		"--addresses='address1,address2'")
	multiSigCmd.Flags().Int64Var(&nonce, "nonce", 0, "random number / access code for the multisig info")
	multiSigCmd.Flags().Uint32Var(&minSignature, "min-signature", 0, "minimum number of signature required for the transaction "+
		"to be valid")
	multiSigCmd.Flags().StringVar(&unsignedTxHex, "unsigned-transaction", "", "hex string of the unsigned transaction bytes")
	multiSigCmd.Flags().StringVar(&txHash, "transaction-hash", "", "hash of transaction being signed by address-signature list (hex)")
	multiSigCmd.Flags().StringToStringVar(&addressSignatures, "address-signatures", make(map[string]string), "address:signature list "+
		"--address1='signature1' --address2='signature2'")
}

// Commands set TXGeneratorCommandsInstance that will used by whole commands
func Commands(sqliteDB *sql.DB) *cobra.Command {
	if txGeneratorCommandsInstance == nil {
		txGeneratorCommandsInstance = &TXGeneratorCommands{DB: sqliteDB}
	}

	sendMoneyCmd.Run = txGeneratorCommandsInstance.SendMoneyProcess()
	txCmd.AddCommand(sendMoneyCmd)
	registerNodeCmdScheduler.Run = txGeneratorCommandsInstance.RegisterNodeProcess()
	txCmd.AddCommand(registerNodeCmdScheduler)
	registerNodeCmd.Run = txGeneratorCommandsInstance.RegisterNodeProcessScheduler()
	txCmd.AddCommand(registerNodeCmd)
	updateNodeCmdScheduler.Run = txGeneratorCommandsInstance.UpdateNodeProcessScheduler()
	txCmd.AddCommand(updateNodeCmdScheduler)
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
	escrowApprovalCmd.Run = txGeneratorCommandsInstance.EscrowApprovalProcess()
	txCmd.AddCommand(escrowApprovalCmd)
	multiSigCmd.Run = txGeneratorCommandsInstance.MultiSignatureProcess()
	txCmd.AddCommand(multiSigCmd)
	return txCmd
}

// SendMoneyProcess for generate TX SendMoney type
func (*TXGeneratorCommands) SendMoneyProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		tx := GenerateBasicTransaction(
			senderSeed,
			senderSignatureType,
			version,
			timestamp,
			fee,
			recipientAccountAddress,
		)
		tx = GenerateTxSendMoney(tx, sendAmount)
		if escrow {
			tx = GenerateEscrowedTransaction(tx)
		}
		PrintTx(GenerateSignedTxBytes(tx, senderSeed, senderSignatureType), outputType)
	}
}

// RegisterNodeProcess for generate TX RegisterNode type
func (txg *TXGeneratorCommands) RegisterNodeProcessScheduler() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		tx := GenerateBasicTransaction(
			senderSeed,
			senderSignatureType,
			version,
			timestamp,
			fee,
			recipientAccountAddress,
		)
		tx = GenerateTxRegisterNodeScheduler(
			tx,
			recipientAccountAddress,
			nodeAddress,
			lockedBalance,
			poowMessageByte,
			signatureByte,
			nodePubKey,
		)
		if escrow {
			tx = GenerateEscrowedTransaction(tx)
		}
		PrintTx(GenerateSignedTxBytes(tx, senderSeed, senderSignatureType), outputType)
	}
}

// RegisterNodeProcess for generate TX RegisterNode type
func (txg *TXGeneratorCommands) RegisterNodeProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		tx := GenerateBasicTransaction(
			senderSeed,
			senderSignatureType,
			version,
			timestamp,
			fee,
			recipientAccountAddress,
		)
		tx = GenerateTxRegisterNode(
			tx,
			nodeOwnerAccountAddress,
			nodeSeed,
			recipientAccountAddress,
			nodeAddress,
			lockedBalance,
			txg.DB,
		)
		if escrow {
			tx = GenerateEscrowedTransaction(tx)
		}
		PrintTx(GenerateSignedTxBytes(tx, senderSeed, senderSignatureType), outputType)
	}
}

// UpdateNodeProcess for generate TX UpdateNode type
func (txg *TXGeneratorCommands) UpdateNodeProcessScheduler() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		tx := GenerateBasicTransaction(
			senderSeed,
			senderSignatureType,
			version,
			timestamp,
			fee,
			recipientAccountAddress,
		)
		tx = GenerateTxUpdateNodeScheduler(
			tx,
			nodeAddress,
			lockedBalance,
			poowMessageByte,
			signatureByte,
			nodePubKey,
		)
		if escrow {
			tx = GenerateEscrowedTransaction(tx)
		}
		PrintTx(GenerateSignedTxBytes(tx, senderSeed, senderSignatureType), outputType)
	}
}

// UpdateNodeProcess for generate TX UpdateNode type
func (txg *TXGeneratorCommands) UpdateNodeProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		tx := GenerateBasicTransaction(
			senderSeed,
			senderSignatureType,
			version,
			timestamp,
			fee,
			recipientAccountAddress,
		)
		tx = GenerateTxUpdateNode(
			tx,
			nodeOwnerAccountAddress,
			nodeSeed,
			nodeAddress,
			lockedBalance,
			txg.DB,
		)
		if escrow {
			tx = GenerateEscrowedTransaction(tx)
		}
		PrintTx(GenerateSignedTxBytes(tx, senderSeed, senderSignatureType), outputType)
	}
}

// RemoveNodeProcessScheduler for generate TX RemoveNode type
func (*TXGeneratorCommands) RemoveNodeProcessScheduler() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		tx := GenerateBasicTransaction(
			senderSeed,
			senderSignatureType,
			version,
			timestamp,
			fee,
			recipientAccountAddress,
		)
		tx = GenerateTxRemoveNodeScheduler(tx, nodePubKey)
		if escrow {
			tx = GenerateEscrowedTransaction(tx)
		}
		PrintTx(GenerateSignedTxBytes(tx, senderSeed, senderSignatureType), outputType)
	}
}

// RemoveNodeProcess for generate TX RemoveNode type
func (*TXGeneratorCommands) RemoveNodeProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		tx := GenerateBasicTransaction(
			senderSeed,
			senderSignatureType,
			version,
			timestamp,
			fee,
			recipientAccountAddress,
		)
		tx = GenerateTxRemoveNode(tx, nodeSeed)
		if escrow {
			tx = GenerateEscrowedTransaction(tx)
		}
		PrintTx(GenerateSignedTxBytes(tx, senderSeed, senderSignatureType), outputType)
	}
}

// ClaimNodeProcess for generate TX ClaimNode type
func (txg *TXGeneratorCommands) ClaimNodeProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		tx := GenerateBasicTransaction(
			senderSeed,
			senderSignatureType,
			version,
			timestamp,
			fee,
			recipientAccountAddress,
		)
		tx = GenerateTxClaimNode(
			tx,
			nodeOwnerAccountAddress,
			nodeSeed,
			recipientAccountAddress,
			txg.DB,
		)
		if escrow {
			tx = GenerateEscrowedTransaction(tx)
		}
		PrintTx(GenerateSignedTxBytes(tx, senderSeed, senderSignatureType), outputType)
	}
}

// SetupAccountDatasetProcess for generate TX SetupAccountDataset type
func (*TXGeneratorCommands) SetupAccountDatasetProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		senderAccountAddress := crypto.NewEd25519Signature().GetAddressFromSeed(senderSeed)
		tx := GenerateBasicTransaction(
			senderSeed,
			senderSignatureType,
			version,
			timestamp,
			fee,
			recipientAccountAddress,
		)
		tx = GenerateTxSetupAccountDataset(tx, senderAccountAddress, recipientAccountAddress, property, value, activeTime)
		if escrow {
			tx = GenerateEscrowedTransaction(tx)
		}
		PrintTx(GenerateSignedTxBytes(tx, senderSeed, senderSignatureType), outputType)
	}
}

// RemoveAccountDatasetProcess for generate TX RemoveAccountDataset type
func (*TXGeneratorCommands) RemoveAccountDatasetProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		senderAccountAddress := crypto.NewEd25519Signature().GetAddressFromSeed(senderSeed)
		tx := GenerateBasicTransaction(
			senderSeed,
			senderSignatureType,
			version,
			timestamp,
			fee,
			recipientAccountAddress,
		)
		tx = GenerateTxRemoveAccountDataset(tx, senderAccountAddress, recipientAccountAddress, property, value)
		if escrow {
			tx = GenerateEscrowedTransaction(tx)
		}
		PrintTx(GenerateSignedTxBytes(tx, senderSeed, senderSignatureType), outputType)
	}
}

// EscrowApprovalProcess for generate TX EscrowApproval type
func (*TXGeneratorCommands) EscrowApprovalProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		tx := GenerateBasicTransaction(
			senderSeed,
			senderSignatureType,
			version,
			timestamp,
			fee,
			recipientAccountAddress,
		)
		tx = GenerateEscrowApprovalTransaction(tx)
		PrintTx(GenerateSignedTxBytes(tx, senderSeed, senderSignatureType), outputType)
	}
}

// MultiSignatureProcess for generate TX MultiSignature type
func (*TXGeneratorCommands) MultiSignatureProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		tx := GenerateBasicTransaction(
			senderSeed,
			senderSignatureType,
			version,
			timestamp,
			fee,
			recipientAccountAddress)

		tx = GeneratedMultiSignatureTransaction(tx, minSignature, nonce, unsignedTxHex, txHash, addressSignatures, addresses)
		if tx == nil {
			fmt.Printf("fail to generate transaction, please check the provided parameter")
		} else {
			PrintTx(GenerateSignedTxBytes(tx, senderSeed, senderSignatureType), outputType)
		}
	}
}
