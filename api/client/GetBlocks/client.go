package main

import (
	"context"
	"log"

	// log "github.com/sirupsen/logrus"

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

	response, err := c.GetBlocks(context.Background(), &rpc_model.GetBlocksRequest{
		BlockSize:   3,
		BlockHeight: 12,
	})

	if err != nil {
		log.Fatalf("error calling remote.GetBlocks: %s", err)
	}

	log.Printf("response from remote.GetBlocks(): %v", response)

}
