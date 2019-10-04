package blockchain

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
	}
)

func GenerateGenesis(logger *logrus.Logger) *cobra.Command {
	// var (
	// 	withDbLastState bool
	// )
	var txCmd = &cobra.Command{
		Use:   "genesis",
		Short: "genesis command used to generate a new genesis.go file",
		Long: `genesis command generate a genesis.go file from a list of accounts and/or from current database.
		the latter is to be used when we want to reset the blockchain mantaining the latest state of accounts and node registrations`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// withDbLastState, err := cmd.Flags().GetBool("withDbLastState")
			// if err != nil {
			// 	logger.Printf("%s", err)
			// }
			if args[0] == "generate" {
				generateGenesisFiles(logger, false)
			} else {
				logger.Error("unknown command")
			}
		},
	}
	// txCmd.Flags().BoolVarP(&withDbLastState, "withDbLastState", "withDb", false,
	// 	"add to genesis all registered nodes and account balances from current database")
	return txCmd
}

func generateGenesisFiles(logger *logrus.Logger, withDbLastState bool) {
	logger.Printf("withDb: %v", withDbLastState)
	// read and execute genesis template, outputting the genesis.go to stdout
	// genesisTmpl, err := helpers.ReadTemplateFile("./genesis.tmpl")
	tmpl, err := template.ParseFiles("./blockchain/genesis.tmpl")
	if err != nil {
		log.Fatalf("Error while reading genesis.tmpl file: %s", err)
	}
	newGenesisPath := "./genesis.go.new"
	os.Remove(newGenesisPath)
	f, err := os.Create(newGenesisPath)
	if err != nil {
		logger.Println("create genesis.go.new file: ", err)
		return
	}
	defer f.Close()

	file, err := ioutil.ReadFile("./blockchain/preRegisteredNodes.json")
	if err != nil {
		log.Fatalf("Error reading preRegisteredNodes.json file: %s", err)
	}
	data := []genesisEntry{}
	err = json.Unmarshal(file, &data)
	if err != nil {
		log.Fatalf("preRegisteredNodes.json parsing error: %s", err)
	}
	for _, entry := range data {
		logger.Printf("entry %v\n", entry)
	}

	config := map[string]interface{}{
		"MainchainGenesisConfig": data,
	}
	err = tmpl.Execute(f, config)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Command executed successfully\ngenesis.go.new has been generated in cmd directory")

}

func (ge *genesisEntry) FormatPubKeyByteString() string {
	if ge.NodePublicKeyB64 == "" {
		return ""
	}
	pubKey, err := base64.StdEncoding.DecodeString(ge.NodePublicKeyB64)
	if err != nil {
		log.Fatalf("Error decoding node public key: %s", err)
	}
	var buffer bytes.Buffer
	for i, b := range pubKey {
		buffer.WriteString(strconv.Itoa(int(b)))
		if i != len(pubKey)-1 {
			if i != 0 && i%18 == 0 {
				buffer.WriteString(", \n\t\t\t\t")
			} else {
				buffer.WriteString(", ")
			}
		}
	}
	return fmt.Sprintf("[]byte{%s}", buffer.String())
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
