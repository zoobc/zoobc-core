package util

import (
	"bytes"

	"github.com/zoobc/zoobc-core/common/model"
)

// GetProofOfOriginBytes serialize ProofOfOrigin struct into bytes
func GetProofOfOriginUnsignedBytes(poown *model.ProofOfOrigin) []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(poown.MessageBytes)
	buffer.Write(ConvertUint64ToBytes(uint64(poown.Timestamp)))
	return buffer.Bytes()
}
