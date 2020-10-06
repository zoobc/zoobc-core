package main

import (
	"context"
	"encoding/json"
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
		configPath = "./"
	)
	dir, _ := os.Getwd()
	if strings.Contains(dir, "api") {
		configPath = "../../../"
	}
	if err := util.LoadConfig(configPath, "config", "toml", ""); err != nil {
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
			AccountAddress: []byte{0, 0, 0, 0, 98, 118, 38, 51, 199, 143, 112, 175, 220, 74, 221, 170, 56, 103, 159, 209, 242, 132, 219, 155,
				169, 123, 104, 77, 139, 18, 224, 166, 162, 83, 125, 96},
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
		log.Fatalf("error calling rpc_service.GetAccountLedgers: %s", err)
	}
	j, _ := json.MarshalIndent(res, "", "  ")
	log.Printf("response from remote GetAccountLedgers.ID(): %s", j)

}
