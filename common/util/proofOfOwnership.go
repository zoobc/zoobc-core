package util

import (
	"bytes"

	"github.com/zoobc/zoobc-core/common/blocker"
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
func GetProofOfOwnershipBytes(poown *model.ProofOfOwnership) []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(poown.MessageBytes)
	buffer.Write(poown.Signature)
	return buffer.Bytes()
}

// ParseProofOfOwnershipBytes parse a byte array into a ProofOfOwnership struct (message + signature)
// poownBytes if true returns size of message + signature
func ParseProofOfOwnershipBytes(poownBytes []byte) (*model.ProofOfOwnership, error) {
	buffer := bytes.NewBuffer(poownBytes)
	poownMessageBytes, err := ReadTransactionBytes(buffer, int(GetProofOfOwnershipSize(false)))
	if err != nil {
		return nil, err
	}
	signature, err := ReadTransactionBytes(buffer, int(constant.NodeSignature+constant.SignatureType))
	if err != nil {
		return nil, err
	}
	return &model.ProofOfOwnership{
		MessageBytes: poownMessageBytes,
		Signature:    signature,
	}, nil
}

// GetProofOfOwnershipMessageBytes serialize ProofOfOwnershipMessage struct into bytes
func GetProofOfOwnershipMessageBytes(poownMessage *model.ProofOfOwnershipMessage) []byte {
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write([]byte(poownMessage.AccountAddress))
	buffer.Write(poownMessage.BlockHash)
	buffer.Write(ConvertUint32ToBytes(poownMessage.BlockHeight))
	return buffer.Bytes()
}

// ParseProofOfOwnershipMessageBytes parse a byte array into a ProofOfOwnershipMessage struct (only the message, no signature)
func ParseProofOfOwnershipMessageBytes(poownMessageBytes []byte) (*model.ProofOfOwnershipMessage, error) {
	buffer := bytes.NewBuffer(poownMessageBytes)
	if buffer.Len() < int(constant.AccountAddress) {
		return nil, blocker.NewBlocker(blocker.ParserErr, "ProofOfOwnershipInvalidMessageFormat")
	}
	accountAddress := buffer.Next(int(constant.AccountAddress))
	if buffer.Len() < int(constant.BlockHash) {
		return nil, blocker.NewBlocker(blocker.ParserErr, "ProofOfOwnershipInvalidMessageFormat")
	}
	blockHash := buffer.Next(int(constant.BlockHash))
	if buffer.Len() < int(constant.Height) {
		return nil, blocker.NewBlocker(blocker.ParserErr, "ProofOfOwnershipInvalidMessageFormat")
	}
	height := ConvertBytesToUint32(buffer.Next(int(constant.Height)))
	return &model.ProofOfOwnershipMessage{
		AccountAddress: string(accountAddress),
		BlockHash:      blockHash,
		BlockHeight:    height,
	}, nil
}
