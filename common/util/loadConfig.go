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

	viper.SetDefault("dbName", "spinechain.db")
	viper.SetDefault("dbPath", "./resource")
	viper.SetDefault("apiRPCPort", 8080)
	viper.SetDefault("apiHTTPPort", 0)
	viper.SetDefault("logLevels", []string{"error", "fatal", "panic"})
	viper.SetConfigName(name)
	viper.SetConfigType(extension)
	viper.AddConfigPath(path)
	viper.AddConfigPath("../../resource")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return nil

}
