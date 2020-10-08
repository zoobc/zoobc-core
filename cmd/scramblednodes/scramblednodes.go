package scramblednodes

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/zoobc/zoobc-core/common/monitoring"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/database"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/core/service"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
)

type (
	// RunCommand represent of output function from rollback commands
	RunCommand func(ccmd *cobra.Command, args []string)
)

var (
	// Flag Command
	wantedBlockHeight uint32
	dBPath, dBName    string
	senderPeerID      int64

	// subcommand, getScrambledNodes blockchain
	getScrambledNodesCmd = &cobra.Command{
		Use:   "scrambledNodes",
		Short: "scrambledNodes subcommand to get the list of nodes in the scrambledNodes at that height",
		Long:  "scrambledNodes subcommand to get the list of nodes in the scrambledNodes at that height",
	}
	getPriorityPeersCmd = &cobra.Command{
		Use:   "priorityPeers",
		Short: "priorityPeers subcommand to get the list of nodes in the priorityPeers at that height",
		Long:  "priorityPeers subcommand to get the list of nodes in the priorityPeers at that height",
	}
)

func init() {
	// Rollback Blockchain flag
	getScrambledNodesCmd.Flags().Uint32Var(&wantedBlockHeight, "height", 0, "Block height at which the scrambled nodes is positioned")
	getScrambledNodesCmd.Flags().StringVar(&dBPath, "db-path", "../resource", "path of DB blockchain wanted to rollback")
	getScrambledNodesCmd.Flags().StringVar(&dBName, "db-name", "zoobc.db", "name of DB blockchain wanted to rollback")

	getPriorityPeersCmd.Flags().Uint32Var(&wantedBlockHeight, "height", 0, "Block height at which the scrambled nodes is positioned")
	getPriorityPeersCmd.Flags().StringVar(&dBPath, "db-path", "../resource", "path of DB blockchain wanted to rollback")
	getPriorityPeersCmd.Flags().StringVar(&dBName, "db-name", "zoobc.db", "name of DB blockchain wanted to rollback")
	getPriorityPeersCmd.Flags().Int64Var(&senderPeerID, "sender-peer-id", 0, "the full id of the sender")
}

// Commands return Instance of rollback command
func Commands() map[string]*cobra.Command {
	getScrambledNodesCmd.Run = func(ccmd *cobra.Command, args []string) {
		scrambledNodes := getScrambledNodesAtHeight()
		j, _ := json.MarshalIndent(scrambledNodes.IndexNodes, "", "  ")
		stringJSON := string(j)
		fmt.Println(stringJSON)
		fmt.Println("scrambledNodes length ", len(scrambledNodes.IndexNodes))
	}

	getPriorityPeersCmd.Run = func(ccmd *cobra.Command, args []string) {
		priorityPeers := getPriorityPeers()
		j, _ := json.MarshalIndent(priorityPeers, "", "  ")
		stringJSON := string(j)
		fmt.Println(stringJSON)
	}

	return map[string]*cobra.Command{"getScrambledNodesCmd": getScrambledNodesCmd, "getPriorityPeersCmd": getPriorityPeersCmd}
}

func getPriorityPeers() map[string]*model.Peer {
	scrambledNodes := getScrambledNodesAtHeight()
	peers, err := p2pUtil.GetPriorityPeersByNodeID(
		senderPeerID,
		scrambledNodes,
	)
	if err != nil {
		panic(err)
	}
	return peers
}

// getScrambledNodesAtHeight func to run rollback to all
func getScrambledNodesAtHeight() *model.ScrambledNodes {
	var (
		dB, err = getSqliteDB(dBPath, dBName)
	)
	if err != nil {
		fmt.Println("Failed get Db")
		panic(err)
	}
	activeNodeRegistryCacheStorage := storage.NewNodeRegistryCacheStorage(monitoring.TypeActiveNodeRegistryStorage, nil)
	pendingNodeRegistryCacheStorage := storage.NewNodeRegistryCacheStorage(monitoring.TypePendingNodeRegistryStorage, nil)
	var (
		queryExecutor          = query.NewQueryExecutor(dB, logrus.New())
		nodeAddressInfoService = service.NewNodeAddressInfoService(
			queryExecutor,
			query.NewNodeAddressInfoQuery(),
			query.NewNodeRegistrationQuery(),
			query.NewBlockQuery(&chaintype.MainChain{}),
			nil,
			storage.NewNodeAddressInfoStorage(),
			nil,
			activeNodeRegistryCacheStorage,
			logrus.New(),
		)

		nodeRegistrationService = service.NewNodeRegistrationService(
			queryExecutor,
			query.NewAccountBalanceQuery(),
			query.NewNodeRegistrationQuery(),
			query.NewParticipationScoreQuery(),
			query.NewNodeAdmissionTimestampQuery(),
			nil,
			nil,
			nodeAddressInfoService,
			nil,
			activeNodeRegistryCacheStorage,
			pendingNodeRegistryCacheStorage,
		)
		scramblecache       = storage.NewScrambleCacheStackStorage()
		scrambleNodeService = service.NewScrambleNodeService(
			nodeRegistrationService, nodeAddressInfoService, queryExecutor, query.NewBlockQuery(&chaintype.MainChain{}), scramblecache)
	)
	err = nodeAddressInfoService.ClearUpdateNodeAddressInfoCache()
	if err != nil {
		panic(err)
	}
	scrambledNodes, err := scrambleNodeService.GetScrambleNodesByHeight(wantedBlockHeight)
	if err != nil {
		panic(err)
	}
	return scrambledNodes
}

// getSqliteDB to get sql.Db of sqlite based on DB path & name
func getSqliteDB(dbPath, dbName string) (*sql.DB, error) {
	var sqliteDbInstance = database.NewSqliteDB()
	if err := sqliteDbInstance.InitializeDB(dbPath, dbName); err != nil {
		return nil, err
	}
	sqliteDB, err := sqliteDbInstance.OpenDB(
		dbPath,
		dbName,
		constant.SQLMaxOpenConnetion,
		constant.SQLMaxIdleConnections,
		constant.SQLMaxConnectionLifetime,
	)
	if err != nil {
		return nil, err
	}
	return sqliteDB, nil
}
