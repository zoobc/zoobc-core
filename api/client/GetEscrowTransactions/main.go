package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	rpcModel "github.com/zoobc/zoobc-core/common/model"
	rpcService "github.com/zoobc/zoobc-core/common/service"
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

	c := rpcService.NewEscrowTransactionServiceClient(conn)

	response, err := c.GetEscrowTransactions(context.Background(), &rpcModel.GetEscrowTransactionsRequest{
		ApproverAddress: []byte{0, 0, 0, 0, 98, 118, 38, 51, 199, 143, 112, 175, 220, 74, 221, 170, 56, 103, 159, 209, 242, 132, 219,
			155, 169, 123, 104, 77, 139, 18, 224, 166, 162, 83, 125, 96},
		Statuses: []rpcModel.EscrowStatus{
			rpcModel.EscrowStatus_Approved,
		},
	})

	if err != nil {
		log.Fatalf("error calling : %s", err)
	}
	j, _ := json.MarshalIndent(response, "", "  ")
	log.Printf("response from remote : %s", j)

}
