package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/cmd/account"
	"github.com/zoobc/zoobc-core/cmd/block"
	"github.com/zoobc/zoobc-core/cmd/genesisblock"
	"github.com/zoobc/zoobc-core/cmd/transaction"
	"github.com/zoobc/zoobc-core/common/database"
	"github.com/zoobc/zoobc-core/common/util"
)

var (
	dbPath, dbName   string
	configPath       = "./resource"
	sqliteDbInstance database.SqliteDBInstance
)

func init() {
	dir, _ := os.Getwd()
	if strings.Contains(dir, "cmd") {
		configPath = "../resource"
	}

	if err := util.LoadConfig(configPath, "config", "toml"); err != nil {
		panic(err)
	}

	dbName = viper.GetString("dbName")
	dbPath = viper.GetString("dbPath")
	if strings.Contains(dir, "cmd") {
		dbPath = filepath.Join("../", viper.GetString("dbPath"))
	}
}
func main() {

	var (
		rootCmd     *cobra.Command
		generateCmd = &cobra.Command{
			Use:   "generate",
			Short: "generate command is a parent command for generating stuffs",
		}
	)

	sqliteDbInstance = database.NewSqliteDB()
	sqliteDB, err := sqliteDbInstance.OpenDB(dbPath, dbName, 10, 20)
	if err != nil {
		log.Fatal(err)
	}

	rootCmd = &cobra.Command{
		Use:   "zoobc",
		Short: "CLI app for zoobc core",
		Long:  "Commandline Tools for zoobc core",
	}
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(genesisblock.Commands())
	generateCmd.AddCommand(account.Commands())
	generateCmd.AddCommand(transaction.Commands(sqliteDB))
	generateCmd.AddCommand(block.Commands())
	_ = rootCmd.Execute()

}
