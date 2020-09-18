package block

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/common/auth"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/database"
	"github.com/zoobc/zoobc-core/common/fee"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/core/smith"
	"github.com/zoobc/zoobc-core/core/smith/strategy"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
	"github.com/zoobc/zoobc-core/observer"
)

type (
	mockBlockchainStatusService struct {
		service.BlockchainStatusService
	}
)

var (
	blocksmith              *model.Blocksmith
	chainType               chaintype.ChainType
	blockProcessor          smith.BlockchainProcessorInterface
	blockService            service.BlockServiceInterface
	nodeRegistrationService service.NodeRegistrationServiceInterface
	blocksmithStrategy      strategy.BlocksmithStrategyInterface
	queryExecutor           query.ExecutorInterface
	migration               database.Migration

	numberOfBlocks         int
	blocksmithSecretPhrase string
	outputPath             string

	blockCmd = &cobra.Command{
		Use:   "block",
		Short: "block command used to manipulate block of node",
		Long: `
			block command is use to manipulate block creation or broadcasting in the node
		`,
	}

	fakeBlockCmd = &cobra.Command{
		Use:   "fake-blocks",
		Short: "fake-blocks command used to create fake blocks",
		Run: func(cmd *cobra.Command, args []string) {
			generateBlocks(numberOfBlocks, blocksmithSecretPhrase, outputPath, chainType)
		},
	}
)

func (*mockBlockchainStatusService) IsFirstDownloadFinished(ct chaintype.ChainType) bool {
	return true
}

func (*mockBlockchainStatusService) IsDownloading(ct chaintype.ChainType) bool {
	return true
}

func init() {
	fakeBlockCmd.Flags().IntVar(
		&numberOfBlocks,
		"numberOfBlocks",
		100,
		"number of account to generate",
	)
	fakeBlockCmd.Flags().StringVar(
		&blocksmithSecretPhrase,
		"blocksmithSecretPhrase",
		"",
		"secret phrase of blocksmith | required",
	)
	fakeBlockCmd.Flags().StringVar(
		&outputPath,
		"out",
		"./testdata/zoobc.db",
		"output path of the database",
	)
	blockCmd.AddCommand(fakeBlockCmd)
}

func Commands() *cobra.Command {
	return blockCmd
}

func initialize(
	secretPhrase, outputPath string,
) {
	signature := crypto.NewSignature()
	nodeAuthValidation := auth.NewNodeAuthValidation(signature)
	transactionUtil := &transaction.Util{}
	receiptUtil := &coreUtil.ReceiptUtil{}
	paths := strings.Split(outputPath, "/")
	dbPath, dbName := strings.Join(paths[:len(paths)-1], "/")+"/", paths[len(paths)-1]
	chainType = &chaintype.MainChain{}
	observerInstance := observer.NewObserver()
	blocksmith = model.NewBlocksmith(secretPhrase, crypto.NewEd25519Signature().GetPublicKeyFromSeed(secretPhrase), 0)
	// initialize/open db and queryExecutor
	dbInstance := database.NewSqliteDB()
	if err := dbInstance.InitializeDB(dbPath, dbName); err != nil {
		panic(err)
	}
	db, err := dbInstance.OpenDB(dbPath, dbName, 10, 10, 20*time.Minute)
	if err != nil {
		panic(err)
	}
	queryExecutor = query.NewQueryExecutor(db)
	mempoolStorage := storage.NewMempoolStorage()

	actionSwitcher := &transaction.TypeSwitcher{
		Executor:            queryExecutor,
		NodeAuthValidation:  nodeAuthValidation,
		MempoolCacheStorage: mempoolStorage,
	}
	blockStorage := storage.NewBlockStateStorage()
	receiptService := service.NewReceiptService(
		query.NewNodeReceiptQuery(),
		query.NewMerkleTreeQuery(),
		query.NewNodeRegistrationQuery(),
		query.NewBlockQuery(chainType),
		queryExecutor,
		nodeRegistrationService,
		crypto.NewSignature(),
		nil,
		receiptUtil,
		nil,
		nil,
		nil,
	)
	mempoolService := service.NewMempoolService(
		transactionUtil,
		chainType,
		queryExecutor,
		query.NewMempoolQuery(chainType),
		query.NewMerkleTreeQuery(),
		actionSwitcher,
		query.NewAccountBalanceQuery(),
		query.NewTransactionQuery(chainType),
		crypto.NewSignature(),
		observerInstance,
		log.New(),
		receiptUtil,
		receiptService,
		nil,
		blockStorage,
		mempoolStorage,
		nil,
	)
	nodeRegistrationService := service.NewNodeRegistrationService(
		queryExecutor,
		query.NewNodeAddressInfoQuery(),
		query.NewAccountBalanceQuery(),
		query.NewNodeRegistrationQuery(),
		query.NewParticipationScoreQuery(),
		query.NewBlockQuery(chainType),
		query.NewNodeAdmissionTimestampQuery(),
		log.New(),
		&mockBlockchainStatusService{},
		nil,
		nil,
		nil,
		nil,
	)
	blocksmithStrategy = strategy.NewBlocksmithStrategyMain(
		queryExecutor, query.NewNodeRegistrationQuery(), query.NewSkippedBlocksmithQuery(), log.New(),
	)
	publishedReceiptUtil := coreUtil.NewPublishedReceiptUtil(
		query.NewPublishedReceiptQuery(),
		queryExecutor,
	)
	feeScaleService := fee.NewFeeScaleService(
		query.NewFeeScaleQuery(),
		query.NewBlockQuery(&chaintype.MainChain{}),
		queryExecutor,
	)
	blockService = service.NewBlockMainService(
		chainType,
		queryExecutor,
		query.NewBlockQuery(chainType),
		query.NewMempoolQuery(chainType),
		query.NewTransactionQuery(chainType),
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
		blocksmithStrategy,
		log.New(),
		query.NewAccountLedgerQuery(),
		service.NewBlockIncompleteQueueService(chainType, observerInstance),
		transactionUtil,
		receiptUtil,
		publishedReceiptUtil,
		service.NewTransactionCoreService(
			nil,
			queryExecutor,
			nil,
			nil,
			query.NewTransactionQuery(chainType),
			nil,
			nil,
			nil,
		),
		nil,
		nil,
		nil,
		nil,
		nil,
		feeScaleService,
		query.GetPruneQuery(chainType),
		nil,
		nil,
	)

	migration = database.Migration{Query: queryExecutor}
}

// generateBlocks used to generate dummy block for testing
// note: now only support mainchain, will implement spinechain implementation details later.
func generateBlocks(numberOfBlocks int, blocksmithSecretPhrase, outputPath string, ct chaintype.ChainType) {
	fmt.Println("initializing dependency and database...")
	initialize(blocksmithSecretPhrase, outputPath)
	fmt.Println("done initializing database")
	blockProcessor = smith.NewBlockchainProcessor(
		blockService.GetChainType(),
		blocksmith,
		blockService,
		log.New(),
		&mockBlockchainStatusService{},
		nodeRegistrationService,
	)
	startTime := time.Now().UnixNano() / 1e6
	fmt.Printf("generating %d blocks\n", numberOfBlocks)
	fmt.Println("initializing database schema migration")
	if err := migration.Init(); err != nil {
		panic(err)
	}

	fmt.Println("applying database schema migration")
	if err := migration.Apply(); err != nil {
		panic(err)
	}
	fmt.Println("checking genesis...")

	if exist, _ := blockService.CheckGenesis(); !exist { // Add genesis if not exist
		fmt.Println("genesis does not exist, adding genesis")
		// genesis account will be inserted in the very beginning
		if err := service.AddGenesisAccount(queryExecutor); err != nil {
			panic(err)
		}

		if err := blockService.AddGenesis(); err != nil {
			panic(err)
		}

		fmt.Println("begin generating blocks")
		if err := blockProcessor.FakeSmithing(numberOfBlocks, true, ct); err != nil {
			panic(err)
		}
	} else {
		// start from last block's timestamp
		fmt.Println("continuing from last database...")
		if err := blockProcessor.FakeSmithing(numberOfBlocks, false, ct); err != nil {
			panic("error in appending block to existing database")
		}
	}
	fmt.Printf("database generated in %s", outputPath)
	fmt.Printf("block generation success in %d miliseconds", (time.Now().UnixNano()/1e6)-startTime)
}
