package main

import (
	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/cmd/account"
	"github.com/zoobc/zoobc-core/common/util"
)

func main() {
	var rootCmd = &cobra.Command{Use: "zoobc"}
	logger, _ := util.InitLogger(".log/", "debug.log")
	rootCmd.AddCommand(account.GenerateAccount(logger))
	_ = rootCmd.Execute()
}
