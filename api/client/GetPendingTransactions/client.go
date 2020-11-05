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

	c := rpc_service.NewMultisigServiceClient(conn)

	response, err := c.GetPendingTransactions(context.Background(),
		&rpc_model.GetPendingTransactionsRequest{
			SenderAddress: "E6u7lDnLgyiPuklLd6rXNQJI3_kGA1Q7e1BEXdJVB1hy",
			Status:        rpc_model.PendingTransactionStatus_PendingTransactionPending,
			Pagination: &rpc_model.Pagination{
				OrderField: "block_height",
				OrderBy:    rpc_model.OrderBy_DESC,
				Page:       2,
				Limit:      1,
			},
		},
	)

	if err != nil {
		log.Fatalf("error calling remote.GetBlocks: %s", err)
	}

	log.Printf("response from remote.GetBlocks(): %v", response)

}
