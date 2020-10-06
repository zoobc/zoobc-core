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

	response, err := c.GetAccountBalances(context.Background(), &rpc_model.GetAccountBalancesRequest{
		AccountAddresses: [][]byte{
			[]byte{0, 0, 0, 0, 98, 118, 38, 51, 199, 143, 112, 175, 220, 74, 221, 170, 56, 103, 159, 209, 242, 132, 219, 155,
				169, 123, 104, 77, 139, 18, 224, 166, 162, 83, 125, 96},
			[]byte{0, 0, 0, 0, 27, 175, 195, 164, 131, 47, 248, 249, 105, 108, 85, 109, 152, 244, 63, 212, 143, 10, 227, 1, 190, 31, 63,
				64, 219, 176, 99, 37, 78, 130, 27, 40},
			[]byte{0, 0, 0, 0, 192, 124, 122, 102, 248, 101, 1, 230, 239, 101, 6, 87, 182, 13, 221, 113, 69, 154, 170, 121, 179, 223, 171,
				177, 193, 38, 101, 24, 132, 147, 122, 9},
		},
	})

	if err != nil {
		log.Fatalf("error calling rpc_service.GetAccountBalances: %s", err)
	}
	j, _ := json.MarshalIndent(response, "", "  ")
	log.Printf("response from remote rpc_service.GetAccountBalances(): %s", j)

}
