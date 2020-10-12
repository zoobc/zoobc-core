package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/common/model"
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
	}

	conn, err := grpc.Dial(fmt.Sprintf(":%d", apiRPCPort), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := rpcService.NewAccountDatasetServiceClient(conn)
	response, err := c.GetAccountDatasets(context.Background(), &model.GetAccountDatasetsRequest{
		SetterAccountAddress: []byte{0, 0, 0, 0, 98, 118, 38, 51, 199, 143, 112, 175, 220, 74, 221, 170, 56, 103, 159, 209, 242, 132, 219,
			155, 169, 123, 104, 77, 139, 18, 224, 166, 162, 83, 125, 96},
		RecipientAccountAddress: []byte{0, 0, 0, 0, 27, 175, 195, 164, 131, 47, 248, 249, 105, 108, 85, 109, 152, 244, 63, 212, 143, 10,
			227, 1, 190, 31, 63, 64, 219, 176, 99, 37, 78, 130, 27, 40},
	})
	if err != nil {
		log.Fatalf("error calling grpc GetAccountDatasets: %s", err.Error())
	}
	j, _ := json.MarshalIndent(response, "", "  ")
	log.Printf("response from remote rpc_service.GetAccountDatasets(): %s", j)
}
