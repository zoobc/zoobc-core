package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/sirupsen/logrus"

	"github.com/spf13/viper"
	rpc_model "github.com/zoobc/zoobc-core/common/model"
	rpc_service "github.com/zoobc/zoobc-core/common/service"
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

	c := rpc_service.NewNodeHardwareServiceClient(conn)

	response, err := c.GetNodeTime(context.Background(),
		&rpc_model.Empty{},
	)

	if err != nil {
		log.Fatalf("error calling rpc.GetNodeTime: %s", err)
	}

	log.Println(response)
	j, _ := json.MarshalIndent(response, "", "  ")

	log.Printf("response from rpc.GetNodeTime(): %v", j)

}
