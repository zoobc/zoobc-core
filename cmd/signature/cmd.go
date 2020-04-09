package signature

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"golang.org/x/crypto/sha3"
)

type (
	// GeneratorCommands represent struct of signature generator commands
	GeneratorCommands struct {
		Signature crypto.SignatureInterface
	}
)

var (
	signatureCmdInstance *GeneratorCommands
	/*
		Signer command line tools
	*/
	signatureCmd = &cobra.Command{
		Use:   "signature",
		Short: "signature command is a parent command for signature stuffs",
	}
	signerCmd = &cobra.Command{
		Use:   "sign",
		Short: "sign provided data",
		Long:  "sign any provided data by using the --seed parameter",
	}

	ed25519SignerCmd = &cobra.Command{
		Use:   "ed25519",
		Short: "sign using ed25519 algoritmn",
		Long:  "sign any provided data by using the --seed parameter",
	}

	verifyCmd = &cobra.Command{
		Use:   "verify",
		Short: "verify provided signature ",
		Long:  "verify provided signature against provided data using public key",
	}
)

func init() {
	signatureCmd.PersistentFlags().StringVar(&dataHex, "data-hex", "", "hex string of the data to sign")
	signatureCmd.PersistentFlags().StringVar(&dataBytes, "data-bytes", "", "data bytes separated by `, `. eg:"+
		"--data-bytes='1, 222, 54, 12, 32'")

	signerCmd.PersistentFlags().StringVar(&seed, "seed", "", "your secret phrase")
	signerCmd.PersistentFlags().BoolVar(&hash, "hash", false, "turn this flag on to hash the data before signing")
	ed25519SignerCmd.Flags().BoolVar(&ed25519UseSpli10, "use-slip10", false, "use slip10 to generate ed25519 private key for signing")

	verifyCmd.Flags().StringVar(&signatureHex, "signature-hex", "", "hex string of the signature")
	verifyCmd.Flags().StringVar(&signatureBytes, "signature-bytes", "", "signature bytes stseparated by `, `. eg:"+
		"--signature-bytes='1, 222, 54, 12, 32'")
	verifyCmd.Flags().StringVar(&accountAddress, "account-address", "", "the address who sign the data")

}

// Commands return main command of signature
func Commands() *cobra.Command {
	if signatureCmdInstance == nil {
		signatureCmdInstance = &GeneratorCommands{
			Signature: &crypto.Signature{},
		}
	}
	ed25519SignerCmd.Run = signatureCmdInstance.SignEd25519
	signerCmd.AddCommand(ed25519SignerCmd)
	signerCmd.Run = signatureCmdInstance.SignEd25519
	signatureCmd.AddCommand(signerCmd)

	verifyCmd.Run = signatureCmdInstance.VerySignature
	signatureCmd.AddCommand(verifyCmd)
	return signatureCmd
}

// SignEd25519 is sign command hanlder using Ed25519 algorithm
func (gc *GeneratorCommands) SignEd25519(*cobra.Command, []string) {
	var (
		unsignedBytes       []byte
		hashedUnsignedBytes [32]byte
		accountAddress      string
		signature           []byte
		err                 error
	)

	if dataHex != "" {
		unsignedBytes, err = hex.DecodeString(dataHex)
		if err != nil {
			panic("failed to decode data hex")
		}
	} else {
		unsignedBytes, err = parseBytesArgs(dataBytes, ", ")
		if err != nil {
			panic("failed to parse data bytes")
		}
	}
	_, _, _, accountAddress, err = gc.Signature.GenerateAccountFromSeed(
		model.SignatureType_DefaultSignature,
		seed,
		ed25519UseSpli10,
	)
	if err != nil {
		panic(err.Error())
	}
	if hash {
		hashedUnsignedBytes = sha3.Sum256(unsignedBytes)
		unsignedBytes = hashedUnsignedBytes[:]
	}
	signature, err = gc.Signature.Sign(
		unsignedBytes,
		model.SignatureType_DefaultSignature,
		seed,
		ed25519UseSpli10,
	)
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("account-address:\t%v\n", accountAddress)
	fmt.Printf("transaction-bytes:\t%v\n", unsignedBytes)
	fmt.Printf("transaction-hash:\t%v\n", hex.EncodeToString(hashedUnsignedBytes[:]))
	fmt.Printf("signature-bytes:\t%v\n", signature)
	fmt.Printf("signature-hex:\t%v\n", hex.EncodeToString(signature))
}

// VerySignature is verify signature command hendler
func (gc *GeneratorCommands) VerySignature(*cobra.Command, []string) {
	var (
		unsignedBytes     []byte
		signature         []byte
		failedVerifyCause = "none"
		isVerified        = true
		err               error
	)
	if dataHex != "" {
		unsignedBytes, err = hex.DecodeString(dataHex)
		if err != nil {
			panic("failed to decode data hex")
		}
	} else {
		unsignedBytes, err = parseBytesArgs(dataBytes, ", ")
		if err != nil {
			panic("failed to parse data bytes")
		}
	}

	if signatureHex != "" {
		signature, err = hex.DecodeString(signatureHex)
		if err != nil {
			panic("failed to decode signature hex")
		}
	} else {
		signature, err = parseBytesArgs(signatureBytes, ", ")
		if err != nil {
			panic("failed to parse data bytes")
		}
	}

	err = gc.Signature.VerifySignature(unsignedBytes, signature, accountAddress)
	if err != nil {
		failedVerifyCause = err.Error()
		isVerified = false
	}

	fmt.Printf("verify-status:\t%v\n", isVerified)
	fmt.Printf("failed-causes:\t%v\n", failedVerifyCause)
	fmt.Printf("address:\t%v\n", accountAddress)
	fmt.Printf("payload-bytes:\t%v\n", unsignedBytes)
	fmt.Printf("payload-hex:\t%v\n", hex.EncodeToString(unsignedBytes))

	fmt.Printf("signature-hex:\t%v\n", hex.EncodeToString(signature))
	fmt.Printf("signature-bytes:%v\n", signature)

}

func parseBytesArgs(argsBytesString, separated string) ([]byte, error) {
	var (
		parsedByte    []byte
		byteCharSlice = strings.Split(dataBytes, separated)
	)
	for _, v := range byteCharSlice {
		byteValue, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		parsedByte = append(parsedByte, byte(byteValue))
	}
	return parsedByte, nil
}
