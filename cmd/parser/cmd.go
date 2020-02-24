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
