package util

// MinUint32 returns the smallest uint32 number supplied
func MinUint32(x, y uint32) uint32 {
	if x < y {
		return x
	}
	return y
}

// MaxUint32 returns the largest uint32 number supplied
func MaxUint32(x, y uint32) uint32 {
	if x > y {
		return x
	}
	return y
}
