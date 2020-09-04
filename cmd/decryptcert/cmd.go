package decryptcert

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
)

var (
	configureCmd = &cobra.Command{
		Use:   "decryptcert",
		Short: "command to decrypt a batch of wallet certificates",
		Long:  "command to decrypt a batch of wallet certificates and output a file with the decrypted data",
	}
)

func init() {
}
func Commands() *cobra.Command {
	configureCmd.Run = decryptCertCommand
	return configureCmd
}
func decryptCertCommand(*cobra.Command, []string) {
	var (
		err error
	)

	file, err := ioutil.ReadFile(path.Join(getRootPath(), "./resource/templates/certificates.json"))
	if err == nil {
		var (
			decryptedCertEntries []certEntry
			encryptedCertEntries []encryptedCertEntry
		)
		err = json.Unmarshal(file, &encryptedCertEntries)
		if err != nil {
			log.Fatalf("json parsing error: %s", err)
		}

		for _, encryptedEntry := range encryptedCertEntries {
			decryptedEntry, err := readCertEntry(encryptedEntry)
			if err != nil {
				log.Fatal(err)
			}
			decryptedCertEntries = append(decryptedCertEntries, *decryptedEntry)
		}

		if len(decryptedCertEntries) > 0 {
			var outPath = path.Join(getRootPath(), "/resource/generated/decrypted")
			if err := os.MkdirAll(outPath, os.ModePerm); err != nil {
				log.Fatalf("can't create folder %s. error: %s", outPath, err)
			}
			generateClusterConfigFile(decryptedCertEntries, path.Join(outPath, "cluster_config_seatSale.json"))
		}
	}

}

func getRootPath() string {
	wd, _ := os.Getwd()
	if strings.Contains(wd, "zoobc-core/") {
		return path.Join(wd, "../")
	}
	return wd
}

func readCertEntry(encryptedEntry encryptedCertEntry) (*certEntry, error) {
	var (
		certBytes []byte
		entry     certEntry
	)
	certBytes, err := crypto.OpenSSLDecrypt(encryptedEntry.Password, encryptedEntry.EncryptedCert)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(certBytes, &entry)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate, %s", err.Error())
	}
	return &entry, nil
}

func generateClusterConfigFile(entries []certEntry, newClusterConfigFilePath string) (clusterConfig []clusterConfigEntry, err error) {
	var sig = crypto.NewSignature()
	for _, genEntry := range entries {
		var nodePubKeyAddress string
		// exclude entries that don't have NodeSeed set from cluster_config.json
		// (they are possibly pre-registered nodes managed by someone, thus they shouldn't be deployed automatically)
		_, _, _, nodePubKeyAddress, err = sig.GenerateAccountFromSeed(model.SignatureType_DefaultSignature, genEntry.NodeSeed)
		if err != nil {
			return nil, err
		}
		if genEntry.NodeSeed != "" {
			entry := clusterConfigEntry{
				NodePublicKey:       nodePubKeyAddress,
				NodeSeed:            genEntry.NodeSeed,
				OwnerAccountAddress: genEntry.OwnerAccountAddress,
				Smithing:            true,
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
	return clusterConfig, nil
}
