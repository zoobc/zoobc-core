package main

import (
	"context"
	"encoding/json"
	"flag"

	log "github.com/sirupsen/logrus"
	rpc_model "github.com/zoobc/zoobc-core/common/model"
	rpc_service "github.com/zoobc/zoobc-core/common/service"
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
		ip = ":3001"
	}
	conn, err = grpc.Dial(ip, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := rpc_service.NewP2PCommunicationClient(conn)

	response, err := c.GetMorePeers(context.Background(), &rpc_model.Empty{})

	if err != nil {
		log.Fatalf("error calling rpc_service.GetPeerInfo: %s", err)
	}

	j, _ := json.MarshalIndent(response, "", "  ")

	log.Printf("response from remote rpc_service.GetPeerInfo(): %s", j)

}
