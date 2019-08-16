package main

import (
	"context"
	"github.com/zoobc/zoobc-core/common/crypto"
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
	poow := &rpcModel.ProofOfOwnership{
		MessageBytes: []byte("HelloBlock"),
		Signature: signature.SignByNode(
			[]byte("HelloBlock"),
			"sprinkled sneak species pork outpost thrift unwind cheesy vexingly dizzy neurology neatness",
		),
	}
	go func() {
		for {
			log.Println("Sleeping...")
			time.Sleep(2 * time.Second)
			log.Println("Sending request...")
			err = stream.Send(&rpcModel.GetNodeHardwareRequest{
				ProofOfOwnership: poow,
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
