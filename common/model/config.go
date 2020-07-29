package model

import (
	"errors"
	"github.com/spf13/viper"
	"os"
)

type (
	Config struct {
		PeerPort, MaxAPIRequestPerSecond                          uint32
		RPCAPIPort, HTTPAPIPort, MonitoringPort, CPUProfilingPort int
		Smithing, IsNodeAddressDynamic, LogOnCli, CliMonitoring   bool
		MyAddress, OwnerAccountAddress                            string
		DatabasePath, DatabaseFileName, ResourcePath, BadgerDbName,
		NodeKeyFileName, NodeSeed, APICertFile, APIKeyFile, SnapshotPath string
		WellknownPeers []string

		NodeKey *NodeKey

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
	cfg.OwnerAccountAddress = viper.GetString("ownerAccountAddress")
	cfg.WellknownPeers = viper.GetStringSlice("wellknownPeers")
	cfg.Smithing = viper.GetBool("smithing")
	cfg.DatabaseFileName = viper.GetString("dbName")
	cfg.BadgerDbName = viper.GetString("badgerDbName")
	cfg.ResourcePath = viper.GetString("resourcePath")
	cfg.NodeKeyFileName = viper.GetString("nodeKeyFile")
	cfg.NodeSeed = viper.GetString("nodeSeed")
	cfg.APICertFile = viper.GetString("apiCertFile")
	cfg.APIKeyFile = viper.GetString("apiKeyFile")
	cfg.SnapshotPath = viper.GetString("snapshotPath")
	cfg.LogOnCli = viper.GetBool("logOnCli")
	cfg.CliMonitoring = viper.GetBool("cliMonitoring")
}

func (cfg *Config) SaveConfig() error {
	var err error
	viper.Set("smithing", cfg.Smithing)
	viper.Set("ownerAddress", cfg.OwnerAccountAddress)
	viper.Set("wellknownPeers", cfg.WellknownPeers)
	viper.Set("peerPort", cfg.PeerPort)
	viper.Set("apiRPCPort", cfg.RPCAPIPort)
	viper.Set("apiHTTPPort", cfg.HTTPAPIPort)
	// todo: code in rush, need refactor later andy-shi88
	_, err = os.Stat("./config.toml")
	if err != nil {
		if ok := os.IsNotExist(err); ok {
			err = viper.SafeWriteConfigAs("./config.toml")
			if err != nil {
				return errors.New("error saving configuration to ./config.toml\terror: " + err.Error())
			}
		} else {
			return err
		}
	}
	err = viper.WriteConfigAs("./config.toml")
	if err != nil {
		return errors.New("error saving configuration to ./config.toml\terror: " + err.Error())
	}
	return nil
}
