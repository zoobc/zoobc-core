package util

import (
	"bytes"
	"github.com/zoobc/zoobc-core/common/accounttype"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
)

// GetProofOfOwnershipSize returns size in bytes of a proof of ownership message
func GetProofOfOwnershipSize(accountAddressType accounttype.AccountTypeInterface, withSignature bool) uint32 {
	var (
		accountAddressSize = constant.AccountAddressTypeLength + accountAddressType.GetAccountPublicKeyLength()
	)
	message := accountAddressSize + constant.BlockHash + constant.Height
	if withSignature {
		return message + constant.NodeSignature
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
	// copy poown bytes and parse first bytes as accountAddress to get the address size
	var tmpPoonBytes = make([]byte, len(poownBytes))
	copy(tmpPoonBytes, poownBytes)
	tmpBuffer := bytes.NewBuffer(tmpPoonBytes)
	accType, err := accounttype.ParseBytesToAccountType(tmpBuffer)
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(poownBytes)
	poownMessageBytes, err := ReadTransactionBytes(buffer, int(GetProofOfOwnershipSize(accType, false)))
	if err != nil {
		return nil, err
	}
	signature, err := ReadTransactionBytes(buffer, int(constant.NodeSignature))
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
	buffer.Write(poownMessage.AccountAddress)
	buffer.Write(poownMessage.BlockHash)
	buffer.Write(ConvertUint32ToBytes(poownMessage.BlockHeight))
	return buffer.Bytes()
}

// ParseProofOfOwnershipMessageBytes parse a byte array into a ProofOfOwnershipMessage struct (only the message, no signature)
func ParseProofOfOwnershipMessageBytes(poownMessageBytes []byte) (*model.ProofOfOwnershipMessage, error) {
	buffer := bytes.NewBuffer(poownMessageBytes)
	account, err := accounttype.ParseBytesToAccountType(buffer)
	if err != nil {
		return nil, err
	}
	blockHash, err := ReadTransactionBytes(buffer, int(constant.BlockHash))
	if err != nil {
		return nil, err
	}
	heightBytes, err := ReadTransactionBytes(buffer, int(constant.Height))
	if err != nil {
		return nil, err
	}
	height := ConvertBytesToUint32(heightBytes)
	accountAddress, err := account.GetAccountAddress()
	if err != nil {
		return nil, err
	}
	return &model.ProofOfOwnershipMessage{
		AccountAddress: accountAddress,
		BlockHash:      blockHash,
		BlockHeight:    height,
	}, nil
}
