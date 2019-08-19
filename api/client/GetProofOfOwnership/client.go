package main

import (
	"bytes"
	"context"
	"github.com/zoobc/zoobc-core/common/constant"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/crypto"
	rpc_model "github.com/zoobc/zoobc-core/common/model"
	rpc_service "github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial(":3001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := rpc_service.NewNodeAdminServiceClient(conn)

	sig := crypto.NewSignature().Sign([]byte("BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE"),
		constant.NodeSignatureTypeDefault,
		"concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved")
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint32ToBytes(1))
	buffer.Write(sig)
	newSig := buffer.Bytes()
	response, err := c.GetProofOfOwnership(context.Background(), &rpc_model.GetProofOfOwnershipRequest{
		AccountAddress: "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
		Signature:      newSig,
	})

	if err != nil {
		log.Fatalf("error calling remote.GetProofOfOwnership: %s", err)
	}

	log.Printf("response from remote.GetProofOfOwnership(): %v", response)

}
