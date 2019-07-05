package api

import (
	"context"
	"flag"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/api/controller"
	"github.com/zoobc/zoobc-core/api/internal"
	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/query"
	rpc_service "github.com/zoobc/zoobc-core/common/schema/service"
	"github.com/zoobc/zoobc-core/common/util"
	"google.golang.org/grpc"
)

var (
	echoEndpoint = flag.String("echo_endpoint", "localhost:8000", "endpoint of YourService")

	apiLogger *log.Logger
)

func run() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	_ = rpc_service.RegisterAccountBalancesServiceHandlerFromEndpoint(ctx, mux, *echoEndpoint, opts)

	apiLogger.Info("listening http:8080")

	_ = http.ListenAndServe(":8080", mux)
}

func init() {
	var err error
	if apiLogger, err = util.InitLogger(".log/", "debug.log"); err != nil {
		panic(err)
	}
}

func Start(queryExecutor *query.Executor) {
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(internal.NewInterceptor(apiLogger)),
	)

	serv, err := net.Listen("tcp", ":8000")
	if err != nil {
		apiLogger.Fatalf("failed to listen: %v\n", err)
		return
	}

	// Set GRPC handler for AccountBalance requests
	rpc_service.RegisterAccountBalancesServiceServer(grpcServer, &controller.AccountBalanceController{
		// inject accountBalanceService instance to AccountBalanceController
		Service: service.NewAccountBalanceService(),
	})

	// Set GRPC handler for Block requests
	rpc_service.RegisterBlockServiceServer(grpcServer, &controller.BlockController{
		Service: service.NewBlockService(queryExecutor),
	})

	// run grpc-gateway controller
	go run()
	if err := grpcServer.Serve(serv); err != nil {
		panic(err)
	}
}
