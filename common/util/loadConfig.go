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
func LoadConfig(path, name, extension string) error {
	if len(path) < 1 || len(name) < 1 || len(extension) < 1 {
		return fmt.Errorf("path and extension cannot be nil")
	}

	viper.SetDefault("dbName", "zoobc.db")
	viper.SetDefault("badgerDbName", "zoobc_kv/")
	viper.SetDefault("nodeKeyFile", "node_keys.json")
	viper.Set("resourcePath", filepath.Join(path, "../resource"))
	viper.SetDefault("peerPort", 8002)
	viper.SetDefault("myAddress", "")
	viper.SetDefault("monitoringPort", 9090)
	viper.SetDefault("apiRPCPort", 7000)
	viper.SetDefault("apiHTTPPort", 7003)
	viper.SetDefault("logLevels", []string{"fatal", "error", "panic"})
	viper.Set("snapshotPath", filepath.Join(path, "../resource/snapshots"))
	viper.SetDefault("logOnCli", false)
	viper.SetDefault("cliMonitoring", false)
	viper.SetDefault("maxAPIRequestPerSecond", 10)

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
