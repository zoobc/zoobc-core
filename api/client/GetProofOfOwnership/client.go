package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	rpcModel "github.com/zoobc/zoobc-core/common/model"
	rpcService "github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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
	c := rpcService.NewNodeAdminServiceClient(conn)

	signature := crypto.Signature{}
	accountSeed := "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved"
	currentTime := uint64(time.Now().Unix())
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint32ToBytes(uint32(rpcModel.RequestType_GetProofOfOwnership)))
	buffer.Write(util.ConvertUint64ToBytes(currentTime))
	sig := signature.Sign(
		buffer.Bytes(),
		constant.NodeSignatureTypeDefault,
		accountSeed,
	)
	buffer.Write(sig)
	ctx := context.Background()
	md := metadata.Pairs("authorization", base64.StdEncoding.EncodeToString(buffer.Bytes()))
	ctx = metadata.NewOutgoingContext(ctx, md)

	response, err := c.GetProofOfOwnership(ctx, &rpcModel.GetProofOfOwnershipRequest{})

	if err != nil {
		log.Fatalf("error calling remote.GetProofOfOwnership: %s", err)
	}

	log.Printf("response from remote.GetProofOfOwnership(): %v", response)

}
