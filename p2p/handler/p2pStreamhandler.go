package handler

import (
	"io"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type P2PStreamHandler struct {
}

func NewP2PStreamHandler() *P2PStreamHandler {
	return &P2PStreamHandler{}
}

func (ss *P2PStreamHandler) SendStreamRequest(stream service.P2PStream_SendDataServer) error {
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			logrus.Println("no more data")
			break
		}
		if err != nil {
			return status.Errorf(codes.Unknown, "cannot receive stream request: %v", err)
		}
		go func() {
			time.Sleep(30 * time.Millisecond)
			err = stream.Send(&service.P2PStreamResponse{})
			if err != nil {
				logrus.Println(err)
			}
		}()
	}

	return nil
}
