package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"

	"github.com/zoobc/zoobc-core/common/crypto"

	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/core/service"

	"github.com/zoobc/zoobc-core/core/smith"

	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/api"
	"github.com/zoobc/zoobc-core/common/database"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

var (
	dbPath, dbName          string
	dbInstance              *database.SqliteDB
	db                      *sql.DB
	nodeSecretPhrase        string
	apiRPCPort, apiHTTPPort int
)

func init() {
	var err error
	if err := util.LoadConfig("./resource", "config", "toml"); err != nil {
		panic(err)
	} else {
		dbPath = viper.GetString("dbPath")
		dbName = viper.GetString("dbName")
		nodeSecretPhrase = viper.GetString("nodeSecretPhrase")
		apiRPCPort = viper.GetInt("apiRPCPort")
		if apiRPCPort == 0 {
			apiRPCPort = 8080
		}
		apiHTTPPort = viper.GetInt("apiHTTPPort")
		if apiHTTPPort == 0 {
			apiHTTPPort = 8000
		}
	}

	dbInstance = database.NewSqliteDB()
	if err := dbInstance.InitializeDB(dbPath, dbName); err != nil {
		panic(err)
	}
	db, err = dbInstance.OpenDB(dbPath, dbName, 10, 20)
	if err != nil {
		panic(err)
	}
}

func startServices(queryExecutor *query.Executor) {
	api.Start(apiRPCPort, apiHTTPPort, queryExecutor)
}

func main() {
	fmt.Println("run")

	queryExecutor := query.NewQueryExecutor(db)

	migration := database.Migration{}
	if err := migration.Init(queryExecutor); err != nil {
		panic(err)
	}

	if err := migration.Apply(); err != nil {
		panic(err)
	}
	mainchain := &chaintype.MainChain{}
	sleepPeriod := int(mainchain.GetChainSmithingDelayTime())
	// todo: read secret phrase from config
	blockchainProcessor := smith.NewBlockchainProcessor(mainchain,
		smith.NewBlocksmith(nodeSecretPhrase),
		service.NewBlockService(mainchain, query.NewQueryExecutor(db), query.NewBlockQuery(mainchain),
			query.NewMempoolQuery(mainchain), query.NewTransactionQuery(mainchain), crypto.NewSignature()),
		service.NewMempoolService(mainchain, query.NewQueryExecutor(db), query.NewMempoolQuery(mainchain)))
	if !blockchainProcessor.CheckGenesis() { // Add genesis if not exist
		addGenesis(queryExecutor) // genesis account will be inserted in the very beginning
		_ = blockchainProcessor.AddGenesis()
	}
	if len(nodeSecretPhrase) > 0 {
		go startSmith(sleepPeriod, blockchainProcessor)
	}

	startServices(queryExecutor)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// When we receive a signal from the OS, shut down everything
	<-sigs

}

func startSmith(sleepPeriod int, processor *smith.BlockchainProcessor) {
	for {
		_ = processor.StartSmithing()
		time.Sleep(time.Duration(sleepPeriod) * time.Second)
	}

}

func addGenesis(executor query.ExecutorInterface) {
	// add genesis account
	genesisAccount := model.Account{
		ID:          util.CreateAccountIDFromAddress(0, constant.GenesisAccountAddress),
		AccountType: 0,
		Address:     constant.GenesisAccountAddress,
	}
	genesisAccountBalance := model.AccountBalance{
		AccountID:        genesisAccount.ID,
		BlockHeight:      0,
		SpendableBalance: 0,
		Balance:          0,
		PopRevenue:       0,
		Latest:           true,
	}
	genesisAccountInsertQ, genesisAccountInsertArgs := query.NewAccountQuery().InsertAccount(&genesisAccount)
	genesisAccountBalanceInsertQ, genesisAccountBalanceInsertArgs := query.NewAccountBalanceQuery().InsertAccountBalance(
		&genesisAccountBalance)
	_, err := executor.ExecuteStatement(genesisAccountInsertQ, genesisAccountInsertArgs...)
	if err != nil {
		panic("fail to add genesis account")
	}
	_, err = executor.ExecuteStatement(genesisAccountBalanceInsertQ, genesisAccountBalanceInsertArgs...)
	if err != nil {
		panic("fail to add genesis account balance")
	}
}
