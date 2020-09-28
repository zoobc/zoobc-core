package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/takama/daemon"
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
	config                                                                 *model.Config
	dbInstance                                                             *database.SqliteDB
	db                                                                     *sql.DB
	nodeShardStorage, mainBlockStateStorage, spineBlockStateStorage        storage.CacheStorageInterface
	nextNodeAdmissionStorage, mempoolStorage, receiptReminderStorage       storage.CacheStorageInterface
	mempoolBackupStorage, batchReceiptCacheStorage                         storage.CacheStorageInterface
	activeNodeRegistryCacheStorage, pendingNodeRegistryCacheStorage        storage.CacheStorageInterface
	nodeAddressInfoStorage                                                 *storage.NodeAddressInfoStorage
	scrambleNodeStorage                                                    storage.CacheStackStorageInterface
	blockStateStorages                                                     = make(map[int32]storage.CacheStorageInterface)
	snapshotChunkUtil                                                      util.ChunkUtilInterface
	p2pServiceInstance                                                     p2p.Peer2PeerServiceInterface
	queryExecutor                                                          *query.Executor
	observerInstance                                                       *observer.Observer
	schedulerInstance                                                      *util.Scheduler
	snapshotSchedulers                                                     *scheduler.SnapshotScheduler
	blockServices                                                          = make(map[int32]service.BlockServiceInterface)
	snapshotBlockServices                                                  = make(map[int32]service.SnapshotBlockServiceInterface)
	mainchainBlockService                                                  *service.BlockService
	spinePublicKeyService                                                  *service.BlockSpinePublicKeyService
	mainBlockSnapshotChunkStrategy                                         service.SnapshotChunkStrategyInterface
	spinechainBlockService                                                 *service.BlockSpineService
	fileDownloader                                                         p2p.FileDownloaderInterface
	mempoolServices                                                        = make(map[int32]service.MempoolServiceInterface)
	blockIncompleteQueueService                                            service.BlockIncompleteQueueServiceInterface
	receiptService                                                         service.ReceiptServiceInterface
	peerServiceClient                                                      client.PeerServiceClientInterface
	peerExplorer                                                           p2pStrategy.PeerExplorerStrategyInterface
	nodeRegistrationService                                                service.NodeRegistrationServiceInterface
	nodeAuthValidationService                                              auth.NodeAuthValidationInterface
	mainchainProcessor                                                     smith.BlockchainProcessorInterface
	spinechainProcessor                                                    smith.BlockchainProcessorInterface
	loggerAPIService, loggerCoreService, loggerP2PService, loggerScheduler *log.Logger
	spinechainSynchronizer, mainchainSynchronizer                          blockchainsync.BlockchainSyncServiceInterface
	spineBlockManifestService                                              service.SpineBlockManifestServiceInterface
	snapshotService                                                        service.SnapshotServiceInterface
	transactionUtil                                                        transaction.UtilInterface
	receiptUtil                                                            = &coreUtil.ReceiptUtil{}
	transactionCoreServiceIns                                              service.TransactionCoreServiceInterface
	pendingTransactionServiceIns                                           service.PendingTransactionServiceInterface
	fileService                                                            service.FileServiceInterface
	mainchain                                                              = &chaintype.MainChain{}
	spinechain                                                             = &chaintype.SpineChain{}
	blockchainStatusService                                                service.BlockchainStatusServiceInterface
	nodeConfigurationService                                               service.NodeConfigurationServiceInterface
	nodeAddressInfoService                                                 service.NodeAddressInfoServiceInterface
	mempoolService                                                         service.MempoolServiceInterface
	mainchainPublishedReceiptService                                       service.PublishedReceiptServiceInterface
	mainchainPublishedReceiptUtil                                          coreUtil.PublishedReceiptUtilInterface
	mainchainCoinbaseService                                               service.CoinbaseServiceInterface
	mainchainBlocksmithService                                             service.BlocksmithServiceInterface
	mainchainParticipationScoreService                                     service.ParticipationScoreServiceInterface
	scrambleNodeService                                                    service.ScrambleNodeServiceInterface
	actionSwitcher                                                         transaction.TypeActionSwitcher
	feeScaleService                                                        fee.FeeScaleServiceInterface
	mainchainDownloader, spinechainDownloader                              blockchainsync.BlockchainDownloadInterface
	mainchainForkProcessor, spinechainForkProcessor                        blockchainsync.ForkingProcessorInterface
	cliMonitoring                                                          monitoring.CLIMonitoringInteface
)
var (
	flagConfigPath, flagConfigPostfix, flagResourcePath string
	flagDebugMode, flagProfiling, flagUseEnv            bool
	daemonCommand                                       = &cobra.Command{
		Use:        "daemon",
		Short:      "Run node on daemon service, which mean running in the background. Similar to launchd or systemd",
		Example:    "daemon install | start | stop | remove | status",
		SuggestFor: []string{"up", "stats", "delete", "deamon", "demon"},
	}
	runCommand = &cobra.Command{
		Use:   "run",
		Short: "Run node without daemon.",
	}
	rootCmd *cobra.Command
)

// goDaemon instance that needed  to implement whole method of daemon
type goDaemon struct {
	daemon.Daemon
}

// initiateMainInstance initiation all instance that must be needed and exists before running the node
func initiateMainInstance() {
	var (
		err error
	)

	// load config for default value to be feed to viper
	if err = util.LoadConfig(flagConfigPath, "config"+flagConfigPostfix, "toml", flagResourcePath); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok && flagUseEnv {
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
	} else {
		// setup wizard don't set node key, meaning ./resource/node_keys.json exist
		nodeKeys, err := nodeAdminKeysService.ParseKeysFile()
		if err != nil {
			log.Fatal("existing node keys has wrong format, please fix it or delete it, then re-run the application")
		}
		config.NodeKey = nodeAdminKeysService.GetLastNodeKey(nodeKeys)
	}

	knownPeersResult, err := p2pUtil.ParseKnownPeers(config.WellknownPeers)
	if err != nil {
		log.Fatalf("ParseKnownPeers Err: %s", err.Error())
	}

	nodeConfigurationService.SetHost(p2pUtil.NewHost(config.MyAddress, config.PeerPort, knownPeersResult,
		constant.ApplicationVersion, constant.ApplicationCodeName))
	nodeConfigurationService.SetIsMyAddressDynamic(config.IsNodeAddressDynamic)
	if config.NodeKey.Seed == "" {
		log.Fatal("node seed is empty")
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
		err = config.SaveConfig(flagConfigPath)
		if err != nil {
			log.Fatal("Fail to save new configuration")
		}
	}
	cliMonitoring = monitoring.NewCLIMonitoring(config)
	monitoring.SetCLIMonitoring(cliMonitoring)
	initLogInstance(fmt.Sprintf("%s/.log", config.ResourcePath))

	// break
	// initialize/open db and queryExecutor
	dbInstance = database.NewSqliteDB()
	if err = dbInstance.InitializeDB(config.ResourcePath, config.DatabaseFileName); err != nil {
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
	queryExecutor = query.NewQueryExecutor(db)

	// initialize cache storage
	mainBlockStateStorage = storage.NewBlockStateStorage()
	spineBlockStateStorage = storage.NewBlockStateStorage()
	blockStateStorages[mainchain.GetTypeInt()] = mainBlockStateStorage
	blockStateStorages[spinechain.GetTypeInt()] = spineBlockStateStorage
	nextNodeAdmissionStorage = storage.NewNodeAdmissionTimestampStorage()
	nodeShardStorage = storage.NewNodeShardCacheStorage()
	mempoolStorage = storage.NewMempoolStorage()
	scrambleNodeStorage = storage.NewScrambleCacheStackStorage()
	receiptReminderStorage = storage.NewReceiptReminderStorage()
	mempoolBackupStorage = storage.NewMempoolBackupStorage()
	batchReceiptCacheStorage = storage.NewBatchReceiptCacheStorage()
	nodeAddressInfoStorage = storage.NewNodeAddressInfoStorage()
	// store current active node registry (not in queue)
	activeNodeRegistryCacheStorage = storage.NewNodeRegistryCacheStorage(
		monitoring.TypeActiveNodeRegistryStorage,
		func(registries []storage.NodeRegistry) {
			sort.SliceStable(registries, func(i, j int) bool {
				// sort by nodeID lowest - highest
				return registries[i].Node.GetNodeID() < registries[j].Node.GetNodeID()
			})
		})
	// store pending node registry
	pendingNodeRegistryCacheStorage = storage.NewNodeRegistryCacheStorage(
		monitoring.TypePendingNodeRegistryStorage,
		func(registries []storage.NodeRegistry) {
			sort.SliceStable(registries, func(i, j int) bool {
				// sort by locked balance highest - lowest
				return registries[i].Node.GetLockedBalance() > registries[j].Node.GetLockedBalance()
			})
		},
	)
	// initialize services
	blockchainStatusService = service.NewBlockchainStatusService(true, loggerCoreService)
	feeScaleService = fee.NewFeeScaleService(query.NewFeeScaleQuery(), query.NewBlockQuery(mainchain), queryExecutor)
	transactionUtil = &transaction.Util{
		FeeScaleService:     feeScaleService,
		MempoolCacheStorage: mempoolStorage,
	}
	// initialize Observer
	observerInstance = observer.NewObserver()
	schedulerInstance = util.NewScheduler(loggerScheduler)
	snapshotChunkUtil = util.NewChunkUtil(sha256.Size, nodeShardStorage, loggerScheduler)
	nodeAuthValidationService = auth.NewNodeAuthValidation(
		crypto.NewSignature(),
	)
	actionSwitcher = &transaction.TypeSwitcher{
		Executor:                   queryExecutor,
		MempoolCacheStorage:        mempoolStorage,
		NodeAddressInfoStorage:     nodeAddressInfoStorage,
		NodeAuthValidation:         nodeAuthValidationService,
		ActiveNodeRegistryStorage:  activeNodeRegistryCacheStorage,
		PendingNodeRegistryStorage: pendingNodeRegistryCacheStorage,
	}

	nodeAddressInfoService = service.NewNodeAddressInfoService(
		queryExecutor,
		query.NewNodeAddressInfoQuery(),
		query.NewNodeRegistrationQuery(),
		query.NewBlockQuery(mainchain),
		crypto.NewSignature(),
		nodeAddressInfoStorage,
		mainBlockStateStorage,
		activeNodeRegistryCacheStorage,
		loggerCoreService,
	)
	nodeRegistrationService = service.NewNodeRegistrationService(
		queryExecutor,
		query.NewAccountBalanceQuery(),
		query.NewNodeRegistrationQuery(),
		query.NewParticipationScoreQuery(),
		query.NewNodeAdmissionTimestampQuery(),
		loggerCoreService,
		blockchainStatusService,
		nodeAddressInfoService,
		nextNodeAdmissionStorage,
		activeNodeRegistryCacheStorage,
		pendingNodeRegistryCacheStorage,
	)
	scrambleNodeService = service.NewScrambleNodeService(
		nodeRegistrationService,
		nodeAddressInfoService,
		queryExecutor,
		query.NewBlockQuery(mainchain),
		scrambleNodeStorage,
	)

	receiptService = service.NewReceiptService(
		query.NewNodeReceiptQuery(),
		query.NewMerkleTreeQuery(),
		query.NewNodeRegistrationQuery(),
		query.NewBlockQuery(mainchain),
		queryExecutor,
		nodeRegistrationService,
		crypto.NewSignature(),
		query.NewPublishedReceiptQuery(),
		receiptUtil,
		mainBlockStateStorage,
		receiptReminderStorage,
		batchReceiptCacheStorage,
		scrambleNodeService,
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
	mainchainBlocksmithService = service.NewBlocksmithService(
		query.NewAccountBalanceQuery(),
		query.NewAccountLedgerQuery(),
		query.NewNodeRegistrationQuery(),
		queryExecutor,
		mainchain,
	)
	mainchainCoinbaseService = service.NewCoinbaseService(
		query.NewNodeRegistrationQuery(),
		queryExecutor,
		mainchain,
	)
	mainchainParticipationScoreService = service.NewParticipationScoreService(
		query.NewParticipationScoreQuery(),
		queryExecutor,
	)
	mainchainPublishedReceiptUtil = coreUtil.NewPublishedReceiptUtil(
		query.NewPublishedReceiptQuery(),
		queryExecutor,
	)
	mainchainPublishedReceiptService = service.NewPublishedReceiptService(
		query.NewPublishedReceiptQuery(),
		receiptUtil,
		mainchainPublishedReceiptUtil,
		receiptService,
		queryExecutor,
	)
	transactionCoreServiceIns = service.NewTransactionCoreService(
		loggerCoreService,
		queryExecutor,
		actionSwitcher,
		transactionUtil,
		query.NewTransactionQuery(mainchain),
		query.NewEscrowTransactionQuery(),
		query.NewLiquidPaymentTransactionQuery(),
	)
	pendingTransactionServiceIns = service.NewPendingTransactionService(
		loggerCoreService,
		queryExecutor,
		actionSwitcher,
		transactionUtil,
		query.NewTransactionQuery(mainchain),
		query.NewPendingTransactionQuery(),
	)

	mempoolService = service.NewMempoolService(
		transactionUtil,
		mainchain,
		queryExecutor,
		query.NewMempoolQuery(mainchain),
		query.NewMerkleTreeQuery(),
		actionSwitcher,
		query.NewAccountBalanceQuery(),
		query.NewTransactionQuery(mainchain),
		crypto.NewSignature(),
		observerInstance,
		loggerCoreService,
		receiptUtil,
		receiptService,
		transactionCoreServiceIns,
		mainBlockStateStorage,
		mempoolStorage,
		mempoolBackupStorage,
	)

	mainchainBlockService = service.NewBlockMainService(
		mainchain,
		queryExecutor,
		query.NewBlockQuery(mainchain),
		query.NewMempoolQuery(mainchain),
		query.NewTransactionQuery(mainchain),
		query.NewSkippedBlocksmithQuery(),
		crypto.NewSignature(),
		mempoolService,
		receiptService,
		nodeRegistrationService,
		nodeAddressInfoService,
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
		transactionUtil, receiptUtil,
		mainchainPublishedReceiptUtil,
		transactionCoreServiceIns,
		pendingTransactionServiceIns,
		mainchainBlockPool,
		mainchainBlocksmithService,
		mainchainCoinbaseService,
		mainchainParticipationScoreService,
		mainchainPublishedReceiptService,
		feeScaleService,
		query.GetPruneQuery(mainchain),
		mainBlockStateStorage,
		blockchainStatusService,
		scrambleNodeService,
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
		actionSwitcher,
		mainchainBlockService,
		nodeRegistrationService,
		scrambleNodeService,
	)

	snapshotService = service.NewSnapshotService(
		spineBlockManifestService,
		blockchainStatusService,
		snapshotBlockServices,
		loggerCoreService,
	)

	spinePublicKeyService = service.NewBlockSpinePublicKeyService(
		crypto.NewSignature(),
		queryExecutor,
		query.NewNodeRegistrationQuery(),
		query.NewSpinePublicKeyQuery(),
		loggerCoreService,
	)
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

	initP2pInstance()

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

	/*
		Snapshot Scheduler initiate
	*/
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
	// assign chain services to the map
	mempoolServices[mainchain.GetTypeInt()] = mempoolService
	blockServices[mainchain.GetTypeInt()] = mainchainBlockService
	blockServices[spinechain.GetTypeInt()] = spinechainBlockService
	// register event listeners
	initObserverListeners()

}

func initLogInstance(logPath string) {
	var (
		err       error
		logLevels = viper.GetStringSlice("logLevels")
		t         = time.Now().Format("01-02-2006_")
	)

	if loggerAPIService, err = util.InitLogger(logPath, t+"APIdebug.log", logLevels, config.LogOnCli); err != nil {
		panic(err)
	}
	if loggerCoreService, err = util.InitLogger(logPath, t+"Coredebug.log", logLevels, config.LogOnCli); err != nil {
		panic(err)
	}
	if loggerP2PService, err = util.InitLogger(logPath, t+"P2Pdebug.log", logLevels, config.LogOnCli); err != nil {
		panic(err)
	}
	if loggerScheduler, err = util.InitLogger(logPath, t+"Scheduler.log", logLevels, config.LogOnCli); err != nil {
		panic(err)
	}
}

func initP2pInstance() {
	// initialize peer client service
	peerServiceClient = client.NewPeerServiceClient(
		queryExecutor, query.NewNodeReceiptQuery(),
		config.NodeKey.PublicKey,
		nodeRegistrationService,
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
		nodeAddressInfoService,
		mainchainBlockService,
		loggerP2PService,
		p2pStrategy.NewPeerStrategyHelper(),
		nodeConfigurationService,
		blockchainStatusService,
		crypto.NewSignature(),
		scrambleNodeService,
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
	observerInstance.AddListener(observer.ReceivedBlockTransactionsValidated, mainchainBlockService.ReceivedValidatedBlockTransactionsListener())
	observerInstance.AddListener(observer.BlockTransactionsRequested, mainchainBlockService.BlockTransactionsRequestedListener())
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
		queryExecutor,
		p2pServiceInstance,
		blockServices,
		nodeRegistrationService,
		nodeAddressInfoService,
		mempoolService,
		scrambleNodeService,
		transactionUtil,
		actionSwitcher,
		blockStateStorages,
		config.RPCAPIPort,
		config.HTTPAPIPort,
		config.OwnerAccountAddress,
		filepath.Join(config.ResourcePath, config.NodeKeyFileName),
		loggerAPIService,
		flagDebugMode,
		config.APICertFile,
		config.APIKeyFile,
		config.MaxAPIRequestPerSecond,
		config.NodeKey.PublicKey,
	)
}

func startNodeMonitoring() {
	log.Infof("starting node monitoring at port:%d...", config.MonitoringPort)
	monitoring.SetMonitoringActive(true)
	monitoring.SetNodePublicKey(config.NodeKey.PublicKey)
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", monitoring.Handler())
		err := http.ListenAndServe(fmt.Sprintf(":%d", config.MonitoringPort), mux)
		if err != nil {
			panic(fmt.Sprintf("failed to start monitoring service: %s", err))
		}
	}()
	// populate node address info counter when node starts
	if nodeCount, err := nodeAddressInfoService.CountRegistredNodeAddressWithAddressInfo(); err == nil {
		monitoring.SetNodeAddressInfoCount(nodeCount)
	}
	if cna, err := nodeAddressInfoService.CountNodesAddressByStatus(); err == nil {
		for status, counter := range cna {
			monitoring.SetNodeAddressStatusCount(counter, status)
		}
	}
}

func startMainchain() {
	var (
		lastBlockAtStart *model.Block
		err              error
		sleepPeriod      = constant.MainChainSmithIdlePeriod
	)
	monitoring.SetBlockchainStatus(mainchain, constant.BlockchainStatusIdle)

	exist, errGenesis := mainchainBlockService.CheckGenesis()
	if errGenesis != nil {
		log.Fatal(errGenesis)
	}
	if !exist { // Add genesis if not exist
		// genesis account will be inserted in the very beginning
		if err = service.AddGenesisAccount(queryExecutor); err != nil {
			loggerCoreService.Fatal("Fail to add genesis account")
		}
		// genesis next node admission timestamp will be inserted in the very beginning
		if err = service.AddGenesisNextNodeAdmission(
			queryExecutor,
			mainchain.GetGenesisBlockTimestamp(),
			nextNodeAdmissionStorage,
		); err != nil {
			loggerCoreService.Fatal(err)
		}
		if err = mainchainBlockService.AddGenesis(); err != nil {
			loggerCoreService.Fatal(err)
		}
	}
	// set all needed cache
	err = mainchainBlockService.UpdateLastBlockCache(nil)
	if err != nil {
		loggerCoreService.Fatal(err)
	}
	err = nodeRegistrationService.UpdateNextNodeAdmissionCache(nil)
	if err != nil {
		loggerCoreService.Fatal(err)
	}
	err = nodeAddressInfoService.ClearUpdateNodeAddressInfoCache()
	if err != nil {
		loggerCoreService.Fatal(err)
	}

	lastBlockAtStart, err = mainchainBlockService.GetLastBlock()
	if err != nil {
		loggerCoreService.Fatal(err)
	}
	monitoring.SetLastBlock(mainchain, lastBlockAtStart)
	// TODO: Check computer/node local time. Comparing with last block timestamp
	// initialize node registry cache
	err = nodeRegistrationService.InitializeCache()
	if err != nil {
		loggerCoreService.Fatalf("InitializeNodeRegistryCacheFail - %v", err)
	}
	// initialize scrambled nodes
	err = scrambleNodeService.InitializeScrambleCache(lastBlockAtStart.GetHeight())
	if err != nil {
		loggerCoreService.Fatalf("InitializeScrambleNodeFail - %v", err)
	}

	if len(config.NodeKey.Seed) > 0 && config.Smithing {
		node, err := nodeRegistrationService.GetNodeRegistrationByNodePublicKey(config.NodeKey.PublicKey)
		if err != nil {
			loggerCoreService.Fatal(err)
		} else if node == nil {
			// no nodes registered with current node public key, only warn the user but we keep running smithing goroutine
			// so it immediately start when register+admitted to the registry
			loggerCoreService.Error(
				"Current node is not in node registry and won't be able to smith until registered!",
			)
		}
		// register node config public key, so node registration service can detect if node has been admitted
		nodeRegistrationService.SetCurrentNodePublicKey(config.NodeKey.PublicKey)
		// default to isBlocksmith=true
		blockchainStatusService.SetIsBlocksmith(true)
		mainchainProcessor = smith.NewBlockchainProcessor(
			mainchainBlockService.GetChainType(),
			model.NewBlocksmith(config.NodeKey.Seed, config.NodeKey.PublicKey, node.GetNodeID()),
			mainchainBlockService,
			loggerCoreService,
			blockchainStatusService,
			nodeRegistrationService,
		)
		mainchainProcessor.Start(sleepPeriod)
	}
	mainchainDownloader = blockchainsync.NewBlockchainDownloader(
		mainchainBlockService,
		peerServiceClient,
		peerExplorer,
		loggerCoreService,
		blockchainStatusService,
	)
	mainchainForkProcessor = &blockchainsync.ForkingProcessor{
		ChainType:             mainchainBlockService.GetChainType(),
		BlockService:          mainchainBlockService,
		QueryExecutor:         queryExecutor,
		ActionTypeSwitcher:    actionSwitcher,
		MempoolService:        mempoolService,
		PeerExplorer:          peerExplorer,
		Logger:                loggerCoreService,
		TransactionUtil:       transactionUtil,
		TransactionCorService: transactionCoreServiceIns,
		MempoolBackupStorage:  mempoolBackupStorage,
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
		err              error
		nodeID           int64
		lastBlockAtStart *model.Block
		sleepPeriod      = constant.SpineChainSmithIdlePeriod
	)
	monitoring.SetBlockchainStatus(spinechain, constant.BlockchainStatusIdle)

	exist, errGenesis := spinechainBlockService.CheckGenesis()
	if errGenesis != nil {
		log.Fatal(errGenesis)
	}
	if !exist { // Add genesis if not exist
		if err = spinechainBlockService.AddGenesis(); err != nil {
			log.Fatal(err)
		}
	}
	// update cache last spine block  block
	err = spinechainBlockService.UpdateLastBlockCache(nil)
	if err != nil {
		loggerCoreService.Fatal(err)
	}
	lastBlockAtStart, err = spinechainBlockService.GetLastBlock()
	if err != nil {
		loggerCoreService.Fatal(err)
	}
	monitoring.SetLastBlock(spinechain, lastBlockAtStart)

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
			nodeRegistrationService,
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
		ChainType:             spinechainBlockService.GetChainType(),
		BlockService:          spinechainBlockService,
		QueryExecutor:         queryExecutor,
		ActionTypeSwitcher:    nil, // no mempool for spine blocks
		MempoolService:        nil, // no transaction types for spine blocks
		PeerExplorer:          peerExplorer,
		Logger:                loggerCoreService,
		TransactionUtil:       transactionUtil,
		TransactionCorService: transactionCoreServiceIns,
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

// start will start all existence instance
func start() {
	if binaryChecksum, err := util.GetExecutableHash(); err == nil {
		log.Printf("binary checksum: %s", hex.EncodeToString(binaryChecksum))
	}

	// start cpu profiling if enabled
	if flagProfiling {
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

	if flagDebugMode {
		startNodeMonitoring()
		blocker.SetIsDebugMode(true)
	}

	// preload-caches
	err := mempoolService.InitMempoolTransaction()
	if err != nil {
		loggerCoreService.Fatalf("fail to load mempool data - error: %v", err)
	}

	mainchainSyncChannel := make(chan bool, 1)
	mainchainSyncChannel <- true
	startSpinechain()
	startMainchain()
	startServices()
	startScheduler()
	go startBlockchainSynchronizers()

	// Shutting Down
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
func init() {
	rootCmd = &cobra.Command{}
	rootCmd.PersistentFlags().StringVar(&flagConfigPostfix, "config-postfix", "", "Configuration version")
	rootCmd.PersistentFlags().StringVar(&flagConfigPath, "config-path", "", "Configuration path")
	rootCmd.PersistentFlags().BoolVar(&flagDebugMode, "debug", false, "Run on debug mode")
	rootCmd.PersistentFlags().BoolVar(&flagProfiling, "profiling", false, "Run with profiling")
	rootCmd.PersistentFlags().BoolVar(&flagUseEnv, "use-env", false, "Running node without configuration file")
	rootCmd.PersistentFlags().StringVar(&flagResourcePath, "resource-path", "", "Resource path location")
	config = model.NewConfig()
}

func main() {

	var god = goDaemon{}

	daemonCommand.Run = func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			var (
				daemonMessage string
				daemonKind    = daemon.SystemDaemon
			)
			if runtime.GOOS == "darwin" {
				daemonKind = daemon.GlobalDaemon
			}
			srvDaemon, err := daemon.New("zoobc.node", "zoobc node service", daemonKind)
			if err != nil {
				loggerCoreService.Fatalf("failed to run daemon: %s", err.Error())
			}
			god = goDaemon{srvDaemon}
			if runtime.GOOS == "darwin" {
				if dErr := god.SetTemplate(constant.PropertyList); dErr != nil {
					log.Fatal(dErr)
				}
			}

			switch args[0] {
			case "install":
				daemonArgs := []string{"daemon", "run"}
				if flagDebugMode {
					daemonArgs = append(daemonArgs, "--debug")
				}
				if flagProfiling {
					daemonArgs = append(daemonArgs, "--profiling")
				}
				if flagConfigPath != "" {
					daemonArgs = append(daemonArgs, fmt.Sprintf("--config-path=%s", flagConfigPath))
				}
				if flagUseEnv {
					daemonArgs = append(daemonArgs, "--use-env")
				}
				if flagResourcePath != "" {
					daemonArgs = append(daemonArgs, fmt.Sprintf("--resource-path=%s", flagResourcePath))
				}
				daemonMessage, err = god.Install(daemonArgs...)
			case "start":
				daemonMessage, err = god.Start()
			case "stop":
				daemonMessage, err = god.Stop()
			case "remove":
				daemonMessage, err = god.Remove()
			case "status":
				daemonMessage, err = god.Status()
			case "run":
				// sub command used by system
				initiateMainInstance()
				start()
			default:
				_ = daemonCommand.Usage()
			}
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(daemonMessage)
			}
		} else {
			_ = daemonCommand.Usage()
		}
	}
	runCommand.Run = func(cmd *cobra.Command, args []string) {
		initiateMainInstance()
		if !config.LogOnCli && config.CliMonitoring {
			go cliMonitoring.Start()
		}
		start()
	}

	rootCmd.AddCommand(runCommand)
	rootCmd.AddCommand(daemonCommand)
	_ = rootCmd.Execute()

}
