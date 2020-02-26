package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"github.com/ugorji/go/codec"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/api"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/database"
	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/blockchainsync"
	"github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/core/smith"
	blockSmithStrategy "github.com/zoobc/zoobc-core/core/smith/strategy"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
	"github.com/zoobc/zoobc-core/observer"
	"github.com/zoobc/zoobc-core/p2p"
	"github.com/zoobc/zoobc-core/p2p/client"
	p2pStrategy "github.com/zoobc/zoobc-core/p2p/strategy"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
)

var (
	dbPath, dbName, badgerDbPath, badgerDbName, nodeSecretPhrase, nodeKeyPath,
	nodeKeyFile, nodePreSeed, ownerAccountAddress, myAddress, nodeKeyFilePath, snapshotPath string
	dbInstance                                      *database.SqliteDB
	badgerDbInstance                                *database.BadgerDB
	db                                              *sql.DB
	badgerDb                                        *badger.DB
	apiRPCPort, apiHTTPPort, monitoringPort         int
	apiCertFile, apiKeyFile                         string
	peerPort                                        uint32
	p2pServiceInstance                              p2p.Peer2PeerServiceInterface
	queryExecutor                                   *query.Executor
	kvExecutor                                      *kvdb.KVExecutor
	observerInstance                                *observer.Observer
	schedulerInstance                               *util.Scheduler
	blockServices                                   = make(map[int32]service.BlockServiceInterface)
	snapshotBlockServices                           = make(map[int32]service.SnapshotBlockServiceInterface)
	mainchainBlockService                           *service.BlockService
	mainBlockSnapshotChunkStrategy                  service.SnapshotChunkStrategyInterface
	spinechainBlockService                          *service.BlockSpineService
	fileDownloadService                             service.FileDownloaderServiceInterface
	mempoolServices                                 = make(map[int32]service.MempoolServiceInterface)
	blockIncompleteQueueService                     service.BlockIncompleteQueueServiceInterface
	receiptService                                  service.ReceiptServiceInterface
	peerServiceClient                               client.PeerServiceClientInterface
	p2pHost                                         *model.Host
	peerExplorer                                    p2pStrategy.PeerExplorerStrategyInterface
	wellknownPeers                                  []string
	smithing, isNodePreSeed, isDebugMode            bool
	nodeRegistrationService                         service.NodeRegistrationServiceInterface
	mainchainProcessor                              smith.BlockchainProcessorInterface
	spinechainProcessor                             smith.BlockchainProcessorInterface
	loggerAPIService                                *log.Logger
	loggerCoreService                               *log.Logger
	loggerP2PService                                *log.Logger
	spinechainSynchronizer, mainchainSynchronizer   *blockchainsync.Service
	spineBlockManifestService                       service.SpineBlockManifestServiceInterface
	snapshotService                                 service.SnapshotServiceInterface
	transactionUtil                                 = &transaction.Util{}
	receiptUtil                                     = &coreUtil.ReceiptUtil{}
	transactionCoreServiceIns                       service.TransactionCoreServiceInterface
	fileService                                     service.FileServiceInterface
	chainTypes                                      = chaintype.GetChainTypes()
	mainchain                                       = &chaintype.MainChain{}
	spinechain                                      = &chaintype.SpineChain{}
	blockTypeStatusService                          service.BlockTypeStatusServiceInterface
	mainchainDownloader, spinechainDownloader       blockchainsync.BlockchainDownloadInterface
	mainchainForkProcessor, spinechainForkProcessor blockchainsync.ForkingProcessorInterface
)

func init() {
	var (
		configPostfix string
		configPath    string
		err           error
	)

	flag.StringVar(&configPostfix, "config-postfix", "", "Usage")
	flag.StringVar(&configPath, "config-path", "./resource", "Usage")
	flag.BoolVar(&isDebugMode, "debug", false, "Usage")
	flag.Parse()

	loadNodeConfig(configPath, "config"+configPostfix, "toml")

	initLogInstance()
	// initialize/open db and queryExecutor
	dbInstance = database.NewSqliteDB()
	if err := dbInstance.InitializeDB(dbPath, dbName); err != nil {
		loggerCoreService.Fatal(err)
	}
	db, err = dbInstance.OpenDB(
		dbPath,
		dbName,
		constant.SQLMaxOpenConnetion,
		constant.SQLMaxIdleConnections,
		constant.SQLMaxConnectionLifetime,
	)

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

	// initialize services
	nodeRegistrationService = service.NewNodeRegistrationService(
		queryExecutor,
		query.NewAccountBalanceQuery(),
		query.NewNodeRegistrationQuery(),
		query.NewParticipationScoreQuery(),
		query.NewBlockQuery(mainchain),
		loggerCoreService,
	)
	receiptService = service.NewReceiptService(
		query.NewNodeReceiptQuery(),
		query.NewBatchReceiptQuery(),
		query.NewMerkleTreeQuery(),
		query.NewNodeRegistrationQuery(),
		query.NewBlockQuery(mainchain),
		kvExecutor,
		queryExecutor,
		nodeRegistrationService,
		crypto.NewSignature(),
		query.NewPublishedReceiptQuery(),
		receiptUtil,
	)
	blockTypeStatusService = service.NewBlockTypeStatusService()
	spineBlockManifestService = service.NewSpineBlockManifestService(
		queryExecutor,
		query.NewSpineBlockManifestQuery(),
		query.NewBlockQuery(spinechain),
		loggerCoreService,
	)
	fileService = service.NewFileService(
		loggerCoreService,
		new(codec.CborHandle),
	)
	mainBlockSnapshotChunkStrategy = service.NewSnapshotBasicChunkStrategy(
		constant.SnapshotChunkSize,
		fileService,
	)
	snapshotBlockServices[mainchain.GetTypeInt()] = service.NewSnapshotMainBlockService(
		snapshotPath,
		queryExecutor,
		loggerCoreService,
		mainBlockSnapshotChunkStrategy,
		query.NewAccountBalanceQuery(),
		query.NewNodeRegistrationQuery(),
		query.NewParticipationScoreQuery(),
		query.NewAccountDatasetsQuery(),
		query.NewEscrowTransactionQuery(),
		query.NewPublishedReceiptQuery(),
		query.GetSnapshotQuery(mainchain),
	)

	fileDownloadService = service.NewFileDownloaderService(
		snapshotPath,
		fileService,
		loggerP2PService,
	)
	snapshotService = service.NewSnapshotService(
		spineBlockManifestService,
		blockTypeStatusService,
		snapshotBlockServices,
		fileDownloadService,
		fileService,
		loggerCoreService,
	)

	transactionCoreServiceIns = service.NewTransactionCoreService(
		queryExecutor,
		query.NewTransactionQuery(mainchain),
		query.NewEscrowTransactionQuery(),
	)

	// initialize Observer
	observerInstance = observer.NewObserver()
	schedulerInstance = util.NewScheduler()
	initP2pInstance()
}

func loadNodeConfig(configPath, configFileName, configExtension string) {
	var (
		seed    string
		nodeKey *model.NodeKey
	)

	if err := util.LoadConfig(configPath, configFileName, configExtension); err != nil {
		panic(err)
	}

	myAddress = viper.GetString("myAddress")
	if myAddress == "" {
		ipAddr, err := util.GetOutboundIP()
		if err != nil {
			myAddress = "127.0.0.1"
		} else {
			myAddress = ipAddr.String()
		}
		viper.Set("myAddress", myAddress)
	}
	peerPort = viper.GetUint32("peerPort")
	monitoringPort = viper.GetInt("monitoringPort")
	apiRPCPort = viper.GetInt("apiRPCPort")
	apiHTTPPort = viper.GetInt("apiHTTPPort")
	ownerAccountAddress = viper.GetString("ownerAccountAddress")
	wellknownPeers = viper.GetStringSlice("wellknownPeers")
	smithing = viper.GetBool("smithing")
	dbPath = viper.GetString("dbPath")
	dbName = viper.GetString("dbName")
	badgerDbPath = viper.GetString("badgerDbPath")
	badgerDbName = viper.GetString("badgerDbName")
	nodeKeyPath = viper.GetString("configPath")
	nodeKeyFile = viper.GetString("nodeKeyFile")
	isNodePreSeed = viper.IsSet("nodeSeed")
	nodePreSeed = viper.GetString("nodeSeed")
	apiCertFile = viper.GetString("apiapiCertFile")
	apiKeyFile = viper.GetString("apiKeyFile")
	snapshotPath = viper.GetString("snapshotPath")

	// get the node private key
	nodeKeyFilePath = filepath.Join(nodeKeyPath, nodeKeyFile)
	nodeAdminKeysService := service.NewNodeAdminService(nil, nil, nil, nil, nodeKeyFilePath)
	nodeKeys, err := nodeAdminKeysService.ParseKeysFile()
	if err != nil {
		if isNodePreSeed {
			seed = nodePreSeed
		} else {
			// generate a node private key if there aren't already configured
			seed = util.GetSecureRandomSeed()
		}
		nodePublicKey, err := nodeAdminKeysService.GenerateNodeKey(seed)
		if err != nil {
			loggerCoreService.Fatal(err)
		}
		nodeKey = &model.NodeKey{
			PublicKey: nodePublicKey,
			Seed:      seed,
		}
	} else {
		nodeKey = nodeAdminKeysService.GetLastNodeKey(nodeKeys)
	}
	if nodeKey == nil {
		loggerCoreService.Fatal(errors.New("NodeKeyIsNil"))
	}
	nodeSecretPhrase = nodeKey.Seed
	// log the b64 encoded node public key
	log.Printf("peerPort: %d", peerPort)
	log.Printf("monitoringPort: %d", monitoringPort)
	log.Printf("apiRPCPort: %d", apiRPCPort)
	log.Printf("apiHTTPPort: %d", apiHTTPPort)
	log.Printf("ownerAccountAddress: %s", ownerAccountAddress)
	log.Printf("nodePublicKey: %s", base64.StdEncoding.EncodeToString(nodeKey.PublicKey))
	log.Printf("wellknownPeers: %s", strings.Join(wellknownPeers, ","))
	log.Printf("smithing: %v", smithing)
	log.Printf("myAddress: %s", myAddress)
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
		query.NewNodeReceiptQuery(),
		nodePublicKey,
		query.NewBatchReceiptQuery(),
		query.NewMerkleTreeQuery(),
		receiptService,
		p2pHost,
		loggerP2PService,
	)

	// peer discovery strategy
	peerExplorer = p2pStrategy.NewPriorityStrategy(
		p2pHost,
		peerServiceClient,
		nodeRegistrationService,
		queryExecutor,
		query.NewBlockQuery(mainchain),
		loggerP2PService,
	)
	p2pServiceInstance, _ = p2p.NewP2PService(
		p2pHost,
		peerServiceClient,
		peerExplorer,
		loggerP2PService,
		transactionUtil,
	)
}

func initObserverListeners() {
	// init observer listeners
	// broadcast block will be different than other listener implementation, since there are few exception condition
	observerInstance.AddListener(observer.BroadcastBlock, p2pServiceInstance.SendBlockListener())
	observerInstance.AddListener(observer.TransactionAdded, p2pServiceInstance.SendTransactionListener())
	observerInstance.AddListener(observer.BlockPushed, snapshotService.StartSnapshotListener())
	observerInstance.AddListener(observer.BlockRequestTransactions, p2pServiceInstance.RequestBlockTransactionsListener())
	observerInstance.AddListener(observer.ReceivedBlockTransactionsValidated, blockServices[0].ReceivedValidatedBlockTransactionsListener())
	observerInstance.AddListener(observer.BlockTransactionsRequested, blockServices[0].BlockTransactionsRequestedListener())
	observerInstance.AddListener(observer.SendBlockTransactions, p2pServiceInstance.SendBlockTransactionsListener())
	observerInstance.AddListener(observer.ExpiringEscrowTransactions, transactionCoreServiceIns.ExpiringEscrowListener())
}

func startServices() {
	p2pServiceInstance.StartP2P(
		myAddress,
		peerPort,
		nodeSecretPhrase,
		queryExecutor,
		blockServices,
		mempoolServices,
		observerInstance,
	)
	api.Start(
		apiRPCPort,
		apiHTTPPort,
		kvExecutor,
		queryExecutor,
		p2pServiceInstance,
		blockServices,
		nodeRegistrationService,
		ownerAccountAddress,
		nodeKeyFilePath,
		loggerAPIService,
		isDebugMode,
		apiCertFile,
		apiKeyFile,
		transactionUtil,
		receiptUtil,
		receiptService,
		transactionCoreServiceIns,
	)

	if isDebugMode {
		go startNodeMonitoring()
	}
}

func startNodeMonitoring() {
	log.Infof("starting node monitoring at port:%d...", monitoringPort)
	monitoring.SetMonitoringActive(true)
	monitoring.SetNodePublicKey(util.GetPublicKeyFromSeed(nodeSecretPhrase))
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(fmt.Sprintf(":%d", monitoringPort), nil)
	if err != nil {
		panic(fmt.Sprintf("failed to start monitoring service: %s", err))
	}
}

func startMainchain() {
	var (
		lastBlockAtStart, blockToBuildScrambleNodes *model.Block
		err                                         error
		sleepPeriod                                 = 500
	)
	monitoring.SetBlockchainStatus(mainchain.GetTypeInt(), constant.BlockchainStatusIdle)
	mempoolService := service.NewMempoolService(
		transactionUtil,
		mainchain,
		kvExecutor,
		queryExecutor,
		query.NewMempoolQuery(mainchain),
		query.NewMerkleTreeQuery(),
		&transaction.TypeSwitcher{Executor: queryExecutor},
		query.NewAccountBalanceQuery(),
		query.NewBlockQuery(mainchain),
		query.NewTransactionQuery(mainchain),
		crypto.NewSignature(),
		observerInstance,
		loggerCoreService,
		receiptUtil,
		receiptService,
		transactionCoreServiceIns,
	)
	mempoolServices[mainchain.GetTypeInt()] = mempoolService

	actionSwitcher := &transaction.TypeSwitcher{
		Executor: queryExecutor,
	}
	blocksmithStrategyMain := blockSmithStrategy.NewBlocksmithStrategyMain(
		queryExecutor,
		query.NewNodeRegistrationQuery(),
		query.NewSkippedBlocksmithQuery(),
		loggerCoreService,
	)
	blockIncompleteQueueService = service.NewBlockIncompleteQueueService(
		mainchain,
		observerInstance,
	)
	mainchainBlockPool := service.NewBlockPoolService()
	mainchainBlocksmithService := service.NewBlocksmithService(
		query.NewAccountBalanceQuery(),
		query.NewAccountLedgerQuery(),
		query.NewNodeRegistrationQuery(),
		queryExecutor,
	)
	mainchainCoinbaseService := service.NewCoinbaseService(
		query.NewNodeRegistrationQuery(),
		queryExecutor,
	)
	mainchainParticipationScoreService := service.NewParticipationScoreService(
		query.NewParticipationScoreQuery(),
		queryExecutor,
	)
	mainchainPublishedReceiptUtil := coreUtil.NewPublishedReceiptUtil(
		query.NewPublishedReceiptQuery(),
		queryExecutor,
	)
	mainchainPublishedReceiptService := service.NewPublishedReceiptService(
		query.NewPublishedReceiptQuery(),
		receiptUtil,
		mainchainPublishedReceiptUtil,
		receiptService,
		queryExecutor,
	)
	mainchainBlockService = service.NewBlockMainService(
		mainchain,
		kvExecutor,
		queryExecutor,
		query.NewBlockQuery(mainchain),
		query.NewMempoolQuery(mainchain),
		query.NewTransactionQuery(mainchain),
		query.NewSkippedBlocksmithQuery(),
		crypto.NewSignature(),
		mempoolService,
		receiptService,
		nodeRegistrationService,
		actionSwitcher,
		query.NewAccountBalanceQuery(),
		query.NewParticipationScoreQuery(),
		query.NewNodeRegistrationQuery(),
		observerInstance,
		blocksmithStrategyMain,
		loggerCoreService,
		query.NewAccountLedgerQuery(),
		blockIncompleteQueueService,
		transactionUtil,
		receiptUtil,
		mainchainPublishedReceiptUtil,
		transactionCoreServiceIns,
		mainchainBlockPool,
		mainchainBlocksmithService,
		mainchainCoinbaseService,
		mainchainParticipationScoreService,
		mainchainPublishedReceiptService,
	)
	blockServices[mainchain.GetTypeInt()] = mainchainBlockService

	if !mainchainBlockService.CheckGenesis() { // Add genesis if not exist
		// genesis account will be inserted in the very beginning
		if err := service.AddGenesisAccount(queryExecutor); err != nil {
			loggerCoreService.Fatal("Fail to add genesis account")
		}

		if err := mainchainBlockService.AddGenesis(); err != nil {
			loggerCoreService.Fatal(err)
		}
	}
	lastBlockAtStart, err = mainchainBlockService.GetLastBlock()
	if err != nil {
		loggerCoreService.Fatal(err)
	}

	// TODO: Check computer/node local time. Comparing with last block timestamp

	// initializing scrambled nodes
	heightToBuildScrambleNodes := nodeRegistrationService.GetBlockHeightToBuildScrambleNodes(lastBlockAtStart.GetHeight())
	blockToBuildScrambleNodes, err = mainchainBlockService.GetBlockByHeight(heightToBuildScrambleNodes)
	if err != nil {
		loggerCoreService.Fatal(err)
	}
	err = nodeRegistrationService.BuildScrambledNodes(blockToBuildScrambleNodes)
	if err != nil {
		loggerCoreService.Fatal(err)
	}

	if len(nodeSecretPhrase) > 0 && smithing {
		nodePublicKey := util.GetPublicKeyFromSeed(nodeSecretPhrase)
		node, err := nodeRegistrationService.GetNodeRegistrationByNodePublicKey(nodePublicKey)
		if err != nil {
			// no nodes registered with current node public key
			loggerCoreService.Error(
				"Current node is not in node registry and won't be able to smith until registered!",
			)
		}
		if node != nil {
			mainchainProcessor = smith.NewBlockchainProcessor(
				model.NewBlocksmith(nodeSecretPhrase, nodePublicKey, node.NodeID),
				mainchainBlockService,
				loggerCoreService,
				blockTypeStatusService,
			)
			mainchainProcessor.Start(sleepPeriod)
		}
	}
	mainchainDownloader = blockchainsync.NewBlockchainDownloader(
		mainchainBlockService,
		peerServiceClient,
		peerExplorer,
		loggerCoreService,
		blockTypeStatusService,
	)
	mainchainForkProcessor = &blockchainsync.ForkingProcessor{
		ChainType:          mainchainBlockService.GetChainType(),
		BlockService:       mainchainBlockService,
		QueryExecutor:      queryExecutor,
		ActionTypeSwitcher: actionSwitcher,
		MempoolService:     mempoolService,
		KVExecutor:         kvExecutor,
		PeerExplorer:       peerExplorer,
		Logger:             loggerCoreService,
		TransactionUtil:    transactionUtil,
		TransactionCorService: service.NewTransactionCoreService(
			queryExecutor,
			query.NewTransactionQuery(mainchain),
			query.NewEscrowTransactionQuery(),
		),
	}
	mainchainSynchronizer = blockchainsync.NewBlockchainSyncService(
		mainchainBlockService,
		peerServiceClient, peerExplorer,
		loggerCoreService,
		blockTypeStatusService,
		mainchainDownloader,
		mainchainForkProcessor,
	)
}

func startSpinechain() {
	var (
		nodeID int64
	)
	monitoring.SetBlockchainStatus(spinechain.GetTypeInt(), constant.BlockchainStatusIdle)
	sleepPeriod := 500
	blocksmithStrategySpine := blockSmithStrategy.NewBlocksmithStrategySpine(
		queryExecutor,
		query.NewSpinePublicKeyQuery(),
		loggerCoreService,
		query.NewBlockQuery(spinechain),
	)
	spinechainBlockService = service.NewBlockSpineService(
		spinechain,
		queryExecutor,
		query.NewBlockQuery(spinechain),
		query.NewSpinePublicKeyQuery(),
		crypto.NewSignature(),
		query.NewNodeRegistrationQuery(),
		observerInstance,
		blocksmithStrategySpine,
		loggerCoreService,
		query.NewSpineBlockManifestQuery(),
	)
	blockServices[spinechain.GetTypeInt()] = spinechainBlockService

	if !spinechainBlockService.CheckGenesis() { // Add genesis if not exist
		if err := spinechainBlockService.AddGenesis(); err != nil {
			loggerCoreService.Fatal(err)
		}
	}

	// Note: spine blocks smith even if smithing is false, because are created by every running node
	// 		 Later we only broadcast (and accumulate) signatures of the ones who can smith
	if len(nodeSecretPhrase) > 0 && smithing {
		nodePublicKey := util.GetPublicKeyFromSeed(nodeSecretPhrase)
		// FIXME: ask @barton double check with him that generating a pseudo random id to compute the blockSeed is ok
		nodeID = int64(binary.LittleEndian.Uint64(nodePublicKey))
		spinechainProcessor = smith.NewBlockchainProcessor(
			model.NewBlocksmith(nodeSecretPhrase, nodePublicKey, nodeID),
			spinechainBlockService,
			loggerCoreService,
			blockTypeStatusService,
		)
		spinechainProcessor.Start(sleepPeriod)
	}
	spinechainDownloader = blockchainsync.NewBlockchainDownloader(
		spinechainBlockService,
		peerServiceClient,
		peerExplorer,
		loggerCoreService,
		blockTypeStatusService,
	)
	spinechainForkProcessor = &blockchainsync.ForkingProcessor{
		ChainType:          spinechainBlockService.GetChainType(),
		BlockService:       spinechainBlockService,
		QueryExecutor:      queryExecutor,
		ActionTypeSwitcher: nil, // no mempool for spine blocks
		MempoolService:     nil, // no transaction types for spine blocks
		KVExecutor:         kvExecutor,
		PeerExplorer:       peerExplorer,
		Logger:             loggerCoreService,
		TransactionUtil:    transactionUtil,
		TransactionCorService: service.NewTransactionCoreService(
			queryExecutor,
			query.NewTransactionQuery(mainchain),
			query.NewEscrowTransactionQuery(),
		),
	}
	spinechainSynchronizer = blockchainsync.NewBlockchainSyncService(
		spinechainBlockService,
		peerServiceClient,
		peerExplorer,
		loggerCoreService,
		blockTypeStatusService,
		spinechainDownloader,
		spinechainForkProcessor,
	)
}

// Scheduler Init
func startScheduler() {
	var (
		mainchainMempoolService = mempoolServices[mainchain.GetTypeInt()]
	)
	// scheduler remove expired mempool transaction
	if err := schedulerInstance.AddJob(
		constant.CheckMempoolExpiration,
		mainchainMempoolService.DeleteExpiredMempoolTransactions,
	); err != nil {
		loggerCoreService.Error("Scheduler Err : ", err.Error())
	}
	// scheduler to generate receipt markle root
	if err := schedulerInstance.AddJob(
		constant.ReceiptGenerateMarkleRootPeriod,
		receiptService.GenerateReceiptsMerkleRoot,
	); err != nil {
		loggerCoreService.Error("Scheduler Err : ", err.Error())
	}
	// scheduler to pruning receipts that was expired
	if err := schedulerInstance.AddJob(
		constant.PruningNodeReceiptPeriod,
		receiptService.PruningNodeReceipts,
	); err != nil {
		loggerCoreService.Error("Scheduler Err: ", err.Error())
	}
	// scheduler to remove block uncomplete queue that already waiting transactions too long
	if err := schedulerInstance.AddJob(
		constant.CheckTimedOutBlock,
		blockIncompleteQueueService.PruneTimeoutBlockQueue,
	); err != nil {
		loggerCoreService.Error("Scheduler Err: ", err.Error())
	}
	// register scan block pool for mainchain
	if err := schedulerInstance.AddJob(
		constant.BlockPoolScanPeriod,
		mainchainBlockService.ScanBlockPool,
	); err != nil {
		loggerCoreService.Error("Scheduler Err: ", err.Error())
	}
	// scheduler to remove block uncomplete queue that already waiting transactions too long
	if err := schedulerInstance.AddJob(
		constant.CheckTimedOutBlock,
		blockIncompleteQueueService.PruneTimeoutBlockQueue,
	); err != nil {
		loggerCoreService.Error("Scheduler Err: ", err.Error())
	}
}

func startBlockchainSyncronizers() {
	go spinechainSynchronizer.Start()
	ticker := time.NewTicker(constant.BlockchainsyncCheckInterval)
	timeout := time.After(constant.BlockchainsyncSpineTimeout)
syncronizersLoop:
	for {
		select {
		case <-ticker.C:
			lastSpineBlock, err := spinechainSynchronizer.BlockService.GetLastBlock()
			if err != nil {
				loggerCoreService.Errorf("cannot get last spine block")
				os.Exit(1)
			}
			if blockTypeStatusService.IsFirstDownloadFinished(spinechain) {
				ticker.Stop()
				// loop through all chain types that support snapshots and download them if we find relative
				// spineBlockManifest
				for i := 0; i < len(chainTypes); i++ {
					ct := chaintype.GetChainType(int32(i))
					// exclude spinechain
					if i == int(spinechain.GetTypeInt()) {
						continue
					}

					lastMainBlock, err := mainchainSynchronizer.BlockService.GetLastBlock()
					if err != nil {
						loggerCoreService.Errorf("cannot get last main block")
						os.Exit(1)
					}
					// only download/apply snapshots first time a node joins the network (for now)
					if lastMainBlock.Height == 0 {
						// snapshot download
						lastSpineBlockManifest, err := spineBlockManifestService.GetLastSpineBlockManifest(ct,
							model.SpineBlockManifestType_Snapshot)
						if err != nil {
							loggerCoreService.Errorf("db error: cannot get last spineBlockManifest for chaintype %s",
								ct.GetName())
							break
						}
						if lastSpineBlockManifest != nil {
							loggerCoreService.Infof("found spineBlockManifest for chaintype %s at spine height %d. "+
								"snapshot taken at block height %d", ct.GetName(), lastSpineBlock.Height,
								lastSpineBlockManifest.SpineBlockManifestHeight)
							if err := snapshotService.DownloadSnapshot(ct, lastSpineBlockManifest); err != nil {
								loggerCoreService.Info(err)
							}
						}
					}
					// download remaining main blocks and start the mainchain synchronizer
					// TODO: generalise this so that we can just inject the chaintype and will start the correct
					//  syncronizer
					switch ct.(type) {
					case *chaintype.MainChain:
						go mainchainSynchronizer.Start()
					default:
						loggerCoreService.Debug("invalid chaintype for snapshot")
					}
				}
				break syncronizersLoop
			}
			loggerCoreService.Infof("downloading spine blocks. last height is %d", lastSpineBlock.Height)
		// spine blocks shouldn't take that long to be downloaded. shutdown the node
		// TODO: add push notification to node owner that the node has shutdown because of network issues
		case <-timeout:
			loggerCoreService.Info("spine blocks sync timed out...")
			os.Exit(1)
		}
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
	startSpinechain()
	startMainchain()
	startServices()
	initObserverListeners()
	startScheduler()
	go startBlockchainSyncronizers()

	shutdownCompleted := make(chan bool, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	loggerCoreService.Info("Shutting down node...")
	if mainchainProcessor != nil {
		mainchainProcessor.Stop()
	}
	if spinechainProcessor != nil {
		spinechainProcessor.Stop()
	}
	ticker := time.NewTicker(50 * time.Millisecond)
	timeout := time.After(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			mcSmithing := false
			scSmithing := false
			if mainchainProcessor != nil {
				mcSmithing, _ = mainchainProcessor.GetBlockChainprocessorStatus()
			}
			if spinechainProcessor != nil {
				scSmithing, _ = spinechainProcessor.GetBlockChainprocessorStatus()
			}
			if !mcSmithing && !scSmithing {
				ticker.Stop()
				shutdownCompleted <- true
				loggerCoreService.Info("All smith processors have stopped")
			}
		case <-timeout:
			loggerCoreService.Info("ZOOBC Shutdown timedout...")
			os.Exit(1)
		case <-shutdownCompleted:
			loggerCoreService.Info("ZOOBC Shutdown complete")
			os.Exit(0)
		}
	}
}
