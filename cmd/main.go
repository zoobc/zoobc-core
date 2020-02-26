package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/cmd/account"
	"github.com/zoobc/zoobc-core/cmd/block"
	"github.com/zoobc/zoobc-core/cmd/genesisblock"
	"github.com/zoobc/zoobc-core/cmd/parser"
	"github.com/zoobc/zoobc-core/cmd/rollback"
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
		parserCmd = &cobra.Command{
			Use:   "parser",
			Short: "parse data to understandable struct",
		}
	)

	sqliteDbInstance = database.NewSqliteDB()
	if err := sqliteDbInstance.InitializeDB(dbPath, dbName); err != nil {
		log.Fatalln("InitializeDB err: ", err.Error())
	}
	sqliteDB, err := sqliteDbInstance.OpenDB(dbPath, dbName, 10, 10, 20*time.Minute)
	if err != nil {
		log.Fatalln("OpenDB err: ", err.Error())
	}

	rootCmd = &cobra.Command{
		Use:   "zoobc",
		Short: "CLI app for zoobc core",
		Long:  "Commandline Tools for zoobc core",
	}
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(genesisblock.Commands())
	rootCmd.AddCommand(rollback.Commands(sqliteDB))
	rootCmd.AddCommand(parserCmd)
	generateCmd.AddCommand(account.Commands())
	generateCmd.AddCommand(transaction.Commands(sqliteDB))
	generateCmd.AddCommand(block.Commands())
	parserCmd.AddCommand(parser.Commands())
	_ = rootCmd.Execute()

}
