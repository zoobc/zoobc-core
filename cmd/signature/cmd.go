package signature

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/zoobc/zoobc-core/common/model"

	"github.com/zoobc/zoobc-core/common/crypto"
	"golang.org/x/crypto/sha3"

	"github.com/spf13/cobra"
)

var (
	/*
		Signer command line tools
	*/
	signerCmd = &cobra.Command{
		Use:   "sign",
		Short: "sign provided data",
		Long:  "sign any provided data by using the --seed parameter",
	}
)

func init() {
	signerCmd.Flags().StringVar(&dataHex, "data-hex", "", "hex string of the data to sign")
	signerCmd.Flags().StringVar(&dataBytes, "data-bytes", "", "data bytes separated by `, `. eg:"+
		"--data-bytes='1, 222, 54, 12, 32'")
	signerCmd.Flags().StringVar(&seed, "seed", "", "your secret phrase")
	signerCmd.Flags().BoolVar(&hash, "hash", false, "turn this flag on to hash the data before signing")
}

func Commands() *cobra.Command {
	signerCmd.Run = SignData
	return signerCmd
}

func SignData(*cobra.Command, []string) {
	var (
		unsignedBytes       []byte
		hashedUnsignedBytes [32]byte
		signature           []byte
	)
	if dataHex != "" {
		unsignedBytes, _ = hex.DecodeString(dataHex)
	} else {
		txByteCharSlice := strings.Split(dataBytes, ", ")
		for _, v := range txByteCharSlice {
			byteValue, err := strconv.Atoi(v)
			if err != nil {
				panic("failed to parse transaction bytes")
			}
			unsignedBytes = append(unsignedBytes, byte(byteValue))
		}
	}
	if hash {
		hashedUnsignedBytes = sha3.Sum256(unsignedBytes)
		signature, _ = (&crypto.Signature{}).Sign(hashedUnsignedBytes[:], model.SignatureType_DefaultSignature, seed)
	} else {
		signature, _ = (&crypto.Signature{}).Sign(unsignedBytes, model.SignatureType_DefaultSignature, seed)
	}
	edUtil := crypto.NewEd25519Signature()
	fmt.Printf("account-address:\t%v\n", edUtil.GetAddressFromSeed(seed))
	fmt.Printf("transaction-bytes:\t%v\n", unsignedBytes)
	fmt.Printf("transaction-hash:\t%v\n", hex.EncodeToString(hashedUnsignedBytes[:]))
	fmt.Printf("signature-bytes:\t%v\n", signature)
	fmt.Printf("signature-hex:\t%v\n", hex.EncodeToString(signature))
}
