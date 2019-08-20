package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/zoobc/zoobc-core/common/constant"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/common/crypto"
	rpc_model "github.com/zoobc/zoobc-core/common/model"
	rpc_service "github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"
	"google.golang.org/grpc"
)

func main() {
	var apiRPCPort int
	if err := util.LoadConfig("../../../resource", "config", "toml"); err != nil {
		logrus.Fatal(err)
	} else {
		apiRPCPort = viper.GetInt("apiRPCPort")
		if apiRPCPort == 0 {
			apiRPCPort = 8080
		}
	}

	conn, err := grpc.Dial(fmt.Sprintf(":%d", apiRPCPort), grpc.WithInsecure())
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
