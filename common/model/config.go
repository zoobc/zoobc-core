package model

import (
	"github.com/spf13/viper"
)

type Config struct {
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
