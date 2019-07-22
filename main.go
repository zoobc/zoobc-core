package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/core/service"

	"github.com/zoobc/zoobc-core/core/smith"

	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/api"
	"github.com/zoobc/zoobc-core/common/database"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/p2p"
	p2pNative "github.com/zoobc/zoobc-core/p2p/native"
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

func p2pService() {
	myAddress := viper.GetString("myAddress")
	peerPort := viper.GetUint32("peerPort")
	wellknownPeers := viper.GetStringSlice("wellknownPeers")
	p2pService := p2p.InitP2P(myAddress, peerPort, wellknownPeers, &p2pNative.Service{})

	// run P2P service with any chaintype
	go p2pService.StartP2P()
}

func startServices(queryExecutor *query.Executor) {
	api.Start(8000, 8080, queryExecutor)
	go p2pService()
}

func main() {
	fmt.Println("run")

	queryExecutor := query.NewQueryExecutor(db)

	migration := database.Migration{}
	if err := migration.Init(queryExecutor); err != nil {
		panic(err)
	}

	if err := migration.Apply(); err != nil {
		panic(err)
	}
	mainchain := &chaintype.MainChain{}
	sleepPeriod := int(mainchain.GetChainSmithingDelayTime())
	blockchainProcessor := smith.NewBlockchainProcessor(mainchain, smith.NewBlocksmith(), service.NewBlockService(mainchain,
		query.NewQueryExecutor(db), query.NewBlockQuery(mainchain)))
	if !blockchainProcessor.CheckGenesis() { // Add genesis if not exist
		_ = blockchainProcessor.AddGenesis()
	}

	go startSmith(sleepPeriod, blockchainProcessor)

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
