package main

import (
	"context"
	"fmt"
	"log"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	rpcModel "github.com/zoobc/zoobc-core/common/model"
	rpcService "github.com/zoobc/zoobc-core/common/service"
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

	c := rpcService.NewEscrowTransactionServiceClient(conn)

	response, err := c.GetEscrowTransaction(context.Background(), &rpcModel.GetEscrowTransactionRequest{
		ID: -2747806915165203447,
	})

	if err != nil {
		log.Fatalf("error calling : %s", err)
	}

	log.Printf("response from remote : %s", response)

}
