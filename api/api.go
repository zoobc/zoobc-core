package api

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/api/controller"
	"github.com/zoobc/zoobc-core/api/internal"
	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/query"
	rpc_service "github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"
	"google.golang.org/grpc"
)

var (
	echoEndpoint = flag.String("echo_endpoint", "localhost:8000", "endpoint of YourService")
	apiLogger    *log.Logger
)

func init() {
	var err error
	if apiLogger, err = util.InitLogger(".log/", "debug.log"); err != nil {
		panic(err)
	}
}

func startRestServer(port int) {
	var err error
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err = rpc_service.RegisterBlockServiceHandlerFromEndpoint(ctx, mux, *echoEndpoint, opts)

	go func() {
		err = http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
		if err != nil {
			panic(fmt.Sprintf("Rest Api failure: %v\n", err))
		}
	}()

	apiLogger.Info(fmt.Sprintf("Rest server listening to http:%d\n", port))
}
func startGrpcServer(port int, queryExecutor *query.Executor) {
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(internal.NewInterceptor(apiLogger)),
	)

	serv, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		apiLogger.Fatalf("failed to listen: %v\n", err)
		return
	}

	// Set GRPC handler for Block requests
	rpc_service.RegisterBlockServiceServer(grpcServer, &controller.BlockController{
		Service: service.NewBlockService(queryExecutor),
	})

	// run grpc-gateway controller
	go func() {
		if err := grpcServer.Serve(serv); err != nil {
			panic(err)
		}
	}()
	apiLogger.Infof("GRPC listening to http:%d\n", port)
}

// Start starts api servers in the given port and passing query executor
func Start(grpcPort int, restPort int, queryExecutor *query.Executor) {
	startGrpcServer(grpcPort, queryExecutor)

	// Rest server is disbled for now
	// startRestServer(restPort)
}
