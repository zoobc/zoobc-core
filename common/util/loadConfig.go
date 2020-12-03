package util

import (
	"fmt"
	"github.com/zoobc/zoobc-core/common/constant"
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
	viper.SetDefault("peerPort", 8002)
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
		return err
	}
	return viper.WriteConfig()
}
