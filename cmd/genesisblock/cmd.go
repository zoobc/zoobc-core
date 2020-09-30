package genesisblock

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/zoobc/zoobc-core/common/monitoring"

	"github.com/spf13/cobra"
	"github.com/zoobc/lib/address"
	"github.com/zoobc/zoobc-core/common/auth"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/database"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/transaction"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/core/service"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
)

var (
	genesisGeneratorCmd = &cobra.Command{
		Use:   "generate",
		Short: "genesis generate command used to generate a new genesis.go and cluster_config.json file",
		Long:  `genesis generate command generate a genesis.go file from a list of accounts and/or from current database.`,
		Run: func(cmd *cobra.Command, args []string) {
			if _, ok := envTargetValue[envTarget]; !ok {
				log.Fatal("Invalid env-target flag given, only: develop,staging,alpha")
			}
			generateGenesisFiles(withDbLastState, dbPath, extraNodesCount)
		},
	}
)

func Commands() *cobra.Command {
	return genesisCmd
}

func init() {
	genesisGeneratorCmd.Flags().StringVarP(&dbPath, "dbPath", "f", "../resource/",
		"path of blockchain's database to be used as data source in case the -w flag is used. If not set, the default resource folder is used")
	genesisGeneratorCmd.Flags().IntVarP(&extraNodesCount, "extraNodes", "n", 0,
		"number of 'extra' autogenerated nodes to be deployed using cluster_config.json")
	genesisGeneratorCmd.Flags().StringVar(&logLevels, "logLevels", "fatal error panic",
		"default log levels for all nodes (for kvConsulScript.sh). example: 'warn info fatal error panic'")
	genesisGeneratorCmd.Flags().StringVar(&wellKnownPeers, "wellKnownPeers", "127.0.0.1:8001",
		"default wellKnownPeers for all nodes (for kvConsulScript.sh). example: 'n0.alpha.proofofparticipation.network n1.alpha."+
			"proofofparticipation.network n2.alpha.proofofparticipation.network'")
	genesisGeneratorCmd.Flags().StringVar(&deploymentName, "deploymentName", "zoobc-alpha",
		"nomad task name associated to this deployment")
	genesisGeneratorCmd.Flags().StringVar(&kvFileCustomConfigFile, "kvFileCustomConfigFile", "",
		"(optional) full path (path + fileName) of a custom cluster_config.json file to use to generate consulKvInitScript."+
			"sh instead of the automatically generated in resource/generated/genesis directory")

	genesisGeneratorCmd.Flags().StringVarP(&envTarget, "env-target", "e", "alpha", "env mode indeed a.k.a develop,staging,alpha")
	genesisGeneratorCmd.Flags().StringVarP(&output, "output", "o", "resource", "output generated files target")
	genesisGeneratorCmd.Flags().IntVarP(&genesisTimestamp, "timestamp", "t", 1596708000,
		"genesis timestamp, in unix epoch time, with resolution in seconds")
	genesisGeneratorCmd.Flags().StringVar(&applicationCodeName, "applicationCodeName", "ZBC_main",
		"application code name")
	genesisGeneratorCmd.Flags().StringVar(&applicationVersion, "applicationVersion", "1.0.0",
		"application code version")
	genesisCmd.AddCommand(genesisGeneratorCmd)
}

// generateGenesisFiles generate genesis files starting from a source json file.
// PreRegisteredNodes contains a list of known nodes-accountOwners' public keys to be included in genesis.
// Data that can be pre-set for node registration and and account balance are:
//   AccountAddress (mandatory): node's owner
//   AccountBalance (for funded accounts only): the balance of that account at genesis block
//   NodeSeed (this should be set only for testing nodes): it will be copied into cluster_config.json to
//       automatically deploy new nodes that are already registered
//   NodePublicKey (mandatory): node public key string format
//   NodeAddress (optional): if known, the node address that will be registered and put in cluster_config.json too
//   LockedBalance (optional): account's locked balance
//   ParticipationScore (optional): set custom initial participation score (mainly for testing the smith process and POP algorithm).
//       if not set, defaults to constant.DefaultParticipationScore
//
// withDbLastState if set to true, we also scan a given blockchain database and extract latest state to be included in genesis
//  (account balances and registered nodes/participation scores)
func generateGenesisFiles(withDbLastState bool, dbPath string, extraNodesCount int) {
	var (
		genesisEntries []genesisEntry
		accountNodes   []accountNodeEntry
		err            error
		bcStateMap     = make(map[string]genesisEntry)
	)

	// import seatSale pre-registered-nodes (nodes generated from a wallet certificate that are going to be hosted by the company)
	file, err := ioutil.ReadFile(path.Join(getRootPath(), fmt.Sprintf("./resource/templates/%s.seatSale.json", envTarget)))
	if err == nil {
		var seatSaleNodes []genesisEntry
		err = json.Unmarshal(file, &seatSaleNodes)
		if err != nil {
			log.Fatalf("seatSale.json parsing error: %s", err)
		}
		log.Println("SeatSale Nodes (Ethereum Contract): ", len(seatSaleNodes))
		bcStateMap = buildPreregisteredNodes(seatSaleNodes, withDbLastState, dbPath)
	}

	// import company-managed nodes and merge with previous entries (overriding duplicates,
	// that might have been added to pre-registered-nodes list)
	filePath := path.Join(getRootPath(), fmt.Sprintf("./resource/templates/%s.preRegisteredNodes.json", envTarget))
	file, err = ioutil.ReadFile(filePath)
	if err == nil {
		var (
			preRegisteredNodes []genesisEntry
		)

		err = json.Unmarshal(file, &preRegisteredNodes)
		if err != nil {
			log.Fatalf("preRegisteredNodes.json parsing error: %s", err)
		}
		log.Println("PreRegistered Nodes (hosted): ", len(preRegisteredNodes))

		for key, preRegisteredNode := range buildPreregisteredNodes(preRegisteredNodes, withDbLastState, dbPath) {
			// make sure genesis gets the public key from the ethereum contract (seatSale.json), if present
			// later on we double check that this public key is valid by verifying we obtain the same value parsing the node seed
			if prevBcStateEntry, ok := bcStateMap[key]; ok {
				preRegisteredNode.NodePublicKey = prevBcStateEntry.NodePublicKey
				preRegisteredNode.NodePublicKeyBytes = prevBcStateEntry.NodePublicKeyBytes
			} else {
				entry := parseErrorLog{
					AccountAddress: preRegisteredNode.AccountAddress,
				}
				log.Printf("Warning, this Address in not in the Ethereum contract but only in %s: %s", filePath, entry)
			}
			bcStateMap[key] = preRegisteredNode
		}
	}
	for _, preRegisteredNode := range bcStateMap {
		genesisEntries = append(genesisEntries, preRegisteredNode)
	}

	var idx int
	for idx = 0; idx < extraNodesCount; idx++ {
		genesisEntries = append(genesisEntries, generateRandomGenesisEntry(""))
	}

	// generate extra nodes from a json file containing only account addresses
	file, err = ioutil.ReadFile(path.Join(getRootPath(), fmt.Sprintf("./resource/templates/%s.genesisAccountAddresses.json", envTarget)))
	if err == nil {
		var preRegisteredAccountAddresses []genesisEntry
		// read custom addresses from file
		err = json.Unmarshal(file, &preRegisteredAccountAddresses)
		if err != nil {
			log.Fatalf("preRegisteredAccountAddresses.json parsing error: %s", err)
		}
		if idx == 0 {
			idx--
		}
		for _, preRegisteredAccountAddress := range preRegisteredAccountAddresses {
			idx++
			genesisEntries = append(genesisEntries, generateRandomGenesisEntry(preRegisteredAccountAddress.AccountAddress))
		}
	}

	// append to preRegistered nodes/accounts previous entries from a blockchain db file
	var outPath = path.Join(getRootPath(), fmt.Sprintf("%s/generated/genesis", output))

	if err := os.MkdirAll(outPath, os.ModePerm); err != nil {
		log.Fatalf("can't create folder %s. error: %s", outPath, err)
	}
	if !validateGenesisFile(genesisEntries) {
		log.Fatal("Genesis files not generated because of invalid input files")
	}
	generateGenesisFile(genesisEntries, fmt.Sprintf("%s/genesis.go", outPath), fmt.Sprintf("%s/genesisSpine.go", outPath))
	clusterConfig := generateClusterConfigFile(genesisEntries, fmt.Sprintf("%s/cluster_config.json", outPath))
	// generate a bash script to init consul key/value data store in case we automatically deploy all nodes in genesis
	generateConsulKvInitScript(clusterConfig, fmt.Sprintf("%s/consulKvInit.sh", outPath))

	// also generate a file to be shared with node owners, so they know from the wallet what node to configure as their own node

	for _, entry := range genesisEntries {
		newEntry := accountNodeEntry{
			AccountAddress: entry.AccountAddress,
			NodePublicKey:  entry.NodePublicKey,
		}
		accountNodes = append(accountNodes, newEntry)
	}
	generateAccountNodesFile(accountNodes, fmt.Sprintf("%s/accountNodes.json", outPath))
	fmt.Printf("Command executed successfully\ngenesis.go.new has been generated in %s\n", outPath)
	fmt.Printf("to apply new genesis to the core-node, please overwrite common/constant/genesis."+
		"go with the new one: %s/genesis.go", outPath)
}

// buildPreregisteredNodes merge duplicates: if preRegisteredNodes contains entries that are in db too,
// add the parameters that are't available in db,
// which is are NodeSeed and Smithing
func buildPreregisteredNodes(preRegisteredNodes []genesisEntry, withDbLastState bool, dbPath string) map[string]genesisEntry {
	var (
		bcState          []genesisEntry
		err              error
		preRegisteredMap = make(map[string]genesisEntry)
	)
	if withDbLastState {
		bcState, err = getDbLastState(dbPath)
		if err != nil {
			log.Fatal(err)
		}
	}

	// merge duplicates: if preRegisteredNodes contains entries that are in db too, add the parameters that are't available in db,
	// which is are NodeSeed and Smithing
	for _, prNode := range preRegisteredNodes {
		found := false
		var pubKey = make([]byte, 32)
		for i, e := range bcState {
			if prNode.AccountAddress != e.AccountAddress {
				continue
			}
			if prNode.NodeSeed != "" {
				bcState[i].NodeSeed = prNode.NodeSeed
			}
			bcState[i].Smithing = prNode.Smithing
			err := address.DecodeZbcID(prNode.NodePublicKey, pubKey)
			if err != nil {
				log.Fatal(err)
			}
			bcState[i].NodePublicKeyBytes = pubKey
			preRegisteredMap[prNode.AccountAddress] = bcState[i]
			found = true
			break
		}
		if !found {
			err := address.DecodeZbcID(prNode.NodePublicKey, pubKey)
			if err != nil {
				log.Fatal(err)
			}
			prNode.NodePublicKeyBytes = pubKey
			bcState = append(bcState, prNode)
			preRegisteredMap[prNode.AccountAddress] = prNode
		}
	}
	return preRegisteredMap
}

// generateRandomGenesisEntry randomly generates a genesis node entry
// note: the account address is mandatory for the node registration, but as there is no wallet connected to it
//       and we are not storing the relative seed, needed to sign transactions, these nodes can smith but their owners
//       can't perform any transaction.
//       This is only useful to test multiple smithing-nodes, for instence in a network stress test of tens of nodes connected together
func generateRandomGenesisEntry(accountAddress string) genesisEntry {
	var (
		ed25519Signature = crypto.NewEd25519Signature()
	)
	if accountAddress == "" {
		var (
			seed       = util.GetSecureRandomSeed()
			privateKey = ed25519Signature.GetPrivateKeyFromSeed(seed)
			publicKey  = privateKey[32:]
		)
		accountAddress, _ = ed25519Signature.GetAddressFromPublicKey(constant.PrefixZoobcDefaultAccount, publicKey)
	}
	var (
		nodeSeed       = util.GetSecureRandomSeed()
		nodePrivateKey = ed25519Signature.GetPrivateKeyFromSeed(nodeSeed)
		nodePublicKey  = nodePrivateKey[32:]
	)
	nodePublicKeyStr, _ := address.EncodeZbcID(constant.PrefixZoobcNodeAccount, nodePublicKey)

	return genesisEntry{
		AccountAddress:     accountAddress,
		NodePublicKeyBytes: nodePublicKey,
		NodePublicKey:      nodePublicKeyStr,
		NodeSeed:           nodeSeed,
		ParticipationScore: constant.GenesisParticipationScore,
		Smithing:           true,
		LockedBalance:      0,
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
	db, err = dbInstance.OpenDB(dbPath, "zoobc.db", 10, 10, 20*time.Minute)
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
	balanceRows, err := queryExecutor.ExecuteSelect(qry, false)
	if err != nil {
		return nil, err
	}
	defer balanceRows.Close()
	accountBalances, err := accountBalanceQuery.BuildModel([]*model.AccountBalance{}, balanceRows)
	if err != nil {
		return nil, err
	}
	for _, acc := range accountBalances {
		if acc.AccountAddress == constant.MainchainGenesisAccountAddress {
			continue
		}

		var nodeRegistrations []*model.NodeRegistration

		bcEntry := new(genesisEntry)
		bcEntry.AccountAddress = acc.AccountAddress
		bcEntry.AccountBalance = acc.Balance

		err := func() error {
			// get node registration for this account, if exists
			qry, args := nodeRegistrationQuery.GetNodeRegistrationByAccountAddress(acc.AccountAddress)
			nrRows, err := queryExecutor.ExecuteSelect(qry, false, args...)
			if err != nil {
				return err
			}
			defer nrRows.Close()

			nodeRegistrations, err = nodeRegistrationQuery.BuildModel([]*model.NodeRegistration{}, nrRows)
			if err != nil {
				return err
			}
			return nil
		}()

		if err != nil {
			return nil, err
		}

		if len(nodeRegistrations) > 0 {
			nr := nodeRegistrations[0]
			bcEntry.LockedBalance = nr.LockedBalance
			bcEntry.NodePublicKeyBytes = nr.NodePublicKey

			bcEntry.NodePublicKey, _ = address.EncodeZbcID(constant.PrefixZoobcNodeAccount, nr.NodePublicKey)

			err := func() error {
				// get the participation score for this node registration
				qry, args := participationScoreQuery.GetParticipationScoreByNodeID(nr.NodeID)
				psRows, err := queryExecutor.ExecuteSelect(qry, false, args...)
				if err != nil {
					return err
				}
				defer psRows.Close()

				participationScores, err := participationScoreQuery.BuildModel([]*model.ParticipationScore{}, psRows)
				if (err != nil) || len(participationScores) > 0 {
					bcEntry.ParticipationScore = participationScores[0].Score
				}
				return nil
			}()
			if err != nil {
				return nil, err
			}
		}
		bcEntries = append(bcEntries, *bcEntry)
	}

	return bcEntries, err
}

func validateGenesisFile(genesisEntries []genesisEntry) bool {
	var (
		numberOfUnmatched = 0
		errorLog          = []*parseErrorLog{}
	)
	ed25519 := crypto.NewEd25519Signature()
	for _, genesisEntry := range genesisEntries {
		if genesisEntry.NodeSeed == "" {
			continue
		}
		// compare the public key we've gotten from the input configuration file with the one generated by the node using node seed
		pbKey, _ := ed25519.GetPublicKeyFromAddress(genesisEntry.NodePublicKey)
		computedPbKey := ed25519.GetPublicKeyFromSeed(genesisEntry.NodeSeed)
		computedPbKeyStr, _ := address.EncodeZbcID(constant.PrefixZoobcNodeAccount, computedPbKey)
		genesisEntry.NodePublicKeyString()
		if !bytes.Equal(pbKey, computedPbKey) {
			errorEntry := &parseErrorLog{
				AccountAddress:    genesisEntry.AccountAddress,
				ComputedPublicKey: computedPbKeyStr,
				ConfigPublicKey:   genesisEntry.NodePublicKey,
			}
			errorLog = append(errorLog, errorEntry)
			numberOfUnmatched++
		}
	}

	if len(errorLog) > 0 {
		filePath := path.Join(getRootPath(), "./resource/generated/genesis/error.log")
		file, err := json.MarshalIndent(errorLog, "", "  ")
		if err != nil {
			log.Fatalf("error marshaling error log %s: %s\n", filePath, err)
		}
		err = ioutil.WriteFile(filePath, file, 0644)
		if err != nil {
			log.Fatalf("create %s file: %s\n", filePath, err)
		}
	}

	log.Println("Ivalid node public keys: ", numberOfUnmatched)
	log.Println("Valid node public keys: ", len(genesisEntries)-numberOfUnmatched)
	log.Println("Total genesis entries: ", len(genesisEntries))
	return numberOfUnmatched == 0
}

// generateGenesisFile generates a genesis file with given entries, starting from a template
func generateGenesisFile(genesisEntries []genesisEntry, newMainGenesisFilePath, newSpineGenesisFilePath string) {
	var (
		mainGenesisTmpl, spineGenesisTmpl *template.Template
		err                               error
		mainBlockID, spineBlockID         = getGenesisBlockID(genesisEntries)
	)
	/**
	Main Genesis
	*/
	// read and execute genesis template, outputting the genesis.go to stdout
	mainGenesisTmpl, err = template.ParseFiles(path.Join(getRootPath(), "./resource/templates/genesis.tmpl"))
	if err != nil {
		log.Fatalf("Error while reading genesis.tmpl file: %s", err)
	}
	err = os.Remove(newMainGenesisFilePath)
	if err != nil {
		log.Printf("remove %s file: %s\n", newMainGenesisFilePath, err)
	}
	mainFile, err := os.Create(newMainGenesisFilePath)
	if err != nil {
		log.Printf("create %s file: %s\n", newMainGenesisFilePath, err)
		return
	}
	config := map[string]interface{}{
		"MainchainGenesisBlockID": mainBlockID,
		"GenesisTimestamp":        genesisTimestamp,
		"ApplicationCodeName":     applicationCodeName,
		"ApplicationVersion":      applicationVersion,
		"MainchainGenesisConfig":  genesisEntries,
	}
	err = mainGenesisTmpl.Execute(mainFile, config)
	if err != nil {
		log.Fatal(err)
	}
	func() {
		defer mainFile.Close()
	}()

	/**
	Spine Genesis
	*/
	// read and execute genesis template, outputting the genesis.go to stdout
	spineGenesisTmpl, err = template.ParseFiles(path.Join(getRootPath(), "./resource/templates/genesisSpine.tmpl"))
	if err != nil {
		log.Fatalf("Error while reading genesis.tmpl file: %s", err)
	}
	err = os.Remove(newSpineGenesisFilePath)
	if err != nil {
		log.Printf("remove %s file: %s\n", newSpineGenesisFilePath, err)
	}
	spineFile, err := os.Create(newSpineGenesisFilePath)
	if err != nil {
		log.Printf("create %s file: %s\n", newSpineGenesisFilePath, err)
		return
	}
	defer spineFile.Close()

	err = spineGenesisTmpl.Execute(spineFile, map[string]interface{}{
		"SpinechainGenesisBlockID": spineBlockID,
		"GenesisTimestamp":         genesisTimestamp,
	})
	if err != nil {
		log.Fatal(err)
	}

}

func getGenesisBlockID(genesisEntries []genesisEntry) (mainBlockID, spineBlockID int64) {
	var (
		signature                 = crypto.NewSignature()
		nodeAuthValidationService = auth.NewNodeAuthValidation(signature)
		mempoolStorage            = storage.NewMempoolStorage()
		genesisConfig             []constant.GenesisConfigEntry
	)
	activeNodeRegistryCacheStorage := storage.NewNodeRegistryCacheStorage(
		monitoring.TypeActiveNodeRegistryStorage,
		func(registries []storage.NodeRegistry) {
			sort.SliceStable(registries, func(i, j int) bool {
				// sort by nodeID lowest - highest
				return registries[i].Node.GetNodeID() < registries[j].Node.GetNodeID()
			})
		})
	// store pending node registry
	pendingNodeRegistryCacheStorage := storage.NewNodeRegistryCacheStorage(
		monitoring.TypePendingNodeRegistryStorage,
		func(registries []storage.NodeRegistry) {
			sort.SliceStable(registries, func(i, j int) bool {
				// sort by locked balance highest - lowest
				return registries[i].Node.GetLockedBalance() > registries[j].Node.GetLockedBalance()
			})
		},
	)
	for _, entry := range genesisEntries {
		cfgEntry := constant.GenesisConfigEntry{
			AccountAddress:     entry.AccountAddress,
			AccountBalance:     entry.AccountBalance,
			LockedBalance:      entry.LockedBalance,
			NodePublicKey:      entry.NodePublicKeyBytes,
			ParticipationScore: entry.ParticipationScore,
		}
		genesisConfig = append(genesisConfig, cfgEntry)
	}

	bs := service.NewBlockMainService(
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
		&transaction.TypeSwitcher{
			MempoolCacheStorage:        mempoolStorage,
			NodeAuthValidation:         nodeAuthValidationService,
			ActiveNodeRegistryStorage:  activeNodeRegistryCacheStorage,
			PendingNodeRegistryStorage: pendingNodeRegistryCacheStorage,
		},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil, nil,
		nil,
		&transaction.Util{},
		&coreUtil.ReceiptUtil{},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		query.GetPruneQuery(&chaintype.MainChain{}),
		nil,
		nil,
		nil,
	)
	block, err := bs.GenerateGenesisBlock(genesisConfig)
	if err != nil {
		log.Fatal(err)
	}
	sb := service.NewBlockSpineService(
		&chaintype.SpineChain{},
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
		nil,
		nil,
		bs,
	)
	spine, err := sb.GenerateGenesisBlock(genesisConfig)
	if err != nil {
		log.Fatal(err)
	}

	return block.ID, spine.ID
}

func generateClusterConfigFile(genesisEntries []genesisEntry, newClusterConfigFilePath string) (clusterConfig []clusterConfigEntry) {
	for _, genEntry := range genesisEntries {
		// exclude entries that don't have NodeSeed set from cluster_config.json
		// (they are possibly pre-registered nodes managed by someone, thus they shouldn't be deployed automatically)
		if genEntry.NodeSeed != "" {
			entry := clusterConfigEntry{
				NodePublicKey:  genEntry.NodePublicKey,
				NodeSeed:       genEntry.NodeSeed,
				AccountAddress: genEntry.AccountAddress,
				Smithing:       genEntry.Smithing,
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
	return clusterConfig
}

func generateAccountNodesFile(accountNodeEntries []accountNodeEntry, configFilePath string) {
	var (
		accountNodes []accountNodeEntry
	)

	for _, e := range accountNodeEntries {
		entry := accountNodeEntry{
			NodePublicKey:  e.NodePublicKey,
			AccountAddress: e.AccountAddress,
		}
		accountNodes = append(accountNodes, entry)
	}
	file, err := json.MarshalIndent(accountNodes, "", "  ")
	if err != nil {
		log.Fatalf("error marshaling json file %s: %s\n", configFilePath, err)
	}
	err = ioutil.WriteFile(configFilePath, file, 0644)
	if err != nil {
		log.Fatalf("create %s file: %s\n", configFilePath, err)
	}
}

// generateGenesisFile generates a genesis file with given entries, starting from a template
func generateConsulKvInitScript(clusterConfigEntries []clusterConfigEntry, consulKvInitScriptPath string) {
	// read and execute genesis template, outputting the genesis.go to stdout
	// genesisTmpl, err := helpers.ReadTemplateFile("./genesis.tmpl")
	tmpl, err := template.ParseFiles(path.Join(getRootPath(), "./resource/templates/consulKvInit.tmpl"))
	if err != nil {
		log.Fatalf("Error while reading consulKvInit.tmpl file: %s", err)
	}
	err = os.Remove(consulKvInitScriptPath)
	if err != nil {
		log.Printf("remove %s file: %s\n", consulKvInitScriptPath, err)
	}
	f, err := os.Create(consulKvInitScriptPath)
	if err != nil {
		log.Fatalf("create %s file: %s\n", consulKvInitScriptPath, err)
	}
	defer f.Close()

	if kvFileCustomConfigFile != "" {
		jsonFile, err := os.Open(kvFileCustomConfigFile)
		if err != nil {
			log.Fatalf("opening file %s: error %s\n", kvFileCustomConfigFile, err)
		}
		defer jsonFile.Close()
		byteValue, err := ioutil.ReadAll(jsonFile)
		if err != nil {
			log.Fatalf("reading file %s: error %s\n", kvFileCustomConfigFile, err)
		}
		err = json.Unmarshal(byteValue, &clusterConfigEntries)
		if err != nil {
			log.Fatalf("parsing file %s: error %s\n", kvFileCustomConfigFile, err)
		}
	}

	config := map[string]interface{}{
		"nomadJobName":         deploymentName,
		"wellKnownPeers":       wellKnownPeers,
		"logLevels":            logLevels,
		"clusterConfigEntries": clusterConfigEntries,
	}
	err = tmpl.Execute(f, config)
	if err != nil {
		log.Fatal(err)
	}
}

func getRootPath() string {
	wd, _ := os.Getwd()
	if strings.Contains(wd, "zoobc-core/") {
		return path.Join(wd, "../")
	}
	return wd
}

func (ge *genesisEntry) NodePublicKeyString() string {
	var pubKey []byte
	_ = address.DecodeZbcID(ge.NodePublicKey, pubKey)
	return util.RenderByteArrayAsString(ge.NodePublicKeyBytes)
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

func (ge *genesisEntry) HasNodePublicKey() bool {
	return ge.NodePublicKey != ""
}
