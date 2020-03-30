package main

import (
	"context"
	"fmt"
	"log"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/common/model"
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
	}

	conn, err := grpc.Dial(fmt.Sprintf(":%d", apiRPCPort), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := rpcService.NewAccountDatasetServiceClient(conn)
	response, err := c.GetAccountDataset(context.Background(), &model.GetAccountDatasetRequest{
		RecipientAccountAddress: "H1ftvv3n6CF5NDzdjmZKLRrBg6yPKHXpmatVUhQ5NWYx",
	})
	if err != nil {
		log.Fatalf("error calling grpc GetAccountDatasets: %s", err.Error())
	}
	log.Printf("response from remote rpc_service.GetTransactions(): %s", response)
}
