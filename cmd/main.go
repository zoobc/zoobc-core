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
package main

import (
	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/cmd/account"
	"github.com/zoobc/zoobc-core/cmd/admin"
	"github.com/zoobc/zoobc-core/cmd/block"
	"github.com/zoobc/zoobc-core/cmd/configure"
	"github.com/zoobc/zoobc-core/cmd/decryptcert"
	"github.com/zoobc/zoobc-core/cmd/genesisblock"
	"github.com/zoobc/zoobc-core/cmd/parser"
	"github.com/zoobc/zoobc-core/cmd/rollback"
	"github.com/zoobc/zoobc-core/cmd/scramblednodes"
	"github.com/zoobc/zoobc-core/cmd/signature"
	"github.com/zoobc/zoobc-core/cmd/snapshot"
	"github.com/zoobc/zoobc-core/cmd/transaction"
)

func main() {
	var (
		rootCmd   *cobra.Command
		parserCmd = &cobra.Command{
			Use:   "parser",
			Short: "parse data to understandable struct",
		}
	)

	rootCmd = &cobra.Command{
		Use:   "zoobc",
		Short: "CLI app for zoobc core",
		Long:  "Commandline Tools for zoobc core",
	}
	rootCmd.AddCommand(genesisblock.Commands())
	rootCmd.AddCommand(rollback.Commands())
	rootCmd.AddCommand(parserCmd)
	rootCmd.AddCommand(signature.Commands())
	rootCmd.AddCommand(snapshot.Commands())
	rootCmd.AddCommand(account.Commands())
	rootCmd.AddCommand(transaction.Commands())
	rootCmd.AddCommand(block.Commands())
	rootCmd.AddCommand(admin.Commands())
	rootCmd.AddCommand(scramblednodes.Commands()["getScrambledNodesCmd"])
	rootCmd.AddCommand(scramblednodes.Commands()["getPriorityPeersCmd"])
	rootCmd.AddCommand(configure.Commands())
	rootCmd.AddCommand(decryptcert.Commands())
	parserCmd.AddCommand(parser.Commands())
	_ = rootCmd.Execute()

}
