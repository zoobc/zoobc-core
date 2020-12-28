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
package model

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type (
	Config struct {
		PeerPort, MaxAPIRequestPerSecond                          uint32
		RPCAPIPort, HTTPAPIPort, MonitoringPort, CPUProfilingPort int
		Smithing, IsNodeAddressDynamic, LogOnCli, CliMonitoring   bool
		WellknownPeers                                            []string
		NodeKey                                                   *NodeKey
		MyAddress, OwnerAccountAddressHex, NodeSeed               string
		OwnerAccountAddress                                       []byte
		OwnerEncodedAccountAddress                                string
		OwnerAccountAddressTypeInt                                int32
		APICertFile, APIKeyFile                                   string
		DatabaseFileName, ResourcePath,
		NodeKeyFileName, SnapshotPath string
		AntiSpamFilter                                      bool
		AntiSpamP2PRequestLimit, AntiSpamCPULimitPercentage int

		// validation fields
		ConfigFileExist bool
	}
	PortType string
)

const (
	PeerPort         PortType = "PeerPort"
	RPCAPIPort       PortType = "RPCAPIPort"
	HTTPAPIPort      PortType = "HTTPAPIPort"
	PortChangePeriod int      = 5
)

func NewConfig() *Config {
	return &Config{
		NodeKey: &NodeKey{},
	}
}

func (cfg *Config) LoadConfigurations() {
	cfg.MyAddress = viper.GetString("myAddress")
	cfg.PeerPort = viper.GetUint32("peerPort")
	cfg.MonitoringPort = viper.GetInt("monitoringPort")
	cfg.RPCAPIPort = viper.GetInt("apiRPCPort")
	cfg.HTTPAPIPort = viper.GetInt("apiHTTPPort")
	cfg.MaxAPIRequestPerSecond = viper.GetUint32("maxAPIRequestPerSecond")
	cfg.CPUProfilingPort = viper.GetInt("cpuProfilingPort")
	cfg.OwnerAccountAddressHex = viper.GetString("ownerAccountAddress")
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

func (cfg *Config) SaveConfig(filePath string) error {
	var err error
	viper.Set("smithing", cfg.Smithing)
	viper.Set("ownerAccountAddress", cfg.OwnerAccountAddressHex)
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
