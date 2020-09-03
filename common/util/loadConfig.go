package util

import (
	"fmt"

	"github.com/spf13/viper"
)

/*
LoadConfig must be called at first time while start the app
*/
func LoadConfig(path, name, extension string) error {
	if len(path) < 1 || len(name) < 1 || len(extension) < 1 {
		return fmt.Errorf("path and extension cannot be nil")
	}

	viper.SetDefault("dbName", "zoobc.db")
	viper.SetDefault("badgerDbName", "zoobc_kv/")
	viper.SetDefault("nodeKeyFile", "node_keys.json")
	viper.SetDefault("resourcePath", "./resource")
	viper.SetDefault("peerPort", 8001)
	viper.SetDefault("myAddress", "")
	viper.SetDefault("monitoringPort", 9090)
	viper.SetDefault("apiRPCPort", 7000)
	viper.SetDefault("apiHTTPPort", 7001)
	viper.SetDefault("logLevels", []string{"fatal", "error", "panic"})
	viper.SetDefault("snapshotPath", "./resource/snapshots")
	viper.SetDefault("logOnCli", false)
	viper.SetDefault("cliMonitoring", false)
	viper.SetDefault("maxAPIRequestPerSecond", 10)

	viper.SetEnvPrefix("zoobc") // will be uppercased automatically
	viper.AutomaticEnv()        // value will be read each time it is accessed

	viper.SetConfigName(name)
	viper.SetConfigType(extension)
	viper.AddConfigPath(path)
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok && path == "./" && name == "config" {
			// Config file not found; ignore error if desired
			return err
		}
		// Config file was found but another error was produced
		return err
	}
	return nil
}
