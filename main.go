package main

import (
	"database/sql"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/crypto"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/core/service"

	"github.com/zoobc/zoobc-core/core/smith"

	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/api"
	"github.com/zoobc/zoobc-core/common/database"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/observer"
	"github.com/zoobc/zoobc-core/p2p"
	p2pNative "github.com/zoobc/zoobc-core/p2p/native"
)

var (
	dbPath, dbName          string
	dbInstance              *database.SqliteDB
	db                      *sql.DB
	nodeSecretPhrase        string
	apiRPCPort, apiHTTPPort int
	p2pServiceInstance      contract.P2PType
	queryExecutor           *query.Executor
)

func init() {
	var configPostfix string
	flag.StringVar(&configPostfix, "config-postfix", "", "Usage")
	flag.Parse()

	var err error
	if err := util.LoadConfig("./resource", "config"+configPostfix, "toml"); err != nil {
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
	queryExecutor = query.NewQueryExecutor(db)
}

func startServices(queryExecutor query.ExecutorInterface) {
	p2pService()
	api.Start(apiRPCPort, apiHTTPPort, queryExecutor, p2pServiceInstance)
}

func p2pService() {
	myAddress := viper.GetString("myAddress")
	peerPort := viper.GetUint32("peerPort")
	wellknownPeers := viper.GetStringSlice("wellknownPeers")
	p2pServiceInstance = p2p.InitP2P(myAddress, peerPort, wellknownPeers, &p2pNative.Service{})

	// run P2P service with any chaintype
	go p2pServiceInstance.StartP2P()
}

func main() {

	migration := database.Migration{Query: queryExecutor}
	if err := migration.Init(); err != nil {
		panic(err)
	}

	if err := migration.Apply(); err != nil {
		panic(err)
	}
	mainchain := &chaintype.MainChain{}
	sleepPeriod := int(mainchain.GetChainSmithingDelayTime())

	blockchainProcessor := smith.NewBlockchainProcessor(
		mainchain,
		smith.NewBlocksmith(nodeSecretPhrase),
		service.NewBlockService(
			mainchain,
			queryExecutor,
			query.NewBlockQuery(mainchain),
			query.NewMempoolQuery(mainchain),
			query.NewTransactionQuery(mainchain),
			crypto.NewSignature(),
			service.NewMempoolService(
				mainchain,
				queryExecutor,
				query.NewMempoolQuery(mainchain),
				&transaction.TypeSwitcher{
					Executor: queryExecutor,
				},
				query.NewAccountBalanceQuery(),
			),
			&transaction.TypeSwitcher{
				Executor: queryExecutor,
			},
			query.NewAccountBalanceQuery(),
		),
	)

	if !blockchainProcessor.BlockService.CheckGenesis() { // Add genesis if not exist

		// genesis account will be inserted in the very beginning
		if err := service.AddGenesisAccount(queryExecutor); err != nil {
			panic("Fail to add genesis account")
		}

		if err := blockchainProcessor.BlockService.AddGenesis(); err != nil {
			panic(err)
		}
	}

	if len(nodeSecretPhrase) > 0 {
		go startSmith(sleepPeriod, blockchainProcessor)
	}

	startServices(queryExecutor)

	// observer
	observer.NewObserver().AddListener(p2pServiceInstance.SendBlockListener(), observer.BlockPushed)

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
