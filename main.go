package main

import (
	"database/sql"
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/crypto"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/core/blockchainsync"
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
	dbPath, dbName,
	nodeSecretPhrase string
	dbInstance              *database.SqliteDB
	db                      *sql.DB
	apiRPCPort, apiHTTPPort int
	p2pServiceInstance      p2p.ServiceInterface
	queryExecutor           *query.Executor
	observerInstance        *observer.Observer
	blockServices           = make(map[int32]service.BlockServiceInterface)
	mempoolServices         = make(map[int32]service.MempoolServiceInterface)
	ownerAccountAddress     string
	nodeKeyFilePath         string
	nodeRegistrationService service.NodeRegistrationServiceInterface
)

func init() {
	var (
		configPostfix string
		err           error
	)

	flag.StringVar(&configPostfix, "config-postfix", "", "Usage")
	flag.Parse()

	if err := util.LoadConfig("./resource", "config"+configPostfix, "toml"); err != nil {
		log.Fatal(err)
		return
	}

	dbPath = viper.GetString("dbPath")
	dbName = viper.GetString("dbName")
	apiRPCPort = viper.GetInt("apiRPCPort")
	apiHTTPPort = viper.GetInt("apiHTTPPort")
	ownerAccountAddress = viper.GetString("ownerAccountAddress")

	configPath := viper.GetString("configPath")
	nodeKeyFile := viper.GetString("nodeKeyFile")

	// initialize/open db and queryExecutor
	dbInstance = database.NewSqliteDB()
	if err := dbInstance.InitializeDB(dbPath, dbName); err != nil {
		log.Fatal(err)
	}
	db, err = dbInstance.OpenDB(dbPath, dbName, 10, 20)
	if err != nil {
		log.Fatal(err)
	}
	queryExecutor = query.NewQueryExecutor(db)

	// get the node private key
	nodeKeyFilePath = filepath.Join(configPath, nodeKeyFile)
	nodeAdminKeysService := service.NewNodeAdminService(nil, nil, nil, nil, nodeKeyFilePath)
	nodeKeys, err := nodeAdminKeysService.ParseKeysFile()
	if err != nil {
		// generate a node private key if there aren't already configured
		seed := util.GetSecureRandomSeed()
		if _, err := nodeAdminKeysService.GenerateNodeKey(seed); err != nil {
			log.Fatal(err)
		}
	}
	nodeKey := nodeAdminKeysService.GetLastNodeKey(nodeKeys)
	if nodeKey != nil {
		nodeSecretPhrase = nodeKey.Seed
	}

	// initialize nodeRegistration service
	nodeRegistrationService = service.NewNodeRegistrationService(
		queryExecutor,
		query.NewAccountBalanceQuery(),
		query.NewNodeRegistrationQuery(),
		query.NewParticipationScoreQuery(),
	)

	// initialize Oberver
	observerInstance = observer.NewObserver()
}

func startServices(queryExecutor query.ExecutorInterface, ownerAccountAddress, nodeKeyFilePath string) {
	startP2pService()
	api.Start(
		apiRPCPort,
		apiHTTPPort,
		queryExecutor,
		p2pServiceInstance,
		blockServices,
		ownerAccountAddress,
		nodeKeyFilePath,
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
	var (
		blockSmithAddress string
	)
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
		query.NewTransactionQuery(mainchain),
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
		query.NewParticipationScoreQuery(),
		observerInstance,
	)
	blockServices[mainchain.GetTypeInt()] = mainchainBlockService

	if !mainchainBlockService.CheckGenesis() { // Add genesis if not exist

		// genesis account will be inserted in the very beginning
		if err := service.AddGenesisAccount(queryExecutor); err != nil {
			log.Fatal("Fail to add genesis account")
		}

		if err := mainchainBlockService.AddGenesis(); err != nil {
			log.Fatal(err)
		}
	}

	// no nodes registered with current node public key
	nodeRegistration, err := nodeRegistrationService.GetNodeRegistrationByNodePublicKey(util.GetPublicKeyFromSeed(nodeSecretPhrase))
	if err != nil {
		log.Errorf("Current node is not in node registry and won't be able to smith until registered!")
	} else {
		blockSmithAddress = nodeRegistration.AccountAddress
	}
	mainchainProcessor := smith.NewBlockchainProcessor(
		mainchain,
		smith.NewBlocksmith(nodeSecretPhrase, blockSmithAddress),
		mainchainBlockService,
	)

	if len(nodeSecretPhrase) > 0 {
		go startSmith(sleepPeriod, mainchainProcessor)
	}
	mainchainSynchronizer := blockchainsync.NewBlockchainSyncService(mainchainBlockService, p2pServiceInstance)
	mainchainSynchronizer.Start(mainchainSyncChannel)
}

func main() {
	migration := database.Migration{Query: queryExecutor}
	if err := migration.Init(); err != nil {
		log.Fatal(err)
	}

	if err := migration.Apply(); err != nil {
		log.Fatal(err)
	}

	startServices(queryExecutor, ownerAccountAddress, nodeKeyFilePath)

	mainchainSyncChannel := make(chan bool, 1)
	mainchainSyncChannel <- true
	startMainchain(mainchainSyncChannel)

	// observer
	observerInstance.AddListener(observer.BlockPushed, p2pServiceInstance.SendBlockListener())
	observerInstance.AddListener(observer.TransactionAdded, p2pServiceInstance.SendTransactionListener())
	for _, blockService := range blockServices {
		observerInstance.AddListener(observer.BlockReceived, blockService.ReceivedBlockListener())
		if _, ok := blockService.GetChainType().(*chaintype.MainChain); ok {
			observerInstance.AddListener(observer.BlockPushed, nodeRegistrationService.NodeRegistryListener())
		}
	}
	for _, mempoolService := range mempoolServices {
		observerInstance.AddListener(observer.TransactionReceived, mempoolService.ReceivedTransactionListener())
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// When we receive a signal from the OS, shut down everything
	<-sigs
}
