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
package decryptcert

import (
	"encoding/json"
	"fmt"
	"github.com/zoobc/zoobc-core/common/accounttype"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zoobc/zoobc-core/common/crypto"
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
				log.Print(err)
				continue
			}
			decryptedCertEntries = append(decryptedCertEntries, *decryptedEntry)
		}

		if len(decryptedCertEntries) > 0 {
			var outPath = path.Join(getRootPath(), "/resource/generated/decrypted")
			if err := os.MkdirAll(outPath, os.ModePerm); err != nil {
				log.Fatalf("can't create folder %s. error: %s", outPath, err)
			}
			fName := "hosted_preRegisteredNodes.json"
			if _, err := generateClusterConfigFile(decryptedCertEntries, path.Join(outPath, fName)); err != nil {
				log.Fatalf("error generating output file. error: %s", err)
			}
			log.Printf("Success! check the file : %s/%s", outPath, fName)
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
		return nil, fmt.Errorf("encrypted entry: %s ERROR: %s", encryptedEntry.EncryptedCert, err)
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
		var nodePubKeyStr string
		// exclude entries that don't have NodeSeed set from cluster_config.json
		// (they are possibly pre-registered nodes managed by someone, thus they shouldn't be deployed automatically)

		// verify that the NodePublicKey from cert = the one parsed by node using node seed
		accType := &accounttype.ZbcAccountType{}
		_, _, nodePubKeyStr, _, _, err = sig.GenerateAccountFromSeed(accType, genEntry.NodeSeed)
		if genEntry.NodePublicKey != nodePubKeyStr {
			log.Printf("invalid node pub key:\npk: %s\ncomputed: %s\nacc: %s",
				genEntry.NodePublicKey, nodePubKeyStr, genEntry.AccountAddress)
			continue
		}
		if err != nil {
			return nil, err
		}
		if genEntry.NodeSeed != "" {
			entry := clusterConfigEntry{
				NodePublicKey:  nodePubKeyStr,
				NodeSeed:       genEntry.NodeSeed,
				AccountAddress: genEntry.AccountAddress,
				Smithing:       true,
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
