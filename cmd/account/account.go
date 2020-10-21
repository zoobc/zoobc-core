package account

import (
	"encoding/hex"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/common/accounttype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/signaturetype"
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
	convAccuntToHexCmd = &cobra.Command{
		Use:   "hexconv",
		Short: "Convert a given (encoded/string) account address to hex format",
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
	convAccuntToHexCmd.Flags().StringVar(&encodedAccountAddress, "encodedAccountAddress", "",
		"formatted/encoded account address. eg. ZBC_F5YUYDXD_WFDJSAV5_K3Y72RCM_GLQP32XI_QDVXOGGD_J7CGSSSK_5VKR7YML")
	convAccuntToHexCmd.Flags().Int32Var(&accountTypeInt, "accountType", 0, "Account type num: 0=default, 1=btc, etc..")
	bitcoinAccuntCmd.Flags().Int32Var(
		&bitcoinPublicKeyFormat,
		"public-key-format",
		int32(model.BitcoinPublicKeyFormat_PublicKeyFormatCompressed),
		"Defines the format of public key Bitcoin want to generate. 0 for uncompressed format & 1 for compressed format",
	)
	// multisig
	multiSigCmd.Flags().StringSliceVar(&multisigAddressesHex, "addresses", []string{},
		"addresses that provides in hex format. decoded accountAddress is in the form of a byte array with: first 4 bytes = accountType, "+
			"remaining bytes = account public key")
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
	convAccuntToHexCmd.Run = accountGeneratorInstance.ConvertEncodedAccountAddressToHex()
	accountCmd.AddCommand(convAccuntToHexCmd)
	multiSigCmd.Run = accountGeneratorInstance.GenerateMultiSignatureAccount()
	accountCmd.AddCommand(multiSigCmd)
	return accountCmd

}

// GenerateMultiSignatureAccount to generate address for multi signature transaction
func (gc *GeneratorCommands) ConvertEncodedAccountAddressToHex() RunCommand {
	return func(cmd *cobra.Command, args []string) {
		var (
			accPubKey []byte
			err       error
		)
		switch accountTypeInt {
		case 0:
			ed25519 := signaturetype.NewEd25519Signature()
			accPubKey, err = ed25519.GetPublicKeyFromEncodedAddress(encodedAccountAddress)
			if err != nil {
				panic(err)
			}
		case 1:
			bitcoinSignature := signaturetype.NewBitcoinSignature(signaturetype.DefaultBitcoinNetworkParams(), signaturetype.DefaultBitcoinCurve())
			accPubKey, err = bitcoinSignature.GetAddressBytes(encodedAccountAddress)
			if err != nil {
				panic(err)
			}
		}
		accType, err := accounttype.NewAccountType(accountTypeInt, accPubKey)
		if err != nil {
			panic(err)
		}
		fullAccountAddress, err := accType.GetAccountAddress()
		if err != nil {
			panic(err)
		}
		fmt.Printf("account address type: %s (%d)\n", model.AccountType_name[accountTypeInt], accountTypeInt)
		fmt.Printf("encoded account address: %s\n", encodedAccountAddress)
		fmt.Printf("public key hex: %s\n", hex.EncodeToString(accPubKey))
		fmt.Printf("public key bytes: %v\n", accPubKey)
		fmt.Printf("full account address: %v\n", fullAccountAddress)
		fmt.Printf("full account address hex: %v\n", hex.EncodeToString(fullAccountAddress))
	}
}

// GenerateMultiSignatureAccount to generate address for multi signature transaction
func (gc *GeneratorCommands) GenerateMultiSignatureAccount() RunCommand {
	var (
		multisigFullAccountAddresses [][]byte
	)
	for _, accAddrHex := range multisigAddressesHex {
		decodedAddr, err := hex.DecodeString(accAddrHex)
		if err != nil {
			panic(err)
		}
		multisigFullAccountAddresses = append(multisigFullAccountAddresses, decodedAddr)
	}
	return func(cmd *cobra.Command, args []string) {
		info := &model.MultiSignatureInfo{
			MinimumSignatures: multisigMinimSigs,
			Nonce:             multiSigNonce,
			Addresses:         multisigFullAccountAddresses,
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
			accountType                                                              = &accounttype.ZbcAccountType{}
			privateKey, publicKey, publicKeyString, address, fullAccountAddress, err = gc.Signature.GenerateAccountFromSeed(
				accountType,
				seed,
				ed25519UseSlip10,
			)
		)
		if err != nil {
			panic(err)
		}
		PrintAccount(
			accountType,
			seed,
			publicKeyString,
			address,
			privateKey,
			publicKey,
			fullAccountAddress,
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
			accountType                                                              = &accounttype.BTCAccountType{}
			privateKey, publicKey, publicKeyString, address, fullAccountAddress, err = gc.Signature.GenerateAccountFromSeed(
				accountType,
				seed,
				model.PrivateKeyBytesLength(bitcoinPrivateKeyLength),
				model.BitcoinPublicKeyFormat(bitcoinPublicKeyFormat),
			)
		)
		if err != nil {
			panic(err)
		}
		PrintAccount(
			accountType,
			seed,
			publicKeyString,
			address,
			privateKey,
			publicKey,
			fullAccountAddress,
		)
	}
}

// PrintAccount print out the generated account
func PrintAccount(
	accountType accounttype.AccountTypeInterface,
	seed, publicKeyString, encodedAddress string,
	privateKey, publicKey, fullAccountAddress []byte,
) {
	fmt.Printf("account type: %s\n", model.AccountType_name[accountType.GetTypeInt()])
	fmt.Printf("signature type: %s\n", model.SignatureType_name[int32(accountType.GetSignatureType())])
	fmt.Printf("seed: %s\n", seed)
	fmt.Printf("public key hex: %s\n", hex.EncodeToString(publicKey))
	fmt.Printf("public key bytes: %v\n", publicKey)
	fmt.Printf("public key string : %v\n", publicKeyString)
	fmt.Printf("private key bytes: %v\n", privateKey)
	fmt.Printf("private key hex: %v\n", hex.EncodeToString(privateKey))
	fmt.Printf("encodedAddress: %s\n", encodedAddress)
	fmt.Printf("full account address: %v\n", fullAccountAddress)
	fmt.Printf("full account address hex: %v\n", hex.EncodeToString(fullAccountAddress))
}
