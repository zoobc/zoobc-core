package util

import "github.com/zoobc/zoobc-core/common/constant"

// GetProofOfOwnershipMessageSize returns size in bytes of a proof of ownership message
func GetProofOfOwnershipMessageSize(withSignature bool) uint32 {
	message := constant.AccountType + constant.AccountAddress + constant.BlockHash + constant.Height
	if withSignature {
		return message + constant.NodeSignature
	}
	return message
}
