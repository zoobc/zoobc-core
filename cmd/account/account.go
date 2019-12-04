package account

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/common/util"
)

var (
	seed string

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
			generateAccountFromSeed(seed)
		},
	}
)

func init() {
	accountCmd.AddCommand(randomAccountCmd)

	fromSeedCmd.Flags().StringVar(&seed, "seed", "", "Seed that is used to generate the account")
	accountCmd.AddCommand(fromSeedCmd)
}

func Commands() *cobra.Command {
	return accountCmd
}

func generateRandomAccount() {
	seed := util.GetSecureRandomSeed()
	generateAccountFromSeed(seed)
}

func generateAccountFromSeed(seed string) {
	var (
		privateKey, _ = util.GetPrivateKeyFromSeed(seed)
		publicKey     = privateKey[32:]
		address, _    = util.GetAddressFromPublicKey(publicKey)
	)
	fmt.Printf("seed: %s\n", seed)
	fmt.Printf("public key hex: %s\n", hex.EncodeToString(publicKey))
	fmt.Printf("public key bytes: %v\n", publicKey)
	fmt.Printf("public key string : %v\n", base64.StdEncoding.EncodeToString(publicKey))
	fmt.Printf("private key bytes: %v\n", privateKey)
	fmt.Printf("private key hex: %v\n", hex.EncodeToString(privateKey))
	fmt.Printf("address: %s\n", address)
}
