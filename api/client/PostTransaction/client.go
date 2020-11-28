package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	rpc_model "github.com/zoobc/zoobc-core/common/model"
	rpc_service "github.com/zoobc/zoobc-core/common/service"
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
		configPath = "../../../"
	}
	if err := util.LoadConfig(configPath, "config", "toml", ""); err != nil {
		log.Fatal(err)
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

	c := rpc_service.NewTransactionServiceClient(conn)

	response, err := c.PostTransaction(context.Background(), &rpc_model.PostTransactionRequest{
		// Sendmoney
		TransactionBytes: []byte{1, 0, 0, 0, 1, 75, 93, 171, 95, 0, 0, 0, 0, 0, 0, 0, 0, 236, 125, 37, 22, 103, 77, 115, 149, 65, 98, 75,
			252, 148, 113, 91, 119, 67, 138, 240, 89, 57, 28, 107, 162, 225, 82, 79, 186, 163, 158, 161, 115, 0, 0, 0, 0, 123, 166, 78,
			235, 41, 31, 17, 3, 254, 32, 33, 149, 6, 209, 16, 250, 23, 74, 126, 200, 54, 255, 196, 135, 192, 128, 218, 130, 73, 31, 171,
			72, 65, 66, 15, 0, 0, 0, 0, 0, 8, 0, 0, 0, 0, 225, 245, 5, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 115, 15, 32, 47, 154, 231, 30,
			42, 41, 117, 125, 241, 125, 149, 43, 42, 139, 73, 40, 69, 199, 95, 82, 16, 241, 158, 229, 122, 86, 55, 7, 48, 201, 105, 197,
			107, 159, 203, 89, 109, 245, 231, 11, 115, 67, 61, 67, 128, 7, 52, 109, 217, 41, 252, 26, 135, 25, 129, 140, 182, 82, 38, 78, 6},
	})

	if err != nil {
		log.Fatalf("error calling rpc_service.PostTransaction: %s", err)
	}

	log.Printf("response from remote rpc_service.PostTransaction(): %s", response)

	time.Sleep(2 * time.Second)
	response, err = c.PostTransaction(context.Background(), &rpc_model.PostTransactionRequest{
		// Sendmoney
		TransactionBytes: []byte{1, 0, 0, 0, 1, 177, 96, 171, 95, 0, 0, 0, 0, 0, 0, 0, 0, 236, 125, 37, 22, 103, 77, 115, 149, 65, 98,
			75, 252, 148, 113, 91, 119, 67, 138, 240, 89, 57, 28, 107, 162, 225, 82, 79, 186, 163, 158, 161, 115, 0, 0, 0, 0, 123, 166,
			78, 235, 41, 31, 17, 3, 254, 32, 33, 149, 6, 209, 16, 250, 23, 74, 126, 200, 54, 255, 196, 135, 192, 128, 218, 130, 73, 31,
			171, 72, 65, 66, 15, 0, 0, 0, 0, 0, 8, 0, 0, 0, 45, 4, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 240, 38, 121, 109, 62, 178,
			154, 15, 61, 224, 134, 100, 47, 143, 174, 4, 59, 194, 37, 172, 23, 22, 138, 253, 117, 3, 248, 239, 207, 133, 3, 226, 77, 175,
			128, 201, 61, 101, 93, 33, 89, 163, 74, 64, 178, 218, 185, 87, 88, 58, 99, 80, 44, 126, 40, 223, 35, 233, 62, 8, 27, 103, 166, 6},
	})

	if err != nil {
		log.Fatalf("error calling rpc_service.PostTransaction: %s", err)
	}

	log.Printf("response from remote rpc_service.PostTransaction(): %s", response)

}
