package main

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	rpc_model "github.com/zoobc/zoobc-core/common/model"
	rpc_service "github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"
	"google.golang.org/grpc"
)

func main() {
	var apiRPCPort int
	if err := util.LoadConfig("../../../resource", "config", "toml"); err != nil {
		log.Fatal(err)
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

	c := rpc_service.NewAccountBalanceServiceClient(conn)
	fmt.Println("LEWATTTT 1")

	accountAddresses := []string{
		"iSJt3H8wFOzlWKsy_UoEWF_OjF6oymHMqthyUMDKSyxb",
		"BCZEGOb3WNx3fDOVf9ZS4EjvOIv_UeW4TVBQJ_6tHKlE",
		"OnEYzI-EMV6UTfoUEzpQUjkSlnqB82-SyRN7469lJTWH",
	}
	fmt.Println("accountAddresses::", accountAddresses)

	response, err := c.GetAccountBalances(context.Background(), &rpc_model.GetAccountBalancesRequest{
		AccountAddresses: accountAddresses,
	})

	if err != nil {
		log.Fatalf("error calling rpc_service.GetAccountBalance: %s", err)
	}

	log.Printf("response from remote rpc_service.GetBlockByID(): %s", response)

}
