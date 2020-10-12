package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/service"
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
		if err := util.LoadConfig("../../../", "config", "toml", ""); err != nil {
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

	c := service.NewMultisigServiceClient(conn)

	response, err := c.GetParticipantsByMultisigAddresses(context.Background(),
		&model.GetParticipantsByMultisigAddressesRequest{
			MultisigAddresses: [][]byte{
				{0, 0, 0, 0, 185, 226, 12, 96, 140, 157, 68, 172, 119, 193, 144, 246, 76, 118, 0, 112, 113, 140, 183, 229, 116, 202,
					211, 235, 190, 224, 217, 238, 63, 223, 225, 162},
				{0, 0, 0, 0, 245, 23, 95, 88, 53, 78, 211, 87, 0, 115, 222, 132, 229, 206, 234, 81, 70, 115, 46, 220, 92, 115, 166,
					144, 116, 194, 185, 157, 222, 210, 106, 233},
			},
			Pagination: &model.Pagination{
				OrderField: "block_height",
				OrderBy:    model.OrderBy_ASC,
			},
		},
	)

	if err != nil {
		log.Fatalf("error calling remote.GetParticipantsByMultisigAddresses: %s", err)
	}

	log.Printf("response from remote.GetParticipantsByMultisigAddresses(): %v", response)

}
