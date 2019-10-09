package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/cmd/account"
	"github.com/zoobc/zoobc-core/cmd/block"
	"github.com/zoobc/zoobc-core/cmd/transaction"
	"github.com/zoobc/zoobc-core/common/util"
)

func main() {
	var (
		rootCmd     *cobra.Command
		logLevels   []string
		generateCmd = &cobra.Command{
			Use:   "generate",
			Short: "generate command is a parent command for generating stuffs",
		}
	)
	rootCmd = &cobra.Command{Use: "zoobc"}
	logLevels = viper.GetStringSlice("logLevels")
	logger, _ := util.InitLogger(".log/", "cmd.debug.log", logLevels)

	rootCmd.AddCommand(generateCmd)
	generateCmd.AddCommand(account.GenerateAccount(logger))
	generateCmd.AddCommand(transaction.Commands())
	generateCmd.AddCommand(block.Commands())
	_ = rootCmd.Execute()
}
