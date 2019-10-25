package main

import (
	"database/sql"
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/dgraph-io/badger"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/api"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/database"
	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/blockchainsync"
	"github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/core/smith"
	"github.com/zoobc/zoobc-core/observer"
	"github.com/zoobc/zoobc-core/p2p"
	"github.com/zoobc/zoobc-core/p2p/client"
	"github.com/zoobc/zoobc-core/p2p/strategy"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
)

var (
	dbPath, dbName, badgerDbPath, badgerDbName, nodeSecretPhrase string
	dbInstance                                                   *database.SqliteDB
	badgerDbInstance                                             *database.BadgerDB
	db                                                           *sql.DB
	badgerDb                                                     *badger.DB
	apiRPCPort, apiHTTPPort                                      int
	peerPort                                                     uint32
	p2pServiceInstance                                           p2p.Peer2PeerServiceInterface
	queryExecutor                                                *query.Executor
	kvExecutor                                                   *kvdb.KVExecutor
	observerInstance                                             *observer.Observer
	schedulerInstance                                            *util.Scheduler
	blockServices                                                = make(map[int32]service.BlockServiceInterface)
	mempoolServices                                              = make(map[int32]service.MempoolServiceInterface)
	receiptService                                               service.ReceiptServiceInterface
	peerServiceClient                                            client.PeerServiceClientInterface
	p2pHost                                                      *model.Host
	peerExplorer                                                 strategy.PeerExplorerStrategyInterface
	ownerAccountAddress, myAddress                               string
	wellknownPeers                                               []string
	nodeKeyFilePath                                              string
	smithing                                                     bool
	nodeRegistrationService                                      service.NodeRegistrationServiceInterface
	sortedBlocksmiths                                            []model.Blocksmith
	mainchainProcessor                                           smith.BlockchainProcessorInterface
	loggerAPIService                                             *log.Logger
	loggerCoreService                                            *log.Logger
	loggerP2PService                                             *log.Logger
)

func init() {
	var (
		configPostfix string
		configDir     string
		err           error
	)

	flag.StringVar(&configPostfix, "config-postfix", "", "Usage")
	flag.StringVar(&configDir, "config-path", "", "Usage")
	flag.Parse()

	if configDir == "" {
		configDir = "./resource"
	}

	if err := util.LoadConfig(configDir, "config"+configPostfix, "toml"); err != nil {
		panic(err)
	}

	dbPath = viper.GetString("dbPath")
	dbName = viper.GetString("dbName")
	badgerDbPath = viper.GetString("badgerDbPath")
	badgerDbName = viper.GetString("badgerDbName")
	apiRPCPort = viper.GetInt("apiRPCPort")
	apiHTTPPort = viper.GetInt("apiHTTPPort")
	ownerAccountAddress = viper.GetString("ownerAccountAddress")
	myAddress = viper.GetString("myAddress")
	peerPort = viper.GetUint32("peerPort")
	wellknownPeers = viper.GetStringSlice("wellknownPeers")

	configPath := viper.GetString("configPath")
	nodeKeyFile := viper.GetString("nodeKeyFile")
	smithing = viper.GetBool("smithing")

	initLogInstance()
	// initialize/open db and queryExecutor
	dbInstance = database.NewSqliteDB()
	if err := dbInstance.InitializeDB(dbPath, dbName); err != nil {
		loggerCoreService.Fatal(err)
	}
	db, err = dbInstance.OpenDB(dbPath, dbName, constant.SQLMaxIdleConnections, constant.SQLMaxConnectionLifetime)
	if err != nil {
		loggerCoreService.Fatal(err)
	}
	// initialize k-v db
	badgerDbInstance = database.NewBadgerDB()
	if err := badgerDbInstance.InitializeBadgerDB(badgerDbPath, badgerDbName); err != nil {
		loggerCoreService.Fatal(err)
	}
	badgerDb, err = badgerDbInstance.OpenBadgerDB(badgerDbPath, badgerDbName)
	if err != nil {
		loggerCoreService.Fatal(err)
	}
	queryExecutor = query.NewQueryExecutor(db)
	kvExecutor = kvdb.NewKVExecutor(badgerDb)

	receiptService = service.NewReceiptService(
		query.NewReceiptQuery(),
		query.NewBatchReceiptQuery(),
		query.NewMerkleTreeQuery(),
		kvExecutor,
		queryExecutor,
	)
	// get the node private key
	nodeKeyFilePath = filepath.Join(configPath, nodeKeyFile)
	nodeAdminKeysService := service.NewNodeAdminService(nil, nil, nil, nil, nodeKeyFilePath)
	nodeKeys, err := nodeAdminKeysService.ParseKeysFile()
	if err != nil {
		// generate a node private key if there aren't already configured
		seed := util.GetSecureRandomSeed()
		if _, err := nodeAdminKeysService.GenerateNodeKey(seed); err != nil {
			loggerCoreService.Fatal(err)
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
		loggerCoreService,
	)

	// initialize Observer
	observerInstance = observer.NewObserver()
	schedulerInstance = util.NewScheduler()
	initP2pInstance()
}

func initLogInstance() {
	var (
		err       error
		logLevels = viper.GetStringSlice("logLevels")
		t         = time.Now().Format("2-Jan-2006_")
	)

	if loggerAPIService, err = util.InitLogger(".log/", t+"APIdebug.log", logLevels); err != nil {
		panic(err)
	}
	if loggerCoreService, err = util.InitLogger(".log/", t+"Coredebug.log", logLevels); err != nil {
		panic(err)
	}
	if loggerP2PService, err = util.InitLogger(".log/", t+"P2Pdebug.log", logLevels); err != nil {
		panic(err)
	}
}

func initP2pInstance() {
	// init p2p instances
	knownPeersResult, err := p2pUtil.ParseKnownPeers(wellknownPeers)
	if err != nil {
		loggerCoreService.Fatal("Initialize P2P Err : ", err.Error())
	}
	p2pHost = p2pUtil.NewHost(myAddress, peerPort, knownPeersResult)

	// initialize peer client service
	nodePublicKey := util.GetPublicKeyFromSeed(nodeSecretPhrase)
	peerServiceClient = client.NewPeerServiceClient(
		queryExecutor,
		query.NewReceiptQuery(),
		nodePublicKey,
		query.NewBatchReceiptQuery(),
		query.NewMerkleTreeQuery(),
		p2pHost,
		loggerP2PService,
	)

	// peer discovery strategy
	peerExplorer = strategy.NewPriorityStrategy(
		p2pHost,
		peerServiceClient,
		queryExecutor,
		query.NewNodeRegistrationQuery(),
		loggerP2PService,
	)
	p2pServiceInstance, _ = p2p.NewP2PService(
		p2pHost,
		peerServiceClient,
		peerExplorer,
		loggerP2PService,
	)
}

func initObserverListeners() {
	// init observer listeners
	// broadcast block will be different than other listener implementation, since there are few exception condition
	observerInstance.AddListener(observer.BroadcastBlock, p2pServiceInstance.SendBlockListener())
	observerInstance.AddListener(observer.BlockPushed, nodeRegistrationService.NodeRegistryListener())
	observerInstance.AddListener(observer.BlockPushed, mainchainProcessor.SortBlocksmith(&sortedBlocksmiths))
	observerInstance.AddListener(observer.BlockPushed, peerExplorer.PeerExplorerListener())
	observerInstance.AddListener(observer.TransactionAdded, p2pServiceInstance.SendTransactionListener())
}

func startServices() {
	p2pServiceInstance.StartP2P(
		myAddress,
		peerPort,
		nodeSecretPhrase,
		queryExecutor,
		blockServices,
		mempoolServices,
	)
	api.Start(
		apiRPCPort,
		apiHTTPPort,
		kvExecutor,
		queryExecutor,
		p2pServiceInstance,
		blockServices,
		ownerAccountAddress,
		nodeKeyFilePath,
		loggerAPIService,
	)
}

func startSmith(sleepPeriod int, processor smith.BlockchainProcessorInterface) {
	for {
		err := processor.StartSmithing()
		if err != nil {
			loggerCoreService.Warn("Smith error: ", err.Error())
		}
		time.Sleep(time.Duration(sleepPeriod) * time.Millisecond)
	}
}

func startMainchain(mainchainSyncChannel chan bool) {
	mainchain := &chaintype.MainChain{}
	sleepPeriod := 500
	mempoolService := service.NewMempoolService(
		mainchain,
		kvExecutor,
		queryExecutor,
		query.NewMempoolQuery(mainchain),
		query.NewMerkleTreeQuery(),
		&transaction.TypeSwitcher{
			Executor: queryExecutor,
		},
		query.NewAccountBalanceQuery(),
		crypto.NewSignature(),
		query.NewTransactionQuery(mainchain),
		observerInstance,
		loggerCoreService,
	)
	mempoolServices[mainchain.GetTypeInt()] = mempoolService

	actionSwitcher := &transaction.TypeSwitcher{
		Executor: queryExecutor,
	}

	mainchainBlockService := service.NewBlockService(
		mainchain,
		kvExecutor,
		queryExecutor,
		query.NewBlockQuery(mainchain),
		query.NewMempoolQuery(mainchain),
		query.NewTransactionQuery(mainchain),
		query.NewMerkleTreeQuery(),
		query.NewPublishedReceiptQuery(),
		crypto.NewSignature(),
		mempoolService,
		receiptService,
		actionSwitcher,
		query.NewAccountBalanceQuery(),
		query.NewParticipationScoreQuery(),
		query.NewNodeRegistrationQuery(),
		observerInstance,
		&sortedBlocksmiths,
		loggerCoreService,
	)
	blockServices[mainchain.GetTypeInt()] = mainchainBlockService
	mainchainProcessor = smith.NewBlockchainProcessor(
		mainchain,
		model.NewBlocksmith(nodeSecretPhrase, util.GetPublicKeyFromSeed(nodeSecretPhrase)),
		mainchainBlockService,
		nodeRegistrationService,
	)

	if !mainchainBlockService.CheckGenesis() { // Add genesis if not exist
		// genesis account will be inserted in the very beginning
		if err := service.AddGenesisAccount(queryExecutor); err != nil {
			loggerCoreService.Fatal("Fail to add genesis account")
		}

		if err := mainchainBlockService.AddGenesis(); err != nil {
			loggerCoreService.Fatal(err)
		}
	}

	// Check computer/node local time. Comparing with last block timestamp
	// NEXT: maybe can check timestamp from last block of blockchain network or network time protocol
	lastBlock, err := mainchainBlockService.GetLastBlock()
	if err != nil {
		loggerCoreService.Fatal(err)
	}
	if time.Now().Unix() < lastBlock.GetTimestamp() {
		loggerCoreService.Fatal("Your computer clock is behind from the correct time")
	}

	// no nodes registered with current node public key
	_, err = nodeRegistrationService.GetNodeRegistrationByNodePublicKey(util.GetPublicKeyFromSeed(nodeSecretPhrase))
	if err != nil {
		loggerCoreService.Error("Current node is not in node registry and won't be able to smith until registered!")
	}

	if len(nodeSecretPhrase) > 0 && smithing {
		go startSmith(sleepPeriod, mainchainProcessor)
	}
	mainchainSynchronizer := blockchainsync.NewBlockchainSyncService(
		mainchainBlockService,
		peerServiceClient,
		peerExplorer,
		queryExecutor,
		mempoolService,
		actionSwitcher,
		loggerCoreService,
	)

	go func() {
		mainchainSynchronizer.Start(mainchainSyncChannel)

	}()
}

// Scheduler Init
func startScheduler() {
	var (
		mainchain               = &chaintype.MainChain{}
		mainchainMempoolService = mempoolServices[mainchain.GetTypeInt()]
	)
	if err := schedulerInstance.AddJob(
		constant.CheckMempoolExpiration,
		mainchainMempoolService.DeleteExpiredMempoolTransactions,
	); err != nil {
		loggerCoreService.Error("Scheduler Err : ", err.Error())
	}
	if err := schedulerInstance.AddJob(
		constant.ReceiptGenerateMarkleRootPeriod,
		receiptService.GenerateReceiptsMerkleRoot,
	); err != nil {
		loggerCoreService.Error("Scheduler Err : ", err.Error())
	}
}

func main() {
	migration := database.Migration{Query: queryExecutor}
	if err := migration.Init(); err != nil {
		loggerCoreService.Fatal(err)
	}

	if err := migration.Apply(); err != nil {
		loggerCoreService.Fatal(err)
	}

	mainchainSyncChannel := make(chan bool, 1)
	mainchainSyncChannel <- true
	startMainchain(mainchainSyncChannel)
	startServices()
	initObserverListeners()
	startScheduler()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	loggerCoreService.Info("ZOOBC Shutdown")
	os.Exit(0)
}
