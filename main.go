package main

import (
	"crypto/sha256"
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
	"syscall"
	"time"

	"github.com/dgraph-io/badger/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/ugorji/go/codec"
	"github.com/zoobc/lib/address"
	"github.com/zoobc/zoobc-core/api"
	"github.com/zoobc/zoobc-core/common/auth"
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
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/blockchainsync"
	"github.com/zoobc/zoobc-core/core/scheduler"
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
	config                                                          *model.Config
	dbInstance                                                      *database.SqliteDB
	badgerDbInstance                                                *database.BadgerDB
	db                                                              *sql.DB
	badgerDb                                                        *badger.DB
	nodeShardStorage, mainBlockStateStorage, spineBlockStateStorage storage.CacheStorageInterface
	snapshotChunkUtil                                               util.ChunkUtilInterface
	p2pServiceInstance                                              p2p.Peer2PeerServiceInterface
	queryExecutor                                                   *query.Executor
	kvExecutor                                                      *kvdb.KVExecutor
	observerInstance                                                *observer.Observer
	schedulerInstance                                               *util.Scheduler
	snapshotSchedulers                                              *scheduler.SnapshotScheduler
	blockServices                                                   = make(map[int32]service.BlockServiceInterface)
	snapshotBlockServices                                           = make(map[int32]service.SnapshotBlockServiceInterface)
	blockStateStorages                                              = make(map[int32]storage.CacheStorageInterface)
	mainchainBlockService                                           *service.BlockService
	spinePublicKeyService                                           *service.BlockSpinePublicKeyService
	mainBlockSnapshotChunkStrategy                                  service.SnapshotChunkStrategyInterface
	spinechainBlockService                                          *service.BlockSpineService
	fileDownloader                                                  p2p.FileDownloaderInterface
	mempoolServices                                                 = make(map[int32]service.MempoolServiceInterface)
	blockIncompleteQueueService                                     service.BlockIncompleteQueueServiceInterface
	receiptService                                                  service.ReceiptServiceInterface
	peerServiceClient                                               client.PeerServiceClientInterface
	peerExplorer                                                    p2pStrategy.PeerExplorerStrategyInterface
	isDebugMode, useEnvVar                                          bool
	nodeRegistrationService                                         service.NodeRegistrationServiceInterface
	nodeAuthValidationService                                       auth.NodeAuthValidationInterface
	mainchainProcessor                                              smith.BlockchainProcessorInterface
	spinechainProcessor                                             smith.BlockchainProcessorInterface
	loggerAPIService                                                *log.Logger
	loggerCoreService                                               *log.Logger
	loggerP2PService                                                *log.Logger
	loggerScheduler                                                 *log.Logger
	spinechainSynchronizer, mainchainSynchronizer                   blockchainsync.BlockchainSyncServiceInterface
	spineBlockManifestService                                       service.SpineBlockManifestServiceInterface
	snapshotService                                                 service.SnapshotServiceInterface
	transactionUtil                                                 transaction.UtilInterface
	receiptUtil                                                     = &coreUtil.ReceiptUtil{}
	transactionCoreServiceIns                                       service.TransactionCoreServiceInterface
	fileService                                                     service.FileServiceInterface
	mainchain                                                       = &chaintype.MainChain{}
	spinechain                                                      = &chaintype.SpineChain{}
	blockchainStatusService                                         service.BlockchainStatusServiceInterface
	nodeConfigurationService                                        service.NodeConfigurationServiceInterface
	nodeAddressInfoService                                          service.NodeAddressInfoServiceInterface
	feeScaleService                                                 fee.FeeScaleServiceInterface
	mainchainDownloader, spinechainDownloader                       blockchainsync.BlockchainDownloadInterface
	mainchainForkProcessor, spinechainForkProcessor                 blockchainsync.ForkingProcessorInterface
	cpuProfile                                                      bool
	cliMonitoring                                                   monitoring.CLIMonitoringInteface
)

func init() {
	var (
		configPostfix string
		configPath    string
		err           error
	)
	// parse custom flag in running the node
	flag.StringVar(&configPostfix, "config-postfix", "", "Usage")
	flag.StringVar(&configPath, "config-path", "./", "Usage")
	flag.BoolVar(&isDebugMode, "debug", false, "Usage")
	flag.BoolVar(&cpuProfile, "cpu-profile", false, "if this flag is used, write cpu profile to file")
	flag.BoolVar(&useEnvVar, "use-env", false, "if this flag is enabled, node can run without config file")
	flag.Parse()

	// spawn config object
	config = model.NewConfig()
	// load config for default value to be feed to viper
	if err := util.LoadConfig(configPath, "config"+configPostfix, "toml"); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok && useEnvVar {
			config.ConfigFileExist = true
		}
	} else {
		config.ConfigFileExist = true
	}
	// assign read configuration to config object
	config.LoadConfigurations()

	// early init configuration service
	nodeConfigurationService = service.NewNodeConfigurationService(loggerCoreService)

	// check and validate configurations
	err = util.NewSetupNode(config).CheckConfig()
	if err != nil {
		log.Fatalf("Unknown error occurred - error: %s", err.Error())
		return
	}
	nodeAdminKeysService := service.NewNodeAdminService(nil, nil, nil, nil,
		filepath.Join(config.ResourcePath, config.NodeKeyFileName))
	if len(config.NodeKey.Seed) > 0 {
		config.NodeKey.PublicKey, err = nodeAdminKeysService.GenerateNodeKey(config.NodeKey.Seed)
		if err != nil {
			log.Fatal("Fail to generate node key")
		}
	} else { // setup wizard don't set node key, meaning ./resource/node_keys.json exist
		nodeKeys, err := nodeAdminKeysService.ParseKeysFile()
		if err != nil {
			log.Fatal("existing node keys has wrong format, please fix it or delete it, then re-run the application")
		}
		config.NodeKey = nodeAdminKeysService.GetLastNodeKey(nodeKeys)
	}

	knownPeersResult, err := p2pUtil.ParseKnownPeers(config.WellknownPeers)
	if err != nil {
		loggerCoreService.Fatal("ParseKnownPeers Err : ", err.Error())
	}

	nodeConfigurationService.SetHost(p2pUtil.NewHost(config.MyAddress, config.PeerPort, knownPeersResult,
		constant.ApplicationVersion, constant.ApplicationCodeName))
	nodeConfigurationService.SetIsMyAddressDynamic(config.IsNodeAddressDynamic)
	if config.NodeKey.Seed == "" {
		loggerCoreService.Fatal("node seed is empty", err.Error())
	}
	nodeConfigurationService.SetNodeSeed(config.NodeKey.Seed)

	if config.OwnerAccountAddress == "" {
		// todo: andy-shi88 refactor this
		ed25519 := crypto.NewEd25519Signature()
		accountPrivateKey, err := ed25519.GetPrivateKeyFromSeedUseSlip10(
			config.NodeKey.Seed,
		)
		if err != nil {
			log.Fatal("Fail to generate account private key")
		}
		publicKey, err := ed25519.GetPublicKeyFromPrivateKeyUseSlip10(accountPrivateKey)
		if err != nil {
			log.Fatal("Fail to generate account public key")
		}
		id, err := address.EncodeZbcID(constant.PrefixZoobcDefaultAccount, publicKey)
		if err != nil {
			log.Fatal("Fail generating address from node's seed")
		}
		config.OwnerAccountAddress = id
		err = config.SaveConfig()
		if err != nil {
			log.Fatal("Fail to save new configuration")
		}
	}
	if binaryChecksum, err := util.GetExecutableHash(); err == nil {
		log.Printf("binary checksum: %s", hex.EncodeToString(binaryChecksum))
	}

	initLogInstance()
	cliMonitoring = monitoring.NewCLIMonitoring(config)
	monitoring.SetCLIMonitoring(cliMonitoring)

	// initialize/open db and queryExecutor
	dbInstance = database.NewSqliteDB()
	if err := dbInstance.InitializeDB(config.ResourcePath, config.DatabaseFileName); err != nil {
		loggerCoreService.Fatal(err)
	}
	db, err = dbInstance.OpenDB(
		config.ResourcePath,
		config.DatabaseFileName,
		constant.SQLMaxOpenConnetion,
		constant.SQLMaxIdleConnections,
		constant.SQLMaxConnectionLifetime,
	)

	if err != nil {
		loggerCoreService.Fatal(err)
	}
	// initialize k-v db
	badgerDbInstance = database.NewBadgerDB()
	if err := badgerDbInstance.InitializeBadgerDB(config.ResourcePath, config.BadgerDbName); err != nil {
		loggerCoreService.Fatal(err)
	}
	badgerDb, err = badgerDbInstance.OpenBadgerDB(config.ResourcePath, config.BadgerDbName)
	if err != nil {
		loggerCoreService.Fatal(err)
	}
	queryExecutor = query.NewQueryExecutor(db)
	kvExecutor = kvdb.NewKVExecutor(badgerDb)

	// initialize services
	mainBlockStateStorage = storage.NewBlockStateStorage()
	spineBlockStateStorage = storage.NewBlockStateStorage()
	blockStateStorages[mainchain.GetTypeInt()] = mainBlockStateStorage
	blockStateStorages[spinechain.GetTypeInt()] = spineBlockStateStorage

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
		config.SnapshotPath,
	)
	mainBlockSnapshotChunkStrategy = service.NewSnapshotBasicChunkStrategy(
		constant.SnapshotChunkSize,
		fileService,
	)
	snapshotBlockServices[mainchain.GetTypeInt()] = service.NewSnapshotMainBlockService(
		config.SnapshotPath,
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

	spinePublicKeyService = service.NewBlockSpinePublicKeyService(
		crypto.NewSignature(),
		queryExecutor,
		query.NewNodeRegistrationQuery(),
		query.NewSpinePublicKeyQuery(),
		loggerCoreService,
	)
	// initialize Observer
	observerInstance = observer.NewObserver()
	schedulerInstance = util.NewScheduler(loggerScheduler)
	snapshotChunkUtil = util.NewChunkUtil(sha256.Size, nodeShardStorage, loggerScheduler)

	initP2pInstance()

	/*
		Snapshot Scheduler initiate
	*/
	nodeShardStorage = storage.NewNodeShardCacheStorage()
	snapshotSchedulers = scheduler.NewSnapshotScheduler(
		spineBlockManifestService,
		fileService,
		snapshotChunkUtil,
		nodeShardStorage,
		mainBlockStateStorage,
		blockServices[0],
		&service.BlockSpinePublicKeyService{
			Signature:             crypto.NewSignature(),
			QueryExecutor:         queryExecutor,
			NodeRegistrationQuery: query.NewNodeRegistrationQuery(),
			SpinePublicKeyQuery:   query.NewSpinePublicKeyQuery(),
			Logger:                loggerCoreService,
		},
		nodeConfigurationService,
		fileDownloader,
	)
}

func initLogInstance() {
	var (
		err       error
		logLevels = viper.GetStringSlice("logLevels")
		t         = time.Now().Format("2-Jan-2006_")
	)

	if loggerAPIService, err = util.InitLogger(".log/", t+"APIdebug.log", logLevels, config.LogOnCli); err != nil {
		panic(err)
	}
	if loggerCoreService, err = util.InitLogger(".log/", t+"Coredebug.log", logLevels, config.LogOnCli); err != nil {
		panic(err)
	}
	if loggerP2PService, err = util.InitLogger(".log/", t+"P2Pdebug.log", logLevels, config.LogOnCli); err != nil {
		panic(err)
	}
	if loggerScheduler, err = util.InitLogger(".log/", t+"Scheduler.log", logLevels, config.LogOnCli); err != nil {
		panic(err)
	}
}

func initP2pInstance() {
	// initialize peer client service
	peerServiceClient = client.NewPeerServiceClient(
		queryExecutor,
		query.NewNodeReceiptQuery(),
		config.NodeKey.PublicKey,
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
		blockchainStatusService,
		spinePublicKeyService,
		snapshotChunkUtil,
		loggerP2PService,
	)
}

func initObserverListeners() {
	// init observer listeners
	// broadcast block will be different than other listener implementation, since there are few exception condition
	observerInstance.AddListener(observer.BroadcastBlock, p2pServiceInstance.SendBlockListener())
	observerInstance.AddListener(observer.TransactionAdded, p2pServiceInstance.SendTransactionListener())
	// only smithing nodes generate snapshots
	if config.Smithing {
		observerInstance.AddListener(observer.BlockPushed, snapshotService.StartSnapshotListener())
	}
	observerInstance.AddListener(observer.BlockRequestTransactions, p2pServiceInstance.RequestBlockTransactionsListener())
	observerInstance.AddListener(observer.ReceivedBlockTransactionsValidated, blockServices[0].ReceivedValidatedBlockTransactionsListener())
	observerInstance.AddListener(observer.BlockTransactionsRequested, blockServices[0].BlockTransactionsRequestedListener())
	observerInstance.AddListener(observer.SendBlockTransactions, p2pServiceInstance.SendBlockTransactionsListener())
}

func startServices() {
	p2pServiceInstance.StartP2P(
		config.MyAddress,
		config.OwnerAccountAddress,
		config.PeerPort,
		config.NodeKey.Seed,
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
		config.RPCAPIPort,
		config.HTTPAPIPort,
		kvExecutor,
		queryExecutor,
		p2pServiceInstance,
		blockServices,
		nodeRegistrationService,
		config.OwnerAccountAddress,
		filepath.Join(config.ResourcePath, config.NodeKeyFileName),
		loggerAPIService,
		isDebugMode,
		config.APICertFile,
		config.APIKeyFile,
		transactionUtil,
		receiptUtil,
		receiptService,
		transactionCoreServiceIns,
		config.MaxAPIRequestPerSecond,
		blockStateStorages,
	)
}

func startNodeMonitoring() {
	log.Infof("starting node monitoring at port:%d...", config.MonitoringPort)
	monitoring.SetMonitoringActive(true)
	monitoring.SetNodePublicKey(config.NodeKey.PublicKey)
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", database.InstrumentBadgerMetrics(monitoring.Handler()))
		err := http.ListenAndServe(fmt.Sprintf(":%d", config.MonitoringPort), mux)
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
		mainchain,
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
		mainBlockStateStorage,
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
	cliMonitoring.UpdateBlockState(mainchain, lastBlockAtStart)

	err = mainBlockStateStorage.SetItem(0, *lastBlockAtStart)
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

	if len(config.NodeKey.Seed) > 0 && config.Smithing {
		node, err := nodeRegistrationService.GetNodeRegistrationByNodePublicKey(config.NodeKey.PublicKey)
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
			nodeRegistrationService.SetCurrentNodePublicKey(config.NodeKey.PublicKey)
			// default to isBlocksmith=true
			blockchainStatusService.SetIsBlocksmith(true)
			mainchainProcessor = smith.NewBlockchainProcessor(
				mainchainBlockService.GetChainType(),
				model.NewBlocksmith(config.NodeKey.Seed, config.NodeKey.PublicKey, node.NodeID),
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
		crypto.NewSignature(),
		observerInstance,
		blocksmithStrategySpine,
		loggerCoreService,
		query.NewSpineBlockManifestQuery(),
		spinechainBlocksmithService,
		snapshotBlockServices[mainchain.GetTypeInt()],
		spineBlockStateStorage,
		blockchainStatusService,
		spinePublicKeyService,
	)
	blockServices[spinechain.GetTypeInt()] = spinechainBlockService

	if !spinechainBlockService.CheckGenesis() { // Add genesis if not exist
		if err := spinechainBlockService.AddGenesis(); err != nil {
			loggerCoreService.Fatal(err)
		}
	}
	lastBlockAtStart, err := spinechainBlockService.GetLastBlock()
	if err != nil {
		loggerCoreService.Fatal(err)
	}
	cliMonitoring.UpdateBlockState(spinechain, lastBlockAtStart)
	err = spineBlockStateStorage.SetItem(0, *lastBlockAtStart)
	if err != nil {
		loggerCoreService.Fatal(err)
	}
	// Note: spine blocks smith even if smithing is false, because are created by every running node
	// 		 Later we only broadcast (and accumulate) signatures of the ones who can smith
	if len(config.NodeKey.Seed) > 0 && config.Smithing {
		// FIXME: ask @barton double check with him that generating a pseudo random id to compute the blockSeed is ok
		nodeID = int64(binary.LittleEndian.Uint64(config.NodeKey.PublicKey))
		spinechainProcessor = smith.NewBlockchainProcessor(
			spinechainBlockService.GetChainType(),
			model.NewBlocksmith(config.NodeKey.Seed, config.NodeKey.PublicKey, nodeID),
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
	// scheduler to generate receipt merkle root
	if err := schedulerInstance.AddJob(
		constant.ReceiptGenerateMarkleRootPeriod,
		receiptService.GenerateReceiptsMerkleRoot,
	); err != nil {
		loggerCoreService.Error("Scheduler Err : ", err.Error())
	}
	// scheduler to remove block uncompleted queue that already waiting transactions too long
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

	if err := schedulerInstance.AddJob(
		constant.SnapshotSchedulerUnmaintainedChunksPeriod,
		snapshotSchedulers.DeleteUnmaintainedChunks,
	); err != nil {
		loggerCoreService.Error("Scheduler Err: ", err.Error())
	}

	if err := schedulerInstance.AddJob(
		constant.SnapshotSchedulerUnmaintainedChunksPeriod,
		snapshotSchedulers.CheckChunksIntegrity,
	); err != nil {
		loggerCoreService.Error("Scheduler Err: ", err.Error())
	}
}

func startBlockchainSynchronizers() {
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
			if err := http.ListenAndServe(fmt.Sprintf(":%d", config.CPUProfilingPort), nil); err != nil {
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
	go startBlockchainSynchronizers()

	if !config.LogOnCli && config.CliMonitoring {
		go cliMonitoring.Start()
	}

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
