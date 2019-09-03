package main

import (
	"database/sql"
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/p2p/client"
	"github.com/zoobc/zoobc-core/p2p/strategy"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"

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
)

var (
	dbPath, dbName, nodeSecretPhrase string
	dbInstance                       *database.SqliteDB
	db                               *sql.DB
	apiRPCPort, apiHTTPPort          int
	peerPort                         uint32
	p2pServiceInstance               p2p.Peer2PeerServiceInterface
	queryExecutor                    *query.Executor
	observerInstance                 *observer.Observer
	blockServices                    = make(map[int32]coreService.BlockServiceInterface)
	mempoolServices                  = make(map[int32]service.MempoolServiceInterface)
	peerServiceClient                client.PeerServiceClientInterface
	p2pHost                          *model.Host
	peerExplorer                     strategy.PeerExplorerStrategyInterface
	ownerAccountAddress, myAddress   string
	wellknownPeers                   []string
	nodeKeyFilePath                  string
	smithing                         bool
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
		apiRPCPort = viper.GetInt("apiRPCPort")
		apiHTTPPort = viper.GetInt("apiHTTPPort")
		ownerAccountAddress = viper.GetString("ownerAccountAddress")
		myAddress = viper.GetString("myAddress")
		peerPort = viper.GetUint32("peerPort")
		wellknownPeers = viper.GetStringSlice("wellknownPeers")
		smithing = viper.GetBool("smithing")
		configPath := viper.GetString("configPath")
		nodeKeyFile := viper.GetString("nodeKeyFile")
		nodeKeyFilePath = filepath.Join(configPath, nodeKeyFile)
		nodeAdminKeysService := coreService.NewNodeAdminService(nil, nil, nil, nil, nodeKeyFilePath)
		nodeKeys, err := nodeAdminKeysService.ParseKeysFile()
		if err != nil {
			// generate a node private key if there aren't already configured
			seed := util.GetSecureRandomSeed()
			if _, err := nodeAdminKeysService.GenerateNodeKey(seed); err != nil {
				logrus.Fatal(err)
			}
		}
		nodeKey := nodeAdminKeysService.GetLastNodeKey(nodeKeys)
		if nodeKey != nil {
			nodeSecretPhrase = nodeKey.Seed
		}
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

	initP2pInstance()

	initObserverListeners()
}

func initP2pInstance() {
	nodePublicKey := util.GetPublicKeyFromSeed(nodeSecretPhrase)
	// initialize peer client service
	peerServiceClient = client.NewPeerServiceClient(
		queryExecutor,
		query.NewReceiptQuery(),
		nodePublicKey,
	)

	// init p2p instances
	knownPeersResult, err := p2pUtil.ParseKnownPeers(wellknownPeers)
	if err != nil {
		logrus.Fatal("fail to start p2p service")
	}

	p2pHost = p2pUtil.NewHost(myAddress, peerPort, knownPeersResult)

	// peer discovery strategy
	peerExplorer = strategy.NewNativeStrategy(
		p2pHost,
		peerServiceClient,
	)
	p2pServiceInstance, _ = p2p.NewP2PService(
		p2pHost,
		peerServiceClient,
		peerExplorer,
	)
}

func initObserverListeners() {
	// init observer listeners
	observerInstance.AddListener(observer.BlockPushed, p2pServiceInstance.SendBlockListener())
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
		queryExecutor,
		p2pServiceInstance,
		blockServices,
		ownerAccountAddress,
		nodeKeyFilePath,
	)
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
		crypto.NewSignature(),
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

	if smithing {
		go startSmith(sleepPeriod, mainchainProcessor)
	}
	mainchainSynchronizer := blockchainsync.NewBlockchainSyncService(
		mainchainBlockService,
		p2pServiceInstance,
		peerServiceClient,
		peerExplorer,
	)
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

	startServices()

	mainchainSyncChannel := make(chan bool, 1)
	mainchainSyncChannel <- true
	startMainchain(mainchainSyncChannel)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// When we receive a signal from the OS, shut down everything
	<-sigs
}
