package main

import (
	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/cmd/account"
	"github.com/zoobc/zoobc-core/cmd/block"
	"github.com/zoobc/zoobc-core/cmd/transaction"
)

func main() {
	var (
		rootCmd     *cobra.Command
		generateCmd = &cobra.Command{
			Use:   "generate",
			Short: "generate command is a parent command for generating stuffs",
		}
	)
	rootCmd = &cobra.Command{Use: "zoobc"}
	rootCmd.AddCommand(generateCmd)
	generateCmd.AddCommand(account.Commands())
	generateCmd.AddCommand(transaction.Commands())
	generateCmd.AddCommand(block.Commands())
	_ = rootCmd.Execute()
}
