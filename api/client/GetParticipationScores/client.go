package main

import (
	"context"
	"encoding/json"
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
	if err := util.LoadConfig("../../../", "config", "toml"); err != nil {
		log.Fatal(err)
	} else {
		apiRPCPort = viper.GetInt("apiRPCPort")
	}

	conn, err := grpc.Dial(fmt.Sprintf(":%d", apiRPCPort), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := rpcService.NewParticipationScoreServiceClient(conn)

	response, err := c.GetParticipationScores(context.Background(), &rpcModel.GetParticipationScoresRequest{
		FromHeight: 0,
		ToHeight:   20,
	})

	if err != nil {
		log.Fatalf("error calling rpc_service.GetParticipationScores: %s", err)
	}

	j, _ := json.MarshalIndent(response, "", "  ")

	log.Printf("response from remote rpc_service.GetParticipationScores(): %s", j)

}
