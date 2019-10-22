package constant

import (
	"fmt"

	"github.com/spf13/viper"
)

var ConfigFile = "././resource/config.toml"
var isDebugFlag bool

func GetDebugFlag(isActive bool) {

	isDebugFlag = isActive
}

func setDebug() bool {

	debugMode := isDebugFlag
	return debugMode
}

func SetCheckVarString(key, defaultVal string) string {
	var Output string
	viper.SetConfigFile(ConfigFile)
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("error : %v\n", err)
	}
	if viper.GetString(key) != "" && setDebug() {
		Output = viper.GetString(key)
	} else {
		Output = defaultVal
	}

	return Output
}

func SetCheckVarInt64(key string, defaultVal int64) int64 {
	var Output int64
	viper.SetConfigFile(ConfigFile)
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("error : %v\n", err)
	}
	if viper.GetInt64(key) != 0 && setDebug() {
		Output = viper.GetInt64(key)
	} else {
		Output = defaultVal
	}

	return Output
}

func SetCheckVarInt32(key string, defaultVal int32) int32 {
	var Output int32
	viper.SetConfigFile(ConfigFile)
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("error : %v\n", err)
	}
	if viper.GetInt32(key) != 0 && setDebug() {
		Output = viper.GetInt32(key)
	} else {
		Output = defaultVal
	}

	return Output
}

func SetCheckVarUint32(key string, defaultVal uint32) uint32 {
	var Output uint32
	viper.SetConfigFile(ConfigFile)
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("error : %v\n", err)
	}
	if viper.GetUint32(key) != 0 && setDebug() {
		Output = viper.GetUint32(key)
	} else {
		Output = defaultVal
	}

	return Output
}

func SetCheckVarUint64(key string, defaultVal uint64) uint64 {
	var Output uint64
	viper.SetConfigFile(ConfigFile)
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("error : %v\n", err)
	}
	if viper.GetUint32(key) != 0 && setDebug() {
		Output = viper.GetUint64(key)
	} else {
		Output = defaultVal
	}

	return Output
}

func SetCheckVarUint(key string, defaultVal uint) uint {
	var Output uint
	viper.SetConfigFile(ConfigFile)
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("error : %v\n", err)
	}
	if viper.GetUint(key) != 0 && setDebug() {
		Output = viper.GetUint(key)
	} else {
		Output = defaultVal
	}

	return Output
}

func SetCheckVarInt(key string, defaultVal int) int {
	var Output int
	viper.SetConfigFile(ConfigFile)
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("error : %v\n", err)
	}
	if viper.GetInt(key) != 0 && setDebug() {
		Output = viper.GetInt(key)
	} else {
		Output = defaultVal
	}

	return Output
}
