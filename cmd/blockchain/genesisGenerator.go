package blockchain

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/database"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/util"
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
		NodePublicKey       string
		NodeSeed            string
		OwnerAccountAddress string
		NodeAddress         string
		Smithing            bool
	}
)

func GenerateGenesis(logger *logrus.Logger) *cobra.Command {
	var (
		withDbLastState bool
		dbPath          string
	)
	var txCmd = &cobra.Command{
		Use:   "genesis",
		Short: "genesis command used to generate a new genesis.go file",
		Long: `genesis command generate a genesis.go file from a list of accounts and/or from current database.
		the latter is to be used when we want to reset the blockchain mantaining the latest state of accounts and node registrations`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			withDbLastState, err := cmd.Flags().GetBool("withDbLastState")
			if err != nil {
				logger.Fatal(err)
			}
			dbPath, err := cmd.Flags().GetString("dbPath")
			if err != nil {
				logger.Printf("%s", err)
			}
			if args[0] == "generate" {
				generateFiles(logger, withDbLastState, dbPath)
			} else {
				logger.Error("unknown command")
			}
		},
	}
	txCmd.Flags().BoolVarP(&withDbLastState, "withDbLastState", "w", false,
		"add to genesis all registered nodes and account balances from current database")
	txCmd.Flags().StringVarP(&dbPath, "dbPath", "f", "../resource/",
		"path of blockchain's database to be used as data source in case the -w flag is used. If not set, the default resource folder is used")
	return txCmd
}

// generateFiles generate genesis files starting from a source json file.
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
func generateFiles(logger *logrus.Logger, withDbLastState bool, dbPath string) {
	var (
		bcState, preRegisteredNodes []genesisEntry
		err                         error
	)

	if withDbLastState {
		bcState, err = getDbLastState(dbPath)
		if err != nil {
			logger.Fatal(err)
		}
	}

	file, err := ioutil.ReadFile("./blockchain/preRegisteredNodes.json")
	if err != nil {
		logger.Fatalf("Error reading preRegisteredNodes.json file: %s", err)
	}
	err = json.Unmarshal(file, &preRegisteredNodes)
	if err != nil {
		logger.Fatalf("preRegisteredNodes.json parsing error: %s", err)
	}

	// merge duplicates: if preRegisteredNodes contains entries that are in db too, add the parameters that are't available in db,
	// which is are NodeSeed and Smithing
	for _, prNode := range preRegisteredNodes {
		found := false
		for i, e := range bcState {
			if prNode.AccountAddress == e.AccountAddress {
				bcState[i].NodeSeed = prNode.NodeSeed
				bcState[i].Smithing = prNode.Smithing
				found = true
				break
			}
		}
		if !found {
			bcState = append(bcState, prNode)
		}
	}

	// append to preRegistered nodes/accounts previous entries from a blockchain db file
	generateGenesisFile(logger, bcState, "./genesis.go.new")
	generateClusterConfigFile(logger, bcState, "./cluster_config.json.new")
	fmt.Println("Command executed successfully\ngenesis.go.new has been generated in cmd directory")

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
	accountBalances := accountBalanceQuery.BuildModel([]*model.AccountBalance{}, rows)
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
		nodeRegistrations := nodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, rows)
		if len(nodeRegistrations) > 0 {
			nr := nodeRegistrations[0]
			bcEntry.LockedBalance = nr.LockedBalance
			bcEntry.NodeAddress = nr.NodeAddress
			bcEntry.NodePublicKey = nr.NodePublicKey
			bcEntry.NodePublicKeyB64 = base64.StdEncoding.EncodeToString(nr.NodePublicKey)
			// get the participation score for this node registration
			qry, args := participationScoreQuery.GetParticipationScoreByNodeID(nr.NodeID)
			rows, err := queryExecutor.ExecuteSelect(qry, false, args...)
			if err != nil {
				return nil, err
			}
			defer rows.Close()
			participationScores := participationScoreQuery.BuildModel([]*model.ParticipationScore{}, rows)
			if len(participationScores) > 0 {
				bcEntry.ParticipationScore = participationScores[0].Score
			}
		}
		bcEntries = append(bcEntries, *bcEntry)
	}

	return bcEntries, err
}

// generateGenesisFile generates a genesis file with given entries, starting from a template
// Note: after applying this new genesis we still have to manually update the MainchainGenesisBlockID by
// running the node, wait for the error (invalid block id) to show and update the constant with the suggested new blockID
func generateGenesisFile(logger *logrus.Logger, genesisEntries []genesisEntry, newGenesisFilePath string) {
	// read and execute genesis template, outputting the genesis.go to stdout
	// genesisTmpl, err := helpers.ReadTemplateFile("./genesis.tmpl")
	tmpl, err := template.ParseFiles("./blockchain/genesis.tmpl")
	if err != nil {
		log.Fatalf("Error while reading genesis.tmpl file: %s", err)
	}
	os.Remove(newGenesisFilePath)
	f, err := os.Create(newGenesisFilePath)
	if err != nil {
		logger.Printf("create %s file: %s\n", newGenesisFilePath, err)
		return
	}
	defer f.Close()

	config := map[string]interface{}{
		"MainchainGenesisConfig": genesisEntries,
	}
	err = tmpl.Execute(f, config)
	if err != nil {
		log.Fatal(err)
	}
}

func generateClusterConfigFile(logger *logrus.Logger, genesisEntries []genesisEntry, newClusterConfigFilePath string) {
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
		logger.Fatalf("error marshaling json file %s: %s\n", newClusterConfigFilePath, err)
	}
	err = ioutil.WriteFile(newClusterConfigFilePath, file, 0644)
	if err != nil {
		logger.Fatalf("create %s file: %s\n", newClusterConfigFilePath, err)
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
