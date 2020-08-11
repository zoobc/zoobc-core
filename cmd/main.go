package main

import (
	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/cmd/account"
	"github.com/zoobc/zoobc-core/cmd/block"
	"github.com/zoobc/zoobc-core/cmd/genesisblock"
	"github.com/zoobc/zoobc-core/cmd/noderegistry"
	"github.com/zoobc/zoobc-core/cmd/parser"
	"github.com/zoobc/zoobc-core/cmd/rollback"
	"github.com/zoobc/zoobc-core/cmd/scramblednodes"
	"github.com/zoobc/zoobc-core/cmd/signature"
	"github.com/zoobc/zoobc-core/cmd/snapshot"
	"github.com/zoobc/zoobc-core/cmd/transaction"
)

func main() {
	var (
		rootCmd     *cobra.Command
		generateCmd = &cobra.Command{
			Use:   "generate",
			Short: "generate command is a parent command for generating stuffs",
		}
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
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(genesisblock.Commands())
	rootCmd.AddCommand(rollback.Commands())
	rootCmd.AddCommand(parserCmd)
	rootCmd.AddCommand(signature.Commands())
	rootCmd.AddCommand(snapshot.Commands())
	generateCmd.AddCommand(account.Commands())
	generateCmd.AddCommand(transaction.Commands())
	generateCmd.AddCommand(block.Commands())
	generateCmd.AddCommand(noderegistry.Commands())
	parserCmd.AddCommand(parser.Commands())
	generateCmd.AddCommand(scramblednodes.Commands()["getScrambledNodesCmd"])
	generateCmd.AddCommand(scramblednodes.Commands()["getPriorityPeersCmd"])
	_ = rootCmd.Execute()

}
