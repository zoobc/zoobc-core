package constant

import v1 "github.com/spf13/viper"

func SetCheckVarString(KeyToBeChecked string, defaultVal string) string {
	var Output string
	if v1.GetString(KeyToBeChecked) != "" {
		Output = v1.GetString(KeyToBeChecked)
	} else {
		Output = defaultVal
	}

	return Output
}

func SetCheckVarInt64(KeyToBeChecked string, defaultVal int64) int64 {
	var Output int64
	if v1.GetInt64(KeyToBeChecked) != 0 {
		Output = v1.GetInt64(KeyToBeChecked)
	} else {
		Output = defaultVal
	}

	return Output
}

func SetCheckVarInt32(KeyToBeChecked string, defaultVal int32) int32 {
	var Output int32
	if v1.GetInt32(KeyToBeChecked) != 0 {
		Output = v1.GetInt32(KeyToBeChecked)
	} else {
		Output = defaultVal
	}

	return Output
}

func SetCheckVarUint32(KeyToBeChecked string, defaultVal uint32) uint32 {
	var Output uint32
	if v1.GetUint32(KeyToBeChecked) != 0 {
		Output = v1.GetUint32(KeyToBeChecked)
	} else {
		Output = defaultVal
	}

	return Output
}

func SetCheckVarUint64(KeyToBeChecked string, defaultVal uint64) uint64 {
	var Output uint64
	if v1.GetUint32(KeyToBeChecked) != 0 {
		Output = v1.GetUint64(KeyToBeChecked)
	} else {
		Output = defaultVal
	}

	return Output
}

func SetCheckVarUint(KeyToBeChecked string, defaultVal uint) uint {
	var Output uint
	if v1.GetUint(KeyToBeChecked) != 0 {
		Output = v1.GetUint(KeyToBeChecked)
	} else {
		Output = defaultVal
	}

	return Output
}

func SetCheckVarInt(KeyToBeChecked string, defaultVal int) int {
	var Output int
	if v1.GetInt(KeyToBeChecked) != 0 {
		Output = v1.GetInt(KeyToBeChecked)
	} else {
		Output = defaultVal
	}

	return Output
}
