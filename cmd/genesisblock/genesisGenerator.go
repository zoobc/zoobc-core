package genesisblock

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/database"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/service"
)

type (
	genesisEntry struct {
		AccountAddress     string
		AccountBalance     int64
		NodeSeed           string
		NodePublicKeyB64   string
		NodePublicKey      []byte
		NodeAddress        string
		LockedBalance      int64
		ParticipationScore int64
		Smithing           bool
	}
	clusterConfigEntry struct {
		NodePublicKey       string `json:"NODE_PUBLIC_KEY"`
		NodeSeed            string `json:"NODE_SEED"`
		OwnerAccountAddress string `json:"OWNER_ACCOUNT_ADDRESS"`
		NodeAddress         string `json:"NODE_ADDRESS,omitempty"`
		Smithing            bool   `json:"SMITHING,omitempty"`
	}
)

var (
	withDbLastState bool
	dbPath          string
	extraNodesCount int

	genesisCmd = &cobra.Command{
		Use:   "genesis",
		Short: "command used to generate a new genesis block.",
	}

	genesisGeneratorCmd = &cobra.Command{
		Use:   "generate",
		Short: "genesis generate command used to generate a new genesis.go and cluster_config.json file",
		Long: `genesis generate command generate a genesis.go file from a list of accounts and/or from current database.
		the latter is to be used when we want to reset the blockchain mantaining the latest state of accounts and node registrations`,
		Run: func(cmd *cobra.Command, args []string) {
			generateGenesisFiles(withDbLastState, dbPath, extraNodesCount)
		},
	}
)

func init() {
	genesisGeneratorCmd.Flags().BoolVarP(&withDbLastState, "withDbLastState", "w", false,
		"add to genesis all registered nodes and account balances from current database")
	genesisGeneratorCmd.Flags().StringVarP(&dbPath, "dbPath", "f", "../resource/",
		"path of blockchain's database to be used as data source in case the -w flag is used. If not set, the default resource folder is used")
	genesisGeneratorCmd.Flags().IntVarP(&extraNodesCount, "extraNodes", "n", 0,
		"number of 'extra' autogenerated nodes to be deployed using cluster_config.json")
	genesisCmd.AddCommand(genesisGeneratorCmd)
}

func Commands() *cobra.Command {
	return genesisCmd
}

// generateGenesisFiles generate genesis files starting from a source json file.
// PreRegisteredNodes contains a list of known nodes-accountOwners' public keys to be included in genesis.
// Data that can be pre-set for node registration and and account balance are:
//   AccountAddress (mandatory): node's owner
//   AccountBalance (for funded accounts only): the balance of that account at genesis block
//   NodeSeed (this should be set only for testing nodes): it will be copied into cluster_config.json to
//       automatically deploy new nodes that are already registered
//   NodePublicKeyB64 (mandatory): base64 encoded node public key to be registered
//   NodeAddress (optional): if known, the node address that will be registered and put in cluster_config.json too
//   LockedBalance (optional): account's locked balance
//   ParticipationScore (optional): set custom initial participation score (mainly for testing the smith process and POP algorithm).
//       if not set, defaults to constant.DefaultParticipationScore
//
// withDbLastState if set to true, we also scan a given blockchain database and extract latest state to be included in genesis
//  (account balances and registered nodes/participation scores)
func generateGenesisFiles(withDbLastState bool, dbPath string, extraNodesCount int) {
	var (
		bcState, preRegisteredNodes []genesisEntry
		err                         error
	)

	if withDbLastState {
		bcState, err = getDbLastState(dbPath)
		if err != nil {
			log.Fatal(err)
		}
	}

	file, err := ioutil.ReadFile("./genesisblock/preRegisteredNodes.json")
	if err != nil {
		log.Fatalf("Error reading preRegisteredNodes.json file: %s", err)
	}
	err = json.Unmarshal(file, &preRegisteredNodes)
	if err != nil {
		log.Fatalf("preRegisteredNodes.json parsing error: %s", err)
	}

	// merge duplicates: if preRegisteredNodes contains entries that are in db too, add the parameters that are't available in db,
	// which is are NodeSeed and Smithing
	for _, prNode := range preRegisteredNodes {
		found := false
		for i, e := range bcState {
			if prNode.AccountAddress != e.AccountAddress {
				continue
			}
			bcState[i].NodeSeed = prNode.NodeSeed
			bcState[i].Smithing = prNode.Smithing
			pubKey, err := base64.StdEncoding.DecodeString(prNode.NodePublicKeyB64)
			if err != nil {
				log.Fatal(err)
			}
			bcState[i].NodePublicKey = pubKey
			found = true
			break
		}
		if !found {
			prNode.NodePublicKey, err = base64.StdEncoding.DecodeString(prNode.NodePublicKeyB64)
			if err != nil {
				log.Fatal(err)
			}
			bcState = append(bcState, prNode)
		}
	}

	// generate extra nodes to be deployed using cluster_config.json
	for i := 0; i < extraNodesCount; i++ {
		bcState = append(bcState, generateRandomGenesisEntry())
	}
	// append to preRegistered nodes/accounts previous entries from a blockchain db file
	generateGenesisFile(bcState, "./genesis.go.new")
	generateClusterConfigFile(bcState, "./cluster_config.json.new")
	fmt.Println("Command executed successfully\ngenesis.go.new has been generated in cmd directory")

}

// generateRandomGenesisEntry randomly generates a genesis node entry
// note: the account address is mandatory for the node registration, but as there is no wallet connected to it
//       and we are not storing the relaitve seed, needed to sign transactions, these nodes can smith but their owners
//       can't perform any transaction.
//       This is only useful to test multiple smithing-nodes, for instence in a network stress test of tens of nodes connected together
func generateRandomGenesisEntry() genesisEntry {
	seed := util.GetSecureRandomSeed()
	privateKey, _ := util.GetPrivateKeyFromSeed(seed)
	publicKey := privateKey[32:]
	address, _ := util.GetAddressFromPublicKey(publicKey)

	nodeSeed := util.GetSecureRandomSeed()
	nodePrivateKey, _ := util.GetPrivateKeyFromSeed(nodeSeed)
	nodePublicKey := nodePrivateKey[32:]

	return genesisEntry{
		AccountAddress:     address,
		NodePublicKey:      nodePublicKey,
		NodePublicKeyB64:   base64.StdEncoding.EncodeToString(nodePublicKey),
		NodeSeed:           nodeSeed,
		ParticipationScore: constant.DefaultParticipationScore,
		Smithing:           true,
		LockedBalance:      10000 * constant.OneZBC,
	}
}

func getDbLastState(dbPath string) (bcEntries []genesisEntry, err error) {
	var (
		db *sql.DB
	)
	_, err = os.Stat(dbPath)
	if ok := os.IsNotExist(err); ok {
		return nil, err
	}

	dbInstance := database.NewSqliteDB()
	db, err = dbInstance.OpenDB(dbPath, "zoobc.db", 10, 20)
	if err != nil {
		log.Fatal(err)
	}
	queryExecutor := query.NewQueryExecutor(db)
	accountBalanceQuery := query.NewAccountBalanceQuery()
	nodeRegistrationQuery := query.NewNodeRegistrationQuery()
	participationScoreQuery := query.NewParticipationScoreQuery()
	// get all account balances
	// get the participation score for this node registration
	qry := accountBalanceQuery.GetAccountBalances()
	rows, err := queryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	accountBalances, err := accountBalanceQuery.BuildModel([]*model.AccountBalance{}, rows)
	if err != nil {
		return nil, err
	}
	for _, acc := range accountBalances {
		if acc.AccountAddress == constant.MainchainGenesisAccountAddress {
			continue
		}
		bcEntry := new(genesisEntry)
		bcEntry.AccountAddress = acc.AccountAddress
		bcEntry.AccountBalance = acc.Balance

		// get node registration for this account, if exists
		qry, args := nodeRegistrationQuery.GetNodeRegistrationByAccountAddress(acc.AccountAddress)
		rows, err := queryExecutor.ExecuteSelect(qry, false, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		nodeRegistrations, err := nodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows)
		if err != nil {
			return nil, err
		}

		if len(nodeRegistrations) > 0 {
			nr := nodeRegistrations[0]
			bcEntry.LockedBalance = nr.LockedBalance
			if nr.NodeAddress.Port > 0 {
				bcEntry.NodeAddress = fmt.Sprintf("%s:%d", nr.NodeAddress.Address, nr.NodeAddress.Port)
			} else {
				bcEntry.NodeAddress = nr.NodeAddress.Address
			}
			bcEntry.NodePublicKey = nr.NodePublicKey
			bcEntry.NodePublicKeyB64 = base64.StdEncoding.EncodeToString(nr.NodePublicKey)
			// get the participation score for this node registration
			qry, args := participationScoreQuery.GetParticipationScoreByNodeID(nr.NodeID)
			rows, err := queryExecutor.ExecuteSelect(qry, false, args...)
			if err != nil {
				return nil, err
			}
			defer rows.Close()

			participationScores, err := participationScoreQuery.BuildModel([]*model.ParticipationScore{}, rows)
			if (err != nil) || len(participationScores) > 0 {
				bcEntry.ParticipationScore = participationScores[0].Score
			}
		}
		bcEntries = append(bcEntries, *bcEntry)
	}

	return bcEntries, err
}

// generateGenesisFile generates a genesis file with given entries, starting from a template
func generateGenesisFile(genesisEntries []genesisEntry, newGenesisFilePath string) {
	// read and execute genesis template, outputting the genesis.go to stdout
	// genesisTmpl, err := helpers.ReadTemplateFile("./genesis.tmpl")
	tmpl, err := template.ParseFiles("./genesisblock/genesis.tmpl")
	if err != nil {
		log.Fatalf("Error while reading genesis.tmpl file: %s", err)
	}
	err = os.Remove(newGenesisFilePath)
	if err != nil {
		log.Printf("remove %s file: %s\n", newGenesisFilePath, err)
		return
	}
	f, err := os.Create(newGenesisFilePath)
	if err != nil {
		log.Printf("create %s file: %s\n", newGenesisFilePath, err)
		return
	}
	defer f.Close()

	config := map[string]interface{}{
		"MainchainGenesisBlockID": getGenesisBlockID(genesisEntries), // mocked value (needs one more pass)
		"MainchainGenesisConfig":  genesisEntries,
	}
	err = tmpl.Execute(f, config)
	if err != nil {
		log.Fatal(err)
	}
}

func getGenesisBlockID(genesisEntries []genesisEntry) int64 {
	var (
		genesisConfig []constant.MainchainGenesisConfigEntry
	)
	for _, entry := range genesisEntries {
		cfgEntry := constant.MainchainGenesisConfigEntry{
			AccountAddress:     entry.AccountAddress,
			AccountBalance:     entry.AccountBalance,
			LockedBalance:      entry.LockedBalance,
			NodeAddress:        entry.NodeAddress,
			NodePublicKey:      entry.NodePublicKey,
			ParticipationScore: entry.ParticipationScore,
		}
		genesisConfig = append(genesisConfig, cfgEntry)
	}
	bs := service.NewBlockService(
		&chaintype.MainChain{},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&transaction.TypeSwitcher{},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	block, err := bs.GenerateGenesisBlock(genesisConfig)
	if err != nil {
		log.Fatal(err)
	}
	return block.ID
}

func generateClusterConfigFile(genesisEntries []genesisEntry, newClusterConfigFilePath string) {
	var (
		clusterConfig []clusterConfigEntry
	)

	for _, genesisEntry := range genesisEntries {
		// exclude entries that don't have NodeSeed set from cluster_config.json
		// (they should be nodes already registered/run by someone, thus they shouldn't be deployed automatically)
		if genesisEntry.NodeSeed != "" {
			entry := clusterConfigEntry{
				NodeAddress:         genesisEntry.NodeAddress,
				NodePublicKey:       genesisEntry.NodePublicKeyB64,
				NodeSeed:            genesisEntry.NodeSeed,
				OwnerAccountAddress: genesisEntry.AccountAddress,
				Smithing:            genesisEntry.Smithing,
			}
			clusterConfig = append(clusterConfig, entry)
		}
	}
	file, err := json.MarshalIndent(clusterConfig, "", "  ")
	if err != nil {
		log.Fatalf("error marshaling json file %s: %s\n", newClusterConfigFilePath, err)
	}
	err = ioutil.WriteFile(newClusterConfigFilePath, file, 0644)
	if err != nil {
		log.Fatalf("create %s file: %s\n", newClusterConfigFilePath, err)
	}
}

func (ge *genesisEntry) FormatPubKeyByteString() string {
	if ge.NodePublicKeyB64 == "" {
		return ""
	}
	pubKey, err := base64.StdEncoding.DecodeString(ge.NodePublicKeyB64)
	if err != nil {
		log.Fatalf("Error decoding node public key: %s", err)
	}
	return util.RenderByteArrayAsString(pubKey)
}

func (ge *genesisEntry) HasParticipationScore() bool {
	return ge.ParticipationScore > 0
}

func (ge *genesisEntry) HasLockedBalance() bool {
	return ge.LockedBalance > 0
}

func (ge *genesisEntry) HasAccountBalance() bool {
	return ge.AccountBalance > 0
}

func (ge *genesisEntry) HasNodeAddress() bool {
	return ge.NodeAddress != ""
}

func (ge *genesisEntry) HasNodePublicKey() bool {
	return ge.NodePublicKeyB64 != ""
}
