package main

import (
	"context"
	"fmt"
	"log"

	"github.com/sirupsen/logrus"

	"github.com/spf13/viper"
	rpc_model "github.com/zoobc/zoobc-core/common/model"
	rpc_service "github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"
	"google.golang.org/grpc"
)

func main() {
	var apiRPCPort int
	if err := util.LoadConfig("../../../", "config", "toml", ""); err != nil {
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

	c := rpc_service.NewLiquidPaymentServiceClient(conn)

	response, err := c.GetLiquidTransactions(context.Background(),
		&rpc_model.GetLiquidTransactionsRequest{
			// Status: rpc_model.LiquidPaymentStatus_LiquidPaymentCompleted,
			Pagination: &rpc_model.Pagination{
				OrderField: "block_height",
				OrderBy:    rpc_model.OrderBy_DESC,
				Page:       1,
				Limit:      10,
			},
		},
	)

	if err != nil {
		log.Fatalf("error calling remote.GetLiquidTransactions: %s", err)
	}

	log.Printf("response from remote.GetLiquidTransactions(): %v", response)

}
