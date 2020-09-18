package api

import (
	"fmt"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"html/template"
	"net"
	"net/http"

	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/zoobc/zoobc-core/api/handler"
	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/interceptor"
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
	queryExecutor query.ExecutorInterface,
	p2pHostService p2p.Peer2PeerServiceInterface,
	blockServices map[int32]coreService.BlockServiceInterface,
	nodeRegistrationService coreService.NodeRegistrationServiceInterface,
	nodeAddressInfoService coreService.NodeAddressInfoServiceInterface,
	mempoolService coreService.MempoolServiceInterface,
	scrambleNodeService coreService.ScrambleNodeServiceInterface,
	transactionUtil transaction.UtilInterface,
	actionTypeSwitcher transaction.TypeActionSwitcher,
	blockStateStorages map[int32]storage.CacheStorageInterface,
	rpcPort, httpPort int,
	ownerAccountAddress, nodefilePath string,
	logger *log.Logger,
	isDebugMode bool,
	apiCertFile, apiKeyFile string,
	maxAPIRequestPerSecond uint32,
	nodePublicKey []byte,
) {
	grpcServer := grpc.NewServer(
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
	serv, err := net.Listen("tcp", fmt.Sprintf(":%d", rpcPort))
	if err != nil {
		logger.Fatalf("failed to listen: %v\n", err)
		return
	}
	participationScoreService := coreService.NewParticipationScoreService(query.NewParticipationScoreQuery(), queryExecutor)

	publishedReceiptUtil := coreUtil.NewPublishedReceiptUtil(query.NewPublishedReceiptQuery(), queryExecutor)
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
		Service: service.NewHostService(queryExecutor, p2pHostService, blockServices, nodeRegistrationService, scrambleNodeService, blockStateStorages),
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
	// Set GRPC handler for node address info requests
	rpcService.RegisterNodeAddressInfoServiceServer(grpcServer, &handler.NodeAddressInfoHandler{
		Service: service.NewNodeAddressInfoAPIService(
			nodeAddressInfoService,
		),
	})
	// Set GRPC handler for node registry request
	rpcService.RegisterNodeRegistrationServiceServer(grpcServer, &handler.NodeRegistryHandler{
		Service:       service.NewNodeRegistryService(queryExecutor),
		NodePublicKey: nodePublicKey,
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
			query.NewMultiSignatureParticipantQuery(),
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
	rpcService.RegisterParticipationScoreServiceServer(grpcServer, &handler.ParticipationScoreHandler{
		Service: service.NewParticipationScoreService(participationScoreService),
	})

	// Set GRPC handler for published receipt
	rpcService.RegisterPublishedReceiptServiceServer(grpcServer, &handler.PublishedReceiptHandler{
		Service: service.NewPublishedReceiptService(publishedReceiptUtil),
	})

	// Set GRPC handler for skipped block smith
	rpcService.RegisterSkippedBlockSmithsServiceServer(grpcServer, &handler.SkippedBlockSmithHandler{
		Service: service.NewSkippedBlockSmithService(
			query.NewSkippedBlocksmithQuery(),
			queryExecutor,
		),
	})
	go func() {
		// serve rpc
		if err := grpcServer.Serve(serv); err != nil {
			panic(err)
		}
	}()
	go func() {
		var tmp = template.Must(template.New("nodeStatus").Parse(constant.NodeStatusHTMLTemplate))
		// serve webrpc
		wrappedServer := grpcweb.WrapServer(
			grpcServer,
			grpcweb.WithCorsForRegisteredEndpointsOnly(true),
			grpcweb.WithOriginFunc(func(origin string) bool {
				return true // origin: '*'
			}))

		httpServer := &http.Server{
			Addr: fmt.Sprintf(":%d", httpPort),
			Handler: h2c.NewHandler(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Access-Control-Allow-Origin", "*")
					w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
					w.Header().Set("Access-Control-Allow-Headers",
						"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, "+
							"Authorization, X-User-Agent, X-Grpc-Web")
					if r.URL.String() == "/status" && r.Method == "GET" {
						data := monitoring.GetNodeStatus()
						_ = tmp.ExecuteTemplate(w, "nodeStatus", data)
					}
					if wrappedServer.IsGrpcWebRequest(r) || wrappedServer.IsAcceptableGrpcCorsRequest(r) {
						wrappedServer.ServeHTTP(w, r)
					}
				}),
				&http2.Server{},
			),
		}
		if apiKeyFile == "" || apiCertFile == "" {
			// no certificate provided, run on http
			if err := httpServer.ListenAndServe(); err != nil {
				panic(err)
			}
		} else {
			if err := httpServer.ListenAndServeTLS(apiCertFile, apiKeyFile); err != nil {
				// invalid or not found certificate, falling back to http
				if err := httpServer.ListenAndServe(); err != nil {
					panic(err)
				}
			}
		}

	}()
	logger.Infof("Client API Served on [rpc] http:%d\t [browser] http:%d", rpcPort, httpPort)
}

// Start starts api servers in the given port and passing query executor
func Start(
	queryExecutor query.ExecutorInterface,
	p2pHostService p2p.Peer2PeerServiceInterface,
	blockServices map[int32]coreService.BlockServiceInterface,
	nodeRegistrationService coreService.NodeRegistrationServiceInterface,
	nodeAddressInfoService coreService.NodeAddressInfoServiceInterface,
	mempoolService coreService.MempoolServiceInterface,
	scrambleNodeService coreService.ScrambleNodeServiceInterface,
	transactionUtil transaction.UtilInterface,
	actionTypeSwitcher transaction.TypeActionSwitcher,
	blockStateStorages map[int32]storage.CacheStorageInterface,
	grpcPort, httpPort int, ownerAccountAddress,
	nodefilePath string,
	logger *log.Logger,
	isDebugMode bool,
	apiCertFile, apiKeyFile string,
	maxAPIRequestPerSecond uint32,
	nodePublicKey []byte,
) {
	startGrpcServer(
		queryExecutor,
		p2pHostService,
		blockServices,
		nodeRegistrationService,
		nodeAddressInfoService,
		mempoolService,
		scrambleNodeService,
		transactionUtil,
		actionTypeSwitcher,
		blockStateStorages,
		grpcPort, httpPort,
		ownerAccountAddress,
		nodefilePath,
		logger,
		isDebugMode,
		apiCertFile, apiKeyFile,
		maxAPIRequestPerSecond,
		nodePublicKey,
	)
}
