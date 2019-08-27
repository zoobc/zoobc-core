package main

import (
	"database/sql"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/crypto"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/core/blockchainsync"
	"github.com/zoobc/zoobc-core/core/service"

	coreService "github.com/zoobc/zoobc-core/core/service"
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
	dbPath, dbName, nodeSecretPhrase string
	dbInstance                       *database.SqliteDB
	db                               *sql.DB
	apiRPCPort, apiHTTPPort          int
	p2pServiceInstance               p2p.ServiceInterface
	queryExecutor                    *query.Executor
	observerInstance                 *observer.Observer
	blockServices                    = make(map[int32]coreService.BlockServiceInterface)
	mempoolServices                  = make(map[int32]service.MempoolServiceInterface)
	ownerAccountAddress              string
)

func init() {
	var (
		configPostfix string
		err           error
	)

	flag.StringVar(&configPostfix, "config-postfix", "", "Usage")
	flag.Parse()

	if err := util.LoadConfig("./resource", "config"+configPostfix, "toml"); err != nil {
		logrus.Fatal(err)
	} else {
		dbPath = viper.GetString("dbPath")
		dbName = viper.GetString("dbName")
		nodeSecretPhrase = viper.GetString("nodeSecretPhrase")
		apiRPCPort = viper.GetInt("apiRPCPort")
		apiHTTPPort = viper.GetInt("apiHTTPPort")
		ownerAccountAddress = viper.GetString("ownerAccountAddress")
	}

	dbInstance = database.NewSqliteDB()
	if err := dbInstance.InitializeDB(dbPath, dbName); err != nil {
		logrus.Fatal(err)
	}
	db, err = dbInstance.OpenDB(dbPath, dbName, 10, 20)
	if err != nil {
		logrus.Fatal(err)
	}
	queryExecutor = query.NewQueryExecutor(db)

	// initialize Oberver
	observerInstance = observer.NewObserver()
}

func startServices(queryExecutor query.ExecutorInterface, ownerAccountAddress string) {
	startP2pService()
	api.Start(
		apiRPCPort,
		apiHTTPPort,
		queryExecutor,
		p2pServiceInstance,
		blockServices,
		ownerAccountAddress,
	)
}

func startP2pService() {
	myAddress := viper.GetString("myAddress")
	peerPort := viper.GetUint32("peerPort")
	wellknownPeers := viper.GetStringSlice("wellknownPeers")
	p2pServiceInstance = p2p.InitP2P(myAddress, peerPort, wellknownPeers, &p2pNative.Service{}, observerInstance)
	p2pServiceInstance.SetBlockServices(blockServices)

	// run P2P service with any chaintype
	go p2pServiceInstance.StartP2P()
}
func startSmith(sleepPeriod int, processor *smith.BlockchainProcessor) {
	for {
		_ = processor.StartSmithing()
		time.Sleep(time.Duration(sleepPeriod) * time.Second)
	}

}

func startMainchain(mainchainSyncChannel chan bool) {
	mainchain := &chaintype.MainChain{}
	sleepPeriod := int(mainchain.GetChainSmithingDelayTime())
	mempoolService := service.NewMempoolService(
		mainchain,
		queryExecutor,
		query.NewMempoolQuery(mainchain),
		&transaction.TypeSwitcher{
			Executor: queryExecutor,
		},
		query.NewAccountBalanceQuery(),
		observerInstance,
	)
	mempoolServices[mainchain.GetTypeInt()] = mempoolService

	mainchainBlockService := service.NewBlockService(
		mainchain,
		queryExecutor,
		query.NewBlockQuery(mainchain),
		query.NewMempoolQuery(mainchain),
		query.NewTransactionQuery(mainchain),
		crypto.NewSignature(),
		mempoolService,
		&transaction.TypeSwitcher{
			Executor: queryExecutor,
		},
		query.NewAccountBalanceQuery(),
		observerInstance,
	)
	blockServices[mainchain.GetTypeInt()] = mainchainBlockService

	mainchainProcessor := smith.NewBlockchainProcessor(
		mainchain,
		smith.NewBlocksmith(nodeSecretPhrase),
		mainchainBlockService,
	)

	if !mainchainProcessor.BlockService.CheckGenesis() { // Add genesis if not exist

		// genesis account will be inserted in the very beginning
		if err := service.AddGenesisAccount(queryExecutor); err != nil {
			logrus.Fatal("Fail to add genesis account")
		}

		if err := mainchainProcessor.BlockService.AddGenesis(); err != nil {
			logrus.Fatal(err)
		}
	}

	if len(nodeSecretPhrase) > 0 {
		go startSmith(sleepPeriod, mainchainProcessor)
	}
	mainchainSynchronizer := blockchainsync.NewBlockchainSyncService(mainchainBlockService, p2pServiceInstance, queryExecutor)
	mainchainSynchronizer.Start(mainchainSyncChannel)
}

func main() {
	migration := database.Migration{Query: queryExecutor}
	if err := migration.Init(); err != nil {
		logrus.Fatal(err)
	}

	if err := migration.Apply(); err != nil {
		logrus.Fatal(err)
	}

	startServices(queryExecutor, ownerAccountAddress)

	mainchainSyncChannel := make(chan bool, 1)
	mainchainSyncChannel <- true
	startMainchain(mainchainSyncChannel)

	// observer
	observerInstance.AddListener(observer.BlockPushed, p2pServiceInstance.SendBlockListener())
	observerInstance.AddListener(observer.TransactionAdded, p2pServiceInstance.SendTransactionListener())
	for _, blockService := range blockServices {
		observerInstance.AddListener(observer.BlockReceived, blockService.ReceivedBlockListener())
	}
	for _, mempoolService := range mempoolServices {
		observerInstance.AddListener(observer.TransactionReceived, mempoolService.ReceivedTransactionListener())
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// When we receive a signal from the OS, shut down everything
	<-sigs
}
