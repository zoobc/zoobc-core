package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/common/model"
	rpcService "github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"
	"google.golang.org/grpc"
)

func main() {
	var (
		apiRPCPort int
		configPath = "./resource"
	)
	dir, _ := os.Getwd()
	if strings.Contains(dir, "api") {
		configPath = "../../../resource"
	}
	if err := util.LoadConfig(configPath, "config", "toml"); err != nil {
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

	res, err := rpcService.NewAccountLedgerServiceClient(conn).
		GetAccountLedgers(context.Background(), &model.GetAccountLedgersRequest{
			AccountAddress: "OnEYzI-EMV6UTfoUEzpQUjkSlnqB82-SyRN7469lJTWH",
			// EventType:      model.EventType_EventAny,
			TimestampStart: 1578548985,
			TimestampEnd:   1578549075,
			Pagination: &model.Pagination{
				OrderField: "account_address",
				OrderBy:    model.OrderBy_ASC,
				Page:       1,
				Limit:      2,
			},
		})
	if err != nil {
		log.Fatalf("error calling rpc_service.GetBlockByID: %s", err)
	}

	log.Printf("response from remote rpc_service.ID(): %s", res)

}
