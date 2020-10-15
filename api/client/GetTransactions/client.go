package main

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	rpcModel "github.com/zoobc/zoobc-core/common/model"
	rpcService "github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"
	"google.golang.org/grpc"
)

func main() {
	var apiRPCPort int
	if err := util.LoadConfig("../../../", "config", "toml", ""); err != nil {
		log.Fatal(err)
	} else {
		apiRPCPort = viper.GetInt("apiRPCPort")
	}

	conn, err := grpc.Dial(fmt.Sprintf(":%d", apiRPCPort), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := rpcService.NewTransactionServiceClient(conn)

	response, err := c.GetTransactions(context.Background(), &rpcModel.GetTransactionsRequest{
		TransactionType: 2,
		AccountAddress: []byte{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112, 113, 140, 183, 229,
			116, 202, 211, 235, 190, 224, 217, 238, 63, 223, 225, 162},
		Pagination: &rpcModel.Pagination{
			Limit: 1,
			Page:  0,
		},
	})

	if err != nil {
		log.Fatalf("error calling rpc_service.GetTransactions: %s", err)
	}

	log.Printf("response from remote rpc_service.GetTransactions(): %s", response)

}
