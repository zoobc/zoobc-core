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

	c := rpc_service.NewBlockServiceClient(conn)

	response, err := c.GetBlock(context.Background(), &rpc_model.GetBlockRequest{
		Height: 3,
	})

	if err != nil {
		log.Fatalf("error calling rpc_service.GetBlockByHeight: %s", err)
	}

	log.Printf("response from remote rpc_service.GetBlockByHeight(): %s", response)

}
