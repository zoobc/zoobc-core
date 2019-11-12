package rollback

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/database"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	RunCommand func(ccmd *cobra.Command, args []string)
)

var (
	// Flag Command
	wantedBlockHeight uint32
	dBPath, dBName    string

	// Root rollback command
	rollbackCmd = &cobra.Command{
		Use:   "rollback",
		Short: "rollback is a developer cli tools to run & simulate rollback query.",
		Long: `rollback is a developer cli tools to run & simulate rollback query.
running 'zoobc rollback subCommand --flag' will call rollback query and show the last status of database after rollback
	`,
	}

	// subcommand, rollback blockchain
	rollbackBlockChainCmd = &cobra.Command{
		Use:   "blockchain",
		Short: "The subcommand of rollback to run & simulate all existing rollback query.",
		Long: `rollbackBlockChain is a developer cli tools to run & simulate all existing rollback query.
running 'zoobc rollback blockchain --to-height numberOfHeight --db-path "./dbPath" heigh --db-name "dbName"'
will call all existing rollback query and show the last status of database after rollback
		`,
	}
)

func init() {
	// Rollback Blockchain flag
	rollbackBlockChainCmd.Flags().Uint32Var(&wantedBlockHeight, "to-height", 0, "Block height state wanted after rollback")
	rollbackBlockChainCmd.Flags().StringVar(&dBPath, "db-path", "", "path of DB blockchain wanted to rollback")
	rollbackBlockChainCmd.Flags().StringVar(&dBName, "db-name", "", "name of DB blockchain wanted to rollback")
}

func Commands(sqliteDB *sql.DB) *cobra.Command {
	rollbackBlockChainCmd.Run = rollbackBlockChain(sqliteDB)
	rollbackCmd.AddCommand(rollbackBlockChainCmd)
	return rollbackCmd
}

// RollbackBlockChain func to run rollback to all
func rollbackBlockChain(defaultDB *sql.DB) RunCommand {
	var (
		err               error
		dB                = defaultDB
		chaintypeRollback = chaintype.GetChainType(0)
	)

	// checking DB path and DB name flag, making sure both must use or both of them must default
	if (dBPath == "" && dBName != "") || (dBPath != "" && dBName == "") {
		panic(errors.New("Please set both db-path and db-name"))
	}

	if dBPath != "" && dBName != "" {
		dB, err = getSqliteDB(dBPath, dBName)
		if err != nil {
			fmt.Println("Failed get Db")
			panic(err)
		}
	}

	return func(ccmd *cobra.Command, args []string) {
		var (
			queryExecutor  = query.NewQueryExecutor(dB)
			derivedQueries = query.GetDerivedQuery(chaintypeRollback)
			blockQuery     = query.NewBlockQuery(chaintypeRollback)
			rowLastBlock   = queryExecutor.ExecuteSelectRow(blockQuery.GetLastBlock())
			lastBlock      model.Block
		)

		err = blockQuery.Scan(&lastBlock, rowLastBlock)
		if err != nil {
			fmt.Println("Failed get last block")
		}

		// checking current block state
		if lastBlock.GetHeight() <= wantedBlockHeight {
			fmt.Printf("No need rollback to height %d, current blockchain state in height %d \n\n",
				wantedBlockHeight, lastBlock.GetHeight())
			return
		}

		err = queryExecutor.BeginTx()
		if err != nil {
			fmt.Println("Failed begin Tx Err: ", err.Error())
			return
		}

		for key, dQuery := range derivedQueries {
			queries := dQuery.Rollback(wantedBlockHeight)
			err = queryExecutor.ExecuteTransactions(queries)
			if err != nil {
				fmt.Println(key)
				fmt.Println("Failed execute rollback queries, ", err.Error())
				err = queryExecutor.RollbackTx()
				if err != nil {
					fmt.Println("Failed to run RollbackTX DB")
				}
				break
			}
		}
		err = queryExecutor.CommitTx()
		if err != nil {
			fmt.Println("Failed to run CommitTx DB, err : ", err.Error())
			return
		}
		fmt.Printf("Rollback blockchain sucessfully! \nNow blockchain state in HEIGHT %d\n\n", wantedBlockHeight)
	}
}

// getSqliteDB to get sql.Db of sqlite based on DB path & name
func getSqliteDB(dbPath, dbName string) (*sql.DB, error) {
	var sqliteDbInstance = database.NewSqliteDB()
	if err := sqliteDbInstance.InitializeDB(dbPath, dbName); err != nil {
		return nil, err
	}
	sqliteDB, err := sqliteDbInstance.OpenDB(dbPath, dbName, 10, 20)
	if err != nil {
		return nil, err
	}
	return sqliteDB, nil
}
