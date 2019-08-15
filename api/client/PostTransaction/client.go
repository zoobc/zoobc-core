package main

import (
	"context"

	log "github.com/sirupsen/logrus"
	rpc_model "github.com/zoobc/zoobc-core/common/model"
	rpc_service "github.com/zoobc/zoobc-core/common/service"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial(":3001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := rpc_service.NewTransactionServiceClient(conn)

	response, err := c.PostTransaction(context.Background(), &rpc_model.PostTransactionRequest{
		//  TransactionBytes: []byte{ // keep this to test multiple transaction in single block.
		//	1, 0, 1, 82, 108, 58, 93, 0, 0, 0, 0, 0, 0, 66, 67, 90, 110, 83, 102, 113, 112, 80, 53, 116, 113, 70, 81, 108, 77, 84, 89,
		//	107, 68, 101, 66, 86, 70, 87, 110, 98, 121, 86, 75, 55, 118, 76, 114, 53, 79, 82, 70, 112, 84, 106, 103, 116, 78, 0, 0, 66,
		//	67, 90, 75, 76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106, 116, 70, 57, 75, 111, 74, 115, 107, 106, 86, 80, 118, 66, 57,
		//	106, 112, 73, 106, 102, 122, 122, 73, 54, 122, 68, 87, 48, 74, 1, 0, 0, 0, 0, 0, 0, 0, 8, 0, 0, 0, 16, 39, 0, 0, 0, 0, 0, 0, 140,
		//	134, 217, 86, 232, 251, 81, 174, 86, 44, 221, 44, 226, 73, 245, 19, 170, 94, 47, 160, 53, 20, 225, 192, 19, 200, 196, 217, 96, 64,
		//	66, 6, 146, 16, 61, 104, 106, 112, 122, 96, 233, 224, 208, 119, 245, 148, 60, 9, 131, 211, 110, 68, 167, 115, 243, 251, 90, 64, 234,
		//	66, 108, 30, 116, 9,
		// },
		TransactionBytes: []byte{
			1, 0, 0, 0, 1, 189, 0, 77, 93, 0, 0, 0, 0, 44, 0, 0, 0, 66, 67, 90, 110, 83, 102, 113, 112, 80, 53, 116, 113, 70, 81, 108,
			77, 84, 89, 107, 68, 101, 66, 86, 70, 87, 110, 98, 121, 86, 75, 55, 118, 76, 114, 53, 79, 82, 70, 112, 84, 106, 103, 116,
			78, 44, 0, 0, 0, 66, 67, 90, 75, 76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106, 116, 70, 57, 75, 111, 74, 115, 107, 106,
			86, 80, 118, 66, 57, 106, 112, 73, 106, 102, 122, 122, 73, 54, 122, 68, 87, 48, 74, 1, 0, 0, 0, 0, 0, 0, 0, 8, 0, 0, 0, 16, 39,
			0, 0, 0, 0, 0, 0, 180, 143, 228, 156, 234, 214, 183, 43, 200, 112, 178, 166, 134, 156, 224, 252, 184, 87, 52, 253, 43, 41, 14, 33,
			164, 186, 47, 208, 46, 245, 86, 159, 153, 230, 238, 139, 175, 149, 30, 83, 185, 193, 20, 75, 208, 93, 146, 154, 84, 241, 156, 125,
			95, 254, 211, 62, 46, 67, 42, 88, 91, 241, 79, 0,
		},
		// add node
		// TransactionBytes: []byte{
		//	2, 0, 0, 0, 1, 77, 2, 85, 93, 0, 0, 0, 0, 44, 0, 0, 0, 66, 67, 90, 110, 83, 102, 113, 112, 80, 53, 116, 113, 70, 81, 108,
		//	77, 84, 89, 107, 68, 101, 66, 86, 70, 87, 110, 98, 121, 86, 75, 55, 118, 76, 114, 53, 79, 82, 70, 112, 84, 106, 103, 116,
		//	78, 44, 0, 0, 0, 66, 67, 90, 75, 76, 118, 103, 85, 89, 90, 49, 75, 75, 120, 45, 106, 116, 70, 57, 75, 111, 74, 115, 107,
		//	106, 86, 80, 118, 66, 57, 106, 112, 73, 106, 102, 122, 122, 73, 54, 122, 68, 87, 48, 74, 1, 0, 0, 0, 0, 0, 0, 0, 179, 0,
		//	0, 0, 0, 14, 6, 218, 170, 54, 60, 50, 2, 66, 130, 119, 226, 235, 126, 203, 5, 12, 152, 194, 170, 146, 43, 63, 224, 101,
		//	127, 241, 62, 152, 187, 255, 44, 0, 0, 0, 66, 67, 90, 110, 83, 102, 113, 112, 80, 53, 116, 113, 70, 81, 108, 77, 84, 89,
		//	107, 68, 101, 66, 86, 70, 87, 110, 98, 121, 86, 75, 55, 118, 76, 114, 53, 79, 82, 70, 112, 84, 106, 103, 116, 78, 9, 0, 0,
		//	0, 49, 50, 55, 46, 48, 46, 48, 46, 49, 160, 134, 1, 0, 0, 0, 0, 0, 72, 101, 108, 108, 111, 66, 108, 111, 99, 107, 0, 0, 0, 0,
		//	199, 197, 72, 124, 13, 17, 101, 202, 66, 255, 200, 80, 230, 26, 232, 97, 114, 251, 243, 2, 160, 39, 241, 127, 134, 183, 154,
		//	155, 151, 223, 41, 5, 43, 100, 106, 180, 105, 192, 249, 103, 238, 62, 74, 180, 114, 42, 210, 236, 201, 241, 198, 237, 233, 66, 198,
		//	203, 18, 201, 201, 61, 39, 167, 91, 10, 229, 148, 34, 207, 159, 48, 187, 20, 133, 227, 205, 56, 234, 23, 245, 89, 41, 246, 210, 194,
		//	132, 107, 152, 95, 198, 96, 238, 88, 8, 76, 24, 42, 139, 71, 249, 178, 252, 59, 30, 57, 74, 146, 163, 211, 36, 110, 221, 219, 218,
		//	57, 63, 79, 55, 216, 214, 139, 85, 125, 62, 129, 158, 16, 108, 3,
		// },
	})

	if err != nil {
		log.Fatalf("error calling rpc_service.PostTransaction: %s", err)
	}

	log.Printf("response from remote rpc_service.PostTransaction(): %s", response)

}
