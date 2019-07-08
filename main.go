package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/api"
	"github.com/zoobc/zoobc-core/common/database"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

var (
	dbPath, dbName string
	dbInstance     *database.SqliteDB
	db             *sql.DB
)

func init() {
	var err error
	if err := util.LoadConfig("./resource", "config", "toml"); err != nil {
		panic(err)
	} else {
		dbPath = viper.GetString("dbPath")
		dbName = viper.GetString("dbName")
	}

	dbInstance = database.NewSqliteDB()
	if err := dbInstance.InitializeDB(dbPath, dbName); err != nil {
		panic(err)
	}
	db, err = dbInstance.OpenDB(dbPath, dbName, 10, 20)
	if err != nil {
		panic(err)
	}
}

func startServices(queryExecutor *query.Executor) {
	api.Start(8000, 8080, queryExecutor)
}

func main() {
	fmt.Println("run")

	queryExecutor := query.NewQueryExecutor(db)

	migration := database.Migration{}
	if err := migration.Init(queryExecutor); err != nil {
		panic(err)
	}

	if err := migration.Apply(); err != nil {
		fmt.Println(err)
	}

	startServices(queryExecutor)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// When we receive a signal from the OS, shut down everything
	<-sigs
}
