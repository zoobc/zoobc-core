package util

import (
	"bytes"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
)

// GetProofOfOwnershipSize returns size in bytes of a proof of ownership message
func GetProofOfOwnershipSize(withSignature bool) uint32 {
	message := constant.AccountAddress + constant.BlockHash + constant.Height
	if withSignature {
		return message + constant.NodeSignature + constant.SignatureType
	}
	return message
}

// GetProofOfOwnershipBytes serialize ProofOfOwnership struct into bytes
// TODO: implement this
func GetProofOfOwnershipBytes(poown *model.ProofOfOwnership) []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(poown.MessageBytes)
	buffer.Write(poown.Signature)
	return buffer.Bytes()
}

// ParseProofOfOwnershipBytes parse a byte array into a ProofOfOwnership struct (message + signature)
// poownBytes if true returns size of message + signature
func ParseProofOfOwnershipBytes(poownBytes []byte) *model.ProofOfOwnership {
	buffer := bytes.NewBuffer(poownBytes)
	poownMessageBytes := buffer.Next(int(GetProofOfOwnershipSize(false)))
	signature := buffer.Next(int(constant.NodeSignature + constant.SignatureType))
	return &model.ProofOfOwnership{
		MessageBytes: poownMessageBytes,
		Signature:    signature,
	}
}

// GetProofOfOwnershipMessageBytes serialize ProofOfOwnershipMessage struct into bytes
// TODO: implement this
func GetProofOfOwnershipMessageBytes(poownMessage *model.ProofOfOwnershipMessage) []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write([]byte(poownMessage.AccountAddress))
	buffer.Write(poownMessage.BlockHash)
	buffer.Write(ConvertUint32ToBytes(poownMessage.BlockHeight))
	return buffer.Bytes()
}

// ParseProofOfOwnershipMessageBytes parse a byte array into a ProofOfOwnershipMessage struct (only the message, no signature)
func ParseProofOfOwnershipMessageBytes(poownMessageBytes []byte) *model.ProofOfOwnershipMessage {
	buffer := bytes.NewBuffer(poownMessageBytes)
	accountAddress := buffer.Next(int(constant.AccountAddress))
	blockHash := buffer.Next(int(constant.BlockHash))
	height := ConvertBytesToUint32(buffer.Next(int(constant.Height)))
	return &model.ProofOfOwnershipMessage{
		AccountAddress: string(accountAddress),
		BlockHash:      blockHash,
		BlockHeight:    height,
	}
}
