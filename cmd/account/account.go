package account

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
)

type (
	// GeneratorCommands represent struct of account generator commands
	GeneratorCommands struct {
		Signature       crypto.SignatureInterface
		TransactionUtil transaction.UtilInterface
	}
	// RunCommand represent of output function from account generator commands
	RunCommand func(ccmd *cobra.Command, args []string)
)

var (
	accountGeneratorInstance *GeneratorCommands

	accountCmd = &cobra.Command{
		Use:   "account",
		Short: "account is a developer cli tools to generate account.",
		Long: `account is a developer cli tools to generate account.
running 'zoobc account generate' will show create an account detail with its public key and
private key both in bytes and hex representation + the secret phrase
	`,
	}
	ed25519AccountCmd = &cobra.Command{
		Use:   "ed25519",
		Short: "Generate account using ed25519 algorithm. This is the default zoobc account",
	}
	bitcoinAccuntCmd = &cobra.Command{
		Use:   "bitcoin",
		Short: "Generate account based on Bitcoin signature that using Elliptic Curve Digital Signature Algorithm",
	}
	multiSigCmd = &cobra.Command{
		Use:        "multisig",
		Aliases:    []string{"musig", "ms"},
		SuggestFor: []string{"mul", "multisignature", "multi-signature"},
		Short:      "Multisig allow to generate multi sig account",
		Long: "multisig allow to generate multi sig account address" +
			"provides account addresses, nonce, and minimum assignment",
	}
)

func init() {
	// ed25519
	ed25519AccountCmd.Flags().StringVar(&seed, "seed", "", "Seed that is used to generate the account")
	ed25519AccountCmd.Flags().BoolVar(&ed25519UseSlip10, "use-slip10", false, "use slip10 to generate ed25519 private key")
	// bitcoin
	bitcoinAccuntCmd.Flags().StringVar(&seed, "seed", "", "Seed that is used to generate the account")
	bitcoinAccuntCmd.Flags().Int32Var(
		&bitcoinPrivateKeyLength,
		"private-key-length",
		int32(model.PrivateKeyBytesLength_PrivateKey256Bits),
		"The length of private key Bitcoin want to generate. supported format are 32, 48 & 64 length",
	)
	bitcoinAccuntCmd.Flags().Int32Var(
		&bitcoinPublicKeyFormat,
		"public-key-format",
		int32(model.BitcoinPublicKeyFormat_PublicKeyFormatCompressed),
		"Defines the format of public key Bitcoin want to generate. 0 for compressed format & 1 for uncompressed format",
	)
	// multisig
	multiSigCmd.Flags().StringSliceVar(&multisigAddresses, "addresses", []string{}, "addresses that provides")
	multiSigCmd.Flags().Uint32Var(&multisigMinimSigs, "min-sigs", 0, "min-sigs that provide minimum signs")
	multiSigCmd.Flags().Int64Var(&multiSigNonce, "nonce", 0, "nonce that provides")
}

// Commands will return the main generate account cmd
func Commands() *cobra.Command {
	if accountGeneratorInstance == nil {
		accountGeneratorInstance = &GeneratorCommands{
			Signature:       &crypto.Signature{},
			TransactionUtil: &transaction.Util{},
		}
	}
	ed25519AccountCmd.Run = accountGeneratorInstance.GenerateEd25519Account()
	accountCmd.AddCommand(ed25519AccountCmd)
	bitcoinAccuntCmd.Run = accountGeneratorInstance.GenerateBitcoinAccount()
	accountCmd.AddCommand(bitcoinAccuntCmd)
	multiSigCmd.Run = accountGeneratorInstance.GenerateMultiSignatureAccount()
	accountCmd.AddCommand(multiSigCmd)
	return accountCmd

}

// GenerateMultiSignatureAccount to generate address for multi signature transaction
func (gc *GeneratorCommands) GenerateMultiSignatureAccount() RunCommand {
	return func(cmd *cobra.Command, args []string) {
		info := &model.MultiSignatureInfo{
			MinimumSignatures: multisigMinimSigs,
			Nonce:             multiSigNonce,
			Addresses:         multisigAddresses,
		}
		address, err := gc.TransactionUtil.GenerateMultiSigAddress(info)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println(address)
		}
	}
}

// GenerateEd25519Account to generate ed25519 account
func (gc *GeneratorCommands) GenerateEd25519Account() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		if seed == "" {
			seed = util.GetSecureRandomSeed()
		}
		var (
			signatureType                                        = model.SignatureType_DefaultSignature
			privateKey, publicKey, publicKeyString, address, err = gc.Signature.GenerateAccountFromSeed(
				signatureType,
				seed,
				ed25519UseSlip10,
			)
		)
		if err != nil {
			panic(err)
		}
		PrintAccount(
			int32(signatureType),
			seed,
			publicKeyString,
			address,
			privateKey,
			publicKey,
		)
	}
}

// GenerateBitcoinAccount to generate bitcoin account
func (gc *GeneratorCommands) GenerateBitcoinAccount() RunCommand {
	return func(ccmd *cobra.Command, args []string) {
		if seed == "" {
			seed = util.GetSecureRandomSeed()
		}
		var (
			signatureType                                        = model.SignatureType_BitcoinSignature
			privateKey, publicKey, publicKeyString, address, err = gc.Signature.GenerateAccountFromSeed(
				signatureType,
				seed,
				model.PrivateKeyBytesLength(bitcoinPrivateKeyLength),
				model.BitcoinPublicKeyFormat(bitcoinPublicKeyFormat),
			)
		)
		if err != nil {
			panic(err)
		}
		PrintAccount(
			int32(signatureType),
			seed,
			publicKeyString,
			address,
			privateKey,
			publicKey,
		)
	}
}

// PrintAccount print out the generated account
func PrintAccount(
	signatureType int32,
	seed, publicKeyString, address string,
	privateKey, publicKey []byte,
) {
	fmt.Printf("signature type: %s\n", model.SignatureType_name[signatureType])
	fmt.Printf("seed: %s\n", seed)
	fmt.Printf("public key hex: %s\n", hex.EncodeToString(publicKey))
	fmt.Printf("public key bytes: %v\n", publicKey)
	fmt.Printf("public key string : %v\n", publicKeyString)
	fmt.Printf("private key bytes: %v\n", privateKey)
	fmt.Printf("private key hex: %v\n", hex.EncodeToString(privateKey))
	fmt.Printf("address: %s\n", address)
}
