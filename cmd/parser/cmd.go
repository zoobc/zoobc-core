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
package parser

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/common/transaction"
)

var (
	/*
		Transaction Parser Command
	*/
	txParserCmd = &cobra.Command{
		Use:   "tx",
		Short: "parse transaction from its hex representation",
		Long:  "transaction parser to check the content of your transaction hex",
	}
)

func init() {
	txParserCmd.Flags().StringVar(&parserTxHex, "transaction-hex", "", "hex string of the transaction bytes")
	txParserCmd.Flags().StringVar(&parserTxBytes, "transaction-bytes", "", "transaction bytes separated by `, `. eg:"+
		"--transaction-bytes='1, 222, 54, 12, 32'")
}

func Commands() *cobra.Command {
	txParserCmd.Run = ParseTransaction
	return txParserCmd
}

func ParseTransaction(*cobra.Command, []string) {
	var txBytes []byte
	if parserTxHex != "" {
		txBytes, _ = hex.DecodeString(parserTxHex)
	} else {
		txByteCharSlice := strings.Split(parserTxBytes, ", ")
		for _, v := range txByteCharSlice {
			byteValue, err := strconv.Atoi(v)
			if err != nil {
				panic("failed to parse transaction bytes")
			}
			txBytes = append(txBytes, byte(byteValue))
		}
	}
	tx, err := (&transaction.Util{}).ParseTransactionBytes(txBytes, false)
	if err != nil {
		panic("error parsing tx" + err.Error())
	}
	tx.TransactionBody, err = (&transaction.MultiSignatureTransaction{}).ParseBodyBytes(tx.TransactionBodyBytes)
	if err != nil {
		panic("error parsing tx body" + err.Error())
	}
	fmt.Printf("transaction:\n%v\n", tx)
}
