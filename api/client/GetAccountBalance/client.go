package main

import (
	"context"

	log "github.com/sirupsen/logrus"
	rpc_model "github.com/zoobc/zoobc-core/common/model"
	rpc_service "github.com/zoobc/zoobc-core/common/service"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial(":3001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := rpc_service.NewAccountBalanceServiceClient(conn)

	response, err := c.GetAccountBalance(context.Background(), &rpc_model.GetAccountBalanceRequest{
		AccountAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
	})

	if err != nil {
		log.Fatalf("error calling rpc_service.GetAccountBalance: %s", err)
	}

	log.Printf("response from remote rpc_service.GetBlockByID(): %s", response)

}
