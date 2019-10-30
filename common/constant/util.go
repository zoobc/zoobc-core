package constant

import (
	"github.com/spf13/viper"
)

func isDebug() bool {
	return viper.GetBool("dflag")
}
func SetCheckVarString(key, defaultVal string) string {
	if viper.GetString(key) != "" && isDebug() {
		return viper.GetString(key)
	}
	return defaultVal
}

func SetCheckVarInt64(key string, defaultVal int64) int64 {
	if viper.GetInt64(key) != 0 && isDebug() {
		return viper.GetInt64(key)
	}
	return defaultVal
}

func SetCheckVarInt32(key string, defaultVal int32) int32 {
	if viper.GetInt32(key) != 0 && isDebug() {
		return viper.GetInt32(key)
	}
	return defaultVal
}

func SetCheckVarUint32(key string, defaultVal uint32) uint32 {
	if viper.GetUint32(key) != 0 && isDebug() {
		return viper.GetUint32(key)
	}
	return defaultVal
}

func SetCheckVarUint64(key string, defaultVal uint64) uint64 {
	if viper.GetUint32(key) != 0 && isDebug() {
		return viper.GetUint64(key)
	}
	return defaultVal
}

func SetCheckVarUint(key string, defaultVal uint) uint {
	if viper.GetUint(key) != 0 && isDebug() {
		return viper.GetUint(key)
	}
	return defaultVal
}

func SetCheckVarInt(key string, defaultVal int) int {
	if viper.GetInt(key) != 0 && isDebug() {
		return viper.GetInt(key)
	}
	return defaultVal
}
