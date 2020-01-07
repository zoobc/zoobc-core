package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/crypto"
	"google.golang.org/grpc/metadata"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/zoobc/zoobc-core/common/model"
	rpcService "github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"
	"google.golang.org/grpc"
)

func main() {
	var (
		apiRPCPort int
		configPath = "./resource"
	)
	dir, _ := os.Getwd()
	if strings.Contains(dir, "api") {
		configPath = "../../../resource"
	}
	if err := util.LoadConfig(configPath, "config", "toml"); err != nil {
		logrus.Fatal(err)
	} else {
		apiRPCPort = viper.GetInt("apiRPCPort")
		if apiRPCPort == 0 {
			apiRPCPort = 8080
		}
	}
	conn, err := grpc.Dial(fmt.Sprintf(":%d", apiRPCPort), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	signature := crypto.Signature{}
	currentTime := uint64(time.Now().Unix())
	buffer := bytes.NewBuffer([]byte{})
	buffer.Write(util.ConvertUint64ToBytes(currentTime))
	buffer.Write(util.ConvertUint32ToBytes(uint32(model.RequestType_GetAccountLedgers)))
	sig := signature.Sign(
		buffer.Bytes(),
		constant.SignatureTypeDefault,
		"concur vocalist rotten busload gap quote stinging undiluted surfer goofiness deviation starved",
	)
	buffer.Write(sig)

	md := metadata.Pairs("authorization", base64.StdEncoding.EncodeToString(buffer.Bytes()))
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	res, err := rpcService.NewAccountLedgerServiceClient(conn).
		GetAccountLedgers(ctx, &model.GetAccountLedgersRequest{
			AccountAddress: "OnEYzI-EMV6UTfoUEzpQUjkSlnqB82-SyRN7469lJTWH",
			EventType:      model.EventType_EventAny,
			Pagination: &model.Pagination{
				OrderField: "account_address",
				OrderBy:    model.OrderBy_ASC,
				Page:       1,
				Limit:      3,
			},
		})
	if err != nil {
		log.Fatalf("error calling rpc_service.GetBlockByID: %s", err)
	}

	log.Printf("response from remote rpc_service.ID(): %s", res)

}
