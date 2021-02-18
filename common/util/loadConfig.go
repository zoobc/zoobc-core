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
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zoobc/lib/address"
	"github.com/zoobc/zoobc-core/common/accounttype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"

	"github.com/spf13/viper"
)

/*
LoadConfig must be called at first time while start the app
*/
func LoadConfig(path, name, extension, resourcePath string) (config *model.Config, err error) {
	err = loadFromFile(path, name, extension, resourcePath)
	if err != nil {
		return config, err
	}

	config = model.NewConfig()
	readConfigurations(config)
	return config, err
}

func loadFromFile(path, name, extension, resourcePath string) error {
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

	// set default config values
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
	viper.SetDefault("antiSpamFilter", false)
	viper.SetDefault("antiSpamP2PRequestLimit", constant.P2PRequestHardLimit)
	viper.SetDefault("antiSpamCPULimitPercentage", constant.FeedbackLimitCPUPercentage)

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
		fmt.Printf("ReadConfig.Err %v\n\n", err)
		return err
	}
	err = viper.WriteConfig()
	if err != nil {
		fmt.Printf("WriteConfig.Err %v\n\n", err)
		return err
	}

	return nil
}

func readConfigurations(cfg *model.Config) {
	var pubKey = make([]byte, 32)
	_ = address.DecodeZbcID(viper.GetString("ownerAccountAddress"), pubKey)
	addressHexString := hex.EncodeToString(append([]byte{0, 0, 0, 0}, pubKey...))

	cfg.MyAddress = viper.GetString("myAddress")
	cfg.PeerPort = viper.GetUint32("peerPort")
	cfg.MonitoringPort = viper.GetInt("monitoringPort")
	cfg.RPCAPIPort = viper.GetInt("apiRPCPort")
	cfg.HTTPAPIPort = viper.GetInt("apiHTTPPort")
	cfg.MaxAPIRequestPerSecond = viper.GetUint32("maxAPIRequestPerSecond")
	cfg.CPUProfilingPort = viper.GetInt("cpuProfilingPort")
	cfg.OwnerAccountAddressHex = addressHexString
	cfg.WellknownPeers = viper.GetStringSlice("wellknownPeers")
	cfg.Smithing = viper.GetBool("smithing")
	cfg.DatabaseFileName = viper.GetString("dbName")
	cfg.ResourcePath = viper.GetString("resourcePath")
	cfg.NodeKeyFileName = viper.GetString("nodeKeyFile")
	cfg.NodeSeed = viper.GetString("nodeSeed")
	cfg.APICertFile = viper.GetString("apiCertFile")
	cfg.APIKeyFile = viper.GetString("apiKeyFile")
	cfg.SnapshotPath = viper.GetString("snapshotPath")
	cfg.LogOnCli = viper.GetBool("logOnCli")
	cfg.CliMonitoring = viper.GetBool("cliMonitoring")
	cfg.AntiSpamFilter = viper.GetBool("antiSpamFilter")
	cfg.AntiSpamP2PRequestLimit = viper.GetInt("antiSpamP2PRequestLimit")
	cfg.AntiSpamCPULimitPercentage = viper.GetInt("antiSpamCPULimitPercentage")
}

func SaveConfig(cfg *model.Config, filePath string) error {
	var err error
	ownerAccountAddress, err := hex.DecodeString(cfg.OwnerAccountAddressHex)
	if err != nil {
		return fmt.Errorf("Invalid OwnerAccountAddress in config. It must be in hex format: %s", err.Error())
	}
	// double check that the decoded account address is valid
	accType, err := accounttype.NewAccountTypeFromAccount(ownerAccountAddress)
	if err != nil {
		return fmt.Errorf("Invalid account type: %s", err.Error())
	}
	encodedAddress, err := accType.GetEncodedAddress()
	if err != nil {
		return fmt.Errorf("Error in generating encoded address: %s", err.Error())
	}

	viper.Set("smithing", cfg.Smithing)
	viper.Set("ownerAccountAddress", encodedAddress)
	viper.Set("wellknownPeers", cfg.WellknownPeers)
	viper.Set("peerPort", cfg.PeerPort)
	viper.Set("apiRPCPort", cfg.RPCAPIPort)
	viper.Set("apiHTTPPort", cfg.HTTPAPIPort)
	viper.Set("maxAPIRequestPerSecond", cfg.MaxAPIRequestPerSecond)
	viper.Set("antiSpamFilter", cfg.AntiSpamFilter)
	viper.Set("antiSpamP2PRequestLimit", cfg.AntiSpamP2PRequestLimit)
	viper.Set("antiSpamCPULimitPercentage", cfg.AntiSpamCPULimitPercentage)
	// todo: code in rush, need refactor later andy-shi88
	_, err = os.Stat(filepath.Join(filePath, "./config.toml"))
	if err != nil {
		if ok := os.IsNotExist(err); ok {
			err = viper.SafeWriteConfigAs(filepath.Join(filePath, "./config.toml"))
			if err != nil {
				return errors.New("error saving configuration to ./config.toml\terror: " + err.Error())
			}
		} else {
			return err
		}
	}
	err = viper.WriteConfigAs(filepath.Join(filePath, "./config.toml"))
	if err != nil {
		return errors.New("error saving configuration to ./config.toml\terror: " + err.Error())
	}
	return nil
}
