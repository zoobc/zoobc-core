package main

import (
	"bytes"
	"context"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/util"
	"time"

	log "github.com/sirupsen/logrus"
	rpcModel "github.com/zoobc/zoobc-core/common/model"
	rpcService "github.com/zoobc/zoobc-core/common/service"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial(":7000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := rpcService.NewNodeHardwareServiceClient(conn)
	stream, err := c.GetNodeHardware(context.Background())
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
			auth := &rpcModel.Auth{
				RequestType: rpcModel.RequestType_GetNodeHardware,
				Timestamp:  currentTime,
				Signature: signature.Sign(
					buffer.Bytes(),
					constant.NodeSignatureTypeDefault,
					"concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved",
				),
			}
			log.Println("Sleeping...")
			time.Sleep(2 * time.Second)
			log.Println("Sending request...")
			err = stream.Send(&rpcModel.GetNodeHardwareRequest{
				Auth: auth,
			})
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
