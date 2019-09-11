package api

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/zoobc/zoobc-core/common/interceptor"
	"github.com/zoobc/zoobc-core/observer"

	"github.com/spf13/viper"

	"github.com/zoobc/zoobc-core/common/chaintype"
	coreService "github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/p2p"

	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/transaction"

	"github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/api/handler"
	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/query"
	rpcService "github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"
	"google.golang.org/grpc"
)

var (
	apiLogger *logrus.Logger
)

func init() {
	var (
		err       error
		logLevels []string
	)
	logLevels = viper.GetStringSlice("logLevels")
	if apiLogger, err = util.InitLogger(".log/", "debug.log", logLevels); err != nil {
		panic(err)
	}
}

func startGrpcServer(port int, queryExecutor query.ExecutorInterface, p2pHostService p2p.Peer2PeerServiceInterface,
	blockServices map[int32]coreService.BlockServiceInterface, ownerAccountAddress, nodefilePath string) {
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.NewServerInterceptor(apiLogger, ownerAccountAddress)),
		grpc.StreamInterceptor(interceptor.NewStreamInterceptor(ownerAccountAddress)),
	)
	actionTypeSwitcher := &transaction.TypeSwitcher{
		Executor: queryExecutor,
	}
	mempoolService := coreService.NewMempoolService(
		&chaintype.MainChain{},
		queryExecutor,
		query.NewMempoolQuery(&chaintype.MainChain{}),
		actionTypeSwitcher,
		query.NewAccountBalanceQuery(),
		crypto.NewSignature(),
		query.NewTransactionQuery(&chaintype.MainChain{}),
		observer.NewObserver(),
	)
	serv, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		apiLogger.Fatalf("failed to listen: %v\n", err)
		return
	}
	// *************************************
	// RPC Services Init
	// *************************************

	// Set GRPC handler for Block requests
	rpcService.RegisterBlockServiceServer(grpcServer, &handler.BlockHandler{
		Service: service.NewBlockService(queryExecutor),
	})
	// Set GRPC handler for Transactions requests
	rpcService.RegisterTransactionServiceServer(grpcServer, &handler.TransactionHandler{
		Service: service.NewTransactionService(
			queryExecutor,
			crypto.NewSignature(),
			actionTypeSwitcher,
			mempoolService,
			observer.NewObserver(),
		),
	})
	// Set GRPC handler for Transactions requests
	rpcService.RegisterHostServiceServer(grpcServer, &handler.HostHandler{
		Service: service.NewHostService(queryExecutor, p2pHostService, blockServices),
	})
	// Set GRPC handler for account balance requests
	rpcService.RegisterAccountBalanceServiceServer(grpcServer, &handler.AccountBalanceHandler{
		Service: service.NewAccountBalanceService(queryExecutor, query.NewAccountBalanceQuery()),
	})
	// Set GRPC handler for mempool requests
	rpcService.RegisterMempoolServiceServer(grpcServer, &handler.MempoolTransactionHandler{
		Service: service.NewMempoolTransactionsService(queryExecutor),
	})
	// Set GRPC handler for node admin requests
	rpcService.RegisterNodeAdminServiceServer(grpcServer, &handler.NodeAdminHandler{
		Service: service.NewNodeAdminService(queryExecutor, ownerAccountAddress, nodefilePath),
	})
	// Set GRPC handler for unconfirmed
	rpcService.RegisterNodeHardwareServiceServer(grpcServer, &handler.NodeHardwareHandler{
		Service: service.NewNodeHardwareService(
			ownerAccountAddress,
			crypto.NewSignature(),
		),
	})
	// run grpc-gateway handler
	go func() {
		if err := grpcServer.Serve(serv); err != nil {
			panic(err)
		}
	}()
	apiLogger.Infof("GRPC listening to http:%d\n", port)
}

// Start starts api servers in the given port and passing query executor
func Start(grpcPort, restPort int, queryExecutor query.ExecutorInterface, p2pHostService p2p.Peer2PeerServiceInterface,
	blockServices map[int32]coreService.BlockServiceInterface, ownerAccountAddress, nodefilePath string) {
	startGrpcServer(grpcPort, queryExecutor, p2pHostService, blockServices, ownerAccountAddress, nodefilePath)
	if restPort > 0 { // only start proxy service if apiHTTPPort set with value > 0
		go func() {
			err := runProxy(restPort, grpcPort)
			if err != nil {
				panic(err)
			}
		}()
	}
}

/**
runProxy only ran when `Debug` flag is set to `true` in `config.toml`
this function open a http endpoint that will be proxy to our rpc service
*/
func runProxy(apiPort, rpcPort int) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	_ = rpcService.RegisterAccountBalanceServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%d", rpcPort), opts)
	_ = rpcService.RegisterBlockServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%d", rpcPort), opts)
	_ = rpcService.RegisterHostServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%d", rpcPort), opts)
	_ = rpcService.RegisterMempoolServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%d", rpcPort), opts)
	_ = rpcService.RegisterNodeHardwareServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%d", rpcPort), opts)
	_ = rpcService.RegisterNodeRegistrationServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%d", rpcPort), opts)
	_ = rpcService.RegisterNodeAdminServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%d", rpcPort), opts)
	_ = rpcService.RegisterTransactionServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%d", rpcPort), opts)

	return http.ListenAndServe(fmt.Sprintf(":%d", apiPort), mux)
}
