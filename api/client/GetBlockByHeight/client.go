package main

import (
	"context"
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
		ip         string
		apiRPCPort int
	)
	flag.StringVar(&ip, "ip", "", "Usage")
	flag.Parse()
	if len(ip) < 1 {
		if err := util.LoadConfig("../../../resource", "config", "toml"); err != nil {
			log.Fatal(err)
		} else {
			ip = fmt.Sprintf(":%d", viper.GetInt("apiRPCPort"))
		}
	}

	conn, err := grpc.Dial(fmt.Sprintf(":%d", apiRPCPort), grpc.WithInsecure())
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
