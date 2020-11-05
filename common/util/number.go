package util

import "math"

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

// MinInt64 returns the smallest int64 number supplied
func MinInt64(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

// MaxInt64 returns the largest int64 number supplied
func MaxInt64(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

// GetNextStep returns the next step of an interval, given current one and the interval
func GetNextStep(curStep, interval int64) int64 {
	rate := math.Ceil(float64(curStep) / float64(interval))
	ret := int64(rate) * interval
	return ret
}
