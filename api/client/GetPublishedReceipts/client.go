package main

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/common/util"

	log "github.com/sirupsen/logrus"
	rpcModel "github.com/zoobc/zoobc-core/common/model"
	rpcService "github.com/zoobc/zoobc-core/common/service"
	"google.golang.org/grpc"
)

func main() {
	var apiRPCPort int
	if err := util.LoadConfig("../../../", "config", "toml"); err != nil {
		log.Fatal(err)
	} else {
		apiRPCPort = viper.GetInt("apiRPCPort")
	}

	conn, err := grpc.Dial(fmt.Sprintf(":%d", apiRPCPort), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := rpcService.NewPublishedReceiptServiceClient(conn)

	response, err := c.GetPublishedReceipts(context.Background(), &rpcModel.GetPublishedReceiptsRequest{
		FromHeight: 369,
		ToHeight:   374,
	})

	if err != nil {
		log.Fatalf("error calling rpc_service.GetTransactions: %s", err)
	}

	for _, receipt := range response.GetPublishedReceipts() {
		fmt.Printf("blockHeight: %d\tindex: %d\n", receipt.GetBlockHeight(), receipt.GetPublishedIndex())
	}
}
