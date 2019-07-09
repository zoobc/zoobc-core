package util

import "encoding/binary"

// ConvertBytesToUint64 takes array of bytes and return the uint64 representation of the given bytes
func ConvertBytesToUint64(bytes []byte) uint64 {
	return binary.LittleEndian.Uint64(bytes)
}

// ConvertBytesToUint32 takes array of bytes and return the uint32 representation of the given bytes
func ConvertBytesToUint32(bytes []byte) uint32 {
	return binary.LittleEndian.Uint32(bytes)
}

// ConvertBytesToUint16 takes array of bytes and return the uint16 representation of the given bytes
func ConvertBytesToUint16(bytes []byte) uint16 {
	return binary.LittleEndian.Uint16(bytes)
}

// ConvertUint64ToBytes takes the uint64 decimal number and return the byte array representation of the given number
func ConvertUint64ToBytes(number uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, number)
	return buf
}

// ConvertUint32ToBytes takes the uint32 decimal number and return the byte array representation of the given number
func ConvertUint32ToBytes(number uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, number)
	return buf
}

// ConvertUint16ToBytes takes the uint16 decimal number and return the byte array representation of the given number
func ConvertUint16ToBytes(number uint16) []byte {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, number)
	return buf
}

// ConvertIntToBytes takes the int decimal number and return the byte array representation of the given number
func ConvertIntToBytes(number int) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(number))
	return buf
}
