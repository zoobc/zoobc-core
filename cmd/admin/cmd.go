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
package admin

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/signaturetype"
	"io/ioutil"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/zoobc/lib/address"
	"github.com/zoobc/zoobc-core/cmd/helper"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/database"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
)

var (
	/*
		Generate Proof of Ownership Node Registry Command
	*/
	generateProofOfOwnerShipCmd = &cobra.Command{
		Use:   "poow",
		Short: "generate proof of ownership for node registry transaction",
		Long: `generate proof of ownership for transaction related with node registry.
			For example:  register node transaction, update node transaction & claim node transaction`,
	}
	generateNodeKeyCmd = &cobra.Command{
		Use:   "node-key",
		Short: "generate node_keys.json",
		Long:  "generate node_keys.json file that needed for proof of ownership. Will store into resource directory",
	}
)

func init() {
	generateProofOfOwnerShipCmd.Flags().StringVar(&outputType, "output-type", "hex",
		"defines the type of the output to be generated [\"hex\", \"bytes\"]")
	generateProofOfOwnerShipCmd.Flags().StringVar(&nodeOwnerAccountAddressHex, "node-owner-account-address", "",
		"Account address of the owner of the node")
	generateProofOfOwnerShipCmd.Flags().StringVar(&nodeSeed, "node-seed", "", "Private key of the node")
	generateProofOfOwnerShipCmd.Flags().StringVar(&databasePath, "db-node-path", "../resource", "Database path of node, "+
		"make sure to download the database from node or run this command on node")
	generateProofOfOwnerShipCmd.Flags().StringVar(&databaseName, "db-node-name", "zoobc.db", "Database name of node, "+
		"make sure to download the database from node or run this command on node")
	generateNodeKeyCmd.Flags().StringVar(&nodeSeed, "node-seed", "", "Private key of the node, empty allowed")
}

// Commands will return  proof of owner ship cmd
func Commands() *cobra.Command {
	generateProofOfOwnerShipCmd.Run = GenerateProofOfOwnerShip
	generateNodeKeyCmd.Run = generateNodeKeysCommand

	commands := &cobra.Command{
		Use:   "node-admin",
		Short: "node admin command",
		Long:  "node admin command stuff, proof of ownership stuff",
	}
	commands.AddCommand(generateProofOfOwnerShipCmd)
	commands.AddCommand(generateNodeKeyCmd)
	return commands
}

// GenerateProofOfOwnerShip for generate Proof of ownership node registry
func GenerateProofOfOwnerShip(*cobra.Command, []string) {
	var (
		poow = GetProofOfOwnerShip(
			databasePath,
			databaseName,
			nodeOwnerAccountAddressHex,
			nodeSeed)
		poowBytes = util.GetProofOfOwnershipBytes(poow)
	)

	switch outputType {
	case "hex":
		fmt.Printf("Proof Of Owner Ship Hex:\n%v\n", hex.EncodeToString(poowBytes))
	case "bytes":
		fmt.Printf("Proof Of Owner Ship Bytes:\n%v\n", poowBytes)
	default:
		panic("Invalid Output type")
	}
}

// GetProofOfOwnerShip will return proof of ownership based on provided nodeOwnerAccountAddressHex & nodeSeed in a DB
func GetProofOfOwnerShip(
	dbPath, dbname, nodeOwnerAccountAddressHex, nodeSeed string,
) *model.ProofOfOwnership {
	var (
		sqliteDbInstance = database.NewSqliteDB()
		sqliteDB, err    = sqliteDbInstance.OpenDB(
			databasePath,
			databaseName,
			constant.SQLMaxOpenConnetion,
			constant.SQLMaxIdleConnections,
			constant.SQLMaxConnectionLifetime,
		)
	)
	if err != nil {
		panic(fmt.Sprintf("OpenDB err: %s", err.Error()))
	}
	lastBlock, err := util.GetLastBlock(query.NewQueryExecutor(sqliteDB), query.NewBlockQuery(chaintype.GetChainType(0)))
	if err != nil {
		panic(err)
	}
	decodedAddress, err := hex.DecodeString(nodeOwnerAccountAddressHex)
	if err != nil {
		panic(err)
	}
	poowMessage := &model.ProofOfOwnershipMessage{
		AccountAddress: decodedAddress,
		BlockHash:      lastBlock.BlockHash,
		BlockHeight:    lastBlock.Height,
	}

	poownMessageBytes := util.GetProofOfOwnershipMessageBytes(poowMessage)
	signature := (&crypto.Signature{}).SignByNode(
		poownMessageBytes,
		nodeSeed)

	return &model.ProofOfOwnership{
		MessageBytes: poownMessageBytes,
		Signature:    signature,
	}
}

func generateNodeKeysCommand(*cobra.Command, []string) {
	GenerateNodeKeysFile(nodeSeed)
}

func GenerateNodeKeysFile(seed string) {

	var (
		err     error
		b       []byte
		nodeKey *model.NodeKeyFromFile
	)

	if len(seed) < 1 {
		seed = util.GetSecureRandomSeed()
	}
	nodeKey = &model.NodeKeyFromFile{
		Seed: seed,
	}
	pubKey := signaturetype.NewEd25519Signature().GetPublicKeyFromSeed(seed)
	publicKeyStr, err := address.EncodeZbcID(constant.PrefixZoobcNodeAccount, pubKey)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
	nodeKey.PublicKey = publicKeyStr

	b, err = json.MarshalIndent([]*model.NodeKeyFromFile{
		nodeKey,
	}, "", "")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}

	err = ioutil.WriteFile(path.Join(helper.GetAbsDBPath(), "/resource/node_keys.json"), b, 0644)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
}
