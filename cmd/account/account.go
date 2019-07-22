package account

import (
	"encoding/hex"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tyler-smith/go-bip39"
	"github.com/zoobc/zoobc-core/common/util"
)

func GenerateAccount(logger *logrus.Logger) *cobra.Command {
	var accountCmd = &cobra.Command{
		Use:   "account",
		Short: "account is a developer cli tools to generate account.",
		Long: `account is a developer cli tools to generate account.
running 'zoobc account generate' will show create an account detail with its public key and
private key both in bytes and hex representation + the secret phrase
		`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if args[0] == "generate" {
				entropy, _ := bip39.NewEntropy(128)
				seed, _ := bip39.NewMnemonic(entropy)
				privateKey, _ := util.GetPrivateKeyFromSeed(seed)
				publicKey := privateKey[32:]
				address, _ := util.GetAddressFromPublicKey(publicKey)
				logger.Infof("seed: %s", seed)
				logger.Infof("public key hex: %s", hex.EncodeToString(publicKey))
				logger.Infof("public key bytes: %v", publicKey)
				logger.Infof("private key bytes: %v", privateKey)
				logger.Infof("private key hex: %v", hex.EncodeToString(privateKey))
				logger.Infof("address: %s", address)

			} else {
				logger.Error("unknown command")
			}
		},
	}
	return accountCmd
}
