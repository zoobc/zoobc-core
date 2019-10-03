package main

import (
	"log"

	transaction2 "github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/observer"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/cmd/account"
	"github.com/zoobc/zoobc-core/cmd/block"
	"github.com/zoobc/zoobc-core/cmd/transaction"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/database"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/core/smith"
)

var (
	blocksmith        *model.Blocksmith
	blockProcessor    smith.BlockchainProcessorInterface
	blockService      service.BlockServiceInterface
	dbPath, dbName    = "./testdata/", "zoobc.db"
	sortedBlocksmiths []model.Blocksmith
	queryExecutor     query.ExecutorInterface
	migration         database.Migration
)

func init() {
	chainType := &chaintype.MainChain{}
	observerInstance := observer.NewObserver()
	secretPhrase := "sprinkled sneak species pork outpost thrift unwind cheesy vexingly dizzy neurology neatness"
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
	actionSwitcher := &transaction2.TypeSwitcher{
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
	nodeRegistrationService := service.NewNodeRegistrationService(
		queryExecutor,
		query.NewAccountBalanceQuery(),
		query.NewNodeRegistrationQuery(),
		query.NewParticipationScoreQuery(),
	)
	blockProcessor = smith.NewBlockchainProcessor(
		chainType,
		blocksmith,
		blockService,
		nodeRegistrationService,
	)
	migration = database.Migration{Query: queryExecutor}
}

func main() {
	var (
		rootCmd   *cobra.Command
		logLevels []string
	)
	rootCmd = &cobra.Command{Use: "zoobc"}
	logLevels = viper.GetStringSlice("logLevels")
	logger, _ := util.InitLogger(".log/", "cmd.debug.log", logLevels)
	rootCmd.AddCommand(account.GenerateAccount(logger))
	rootCmd.AddCommand(transaction.GenerateTransactionBytes(logger, &crypto.Signature{}))
	rootCmd.AddCommand(block.GenerateBlocks(logger, blockProcessor, blockService, queryExecutor, migration))
	_ = rootCmd.Execute()
}
