package main

import (
	"bytes"
	"context"
	"fmt"
	"time"

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

	accountSeed := "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved"
	accountAddress := "BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE"
	timestamp := time.Now().Unix()
	message := append([]byte(accountAddress), util.ConvertUint64ToBytes(uint64(timestamp))...)
	sig := crypto.NewSignature().Sign(message,
		constant.NodeSignatureTypeDefault,
		accountSeed)
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint32ToBytes(1))
	buffer.Write(sig)
	response, err := c.GenerateNodeKey(context.Background(), &rpc_model.GenerateNodeKeyRequest{
		AccountAddress: accountAddress,
		Timestamp:      timestamp,
		Signature:      sig,
	})

	if err != nil {
		log.Fatalf("error calling remote.GenerateNodeKey: %s", err)
	}

	log.Printf("response from remote.GenerateNodeKey(): %v", response)

}
