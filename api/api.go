package api

import (
	"context"
	"fmt"
	"net"
	"net/http"

	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"

	"github.com/zoobc/zoobc-core/api/handler"
	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/interceptor"
	"github.com/zoobc/zoobc-core/common/kvdb"
	"github.com/zoobc/zoobc-core/common/query"
	rpcService "github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/transaction"
	coreService "github.com/zoobc/zoobc-core/core/service"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
	"github.com/zoobc/zoobc-core/observer"
	"github.com/zoobc/zoobc-core/p2p"
)

func startGrpcServer(
	port int,
	kvExecutor kvdb.KVExecutorInterface,
	queryExecutor query.ExecutorInterface,
	p2pHostService p2p.Peer2PeerServiceInterface,
	blockServices map[int32]coreService.BlockServiceInterface,
	nodeRegistrationService coreService.NodeRegistrationServiceInterface,
	ownerAccountAddress, nodefilePath string,
	logger *log.Logger,
	isDebugMode bool,
	apiCertFile, apiKeyFile string,
	transactionUtil transaction.UtilInterface,
	receiptUtil coreUtil.ReceiptUtilInterface,
	receiptService coreService.ReceiptServiceInterface,
	transactionCoreService coreService.TransactionCoreServiceInterface,
	maxAPIRequestPerSecond uint32,
	blockStateCache storage.CacheStorageInterface,
) {

	chainType := chaintype.GetChainType(0)

	// load/enable TLS over grpc
	creds, err := credentials.NewServerTLSFromFile(apiCertFile, apiKeyFile)
	if err != nil {
		logger.Infof("Failed to generate credentials %v. TLS encryption won't be enabled for grpc api", err)
	} else {
		logger.Info("TLS certificate loaded. rpc api will use encryption at transport level")
	}
	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(grpcMiddleware.ChainUnaryServer(
			interceptor.NewServerRateLimiterInterceptor(maxAPIRequestPerSecond),
			interceptor.NewServerInterceptor(
				logger,
				ownerAccountAddress,
				map[codes.Code]string{
					codes.Unavailable:     "indicates the destination service is currently unavailable",
					codes.InvalidArgument: "indicates the argument request is invalid",
					codes.Unauthenticated: "indicates the request is unauthenticated",
				},
			),
		),
		),
		grpc.StreamInterceptor(interceptor.NewStreamInterceptor(ownerAccountAddress)),
	)
	actionTypeSwitcher := &transaction.TypeSwitcher{
		Executor: queryExecutor,
	}
	mempoolService := coreService.NewMempoolService(
		transactionUtil,
		chainType,
		kvExecutor,
		queryExecutor,
		query.NewMempoolQuery(chainType),
		query.NewMerkleTreeQuery(),
		actionTypeSwitcher,
		query.NewAccountBalanceQuery(),
		query.NewBlockQuery(chainType),
		query.NewTransactionQuery(chainType),
		crypto.NewSignature(),
		observer.NewObserver(),
		logger,
		receiptUtil,
		receiptService,
		transactionCoreService,
	)
	serv, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Fatalf("failed to listen: %v\n", err)
		return
	}
	// *************************************
	// RPC Services Init
	// *************************************

	// Set GRPC handler for Block requests
	rpcService.RegisterBlockServiceServer(grpcServer, &handler.BlockHandler{
		Service: service.NewBlockService(queryExecutor, blockServices, isDebugMode),
	})
	// Set GRPC handler for Transactions requests
	rpcService.RegisterTransactionServiceServer(grpcServer, &handler.TransactionHandler{
		Service: service.NewTransactionService(
			queryExecutor, crypto.NewSignature(),
			actionTypeSwitcher,
			mempoolService,
			observer.NewObserver(),
			transactionUtil,
		),
	})
	// Set GRPC handler for Transactions requests
	rpcService.RegisterHostServiceServer(grpcServer, &handler.HostHandler{
		Service: service.NewHostService(queryExecutor, p2pHostService, blockServices, nodeRegistrationService, blockStateCache),
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
		Service: service.NewNodeAdminService(
			queryExecutor,
			blockServices[(&chaintype.MainChain{}).GetTypeInt()],
			ownerAccountAddress, nodefilePath),
	})

	// Set GRPC handler for unconfirmed
	rpcService.RegisterNodeHardwareServiceServer(grpcServer, &handler.NodeHardwareHandler{
		Service: service.NewNodeHardwareService(
			ownerAccountAddress,
			crypto.NewSignature(),
		),
	})
	// Set GRPC handler for node registry request
	rpcService.RegisterNodeRegistrationServiceServer(grpcServer, &handler.NodeRegistryHandler{
		Service: service.NewNodeRegistryService(queryExecutor),
	})

	// Set GRPC handler for account ledger request
	rpcService.RegisterAccountLedgerServiceServer(grpcServer, &handler.AccountLedgerHandler{
		Service: service.NewAccountLedgerService(queryExecutor),
	})

	// Set GRPC handler for escrow transaction request
	rpcService.RegisterEscrowTransactionServiceServer(grpcServer, &handler.EscrowTransactionHandler{
		Service: service.NewEscrowTransactionService(queryExecutor),
	})
	// Set GRPC handler for multisig information request
	rpcService.RegisterMultisigServiceServer(grpcServer, &handler.MultisigHandler{
		MultisigService: service.NewMultisigService(
			queryExecutor,
			blockServices[(&chaintype.MainChain{}).GetTypeInt()],
			query.NewPendingTransactionQuery(),
			query.NewPendingSignatureQuery(),
			query.NewMultisignatureInfoQuery(),
		)})
	// Set GRPC handler for health check
	rpcService.RegisterHealthCheckServiceServer(grpcServer, &handler.HealthCheckHandler{})

	// Set GRPC handler for account dataset
	rpcService.RegisterAccountDatasetServiceServer(grpcServer, &handler.AccountDatasetHandler{
		Service: service.NewAccountDatasetService(
			query.NewAccountDatasetsQuery(),
			queryExecutor,
		),
	})
	// run grpc-gateway handler
	go func() {
		if err := grpcServer.Serve(serv); err != nil {
			panic(err)
		}
	}()
	logger.Infof("GRPC listening to http:%d\n", port)
}

// Start starts api servers in the given port and passing query executor
func Start(
	grpcPort, restPort int,
	kvExecutor kvdb.KVExecutorInterface,
	queryExecutor query.ExecutorInterface,
	p2pHostService p2p.Peer2PeerServiceInterface,
	blockServices map[int32]coreService.BlockServiceInterface,
	nodeRegistrationService coreService.NodeRegistrationServiceInterface,
	ownerAccountAddress, nodefilePath string,
	logger *log.Logger,
	isDebugMode bool,
	apiCertFile, apiKeyFile string,
	transactionUtil transaction.UtilInterface,
	receiptUtil coreUtil.ReceiptUtilInterface,
	receiptService coreService.ReceiptServiceInterface,
	transactionCoreService coreService.TransactionCoreServiceInterface,
	maxAPIRequestPerSecond uint32,
	blockStateCache storage.CacheStorageInterface,
) {
	startGrpcServer(
		grpcPort,
		kvExecutor,
		queryExecutor,
		p2pHostService,
		blockServices,
		nodeRegistrationService,
		ownerAccountAddress,
		nodefilePath,
		logger,
		isDebugMode,
		apiCertFile,
		apiKeyFile,
		transactionUtil,
		receiptUtil,
		receiptService,
		transactionCoreService,
		maxAPIRequestPerSecond,
		blockStateCache,
	)
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
	_ = rpcService.RegisterAccountLedgerServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%d", rpcPort), opts)
	_ = rpcService.RegisterEscrowTransactionServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%d", rpcPort), opts)
	_ = rpcService.RegisterMultisigServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%d", rpcPort), opts)
	_ = rpcService.RegisterHealthCheckServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%d", rpcPort), opts)
	_ = rpcService.RegisterAccountDatasetServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%d", rpcPort), opts)
	return http.ListenAndServe(fmt.Sprintf(":%d", apiPort), mux)
}
