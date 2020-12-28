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
package rollback

import (
	"fmt"
	"github.com/zoobc/zoobc-core/cmd/helper"

	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
)

type (
	// RunCommand represent of output function from rollback commands
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
	rollbackBlockChainCmd.Flags().StringVar(&dBPath, "db-path", "../resource", "path of DB blockchain wanted to rollback")
	rollbackBlockChainCmd.Flags().StringVar(&dBName, "db-name", "zoobc.db", "name of DB blockchain wanted to rollback")
}

// Commands return Instance of rollback command
func Commands() *cobra.Command {
	rollbackBlockChainCmd.Run = rollbackBlockChain()
	rollbackCmd.AddCommand(rollbackBlockChainCmd)
	return rollbackCmd
}

// RollbackBlockChain func to run rollback to all
func rollbackBlockChain() RunCommand {
	var (
		chaintypeRollback = chaintype.GetChainType(0)
	)

	return func(ccmd *cobra.Command, args []string) {
		var (
			derivedQueries  = query.GetDerivedQuery(chaintypeRollback)
			blockQuery      = query.NewBlockQuery(chaintypeRollback)
			dB, err         = helper.GetSqliteDB(dBPath, dBName)
			queryExecutor   = query.NewQueryExecutor(dB)
			rowLastBlock, _ = queryExecutor.ExecuteSelectRow(blockQuery.GetLastBlock(), false)
			lastBlock       model.Block
		)
		if err != nil {
			fmt.Println("Failed get Db")
			panic(err)
		}

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
				return
			}
		}
		err = queryExecutor.CommitTx()
		if err != nil {
			fmt.Println("Failed to run CommitTx DB, err : ", err.Error())
			return
		}
		fmt.Printf("Rollback blockchain successfully! \nNow blockchain state in HEIGHT %d\n\n", wantedBlockHeight)
	}
}
