package admin

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/zoobc/lib/address"
	"github.com/zoobc/zoobc-core/cmd/helper"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
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
	pubKey := crypto.NewEd25519Signature().GetPublicKeyFromSeed(seed)
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
