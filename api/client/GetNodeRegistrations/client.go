package main

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	rpcModel "github.com/zoobc/zoobc-core/common/model"
	rpcService "github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"
	"google.golang.org/grpc"
)

func main() {
	var apiRPCPort int
	if err := util.LoadConfig("../../../", "config", "toml", false); err != nil {
		log.Fatal(err)
	} else {
		apiRPCPort = viper.GetInt("apiRPCPort")
	}

	conn, err := grpc.Dial(fmt.Sprintf(":%d", apiRPCPort), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := rpcService.NewNodeRegistrationServiceClient(conn)

	response, err := c.GetNodeRegistrations(context.Background(),
		&rpcModel.GetNodeRegistrationsRequest{
			MinRegistrationHeight: 0,
			MaxRegistrationHeight: 1,
			RegistrationStatuses:  []uint32{1, 2},
		})

	if err != nil {
		log.Fatalf("error calling rpc_service.GetNodeRegistrations: %v", err)
	}

	log.Printf("response from remote rpc_service.GetNodeRegistrations(): %v", response)

}
