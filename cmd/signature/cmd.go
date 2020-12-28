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
package signature

import (
	"encoding/hex"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/cmd/helper"
	"github.com/zoobc/zoobc-core/common/accounttype"
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
	ed25519SignerCmd.Flags().BoolVar(&ed25519UseSlip10, "use-slip10", false, "use slip10 to generate ed25519 private key for signing")

	verifyCmd.Flags().StringVar(&signatureHex, "signature-hex", "", "hex string of the signature")
	verifyCmd.Flags().StringVar(&signatureBytes, "signature-bytes", "", "signature bytes stseparated by `, `. eg:"+
		"--signature-bytes='1, 222, 54, 12, 32'")
	verifyCmd.Flags().StringVar(&accountAddressHex, "account-address", "", "the address who sign the data")

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

// SignEd25519 is sign command handler using Ed25519 algorithm
func (gc *GeneratorCommands) SignEd25519(*cobra.Command, []string) {
	var (
		unsignedBytes         []byte
		hashedUnsignedBytes   [32]byte
		encodedAccountAddress string
		fullAccountAddress    []byte
		signature             []byte
		err                   error
	)

	if dataHex != "" {
		unsignedBytes, err = hex.DecodeString(dataHex)
		if err != nil {
			panic("failed to decode data hex")
		}
	} else {
		unsignedBytes, err = helper.ParseBytesArgument(dataBytes, ", ")
		if err != nil {
			panic("failed to parse data bytes")
		}
	}
	accType := &accounttype.ZbcAccountType{}
	_, _, _, encodedAccountAddress, fullAccountAddress, err = gc.Signature.GenerateAccountFromSeed(
		accType,
		seed,
		ed25519UseSlip10,
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
		model.AccountType_ZbcAccountType,
		seed,
		ed25519UseSlip10,
	)
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("account-address type:\t%d\n", accType.GetTypeInt())
	fmt.Printf("encoded account-address:\t%v\n", encodedAccountAddress)
	fmt.Printf("account-address:\t%v\n", fullAccountAddress)
	fmt.Printf("account-address hex:\t%v\n", hex.EncodeToString(fullAccountAddress))
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
		unsignedBytes, err = helper.ParseBytesArgument(dataBytes, ", ")
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
		signature, err = helper.ParseBytesArgument(signatureBytes, ", ")
		if err != nil {
			panic("failed to parse data bytes")
		}
	}

	decodedAccountAddress, err := hex.DecodeString(accountAddressHex)
	if err != nil {
		panic(err)
	}
	err = gc.Signature.VerifySignature(unsignedBytes, signature, decodedAccountAddress)
	if err != nil {
		failedVerifyCause = err.Error()
		isVerified = false
	}

	fmt.Printf("verify-status:\t%v\n", isVerified)
	fmt.Printf("failed-causes:\t%v\n", failedVerifyCause)
	fmt.Printf("address:\t%v\n", accountAddressHex)
	fmt.Printf("payload-bytes:\t%v\n", unsignedBytes)
	fmt.Printf("payload-hex:\t%v\n", hex.EncodeToString(unsignedBytes))

	fmt.Printf("signature-hex:\t%v\n", hex.EncodeToString(signature))
	fmt.Printf("signature-bytes:%v\n", signature)

}
