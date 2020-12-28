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
package util

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

/*
LoadConfig must be called at first time while start the app
*/
func LoadConfig(path, name, extension, resourcePath string) error {
	if path == "" {
		p, err := GetRootPath()
		if err != nil {
			path = "./"
		} else {
			path = p
		}
	}

	if resourcePath == "" {
		resourcePath = filepath.Join(path, "./resource")
	}
	if len(path) < 1 || len(name) < 1 || len(extension) < 1 {
		return fmt.Errorf("path and extension cannot be nil")
	}

	viper.SetDefault("dbName", "zoobc.db")
	viper.SetDefault("nodeKeyFile", "node_keys.json")
	viper.Set("resourcePath", filepath.Join(resourcePath))
	viper.SetDefault("peerPort", 8001)
	viper.SetDefault("myAddress", "")
	viper.SetDefault("monitoringPort", 9090)
	viper.SetDefault("apiRPCPort", 7000)
	viper.SetDefault("apiHTTPPort", 7001)
	viper.SetDefault("logLevels", []string{"fatal", "error", "panic"})
	viper.Set("snapshotPath", filepath.Join(resourcePath, "./snapshots"))
	viper.SetDefault("logOnCli", false)
	viper.SetDefault("cliMonitoring", true)
	viper.SetDefault("maxAPIRequestPerSecond", 10)
	viper.SetDefault("antiSpamFilter", true)

	viper.SetEnvPrefix("zoobc") // will be uppercased automatically
	viper.AutomaticEnv()        // value will be read each time it is accessed

	viper.SetConfigName(name)
	viper.SetConfigType(extension)
	viper.AddConfigPath(path)
	viper.AddConfigPath("$HOME/zoobc")

	configFile, err := os.Open(filepath.Join(path, fmt.Sprintf("%s.%s", name, extension)))
	if err != nil {
		fmt.Printf("Config not found : %s\n", err.Error())
		return err
	}
	defer configFile.Close()

	err = viper.ReadConfig(configFile)
	if err != nil {
		return err
	}
	return viper.WriteConfig()
}
