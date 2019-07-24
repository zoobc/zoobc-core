package main

import (
	"database/sql"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zoobc/zoobc-core/common/crypto"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/core/service"

	"github.com/zoobc/zoobc-core/core/smith"

	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/api"
	"github.com/zoobc/zoobc-core/common/database"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

var (
	dbPath, dbName          string
	dbInstance              *database.SqliteDB
	db                      *sql.DB
	nodeSecretPhrase        string
	apiRPCPort, apiHTTPPort int
)

func init() {
	var err error
	if err := util.LoadConfig("./resource", "config", "toml"); err != nil {
		panic(err)
	} else {
		dbPath = viper.GetString("dbPath")
		dbName = viper.GetString("dbName")
		nodeSecretPhrase = viper.GetString("nodeSecretPhrase")
		apiRPCPort = viper.GetInt("apiRPCPort")
		if apiRPCPort == 0 {
			apiRPCPort = 8080
		}
		apiHTTPPort = viper.GetInt("apiHTTPPort")
		if apiHTTPPort == 0 {
			apiHTTPPort = 8000
		}
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
	api.Start(apiRPCPort, apiHTTPPort, queryExecutor)
}

func main() {
	queryExecutor := query.NewQueryExecutor(db)

	migration := database.Migration{Query: queryExecutor}
	if err := migration.Init(); err != nil {
		panic(err)
	}

	if err := migration.Apply(); err != nil {
		panic(err)
	}
	mainchain := &chaintype.MainChain{}
	sleepPeriod := int(mainchain.GetChainSmithingDelayTime())
	blockchainProcessor := smith.NewBlockchainProcessor(mainchain,
		smith.NewBlocksmith(nodeSecretPhrase),
		service.NewBlockService(mainchain, query.NewQueryExecutor(db), query.NewBlockQuery(mainchain),
			query.NewMempoolQuery(mainchain), query.NewTransactionQuery(mainchain), crypto.NewSignature()),
		service.NewMempoolService(mainchain, query.NewQueryExecutor(db), query.NewMempoolQuery(mainchain)))
	if !blockchainProcessor.CheckGenesis() { // Add genesis if not exist
		err := service.AddGenesisAccount(queryExecutor) // genesis account will be inserted in the very beginning
		if err != nil {
			panic("Fail to add genesis account")
		}
		_ = blockchainProcessor.AddGenesis()
	}
	if len(nodeSecretPhrase) > 0 {
		go startSmith(sleepPeriod, blockchainProcessor)
	}

	startServices(queryExecutor)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// When we receive a signal from the OS, shut down everything
	<-sigs

}

func startSmith(sleepPeriod int, processor *smith.BlockchainProcessor) {
	for {
		_ = processor.StartSmithing()
		time.Sleep(time.Duration(sleepPeriod) * time.Second)
	}

}
