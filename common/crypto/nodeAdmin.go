package crypto

import (
	"bytes"
	"encoding/base64"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
)

var LastRequestTimestamp uint64

// VerifyAuthAPI verify the request body and the signature of the request message, checking include
// request type checking, and the validity of the signature to the owner address
// return nil if valid, and Blocker object otherwise
func VerifyAuthAPI(
	ownerAddress []byte,
	auth string,
	requestType model.RequestType,
) error {
	// parse
	var (
		authTimestamp   uint64
		authRequestType int32
	)
	signature := NewSignature()
	authBytes, err := base64.StdEncoding.DecodeString(auth)
	if err != nil {
		return err
	}
	authBytesBuffer := bytes.NewBuffer(authBytes)
	authTimestamp = util.ConvertBytesToUint64(authBytesBuffer.Next(constant.AuthTimestamp))
	authRequestType = int32(util.ConvertBytesToUint32(authBytesBuffer.Next(constant.AuthRequestType)))
	if authRequestType != int32(requestType) {
		return blocker.NewBlocker(
			blocker.RequestParameterErr,
			"invalid request type",
		)
	}
	if authTimestamp <= LastRequestTimestamp {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"timestamp is in the past",
		)
	}
	err = signature.VerifySignature(
		authBytes[:constant.AuthRequestType+constant.AuthTimestamp],
		authBytes[constant.AuthRequestType+constant.AuthTimestamp:],
		ownerAddress,
	)
	if err != nil {
		return err
	}
	// if signature valid, update last request timestamp
	LastRequestTimestamp = authTimestamp
	return nil
}
