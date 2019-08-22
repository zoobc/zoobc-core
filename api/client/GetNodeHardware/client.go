package main

import (
	"bytes"
	"context"
	"encoding/base64"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	rpcModel "github.com/zoobc/zoobc-core/common/model"
	rpcService "github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"time"
)

func main() {
	conn, err := grpc.Dial(":7000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := rpcService.NewNodeHardwareServiceClient(conn)
	var stream rpcService.NodeHardwareService_GetNodeHardwareClient

	if err != nil {
		log.Fatalf("error calling rpcService.GetAccountBalance: %s", err)
	}
	waitC := make(chan struct{})
	signature := crypto.Signature{}

	go func() {
		for {
			currentTime := uint64(time.Now().Unix())
			buffer := bytes.NewBuffer([]byte{})
			buffer.Write(util.ConvertUint32ToBytes(uint32(rpcModel.RequestType_GetNodeHardware)))
			buffer.Write(util.ConvertUint64ToBytes(currentTime))
			sig := signature.Sign(
				buffer.Bytes(),
				constant.NodeSignatureTypeDefault,
				"concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved",
			)
			buffer.Write(sig)

			ctx := context.Background()

			md := metadata.Pairs("authorization", base64.StdEncoding.EncodeToString(buffer.Bytes()))
			ctx = metadata.NewOutgoingContext(ctx, md)
			stream, err = c.GetNodeHardware(ctx)
			log.Println("Sleeping...")
			time.Sleep(2 * time.Second)
			log.Println("Sending request...")
			err = stream.Send(&rpcModel.GetNodeHardwareRequest{})
			if err != nil {
				log.Fatalf("error sending request to rpcService.GetNodeHardware: %s", err)
			}
			response, err := stream.Recv()
			if err != nil {
				log.Fatalf("error receiving response from rpcService.GetNodeHardware: %s", err)
			}
			log.Printf("response from remote rpcService.GetNodeHardware(): %s", response)
		}
	}()
	<-waitC
	_ = stream.CloseSend()
}
