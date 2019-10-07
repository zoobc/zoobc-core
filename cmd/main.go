package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/cmd/account"
	"github.com/zoobc/zoobc-core/cmd/block"
	"github.com/zoobc/zoobc-core/cmd/blockchain"
	"github.com/zoobc/zoobc-core/cmd/transaction"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/util"
)

func main() {
	var (
		rootCmd   *cobra.Command
		logLevels []string
	)
	rootCmd = &cobra.Command{Use: "zoobc"}
	logLevels = viper.GetStringSlice("logLevels")
	logger, _ := util.InitLogger(".log/", "cmd.debug.log", logLevels)
	rootCmd.AddCommand(account.GenerateAccount(logger))
	rootCmd.AddCommand(transaction.GenerateTransactionBytes(logger, &crypto.Signature{}))
	rootCmd.AddCommand(blockchain.GenerateGenesis(logger))
	rootCmd.AddCommand(block.Commands())
	_ = rootCmd.Execute()
}
