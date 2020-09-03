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
		ip   string
		conn *grpc.ClientConn
		err  error
	)
	flag.StringVar(&ip, "ip", "", "Usage")
	flag.Parse()
	if len(ip) < 1 {
		if err := util.LoadConfig("../../../", "config", "toml"); err != nil {
			log.Fatal(err)
		} else {
			ip = fmt.Sprintf(":%d", viper.GetInt("apiRPCPort"))
		}
	}
	conn, err = grpc.Dial(ip, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := rpc_service.NewSkippedBlockSmithsServiceClient(conn)

	response, err := c.GetSkippedBlockSmiths(
		context.Background(),
		&rpc_model.GetSkippedBlocksmithsRequest{
			BlockHeightStart: 1,
			BlockHeightEnd:   10,
		},
	)

	if err != nil {
		log.Fatalf("error calling rpc_service.GetSkippedBlockSmiths: %s", err)
	}

	j, _ := json.MarshalIndent(response, "", "  ")
	log.Printf("response from remote rpc_service.GetSkippedBlockSmiths(): %s", j)

}
