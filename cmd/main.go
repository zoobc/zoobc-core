package main

import (
	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/cmd/account"
	"github.com/zoobc/zoobc-core/cmd/admin"
	"github.com/zoobc/zoobc-core/cmd/block"
	"github.com/zoobc/zoobc-core/cmd/configure"
	"github.com/zoobc/zoobc-core/cmd/genesisblock"
	"github.com/zoobc/zoobc-core/cmd/parser"
	"github.com/zoobc/zoobc-core/cmd/rollback"
	"github.com/zoobc/zoobc-core/cmd/scramblednodes"
	"github.com/zoobc/zoobc-core/cmd/signature"
	"github.com/zoobc/zoobc-core/cmd/snapshot"
	"github.com/zoobc/zoobc-core/cmd/transaction"
)

func main() {
	var (
		rootCmd   *cobra.Command
		parserCmd = &cobra.Command{
			Use:   "parser",
			Short: "parse data to understandable struct",
		}
	)

	rootCmd = &cobra.Command{
		Use:   "zoobc",
		Short: "CLI app for zoobc core",
		Long:  "Commandline Tools for zoobc core",
	}
	rootCmd.AddCommand(genesisblock.Commands())
	rootCmd.AddCommand(rollback.Commands())
	rootCmd.AddCommand(parserCmd)
	rootCmd.AddCommand(signature.Commands())
	rootCmd.AddCommand(snapshot.Commands())
	rootCmd.AddCommand(account.Commands())
	rootCmd.AddCommand(transaction.Commands())
	rootCmd.AddCommand(block.Commands())
	rootCmd.AddCommand(admin.Commands())
	rootCmd.AddCommand(scramblednodes.Commands()["getScrambledNodesCmd"])
	rootCmd.AddCommand(scramblednodes.Commands()["getPriorityPeersCmd"])
	rootCmd.AddCommand(configure.Commands())
	parserCmd.AddCommand(parser.Commands())
	_ = rootCmd.Execute()

}
