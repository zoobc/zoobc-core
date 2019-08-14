package main

import (
	"context"

	log "github.com/sirupsen/logrus"
	rpcModel "github.com/zoobc/zoobc-core/common/model"
	rpcService "github.com/zoobc/zoobc-core/common/service"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial(":3001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := rpcService.NewTransactionServiceClient(conn)

	response, err := c.GetTransactions(context.Background(), &rpcModel.GetTransactionsRequest{
		AccountAddress: "BCZD_VxfO2S9aziIL3cn_cXW7uPDVPOrnXuP98GEAUC7",
		Limit:          1,
		Page:           0,
	})

	if err != nil {
		log.Fatalf("error calling rpc_service.GetTransactions: %s", err)
	}

	log.Printf("response from remote rpc_service.GetTransactions(): %s", response)

}
