package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"time"

	"github.com/zoobc/zoobc-core/common/crypto"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	rpcModel "github.com/zoobc/zoobc-core/common/model"
	rpcService "github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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
	c := rpcService.NewNodeAdminServiceClient(conn)

	signature := crypto.Signature{}
	accountSeed := "concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved"
	currentTime := uint64(time.Now().Unix())
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint64ToBytes(currentTime))
	buffer.Write(util.ConvertUint32ToBytes(uint32(rpcModel.RequestType_GetProofOfOwnership)))
	sig, err := signature.Sign(
		buffer.Bytes(),
		rpcModel.SignatureType_DefaultSignature,
		accountSeed,
	)
	if err != nil {
		log.Fatalf("error signing payload: %s", err)
	}

	buffer.Write(sig)
	ctx := context.Background()
	md := metadata.Pairs("authorization", base64.StdEncoding.EncodeToString(buffer.Bytes()))

	ctx = metadata.NewOutgoingContext(ctx, md)

	response, err := c.GetProofOfOwnership(ctx, &rpcModel.GetProofOfOwnershipRequest{})

	if err != nil {
		log.Fatalf("error calling remote.GetProofOfOwnership: %s", err)
	}

	log.Printf("response from remote.GetProofOfOwnership(): %v", response)

}
