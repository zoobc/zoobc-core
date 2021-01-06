// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
package scramblednodes

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/database"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
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
		queryExecutor          = query.NewQueryExecutor(dB)
		nodeAddressInfoService = service.NewNodeAddressInfoService(
			queryExecutor,
			query.NewNodeAddressInfoQuery(),
			query.NewNodeRegistrationQuery(),
			query.NewBlockQuery(&chaintype.MainChain{}),
			nil,
			storage.NewNodeAddressInfoStorage(),
			nil,
			activeNodeRegistryCacheStorage,
			nil,
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
