package main

import (
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/zoobc/zoobc-core/common/auth"

	"github.com/zoobc/lib/address"

	badger "github.com/dgraph-io/badger/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/ugorji/go/codec"

	"github.com/zoobc/zoobc-core/api"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/database"
	"github.com/zoobc/zoobc-core/common/fee"
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
	nodeKeyFile, ownerAccountAddress, myAddress, nodeKeyFilePath, snapshotPath string
	nodeAddressDynamic                              bool
	dbInstance                                      *database.SqliteDB
	badgerDbInstance                                *database.BadgerDB
	db                                              *sql.DB
	badgerDb                                        *badger.DB
	apiRPCPort, monitoringPort                      int
	cpuProfilingPort                                int
	apiCertFile, apiKeyFile                         string
	peerPort                                        uint32
	maxAPIRequestPerSecond                          uint32
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
	fileDownloader                                  p2p.FileDownloaderInterface
	mempoolServices                                 = make(map[int32]service.MempoolServiceInterface)
	blockIncompleteQueueService                     service.BlockIncompleteQueueServiceInterface
	receiptService                                  service.ReceiptServiceInterface
	peerServiceClient                               client.PeerServiceClientInterface
	peerExplorer                                    p2pStrategy.PeerExplorerStrategyInterface
	wellknownPeers                                  []string
	smithing, isDebugMode                           bool
	nodeRegistrationService                         service.NodeRegistrationServiceInterface
	nodeAuthValidationService                       auth.NodeAuthValidationInterface
	mainchainProcessor                              smith.BlockchainProcessorInterface
	spinechainProcessor                             smith.BlockchainProcessorInterface
	loggerAPIService                                *log.Logger
	loggerCoreService                               *log.Logger
	loggerP2PService                                *log.Logger
	spinechainSynchronizer, mainchainSynchronizer   blockchainsync.BlockchainSyncServiceInterface
	spineBlockManifestService                       service.SpineBlockManifestServiceInterface
	snapshotService                                 service.SnapshotServiceInterface
	transactionUtil                                 transaction.UtilInterface
	receiptUtil                                     = &coreUtil.ReceiptUtil{}
	transactionCoreServiceIns                       service.TransactionCoreServiceInterface
	fileService                                     service.FileServiceInterface
	mainchain                                       = &chaintype.MainChain{}
	spinechain                                      = &chaintype.SpineChain{}
	blockchainStatusService                         service.BlockchainStatusServiceInterface
	nodeConfigurationService                        service.NodeConfigurationServiceInterface
	nodeAddressInfoService                          service.NodeAddressInfoServiceInterface
	feeScaleService                                 fee.FeeScaleServiceInterface
	mainchainDownloader, spinechainDownloader       blockchainsync.BlockchainDownloadInterface
	mainchainForkProcessor, spinechainForkProcessor blockchainsync.ForkingProcessorInterface
	defaultSignatureType                            *crypto.Ed25519Signature
	nodeKey                                         *model.NodeKey
	cpuProfile                                      bool
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
	flag.BoolVar(&cpuProfile, "cpu-profile", false, "if this flag is used, write cpu profile to file")
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

	knownPeersResult, err := p2pUtil.ParseKnownPeers(wellknownPeers)
	if err != nil {
		loggerCoreService.Fatal("ParseKnownPeers Err : ", err.Error())
	}
	// initialize services
	nodeConfigurationService = service.NewNodeConfigurationService(
		nodeAddressDynamic,
		nodeSecretPhrase,
		loggerCoreService,
		p2pUtil.NewHost(myAddress, peerPort, knownPeersResult),
	)
	blockchainStatusService = service.NewBlockchainStatusService(true, loggerCoreService)
	feeScaleService = fee.NewFeeScaleService(query.NewFeeScaleQuery(), query.NewBlockQuery(mainchain), queryExecutor)
	transactionUtil = &transaction.Util{
		FeeScaleService: feeScaleService,
	}

	nodeAddressInfoService = service.NewNodeAddressInfoService(
		queryExecutor,
		query.NewNodeRegistrationQuery(),
		query.NewNodeAddressInfoQuery(),
		loggerCoreService,
	)

	nodeRegistrationService = service.NewNodeRegistrationService(
		queryExecutor,
		query.NewNodeAddressInfoQuery(),
		query.NewAccountBalanceQuery(),
		query.NewNodeRegistrationQuery(),
		query.NewParticipationScoreQuery(),
		query.NewBlockQuery(mainchain),
		query.NewNodeAdmissionTimestampQuery(),
		loggerCoreService,
		blockchainStatusService,
		crypto.NewSignature(),
		nodeAddressInfoService,
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
	spineBlockManifestService = service.NewSpineBlockManifestService(
		queryExecutor,
		query.NewSpineBlockManifestQuery(),
		query.NewBlockQuery(spinechain),
		loggerCoreService,
	)
	fileService = service.NewFileService(
		loggerCoreService,
		new(codec.CborHandle),
		snapshotPath,
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
		query.NewPendingTransactionQuery(),
		query.NewPendingSignatureQuery(),
		query.NewMultisignatureInfoQuery(),
		query.NewSkippedBlocksmithQuery(),
		query.NewFeeScaleQuery(),
		query.NewFeeVoteCommitmentVoteQuery(),
		query.NewFeeVoteRevealVoteQuery(),
		query.NewLiquidPaymentTransactionQuery(),
		query.NewNodeAdmissionTimestampQuery(),
		query.NewBlockQuery(mainchain),
		query.GetSnapshotQuery(mainchain),
		query.GetBlocksmithSafeQuery(mainchain),
		query.GetDerivedQuery(mainchain),
		transactionUtil,
		&transaction.TypeSwitcher{Executor: queryExecutor},
	)

	snapshotService = service.NewSnapshotService(
		spineBlockManifestService,
		blockchainStatusService,
		snapshotBlockServices,
		loggerCoreService,
	)

	transactionCoreServiceIns = service.NewTransactionCoreService(
		loggerCoreService,
		queryExecutor,
		&transaction.TypeSwitcher{
			Executor: queryExecutor,
		},
		transactionUtil,
		query.NewTransactionQuery(mainchain),
		query.NewEscrowTransactionQuery(),
		query.NewPendingTransactionQuery(),
		query.NewLiquidPaymentTransactionQuery(),
	)

	nodeAuthValidationService = auth.NewNodeAuthValidation(
		crypto.NewSignature(),
	)

	defaultSignatureType = crypto.NewEd25519Signature()

	// initialize Observer
	observerInstance = observer.NewObserver()
	schedulerInstance = util.NewScheduler()
	initP2pInstance()
}

func loadNodeKey() {
	var (
		seed     string
		nodeKeys []*model.NodeKey
		err      error
	)
	// get the node private key
	nodeKeyFilePath = filepath.Join(nodeKeyPath, nodeKeyFile)
	nodeAdminKeysService := service.NewNodeAdminService(nil, nil, nil, nil, nodeKeyFilePath)
	nodeKeys, err = nodeAdminKeysService.ParseKeysFile()
	if err != nil {
		// fail parsing node_keys.json, could be wrong format or no file found, generating new one anyway
		seed = util.GetSecureRandomSeed()
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
		log.Fatal("could not find or generate node key")
	}
	nodeSecretPhrase = nodeKey.Seed
}

func loadNodeConfig(configPath, configFileName, configExtension string) {

	if err := util.LoadConfig(configPath, configFileName, configExtension); err != nil {
		panic(err)
	}

	myAddress = viper.GetString("myAddress")
	if myAddress == "" {
		ipAddr, err := (&util.IPUtil{}).DiscoverNodeAddress()
		if ipAddr == nil {
			// panic if we can't set an IP address for the node
			panic(err)
		} else if err != nil {
			// notify user that something went wrong in net address discovery process and its node might not behave properly on the network
			log.Print(err)
		}
		myAddress = ipAddr.String()
		nodeAddressDynamic = true
		viper.Set("myAddress", myAddress)
	}
	peerPort = viper.GetUint32("peerPort")
	monitoringPort = viper.GetInt("monitoringPort")
	apiRPCPort = viper.GetInt("apiRPCPort")
	maxAPIRequestPerSecond = viper.GetUint32("maxAPIRequestPerSecond")
	cpuProfilingPort = viper.GetInt("cpuProfilingPort")
	ownerAccountAddress = viper.GetString("ownerAccountAddress")
	wellknownPeers = viper.GetStringSlice("wellknownPeers")
	smithing = viper.GetBool("smithing")
	dbPath = viper.GetString("dbPath")
	dbName = viper.GetString("dbName")
	badgerDbPath = viper.GetString("badgerDbPath")
	badgerDbName = viper.GetString("badgerDbName")
	nodeKeyPath = viper.GetString("configPath")
	nodeKeyFile = viper.GetString("nodeKeyFile")
	apiCertFile = viper.GetString("apiapiCertFile")
	apiKeyFile = viper.GetString("apiKeyFile")
	snapshotPath = viper.GetString("snapshotPath")
	loadNodeKey()
	nodeAddress, err := address.EncodeZbcID(constant.PrefixZoobcNodeAccount, nodeKey.PublicKey)
	if err != nil {
		log.Fatal("fail to encode node public key")
	}
	// log the b64 encoded node public key
	log.Printf("peer to peer port: %d", peerPort)
	log.Printf("node monitoring port: %d", monitoringPort)
	log.Printf("client / wallet API port: %d", apiRPCPort)
	if cpuProfile {
		log.Printf("cpu profiling port: %d", cpuProfilingPort)
	}
	log.Printf("node's owner account address: %s", ownerAccountAddress)

	log.Printf("node's public key: %s", nodeAddress)
	log.Printf("well known peers: %s", strings.Join(wellknownPeers, ","))
	log.Printf("smithing: %v", smithing)
	if nodeAddressDynamic {
		log.Printf("node's ip address: %s (%s)", myAddress, "automatically discovered")
	} else {
		log.Printf("node's ip address: %s (%s)", myAddress, "set in configuration file")
	}
	if binaryChecksum, err := util.GetExecutableHash(); err == nil {
		log.Printf("binary checksum: %s", hex.EncodeToString(binaryChecksum))
	}
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
	// nodeConfigurationService.SetHost(p2pUtil.NewHost(myAddress, peerPort, knownPeersResult))

	// initialize peer client service
	nodePublicKey := defaultSignatureType.GetPublicKeyFromSeed(nodeSecretPhrase)
	peerServiceClient = client.NewPeerServiceClient(
		queryExecutor,
		query.NewNodeReceiptQuery(),
		nodePublicKey,
		nodeRegistrationService,
		query.NewBatchReceiptQuery(),
		query.NewMerkleTreeQuery(),
		receiptService,
		nodeConfigurationService,
		nodeAuthValidationService,
		loggerP2PService,
	)

	// peer discovery strategy
	peerExplorer = p2pStrategy.NewPriorityStrategy(
		peerServiceClient,
		nodeRegistrationService,
		queryExecutor,
		query.NewBlockQuery(mainchain),
		loggerP2PService,
		p2pStrategy.NewPeerStrategyHelper(),
		nodeConfigurationService,
		blockchainStatusService,
		crypto.NewSignature(),
	)
	p2pServiceInstance, _ = p2p.NewP2PService(
		peerServiceClient,
		peerExplorer,
		loggerP2PService,
		transactionUtil,
		fileService,
		nodeRegistrationService,
		nodeConfigurationService,
	)
	fileDownloader = p2p.NewFileDownloader(
		p2pServiceInstance,
		fileService,
		loggerP2PService,
		blockchainStatusService,
	)
}

func initObserverListeners() {
	// init observer listeners
	// broadcast block will be different than other listener implementation, since there are few exception condition
	observerInstance.AddListener(observer.BroadcastBlock, p2pServiceInstance.SendBlockListener())
	observerInstance.AddListener(observer.TransactionAdded, p2pServiceInstance.SendTransactionListener())
	// only smithing nodes generate snapshots
	if smithing {
		observerInstance.AddListener(observer.BlockPushed, snapshotService.StartSnapshotListener())
	}
	observerInstance.AddListener(observer.BlockRequestTransactions, p2pServiceInstance.RequestBlockTransactionsListener())
	observerInstance.AddListener(observer.ReceivedBlockTransactionsValidated, blockServices[0].ReceivedValidatedBlockTransactionsListener())
	observerInstance.AddListener(observer.BlockTransactionsRequested, blockServices[0].BlockTransactionsRequestedListener())
	observerInstance.AddListener(observer.SendBlockTransactions, p2pServiceInstance.SendBlockTransactionsListener())
}

func startServices() {
	p2pServiceInstance.StartP2P(
		myAddress,
		ownerAccountAddress,
		peerPort,
		nodeSecretPhrase,
		queryExecutor,
		blockServices,
		mempoolServices,
		fileService,
		nodeRegistrationService,
		nodeConfigurationService,
		nodeAddressInfoService,
		observerInstance,
	)
	api.Start(
		apiRPCPort,
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
		maxAPIRequestPerSecond,
	)
}

func startNodeMonitoring() {
	log.Infof("starting node monitoring at port:%d...", monitoringPort)
	monitoring.SetMonitoringActive(true)
	monitoring.SetNodePublicKey(defaultSignatureType.GetPublicKeyFromSeed(nodeSecretPhrase))
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", database.InstrumentBadgerMetrics(monitoring.Handler()))
		err := http.ListenAndServe(fmt.Sprintf(":%d", monitoringPort), mux)
		if err != nil {
			panic(fmt.Sprintf("failed to start monitoring service: %s", err))
		}
	}()
	// populate node address info counter when node starts
	if registeredNodesWithAddress, err := nodeRegistrationService.GetRegisteredNodesWithNodeAddress(); err == nil {
		monitoring.SetNodeAddressInfoCount(len(registeredNodesWithAddress))
	}
	if cna, err := nodeRegistrationService.CountNodesAddressByStatus(); err == nil {
		for status, counter := range cna {
			monitoring.SetNodeAddressStatusCount(counter, status)
		}
	}
}

func startMainchain() {
	var (
		lastBlockAtStart, blockToBuildScrambleNodes *model.Block
		err                                         error
		sleepPeriod                                 = constant.MainChainSmithIdlePeriod
	)
	monitoring.SetBlockchainStatus(mainchain, constant.BlockchainStatusIdle)
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
		mainchain,
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
		query.NewFeeVoteRevealVoteQuery(),
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
		feeScaleService,
		query.GetPruneQuery(mainchain),
		blockchainStatusService,
	)
	blockServices[mainchain.GetTypeInt()] = mainchainBlockService

	if !mainchainBlockService.CheckGenesis() { // Add genesis if not exist
		// genesis account will be inserted in the very beginning
		if err := service.AddGenesisAccount(queryExecutor); err != nil {
			loggerCoreService.Fatal("Fail to add genesis account")
		}
		// genesis next node admission timestamp will be inserted in the very beginning
		if err := service.AddGenesisNextNodeAdmission(
			queryExecutor,
			mainchain.GetGenesisBlockTimestamp(),
		); err != nil {
			loggerCoreService.Fatal(err)
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
		nodePublicKey := defaultSignatureType.GetPublicKeyFromSeed(nodeSecretPhrase)
		node, err := nodeRegistrationService.GetNodeRegistrationByNodePublicKey(nodePublicKey)
		if err != nil {
			loggerCoreService.Fatal(err)
		} else if node == nil {
			// no nodes registered with current node public key
			loggerCoreService.Error(
				"Current node is not in node registry and won't be able to smith until registered!",
			)
		}
		if node != nil {
			// register node config public key, so node registration service can detect if node has been admitted
			nodeRegistrationService.SetCurrentNodePublicKey(nodePublicKey)
			// default to isBlocksmith=true
			blockchainStatusService.SetIsBlocksmith(true)
			mainchainProcessor = smith.NewBlockchainProcessor(
				mainchainBlockService.GetChainType(),
				model.NewBlocksmith(nodeSecretPhrase, nodePublicKey, node.NodeID),
				mainchainBlockService,
				loggerCoreService,
				blockchainStatusService,
			)
			mainchainProcessor.Start(sleepPeriod)
		}
	}
	mainchainDownloader = blockchainsync.NewBlockchainDownloader(
		mainchainBlockService,
		peerServiceClient,
		peerExplorer,
		loggerCoreService,
		blockchainStatusService,
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
			loggerCoreService,
			queryExecutor,
			&transaction.TypeSwitcher{
				Executor: queryExecutor,
			},
			transactionUtil,
			query.NewTransactionQuery(mainchain),
			query.NewEscrowTransactionQuery(),
			query.NewPendingTransactionQuery(),
			query.NewLiquidPaymentTransactionQuery(),
		),
	}
	mainchainSynchronizer = blockchainsync.NewBlockchainSyncService(
		mainchainBlockService,
		peerServiceClient, peerExplorer,
		loggerCoreService,
		blockchainStatusService,
		mainchainDownloader,
		mainchainForkProcessor,
	)
}

func startSpinechain() {
	var (
		nodeID      int64
		sleepPeriod = constant.SpineChainSmithIdlePeriod
	)
	monitoring.SetBlockchainStatus(spinechain, constant.BlockchainStatusIdle)
	blocksmithStrategySpine := blockSmithStrategy.NewBlocksmithStrategySpine(
		queryExecutor,
		query.NewSpinePublicKeyQuery(),
		loggerCoreService,
		query.NewBlockQuery(spinechain),
	)
	spinechainBlocksmithService := service.NewBlocksmithService(
		query.NewAccountBalanceQuery(),
		query.NewAccountLedgerQuery(),
		query.NewNodeRegistrationQuery(),
		queryExecutor,
		spinechain,
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
		spinechainBlocksmithService,
		snapshotBlockServices[mainchain.GetTypeInt()],
		blockchainStatusService,
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
		nodePublicKey := defaultSignatureType.GetPublicKeyFromSeed(nodeSecretPhrase)
		// FIXME: ask @barton double check with him that generating a pseudo random id to compute the blockSeed is ok
		nodeID = int64(binary.LittleEndian.Uint64(nodePublicKey))
		spinechainProcessor = smith.NewBlockchainProcessor(
			spinechainBlockService.GetChainType(),
			model.NewBlocksmith(nodeSecretPhrase, nodePublicKey, nodeID),
			spinechainBlockService,
			loggerCoreService,
			blockchainStatusService,
		)
		spinechainProcessor.Start(sleepPeriod)
	}
	spinechainDownloader = blockchainsync.NewBlockchainDownloader(
		spinechainBlockService,
		peerServiceClient,
		peerExplorer,
		loggerCoreService,
		blockchainStatusService,
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
			loggerCoreService,
			queryExecutor,
			&transaction.TypeSwitcher{
				Executor: queryExecutor,
			},
			transactionUtil,
			query.NewTransactionQuery(mainchain),
			query.NewEscrowTransactionQuery(),
			query.NewPendingTransactionQuery(),
			query.NewLiquidPaymentTransactionQuery(),
		),
	}
	spinechainSynchronizer = blockchainsync.NewBlockchainSyncService(
		spinechainBlockService,
		peerServiceClient,
		peerExplorer,
		loggerCoreService,
		blockchainStatusService,
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
	blockchainOrchestrator := blockchainsync.NewBlockchainOrchestratorService(
		spinechainSynchronizer,
		mainchainSynchronizer,
		blockchainStatusService,
		spineBlockManifestService,
		fileDownloader,
		snapshotBlockServices[mainchain.GetTypeInt()],
		loggerCoreService)
	go func() {
		err := blockchainOrchestrator.Start()
		if err != nil {
			loggerCoreService.Fatal(err.Error())
			os.Exit(1)
		}
	}()
}

func main() {
	// start cpu profiling if enabled
	if cpuProfile {
		go func() {
			if err := http.ListenAndServe(fmt.Sprintf(":%d", cpuProfilingPort), nil); err != nil {
				log.Fatalf(fmt.Sprintf("failed to start profiling http server: %s", err))
			}
		}()
	}

	migration := database.Migration{Query: queryExecutor}
	if err := migration.Init(); err != nil {
		loggerCoreService.Fatal(err)
	}

	if err := migration.Apply(); err != nil {
		loggerCoreService.Fatal(err)
	}

	if isDebugMode {
		startNodeMonitoring()
		blocker.SetIsDebugMode(true)
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
				loggerCoreService.Info("All smith processors have stopped")
				shutdownCompleted <- true
			}
		case <-timeout:
			loggerCoreService.Info("ZOOBC Shutdown timedout...")
			os.Exit(1)
		case <-shutdownCompleted:
			loggerCoreService.Info("ZOOBC Shutdown complete")
			ticker.Stop()
			os.Exit(0)
		}
	}
}
