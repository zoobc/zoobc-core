package main

import (
	"context"
	"fmt"
	"os"
	"strings"

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
		configPath = "../../../resource"
	}
	if err := util.LoadConfig(configPath, "config", "toml", false); err != nil {
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

		// Escrow
		TransactionBytes: []byte{1, 0, 0, 0, 1, 114, 186, 68, 94, 0, 0, 0, 0, 44, 0, 0, 0, 72, 108, 90, 76, 104, 51, 86, 99, 110, 78, 108,
			118, 66, 121, 87, 111, 65, 122, 88, 79, 81, 50, 106, 65, 108, 119, 70, 79, 105, 121, 79, 57, 95, 110, 106, 73, 51, 111,
			113, 53, 89, 103, 104, 97, 44, 0, 0, 0, 110, 75, 95, 111, 117, 120, 100, 68, 68, 119, 117, 74, 105, 111, 103, 105, 68, 65,
			105, 95, 122, 115, 49, 76, 113, 101, 78, 55, 102, 53, 90, 115, 88, 98, 70, 116, 88, 71, 113, 71, 99, 48, 80, 100, 1, 0, 0,
			0, 0, 0, 0, 0, 8, 0, 0, 0, 87, 4, 0, 0, 0, 0, 0, 0, 44, 0, 0, 0, 66, 67, 90, 69, 71, 79, 98, 51, 87, 78, 120, 51, 102, 68,
			79, 86, 102, 57, 90, 83, 52, 69, 106, 118, 79, 73, 118, 95, 85, 101, 87, 52, 84, 86, 66, 81, 74, 95, 54, 116, 72, 75, 108,
			69, 111, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 23, 0, 0, 0, 97, 109, 111, 117, 110, 116, 32, 109, 117, 115, 116, 32,
			109, 111, 114, 101, 32, 116, 104, 97, 110, 32, 49, 0, 0, 0, 0, 178, 66, 128, 131, 168, 110, 230, 13, 235, 177, 8, 220, 123,
			148, 159, 170, 237, 219, 168, 84, 207, 106, 112, 50, 2, 39, 139, 246, 51, 100, 142, 98, 198, 14, 196, 147, 248, 167, 20,
			150, 114, 204, 47, 56, 215, 165, 36, 81, 178, 159, 224, 190, 147, 117, 103, 120, 246, 25, 24, 79, 142, 209, 39, 2},
		// Approval
		// TransactionBytes: []byte{4, 0, 0, 0, 1, 162, 32, 65, 94, 0, 0, 0, 0, 44, 0, 0, 0, 66, 67, 90, 69, 71, 79, 98, 51, 87, 78, 120, 51, 102,
		// 	68, 79, 86, 102, 57, 90, 83, 52, 69, 106, 118, 79, 73, 118, 95, 85, 101, 87, 52, 84, 86, 66, 81, 74, 95, 54, 116, 72, 75, 108,
		// 	69, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 12, 0, 0, 0, 0, 0, 0, 0, 170, 224, 210, 110, 59, 66, 173, 125, 0, 0, 0, 0, 0, 0, 0, 0,
		// 	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 225, 156, 122, 176, 20, 82, 143, 126, 10, 21, 142, 236, 89, 163, 58,
		// 	213, 187, 21, 26, 50, 231, 200, 20, 175, 42, 2, 195, 141, 194, 171, 190, 211, 52, 238, 10, 68, 86, 247, 27, 135, 130, 123, 238,
		// 	139, 49, 149, 25, 34, 164, 61, 123, 219, 124, 6, 178, 118, 157, 76, 227, 28, 14, 162, 60, 5},
	})

	if err != nil {
		log.Fatalf("error calling rpc_service.PostTransaction: %s", err)
	}

	log.Printf("response from remote rpc_service.PostTransaction(): %s", response)

}
