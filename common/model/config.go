package model

import (
	"github.com/spf13/viper"
)

type Config struct {
	PeerPort, MaxAPIRequestPerSecond                uint32
	ClientAPIPort, MonitoringPort, CPUProfilingPort int
	Smithing, IsNodeAddressDynamic                  bool
	MyAddress, OwnerAccountAddress                  string
	DatabasePath, DatabaseFileName, ResourcePath, BadgerDbName,
	NodeKeyPath, NodeKeyFileName, APICertFile, APIKeyFile, SnapshotPath string
	WellknownPeers []string

	NodeKey *NodeKey

	// validation fields
	ConfigFileExist bool
}

func NewConfig() *Config {
	return &Config{
		NodeKey: &NodeKey{},
	}
}

func (cfg *Config) LoadConfigurations() {
	cfg.MyAddress = viper.GetString("myAddress")
	cfg.PeerPort = viper.GetUint32("peerPort")
	cfg.MonitoringPort = viper.GetInt("monitoringPort")
	cfg.ClientAPIPort = viper.GetInt("apiRPCPort")
	cfg.MaxAPIRequestPerSecond = viper.GetUint32("maxAPIRequestPerSecond")
	cfg.CPUProfilingPort = viper.GetInt("cpuProfilingPort")
	cfg.OwnerAccountAddress = viper.GetString("ownerAccountAddress")
	cfg.WellknownPeers = viper.GetStringSlice("wellknownPeers")
	cfg.Smithing = viper.GetBool("smithing")
	cfg.DatabaseFileName = viper.GetString("dbName")
	cfg.BadgerDbName = viper.GetString("badgerDbName")
	cfg.ResourcePath = viper.GetString("resourcePath")
	cfg.NodeKeyFileName = viper.GetString("nodeKeyFile")
	cfg.APICertFile = viper.GetString("apiapiCertFile")
	cfg.APIKeyFile = viper.GetString("apiKeyFile")
	cfg.SnapshotPath = viper.GetString("snapshotPath")
}
