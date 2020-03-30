package noderegistry

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
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
		Short: "geneate proof of ownership for node registry transaction",
		Long: `geneate proof of ownership for transaction related with node registry.
			For example:  register node transaction, update node transaction & claim node transaction`,
	}
)

func init() {
	generateProofOfOwnerShipCmd.Flags().StringVar(&outputType, "output-type", "hex",
		"defines the type of the output to be generated [\"hex\", \"bytes\"]")
	generateProofOfOwnerShipCmd.Flags().StringVar(&nodeOwnerAccountAddress, "node-owner-account-address", "",
		"Account address of the owner of the node")
	generateProofOfOwnerShipCmd.Flags().StringVar(&nodeSeed, "node-seed", "", "Private key of the node")
	generateProofOfOwnerShipCmd.Flags().StringVar(&databasePath, "db-node-path", "../resource", "Database path of node, "+
		"make sure to download the database from node or run this command on node")
	generateProofOfOwnerShipCmd.Flags().StringVar(&databaseName, "db-node-name", "zoobc.db", "Database name of node, "+
		"make sure to download the database from node or run this command on node")
}

// Commands will return  proof of owner ship cmd
func Commands() *cobra.Command {
	generateProofOfOwnerShipCmd.Run = GenerateProofOfOwnerShip
	return generateProofOfOwnerShipCmd
}

// GenerateProofOfOwnerShip for generate Proof of ownership node registry
func GenerateProofOfOwnerShip(*cobra.Command, []string) {
	var (
		poow = GetProofOfOwnerShip(
			databasePath,
			databaseName,
			nodeOwnerAccountAddress,
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

// GetProofOfOwnerShip will reuturn proof of ownership basd on provided nodeOwnerAccountAddress & nodeSeed in a DB
func GetProofOfOwnerShip(
	dbPath, dbname, nodeOwnerAccountAddress, nodeSeed string,
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
	poowMessage := &model.ProofOfOwnershipMessage{
		AccountAddress: nodeOwnerAccountAddress,
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
