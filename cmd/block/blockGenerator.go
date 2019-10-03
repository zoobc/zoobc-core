package block

import (
	"log"
	"time"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/observer"

	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/core/service"

	"github.com/zoobc/zoobc-core/common/database"

	"github.com/zoobc/zoobc-core/core/smith"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	blocksmith              *model.Blocksmith
	chainType               chaintype.ChainType
	blockProcessor          smith.BlockchainProcessorInterface
	blockService            service.BlockServiceInterface
	nodeRegistrationService service.NodeRegistrationServiceInterface
	dbPath, dbName          = "./testdata/", "zoobc.db"
	sortedBlocksmiths       []model.Blocksmith
	queryExecutor           query.ExecutorInterface
	migration               database.Migration
)

func initialize(
	secretPhrase string,
) {
	chainType = &chaintype.MainChain{}
	observerInstance := observer.NewObserver()
	blocksmith = model.NewBlocksmith(
		secretPhrase,
		util.GetPublicKeyFromSeed(secretPhrase),
	)
	// initialize/open db and queryExecutor
	dbInstance := database.NewSqliteDB()
	if err := dbInstance.InitializeDB(dbPath, dbName); err != nil {
		log.Fatal(err)
	}
	db, err := dbInstance.OpenDB(dbPath, dbName, 10, 20)
	if err != nil {
		log.Fatal(err)
	}
	queryExecutor = query.NewQueryExecutor(db)
	actionSwitcher := &transaction.TypeSwitcher{
		Executor: queryExecutor,
	}
	mempoolService := service.NewMempoolService(
		chainType,
		queryExecutor,
		query.NewMempoolQuery(chainType),
		actionSwitcher,
		query.NewAccountBalanceQuery(),
		crypto.NewSignature(),
		query.NewTransactionQuery(chainType),
		observerInstance,
	)
	blockService = service.NewBlockService(
		chainType,
		queryExecutor,
		query.NewBlockQuery(chainType),
		query.NewMempoolQuery(chainType),
		query.NewTransactionQuery(chainType),
		crypto.NewSignature(),
		mempoolService,
		actionSwitcher,
		query.NewAccountBalanceQuery(),
		query.NewParticipationScoreQuery(),
		query.NewNodeRegistrationQuery(),
		observerInstance,
		&sortedBlocksmiths,
	)
	nodeRegistrationService = service.NewNodeRegistrationService(
		queryExecutor,
		query.NewAccountBalanceQuery(),
		query.NewNodeRegistrationQuery(),
		query.NewParticipationScoreQuery(),
	)

	migration = database.Migration{Query: queryExecutor}
}

func GenerateBlocks(logger *logrus.Logger) *cobra.Command {
	var (
		numberOfBlocks         int
		blocksmithSecretPhrase string
	)
	var blockCmd = &cobra.Command{
		Use:   "block",
		Short: "block command used to manipulate block of node",
		Long: `
			block command is use to manipulate block creation or broadcasting in the node
		`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("initializing dependency and database...")
			initialize(blocksmithSecretPhrase)
			log.Println("done initializing database")
			blockProcessor = smith.NewBlockchainProcessor(
				chainType,
				blocksmith,
				blockService,
				nodeRegistrationService,
			)
			if args[0] == "generate-fake" {
				startTime := time.Now().Unix()
				log.Printf("generating %d blocks\n", numberOfBlocks)
				log.Println("initializing database schema migration")
				if err := migration.Init(); err != nil {
					log.Fatal(err)
				}

				log.Println("applying database schema migration")
				if err := migration.Apply(); err != nil {
					log.Fatal(err)
				}
				log.Println("checking genesis...")
				if !blockService.CheckGenesis() { // Add genesis if not exist
					log.Println("genesis does not exist, adding genesis")
					// genesis account will be inserted in the very beginning
					if err := service.AddGenesisAccount(queryExecutor); err != nil {
						log.Fatal("Fail to add genesis account")
					}

					if err := blockService.AddGenesis(); err != nil {
						log.Fatalf("error in adding genesis: %v", err)
					}

					log.Println("begin generating blocks")
					if err := blockProcessor.FakeSmithing(numberOfBlocks); err != nil {
						log.Fatalf("error in fake smithing: %v", err)
					}
					log.Printf("block generation success in %d seconds", time.Now().Unix()-startTime)
				} else {
					log.Fatal("previous generated database still exist, move them")
				}
				log.Printf("database generated in %s", dbPath+dbName)
			} else {
				logger.Error("unknown command")
			}
		},
	}
	blockCmd.Flags().IntVar(
		&numberOfBlocks,
		"numberOfBlocks",
		100,
		"number of account to generate",
	)
	blockCmd.Flags().StringVar(
		&blocksmithSecretPhrase,
		"blocksmithSecretPhrase",
		"sprinkled sneak species pork outpost thrift unwind cheesy vexingly dizzy neurology neatness",
		"secret phrase of blocksmith",
	)
	return blockCmd
}
