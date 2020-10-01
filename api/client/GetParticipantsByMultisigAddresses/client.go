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
			MultisigAddresses: []string{
				"ZBC_XHRAYYEM_TVCKY56B_SD3EY5QA_OBYYZN7F_OTFNH256_4DM64P67_4GRBA673",
				"ZBC_6ULV6WBV_J3JVOADT_32COLTXK_KFDHGLW4_LRZ2NEDU_YK4Z3XWS_NLU2VOMX",
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
