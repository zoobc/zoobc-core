package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	rpc_model "github.com/zoobc/zoobc-core/common/model"
	rpc_service "github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"
	"google.golang.org/grpc"
)

func main() {
	var ip string
	flag.StringVar(&ip, "ip", "", "Usage")
	flag.Parse()
	if len(ip) < 1 {
		config, err := util.LoadConfig("../../../", "config", "toml", "", false)
		if err != nil {
			log.Fatal(err)
		} else {
			ip = fmt.Sprintf(":%d", config.RPCAPIPort)
		}
	}
	conn, err := grpc.Dial(ip, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := rpc_service.NewLiquidPaymentServiceClient(conn)

	response, err := c.GetLiquidTransactions(context.Background(),
		&rpc_model.GetLiquidTransactionsRequest{
			// Status: rpc_model.LiquidPaymentStatus_LiquidPaymentCompleted,
			Pagination: &rpc_model.Pagination{
				OrderField: "block_height",
				OrderBy:    rpc_model.OrderBy_DESC,
				Page:       1,
				Limit:      10,
			},
		},
	)

	if err != nil {
		log.Fatalf("error calling remote.GetLiquidTransactions: %s", err)
	}

	log.Printf("response from remote.GetLiquidTransactions(): %v", response)

}
