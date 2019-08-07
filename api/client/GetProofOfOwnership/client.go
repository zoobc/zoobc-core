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

	c := rpc_service.NewGeneratePoownServiceClient(conn)

	response, err := c.GetProofOfOwnership(context.Background(), &rpc_model.GetProofOfOwnershipRequest{})

	if err != nil {
		log.Fatalf("error calling remote.GetProofOfOwnership: %s", err)
	}

	log.Printf("response from remote.GetProofOfOwnership(): %v", response)

}
