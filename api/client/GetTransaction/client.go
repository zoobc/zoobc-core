package main

import (
	"context"

	log "github.com/sirupsen/logrus"
	rpc_model "github.com/zoobc/zoobc-core/common/model"
	rpc_service "github.com/zoobc/zoobc-core/common/service"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial(":8000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := rpc_service.NewTransactionServiceClient(conn)

	response, err := c.GetTransaction(context.Background(), &rpc_model.GetTransactionRequest{
		ID: 31234567888765,
	})

	if err != nil {
		log.Fatalf("error calling rpc_service.GetTransaction: %s", err)
	}

	log.Printf("response from remote rpc_service.GetTransaction(): %s", response)

}
