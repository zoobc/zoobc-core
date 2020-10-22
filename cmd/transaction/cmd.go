package transaction

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/cmd/helper"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/database"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/signaturetype"
	"github.com/zoobc/zoobc-core/common/transaction"
	commonUtil "github.com/zoobc/zoobc-core/common/util"
	"golang.org/x/crypto/sha3"
)

type (
	// TXGeneratorCommands represent struct of transaction generator commands
	TXGeneratorCommands struct{}
	// RunCommand represent of output function from transaction generator commands
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
		Short: "register-node command is used to generate \"node registration\" transaction",
	}
	updateNodeCmd = &cobra.Command{
		Use:   "update-node",
		Short: "update-node command used to generate \"update node registration\" transaction",
	}
	removeNodeCmd = &cobra.Command{
		Use:   "remove-node",
		Short: "remove-node command used to generate \"remove node registration\" transaction",
	}
	claimNodeCmd = &cobra.Command{
		Use:   "claim-node",
		Short: "claim-node command used to generate \"claim node registration\" transaction",
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
	feeVoteCommitmentCmd = &cobra.Command{
		Use:   "fee-vote-commit",
		Short: "transaction sub command used to generate 'fee vote commitment vote' transaction",
		Long:  "transaction sub command used to generate 'fee vote commitment vote' transaction that require the hash of vote object ",
	}
	feeVoteRevealCmd = &cobra.Command{
		Use:   "fee-vote-reveal",
		Short: "transaction sub command used to generate 'fee vote reveal phase' transaction",
		Long:  "transaction sub command used to generate 'fee vote reveal phase' transaction. part of fee vote do this after commitment vote",
	}
	liquidPaymentCmd = &cobra.Command{
		Use:   "liquid-payment",
		Short: "transaction sub command used to generate 'liquid payment' transaction",
		Long:  "transaction sub command used to generate 'liquid payment' transaction whose payment is based on at what time the payment is stopped",
	}
	liquidPaymentStopCmd = &cobra.Command{
		Use:   "liquid-payment-stop",
		Short: "transaction sub command used to generate 'liquid payment stop' transaction",
		Long:  "transaction sub command used to generate 'liquid payment stop' transaction used to stop a particular liquid payment",
	}
	atomicCmd = &cobra.Command{
		Use:   "atomic",
		Short: "transaction sub command used to generate `atomic` transaction",
		Long:  "transaction sub command used to generate `atomic` transaction",
	}
)

func init() {
	/*
		TXCommandRoot
	*/
	txCmd.PersistentFlags().StringVar(
		&message,
		"message",
		"",
		"--message \"test message\". Include an arbitrary message in the transaction (max 256 bytes)",
	)
	txCmd.PersistentFlags().BoolVarP(&sign, "sign", "s", true, "--sign false. Defines transaction should be signed")
	txCmd.PersistentFlags().StringVar(&outputType, "output", "bytes", "--output hex. Defines the type of the output to be generated")
	txCmd.PersistentFlags().Uint32Var(&version, "version", 1, "--version 2. Defines version of the transaction")
	txCmd.PersistentFlags().Int64Var(&timestamp, "timestamp", time.Now().Unix(), "--timestamp 23456789. defines timestamp of the transaction")
	txCmd.PersistentFlags().StringVar(&senderSeed, "sender-seed", "",
		"--sender-seed mnemonic. Defines the sender seed that's used to sign transaction and whose public key will be used in the"+
			"`Sender Account Address` field of the transaction")
	txCmd.PersistentFlags().StringVar(
		&recipientAccountAddressHex,
		"recipient",
		"",
		"--recipient 000000006f5461d22d8b6ce3c9b76e2c1078bb1b475a9f5280d30b77353ff53514e5c073. Provide recipient transaction hex format",
	)
	txCmd.PersistentFlags().Int64Var(&fee, "fee", 1, "--fee 1. Defines the fee of the transaction")
	txCmd.PersistentFlags().BoolVar(&post, "post", false, "--post. Post generated bytes to [127.0.0.1:7000](default)")
	txCmd.PersistentFlags().StringVar(&postHost, "post-host", "127.0.0.1:7000", "--post-host \"127.0.0.1:7000\". Destination of post action")
	txCmd.PersistentFlags().StringVar(
		&senderAddressHex,
		"sender-address",
		"",
		"--sender-address 000000006f5461d22d8b6ce3c9b76e2c1078bb1b475a9f5280d30b77353ff53514e5c073. "+
			"Transaction's sender address, [no necessary if contain --sender-seed]",
	)
	txCmd.PersistentFlags().StringVarP(&dbPath, "db-path", "p", "resource", "--db-path \"resource\". Defines database path location")
	txCmd.PersistentFlags().StringVarP(&dBName, "db-name", "n", "zoobc.db", "--db-name \"zoobc.db\". Defines database name {name}.db")
	/*
		SendMoney Command
	*/
	sendMoneyCmd.Flags().Int64Var(&sendAmount, "amount", 0, "Amount of money we want to send")
	sendMoneyCmd.Flags().BoolVar(&escrow, "escrow", true, "Escrowable transaction ? need approver-address if yes")
	sendMoneyCmd.Flags().StringVar(&esApproverAddressHex, "approver-address", "", "Escrow fields: Approver account address, require if escrow true")
	sendMoneyCmd.Flags().Uint64Var(&esTimeout, "timeout", 0, "Escrow fields: Timeout transaction id")
	sendMoneyCmd.Flags().Int64Var(&esCommission, "commission", 0, "Escrow fields: Commission")
	sendMoneyCmd.Flags().StringVar(&esInstruction, "instruction", "", "Escrow fields: Instruction")

	/*
		RegisterNode Command
	*/
	registerNodeCmd.Flags().StringVar(&nodeOwnerAccountAddressHex, "node-owner-account-address", "", "Account address of the owner of the node")
	registerNodeCmd.Flags().StringVar(&nodeSeed, "node-seed", "", "Private key of the node")
	registerNodeCmd.Flags().Int64Var(&lockedBalance, "locked-balance", 0, "Amount of money wanted to be locked")
	registerNodeCmd.Flags().StringVar(&proofOfOwnershipHex, "proof-of-ownership-hex", "", "the hex string proof of owenership bytes")
	// db path & db name is needed to get last block of node for making sure generate a valid Proof Of Ownership
	registerNodeCmd.Flags().StringVar(&databasePath, "db-node-path", "../resource", "Database path of node, "+
		"make sure to download the database from node or run this command on node")
	registerNodeCmd.Flags().StringVar(&databaseName, "db-node-name", "zoobc.db", "Database name of node, "+
		"make sure to download the database from node or run this command on node")

	/*
		UpdateNode Command
	*/
	updateNodeCmd.Flags().StringVar(&nodeOwnerAccountAddressHex, "node-owner-account-address", "", "Account address of the owner of the node")
	updateNodeCmd.Flags().StringVar(&nodeSeed, "node-seed", "", "Private key of the node")
	updateNodeCmd.Flags().Int64Var(&lockedBalance, "locked-balance", 0, "Amount of money wanted to be locked")
	updateNodeCmd.Flags().StringVar(&proofOfOwnershipHex, "poow-hex", "", "the hex string proof of owenership bytes")
	// db path & db name is needed to get last block of node for making sure generate a valid Proof Of Ownership
	updateNodeCmd.Flags().StringVar(&databasePath, "db-node-path", "../resource", "Database path of node, "+
		"make sure to download the database from node or run this command on node")
	updateNodeCmd.Flags().StringVar(&databaseName, "db-node-name", "zoobc.db", "Database name of node, "+
		"make sure to download the database from node or run this command on node")

	/*
		RemoveNode Command
	*/
	removeNodeCmd.Flags().StringVar(&nodeSeed, "node-seed", "", "Private key of the node")

	/*
		ClaimNode Command
	*/
	claimNodeCmd.Flags().StringVar(&nodeOwnerAccountAddressHex, "node-owner-account-address", "", "Account address of the owner of the node")
	claimNodeCmd.Flags().StringVar(&nodeSeed, "node-seed", "", "Private key of the node")
	claimNodeCmd.Flags().StringVar(&proofOfOwnershipHex, "poow-hex", "", "the hex string proof of owenership bytes")
	// db path & db name is needed to get last block of node for making sure generate a valid Proof Of Ownership
	claimNodeCmd.Flags().StringVar(&databasePath, "db-node-path", "../resource", "Database path of node, "+
		"make sure to download the database from node or run this command on node")
	claimNodeCmd.Flags().StringVar(&databaseName, "db-node-name", "zoobc.db", "Database name of node, "+
		"make sure to download the database from node or run this command on node")

	/*
		SetupAccountDataset Command
	*/
	setupAccountDatasetCmd.Flags().StringVar(&property, "property", "", "Property of dataset wanted to be set")
	setupAccountDatasetCmd.Flags().StringVar(&value, "value", "", "Value of dataset wanted to be set")

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
	multiSigCmd.Flags().StringSliceVar(&addressesHex, "addressesHex", []string{}, "list of participants "+
		"--addressesHex='address1,address2'")
	multiSigCmd.Flags().Int64Var(&nonce, "nonce", 0, "random number / access code for the multisig info")
	multiSigCmd.Flags().Uint32Var(&minSignature, "min-signature", 0, "minimum number of signature required for the transaction "+
		"to be valid")
	multiSigCmd.Flags().StringVar(&unsignedTxHex, "unsigned-transaction", "", "hex string of the unsigned transaction bytes")
	multiSigCmd.Flags().StringVar(&txHash, "transaction-hash", "", "hash of transaction being signed by address-signature list (hex)")
	multiSigCmd.Flags().StringToStringVar(&addressSignatures, "address-signatures", make(map[string]string), "address:signature list "+
		"--address1='signature1' --address2='signature2'")

	/*
		Fee Vote Commitment Command
	*/
	feeVoteCommitmentCmd.Flags().Int64VarP(&feeVote, "fee-vote", "f", 0, "fee-vote which is how much fee wanna be")

	/*
		Fee Vote Reveal Command
	*/
	feeVoteRevealCmd.Flags().Uint32VarP(&recentBlockHeight, "recent-block-height", "b", 0,
		"recent-block-height which is the recent block hash reference")
	feeVoteRevealCmd.Flags().Int64VarP(&feeVote, "fee-vote", "f", 0, "fee-vote which is how much fee wanna be")

	/*
		liquidPaymentCmd
	*/
	liquidPaymentCmd.Flags().Int64Var(&sendAmount, "amount", 0, "Amount of money we want to send with liquid payment")
	liquidPaymentCmd.Flags().Uint64Var(&completeMinutes, "complete-minutes", 0, "In how long the span we want to send the liquid payment (in minutes)")

	/*
		liquidPaymentStopCmd
	*/
	liquidPaymentStopCmd.Flags().Int64Var(&transactionID, "transaction-id", 0, "liquid payment stop transaction body field which is int64")
	/*
		atomicCmd
	*/
	atomicCmd.Flags().Uint32VarP(&inners, "inners", "i", 1, "inners 2, indicate want to have 2 inner transaction")
}

// Commands set TXGeneratorCommandsInstance that will used by whole commands
func Commands() *cobra.Command {
	if txGeneratorCommandsInstance == nil {
		txGeneratorCommandsInstance = &TXGeneratorCommands{}
	}

	sendMoneyCmd.Run = txGeneratorCommandsInstance.SendMoneyProcess()
	txCmd.AddCommand(sendMoneyCmd)
	registerNodeCmd.Run = txGeneratorCommandsInstance.RegisterNodeProcess()
	txCmd.AddCommand(registerNodeCmd)
	updateNodeCmd.Run = txGeneratorCommandsInstance.UpdateNodeProcess()
	txCmd.AddCommand(updateNodeCmd)
	removeNodeCmd.Run = txGeneratorCommandsInstance.RemoveNodeProcess()
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
	feeVoteCommitmentCmd.Run = txGeneratorCommandsInstance.feeVoteCommitmentProcess()
	txCmd.AddCommand(feeVoteCommitmentCmd)
	feeVoteRevealCmd.Run = txGeneratorCommandsInstance.feeVoteRevealProcess()
	txCmd.AddCommand(feeVoteRevealCmd)
	liquidPaymentCmd.Run = txGeneratorCommandsInstance.LiquidPaymentProcess()
	txCmd.AddCommand(liquidPaymentCmd)
	liquidPaymentStopCmd.Run = txGeneratorCommandsInstance.LiquidPaymentStopProcess()
	txCmd.AddCommand(liquidPaymentStopCmd)
	atomicCmd.Run = txGeneratorCommandsInstance.AtomicProcess()
	txCmd.AddCommand(atomicCmd)

	return txCmd
}

// SendMoneyProcess for generate TX SendMoney type
func (*TXGeneratorCommands) SendMoneyProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		tx := GenerateBasicTransaction(
			senderAddressHex,
			senderSeed,
			version,
			timestamp,
			fee,
			recipientAccountAddressHex,
			message,
		)
		tx = GenerateTxSendMoney(tx, sendAmount)
		if escrow {
			tx = GenerateEscrowedTransaction(tx)
		}
		senderAccountType := getAccountTypeFromAccountHex(hex.EncodeToString(tx.GetSenderAccountAddress())).GetTypeInt()
		PrintTx(GenerateSignedTxBytes(tx, senderSeed, senderAccountType, sign), outputType)
	}
}

// RegisterNodeProcess for generate TX RegisterNode type
func (*TXGeneratorCommands) RegisterNodeProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		var (
			tx = GenerateBasicTransaction(
				senderAddressHex,
				senderSeed,
				version,
				timestamp,
				fee,
				recipientAccountAddressHex,
				message,
			)
			nodePubKey = signaturetype.NewEd25519Signature().GetPublicKeyFromSeed(nodeSeed)
			poow       = GenerateProofOfOwnership(
				databasePath,
				databaseName,
				nodeOwnerAccountAddressHex,
				nodeSeed,
				proofOfOwnershipHex,
			)
		)
		tx = GenerateTxRegisterNode(
			tx,
			lockedBalance,
			nodePubKey,
			poow,
		)
		if escrow {
			tx = GenerateEscrowedTransaction(tx)
		}
		senderAccountType := getAccountTypeFromAccountHex(senderAddressHex).GetTypeInt()
		PrintTx(GenerateSignedTxBytes(tx, senderSeed, senderAccountType, sign), outputType)
	}
}

// UpdateNodeProcess for generate TX UpdateNode type
func (*TXGeneratorCommands) UpdateNodeProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		var (
			tx = GenerateBasicTransaction(
				senderAddressHex,
				senderSeed,
				version,
				timestamp,
				fee,
				recipientAccountAddressHex,
				message,
			)
			nodePubKey = signaturetype.NewEd25519Signature().GetPublicKeyFromSeed(nodeSeed)
			poow       = GenerateProofOfOwnership(
				databasePath,
				databaseName,
				nodeOwnerAccountAddressHex,
				nodeSeed,
				proofOfOwnershipHex,
			)
		)

		tx = GenerateTxUpdateNode(
			tx,
			lockedBalance,
			nodePubKey,
			poow,
		)
		if escrow {
			tx = GenerateEscrowedTransaction(tx)
		}
		senderAccountType := getAccountTypeFromAccountHex(senderAddressHex).GetTypeInt()
		PrintTx(GenerateSignedTxBytes(tx, senderSeed, senderAccountType, sign), outputType)
	}
}

// RemoveNodeProcess for generate TX RemoveNode type
func (*TXGeneratorCommands) RemoveNodeProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		tx := GenerateBasicTransaction(
			senderAddressHex,
			senderSeed,
			version,
			timestamp,
			fee,
			recipientAccountAddressHex,
			message,
		)
		nodePubKey := signaturetype.NewEd25519Signature().GetPublicKeyFromSeed(nodeSeed)
		tx = GenerateTxRemoveNode(tx, nodePubKey)
		if escrow {
			tx = GenerateEscrowedTransaction(tx)
		}
		senderAccountType := getAccountTypeFromAccountHex(senderAddressHex).GetTypeInt()
		PrintTx(GenerateSignedTxBytes(tx, senderSeed, senderAccountType, sign), outputType)
	}
}

// ClaimNodeProcess for generate TX ClaimNode type
func (*TXGeneratorCommands) ClaimNodeProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		var (
			tx = GenerateBasicTransaction(
				senderAddressHex,
				senderSeed,
				version,
				timestamp,
				fee,
				recipientAccountAddressHex,
				message,
			)
			nodePubKey = signaturetype.NewEd25519Signature().GetPublicKeyFromSeed(nodeSeed)
			poow       = GenerateProofOfOwnership(
				databasePath,
				databaseName,
				nodeOwnerAccountAddressHex,
				nodeSeed,
				proofOfOwnershipHex,
			)
		)
		tx = GenerateTxClaimNode(
			tx,
			nodePubKey,
			poow,
		)
		if escrow {
			tx = GenerateEscrowedTransaction(tx)
		}
		senderAccountType := getAccountTypeFromAccountHex(senderAddressHex).GetTypeInt()
		PrintTx(GenerateSignedTxBytes(tx, senderSeed, senderAccountType, sign), outputType)
	}
}

// SetupAccountDatasetProcess for generate TX SetupAccountDataset type
func (*TXGeneratorCommands) SetupAccountDatasetProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		tx := GenerateBasicTransaction(
			senderAddressHex,
			senderSeed,
			version,
			timestamp,
			fee,
			recipientAccountAddressHex,
			message,
		)

		// Recipient required while property set as AccountDatasetEscrowApproval
		_, ok := model.AccountDatasetProperty_value[property]
		if ok && recipientAccountAddressHex == "" {
			println("--recipient is required while property as AccountDatasetEscrowApproval")
			return
		}
		tx = GenerateTxSetupAccountDataset(tx, property, value)
		if escrow {
			tx = GenerateEscrowedTransaction(tx)
		}
		senderAccountType := getAccountTypeFromAccountHex(senderAddressHex).GetTypeInt()
		PrintTx(GenerateSignedTxBytes(tx, senderSeed, senderAccountType, sign), outputType)
	}
}

// RemoveAccountDatasetProcess for generate TX RemoveAccountDataset type
func (*TXGeneratorCommands) RemoveAccountDatasetProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		tx := GenerateBasicTransaction(
			senderAddressHex,
			senderSeed,
			version,
			timestamp,
			fee,
			recipientAccountAddressHex,
			message,
		)
		tx = GenerateTxRemoveAccountDataset(tx, property, value)
		if escrow {
			tx = GenerateEscrowedTransaction(tx)
		}
		senderAccountType := getAccountTypeFromAccountHex(senderAddressHex).GetTypeInt()
		PrintTx(GenerateSignedTxBytes(tx, senderSeed, senderAccountType, sign), outputType)
	}
}

// EscrowApprovalProcess for generate TX EscrowApproval type
func (*TXGeneratorCommands) EscrowApprovalProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		tx := GenerateBasicTransaction(
			senderAddressHex,
			senderSeed,
			version,
			timestamp,
			fee,
			recipientAccountAddressHex,
			message,
		)
		tx = GenerateEscrowApprovalTransaction(tx)
		senderAccountType := getAccountTypeFromAccountHex(senderAddressHex).GetTypeInt()
		PrintTx(GenerateSignedTxBytes(tx, senderSeed, senderAccountType, sign), outputType)
	}
}

// MultiSignatureProcess for generate TX MultiSignature type
func (*TXGeneratorCommands) MultiSignatureProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		tx := GenerateBasicTransaction(
			senderAddressHex,
			senderSeed,
			version,
			timestamp,
			fee,
			recipientAccountAddressHex,
			message,
		)

		tx = GeneratedMultiSignatureTransaction(tx, minSignature, nonce, unsignedTxHex, txHash, addressSignatures, addressesHex)
		if tx == nil {
			fmt.Printf("fail to generate transaction, please check the provided parameter")
		} else {
			senderAccountType := getAccountTypeFromAccountHex(senderAddressHex).GetTypeInt()
			PrintTx(GenerateSignedTxBytes(tx, senderSeed, senderAccountType, sign), outputType)
		}
	}
}

// feeVoteCommitmentProcess for generate TX  commitment vote of fee vote
func (*TXGeneratorCommands) feeVoteCommitmentProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		var (
			err         error
			feeVoteInfo model.FeeVoteInfo
			sqliteDB    *sql.DB
			// voteHash    []byte
			tx = GenerateBasicTransaction(
				senderAddressHex,
				senderSeed,
				version,
				timestamp,
				fee,
				recipientAccountAddressHex,
				message,
			)
		)

		dbInstance := database.NewSqliteDB()
		dbPath = path.Join(helper.GetAbsDBPath(), dbPath)
		err = dbInstance.InitializeDB(dbPath, dBName)
		if err != nil {
			_ = feeVoteCommitmentCmd.Help()
			logrus.Errorf("Getting last block failed: %s", err.Error())
			os.Exit(1)
		}
		sqliteDB, err = dbInstance.OpenDB(
			dbPath,
			dBName,
			constant.SQLMaxOpenConnetion,
			constant.SQLMaxIdleConnections,
			constant.SQLMaxConnectionLifetime,
		)
		if err != nil {
			_ = feeVoteCommitmentCmd.Help()
			logrus.Errorf("Getting last block failed: %s", err.Error())
			os.Exit(1)
		}

		lastBlock, err := commonUtil.GetLastBlock(
			query.NewQueryExecutor(sqliteDB),
			query.NewBlockQuery(&chaintype.MainChain{}),
		)
		if err != nil {
			_ = feeVoteCommitmentCmd.Help()
			logrus.Errorf("Getting last block failed: %s", err.Error())
			os.Exit(1)
		}
		feeVoteInfo = model.FeeVoteInfo{
			RecentBlockHeight: lastBlock.GetHeight(),
			RecentBlockHash:   lastBlock.GetBlockHash(),
			FeeVote:           feeVote,
		}
		fb := (&transaction.FeeVoteRevealTransaction{
			Body: &model.FeeVoteRevealTransactionBody{
				FeeVoteInfo: &feeVoteInfo,
			},
		}).GetFeeVoteInfoBytes()

		digest := sha3.New256()
		_, err = digest.Write(fb)
		if err != nil {
			_ = feeVoteCommitmentCmd.Help()
			logrus.Errorf("GetLast block failed: %s", err.Error())
			os.Exit(1)
		}
		tx = GenerateTxFeeVoteCommitment(tx, digest.Sum([]byte{}))
		if tx == nil {
			fmt.Printf("fail to generate transaction, please check the provided parameter")
		} else {
			senderAccountType := getAccountTypeFromAccountHex(senderAddressHex).GetTypeInt()
			PrintTx(GenerateSignedTxBytes(tx, senderSeed, senderAccountType, sign), outputType)
		}
	}
}

func (*TXGeneratorCommands) feeVoteRevealProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		var (
			feeVoteInfo   model.FeeVoteInfo
			feeVoteSigned []byte
			err           error
			tx            = GenerateBasicTransaction(
				senderAddressHex,
				senderSeed,
				version,
				timestamp,
				fee,
				recipientAccountAddressHex,
				message)
		)

		if recentBlockHeight != 0 {
			var (
				dbInstance = database.NewSqliteDB()
				sqliteDB   *sql.DB
				row        *sql.Row
				block      model.Block
				blockQuery = query.NewBlockQuery(&chaintype.MainChain{})
			)
			dbPath = path.Join(helper.GetAbsDBPath(), dbPath)
			err = dbInstance.InitializeDB(dbPath, dBName)
			if err != nil {
				_ = feeVoteRevealCmd.Help()
				logrus.Errorf("Getting last block failed: %s", err.Error())
				os.Exit(1)
			}
			sqliteDB, err = dbInstance.OpenDB(
				dbPath,
				dBName,
				constant.SQLMaxOpenConnetion,
				constant.SQLMaxIdleConnections,
				constant.SQLMaxConnectionLifetime,
			)
			if err != nil {
				_ = feeVoteRevealCmd.Help()
				logrus.Errorf("Getting last block failed: %s", err.Error())
				os.Exit(1)
			}
			row, err = query.NewQueryExecutor(sqliteDB).ExecuteSelectRow(
				blockQuery.GetBlockByHeight(recentBlockHeight),
				false,
			)
			if err != nil {
				_ = feeVoteRevealCmd.Help()
				logrus.Errorf("Getting last block failed: %s", err.Error())
				return
			}
			err = blockQuery.Scan(&block, row)
			if err != nil {
				_ = feeVoteRevealCmd.Help()
				logrus.Errorf("Getting last block failed: %s", err.Error())
				return
			}
			feeVoteInfo.RecentBlockHash = block.GetBlockHash()
			feeVoteInfo.RecentBlockHeight = recentBlockHeight
		}

		feeVoteInfo.FeeVote = feeVote
		fb := (&transaction.FeeVoteRevealTransaction{
			Body: &model.FeeVoteRevealTransactionBody{
				FeeVoteInfo: &feeVoteInfo,
			},
		}).GetFeeVoteInfoBytes()
		feeVoteSigned, err = signature.Sign(
			fb,
			model.AccountType_ZbcAccountType,
			senderSeed,
		)
		if err != nil {
			_ = feeVoteRevealCmd.Help()
			logrus.Error("Failed to sign fee vote info, check seed")
			return
		}
		tx = GenerateTxFeeVoteRevealPhase(tx, &feeVoteInfo, feeVoteSigned)

		senderAccountType := getAccountTypeFromAccountHex(senderAddressHex).GetTypeInt()
		PrintTx(GenerateSignedTxBytes(tx, senderSeed, senderAccountType, sign), outputType)
	}
}

// LiquidPaymentProcess for generate TX LiquidPayment type
func (*TXGeneratorCommands) LiquidPaymentProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		tx := GenerateBasicTransaction(
			senderAddressHex,
			senderSeed,
			version,
			timestamp,
			fee,
			recipientAccountAddressHex,
			message,
		)
		tx = GenerateTxLiquidPayment(tx, sendAmount, completeMinutes)
		senderAccountType := getAccountTypeFromAccountHex(senderAddressHex).GetTypeInt()
		PrintTx(GenerateSignedTxBytes(tx, senderSeed, senderAccountType, sign), outputType)
	}
}

// LiquidPaymentStopProcess for generate TX LiquidPaymentStop type
func (*TXGeneratorCommands) LiquidPaymentStopProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		tx := GenerateBasicTransaction(
			senderAddressHex,
			senderSeed,
			version,
			timestamp,
			fee,
			recipientAccountAddressHex,
			message,
		)
		tx = GenerateTxLiquidPaymentStop(tx, transactionID)
		senderAccountType := getAccountTypeFromAccountHex(senderAddressHex).GetTypeInt()
		PrintTx(GenerateSignedTxBytes(tx, senderSeed, senderAccountType, sign), outputType)
	}
}

func (*TXGeneratorCommands) AtomicProcess() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		tx := GenerateBasicTransaction(
			senderAddressHex,
			senderSeed,
			version,
			timestamp,
			fee,
			recipientAccountAddressHex,
			message,
		)

		tx = GenerateAtomic(tx, inners)
		senderAccountType := getAccountTypeFromAccountHex(hex.EncodeToString(tx.GetSenderAccountAddress())).GetTypeInt()
		PrintTx(GenerateSignedTxBytes(tx, senderSeed, senderAccountType, sign), outputType)
	}
}
