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
}

func (cfg *Config) SaveConfig(filePath string) error {
	var err error
	viper.Set("smithing", cfg.Smithing)
	viper.Set("ownerAccountAddress", cfg.OwnerAccountAddress)
	viper.Set("wellknownPeers", cfg.WellknownPeers)
	viper.Set("peerPort", cfg.PeerPort)
	viper.Set("apiRPCPort", cfg.RPCAPIPort)
	viper.Set("apiHTTPPort", cfg.HTTPAPIPort)
	viper.Set("maxAPIRequestPerSecond", cfg.MaxAPIRequestPerSecond)
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
