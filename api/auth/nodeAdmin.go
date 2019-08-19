package auth

import (
	"bytes"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/util"
)

var LastRequestTimestamp uint64

// VerifyAuthAPI verify the request body and the signature of the request message, checking include
// request type checking, and the validity of the signature to the owner address
// return nil if valid, and Blocker object otherwise
func VerifyAuthAPI(
	ownerAddress string,
	auth *model.Auth,
	requestType model.RequestType,
	signature crypto.SignatureInterface) error {
	if auth.RequestType != requestType {
		return blocker.NewBlocker(
			blocker.RequestParameterErr,
			"invalid request type",
		)
	}
	if auth.Timestamp <= LastRequestTimestamp {
		return blocker.NewBlocker(
			blocker.ValidationErr,
			"timestamp is in the past",
		)
	}
	LastRequestTimestamp = auth.Timestamp
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint32ToBytes(uint32(auth.RequestType)))
	buffer.Write(util.ConvertUint64ToBytes(auth.Timestamp))
	signatureValid := signature.VerifySignature(
		buffer.Bytes(),
		auth.Signature,
		ownerAddress,
	)
	if !signatureValid {
		return blocker.NewBlocker(blocker.ValidationErr, "request signature invalid")
	}
	return nil
}
