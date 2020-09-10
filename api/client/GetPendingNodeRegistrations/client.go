package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/common/crypto"
	rpcModel "github.com/zoobc/zoobc-core/common/model"
	rpcService "github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func main() {
	var (
		ip         string
		apiRPCPort = 7000
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
	conn, err := grpc.Dial(fmt.Sprintf(":%d", apiRPCPort), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := rpcService.NewNodeRegistrationServiceClient(conn)
	var stream rpcService.NodeRegistrationService_GetPendingNodeRegistrationsClient

	if err != nil {
		log.Fatalf("error calling rpcService.GetPendingNodeRegistrations: %s", err)
	}
	waitC := make(chan struct{})
	signature := crypto.Signature{}
	currentTime := uint64(time.Now().Unix())
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint64ToBytes(currentTime))
	buffer.Write(util.ConvertUint32ToBytes(uint32(rpcModel.RequestType_GetPendingNodeRegistrationsStream)))
	sig, err := signature.Sign(
		buffer.Bytes(),
		rpcModel.SignatureType_DefaultSignature,
		"concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved",
	)
	if err != nil {
		log.Fatalf("error signing payload: %s", err)
	}
	buffer.Write(sig)

	ctx := context.Background()

	md := metadata.Pairs("authorization", base64.StdEncoding.EncodeToString(buffer.Bytes()))
	ctx = metadata.NewOutgoingContext(ctx, md)
	stream, err = c.GetPendingNodeRegistrations(ctx)
	if err != nil {
		log.Fatal(err)
	}
	err = stream.Send(&rpcModel.GetPendingNodeRegistrationsRequest{
		Limit: 2,
	})
	if err != nil {
		log.Fatalf("error sending request to rpcService.GetPendingNodeRegistrations: %s", err)
	}
	go func() {
		for {
			response, err := stream.Recv()
			if err != nil {
				log.Fatalf("error receiving response from rpcService.GetPendingNodeRegistrations: %s", err)
			}
			j, _ := json.MarshalIndent(response, "", "  ")
			log.Printf("response from remote rpcService.GetPendingNodeRegistrations(): %s", j)
		}
	}()
	<-waitC
	_ = stream.CloseSend()
}
