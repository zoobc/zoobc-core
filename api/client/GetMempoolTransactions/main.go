package main

import (
	"context"
	"log"

	rpcModel "github.com/zoobc/zoobc-core/common/model"
	rpcService "github.com/zoobc/zoobc-core/common/service"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial(":7000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := rpcService.NewMempoolServiceClient(conn)
	response, err := c.GetMempoolTransactions(context.Background(), &rpcModel.GetMempoolTransactionsRequest{
		Address: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
	})

	if err != nil {
		log.Fatalf("error calling rpc_service.GetTransactions: %s", err)
	}

	log.Printf("response from remote rpc_service.GetTransactions(): %s", response)

}
