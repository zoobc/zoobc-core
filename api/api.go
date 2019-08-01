package api

import (
	"fmt"
	"net"

	"github.com/zoobc/zoobc-core/common/chaintype"
	core_service "github.com/zoobc/zoobc-core/core/service"

	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/transaction"

	"github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/api/handler"
	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/query"
	rpc_service "github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"
	"github.com/zoobc/zoobc-core/observer"
	"google.golang.org/grpc"
)

var (
	apiLogger *logrus.Logger
)

func init() {
	var err error
	if apiLogger, err = util.InitLogger(".log/", "debug.log"); err != nil {
		panic(err)
	}
}

func startGrpcServer(port int, queryExecutor query.ExecutorInterface, p2pHostService contract.P2PType) {
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(util.NewServerInterceptor(apiLogger)),
	)
	actionTypeSwitcher := &transaction.TypeSwitcher{
		Executor: queryExecutor,
	}
	mempoolService := core_service.NewMempoolService(
		&chaintype.MainChain{},
		queryExecutor,
		query.NewMempoolQuery(&chaintype.MainChain{}),
		actionTypeSwitcher,
		query.NewAccountBalanceQuery(),
		observer.NewObserver())

	serv, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		apiLogger.Fatalf("failed to listen: %v\n", err)
		return
	}

	// Set GRPC handler for Block requests
	rpc_service.RegisterBlockServiceServer(grpcServer, &handler.BlockHandler{
		Service: service.NewBlockService(queryExecutor),
	})

	// Set GRPC handler for Transactions requests
	rpc_service.RegisterTransactionServiceServer(grpcServer, &handler.TransactionHandler{
		Service: service.NewTransactionService(
			queryExecutor,
			&crypto.Signature{},
			actionTypeSwitcher,
			mempoolService,
			apiLogger,
		),
	})

	// Set GRPC handler for Transactions requests
	rpc_service.RegisterHostServiceServer(grpcServer, &handler.HostHandler{
		Service:        service.NewHostService(queryExecutor),
		P2pHostService: p2pHostService,
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
func Start(grpcPort, restPort int, queryExecutor query.ExecutorInterface, p2pHostService contract.P2PType) {
	startGrpcServer(grpcPort, queryExecutor, p2pHostService)
}
