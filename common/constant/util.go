package constant

import v1 "github.com/spf13/viper"

func SetCheckVarString(key, defaultVal string) string {
	var Output string
	if v1.GetString(key) != "" {
		Output = v1.GetString(key)
	} else {
		Output = defaultVal
	}

	return Output
}

func SetCheckVarInt64(key string, defaultVal int64) int64 {
	var Output int64
	if v1.GetInt64(key) != 0 {
		Output = v1.GetInt64(key)
	} else {
		Output = defaultVal
	}

	return Output
}

func SetCheckVarInt32(key string, defaultVal int32) int32 {
	var Output int32
	if v1.GetInt32(key) != 0 {
		Output = v1.GetInt32(key)
	} else {
		Output = defaultVal
	}

	return Output
}

func SetCheckVarUint32(key string, defaultVal uint32) uint32 {
	var Output uint32
	if v1.GetUint32(key) != 0 {
		Output = v1.GetUint32(key)
	} else {
		Output = defaultVal
	}

	return Output
}

func SetCheckVarUint64(key string, defaultVal uint64) uint64 {
	var Output uint64
	if v1.GetUint32(key) != 0 {
		Output = v1.GetUint64(key)
	} else {
		Output = defaultVal
	}

	return Output
}

func SetCheckVarUint(key string, defaultVal uint) uint {
	var Output uint
	if v1.GetUint(key) != 0 {
		Output = v1.GetUint(key)
	} else {
		Output = defaultVal
	}

	return Output
}

func SetCheckVarInt(key string, defaultVal int) int {
	var Output int
	if v1.GetInt(key) != 0 {
		Output = v1.GetInt(key)
	} else {
		Output = defaultVal
	}

	return Output
}
