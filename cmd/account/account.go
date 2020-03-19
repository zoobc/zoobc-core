package account

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	slip10 "github.com/zoobc/zoo-slip10"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
)

var (
	accountCmd = &cobra.Command{
		Use:   "account",
		Short: "account is a developer cli tools to generate account.",
		Long: `account is a developer cli tools to generate account.
running 'zoobc account generate' will show create an account detail with its public key and
private key both in bytes and hex representation + the secret phrase
	`,
		Run: func(cmd *cobra.Command, args []string) {
			generateRandomAccount()
		},
	}

	randomAccountCmd = &cobra.Command{
		Use:   "random",
		Short: "random defines to generate random account.",
		Run: func(cmd *cobra.Command, args []string) {
			generateRandomAccount()
		},
	}

	fromSeedCmd = &cobra.Command{
		Use:   "from-seed",
		Short: "from-seed defines to generate account from provided seed.",
		Run: func(cmd *cobra.Command, args []string) {
			generateAccountFromSeed(signatureType, seed)
		},
	}

	multiSigCmd = &cobra.Command{
		Use:        "multisig",
		Aliases:    []string{"musig", "ms"},
		SuggestFor: []string{"mul", "multisignature", "multi-signature"},
		Short:      "multisig allow to generate multi sig account",
		Long: "multisig allow to generate multi sig account address" +
			"provides account addresses, nonce, and minimum assignment",
		Run: func(cmd *cobra.Command, args []string) {
			info := &model.MultiSignatureInfo{
				MinimumSignatures: multisigMinimSigs,
				Nonce:             multiSigNonce,
				Addresses:         multisigAddresses,
			}
			address, err := (&transaction.Util{}).GenerateMultiSigAddress(info)
			if err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println(address)
			}

		},
	}
)

func init() {
	accountCmd.PersistentFlags().Int32Var(
		&signatureType,
		"signature-type",
		int32(model.SignatureType_DefaultSignature),
		"signature-type that provide type of signature want to use to generate the account",
	)
	accountCmd.AddCommand(randomAccountCmd)

	fromSeedCmd.Flags().StringVar(&seed, "seed", "", "Seed that is used to generate the account")
	fromSeedCmd.Flags().BoolVar(&hd, "hd", true, "--hd allow to generate account HD")
	accountCmd.AddCommand(fromSeedCmd)

	// multisig
	multiSigCmd.Flags().StringSliceVar(&multisigAddresses, "addresses", []string{}, "addresses that provides")
	multiSigCmd.Flags().Uint32Var(&multisigMinimSigs, "min-sigs", 0, "min-sigs that provide minimum signs")
	multiSigCmd.Flags().Int64Var(&multiSigNonce, "nonce", 0, "nonce that provides")
	accountCmd.AddCommand(multiSigCmd)
}

// Commands will return the main generate account cmd
func Commands() *cobra.Command {
	return accountCmd
}

func generateRandomAccount() {
	seed = util.GetSecureRandomSeed()
	generateAccountFromSeed(signatureType, seed)
}

func generateAccountFromSeed(signatureType int32, seed string) {
	var (
		signature                        = crypto.Signature{}
		seedBytes, privateKey, publicKey []byte
		publicKeyString, address         string
		err                              error
		slip10Key                        *slip10.Key
	)

	if hd {

		seedBytes = slip10.NewSeed(seed, slip10.DefaultPassword)
		slip10Key, err = slip10.DeriveForPath(slip10.StellarPrimaryAccountPath, seedBytes)
		if err != nil {
			panic(err)
		}

		privateKey = slip10Key.Key
		publicKey, _ = slip10Key.PublicKey()
		publicKeyString = base64.StdEncoding.EncodeToString(publicKey)
		address, _ = crypto.NewEd25519Signature().GetAddressFromPublicKey(publicKey)
	} else {
		privateKey, publicKey, publicKeyString, address, err = signature.GenerateAccountFromSeed(model.SignatureType(signatureType), seed)
		if err != nil {
			panic(err)
		}
	}

	fmt.Printf("signature type: %s\n", model.SignatureType_name[signatureType])
	fmt.Printf("seed: %s\n", seed)
	fmt.Printf("public key hex: %s\n", hex.EncodeToString(publicKey))
	fmt.Printf("public key bytes: %v\n", publicKey)
	fmt.Printf("public key string : %v\n", publicKeyString)
	fmt.Printf("private key bytes: %v\n", privateKey)
	fmt.Printf("private key hex: %v\n", hex.EncodeToString(privateKey))
	fmt.Printf("address: %s\n", address)
}
