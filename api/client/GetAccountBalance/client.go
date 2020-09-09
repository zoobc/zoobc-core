package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	rpc_model "github.com/zoobc/zoobc-core/common/model"
	rpc_service "github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"
	"google.golang.org/grpc"
)

func main() {
	var (
		ip string
	)
	flag.StringVar(&ip, "ip", "", "Usage")
	flag.Parse()
	if len(ip) < 1 {
		if err := util.LoadConfig("../../../", "config", "toml", ""); err != nil {
			log.Fatal(err)
		} else {
			ip = fmt.Sprintf(":%d", viper.GetInt("apiRPCPort"))
		}
	}

	conn, err := grpc.Dial(ip, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := rpc_service.NewAccountBalanceServiceClient(conn)

	response, err := c.GetAccountBalance(context.Background(), &rpc_model.GetAccountBalanceRequest{
		AccountAddress: "BCZnSfqpP5tqFQlMTYkDeBVFWnbyVK7vLr5ORFpTjgtN",
	})

	if err != nil {
		log.Fatalf("error calling rpc_service.GetAccountBalance: %s", err)
	}

	j, _ := json.MarshalIndent(response, "", "  ")
	log.Printf("response from remote rpc_service.GetAccountBalance(): %s", j)

}
