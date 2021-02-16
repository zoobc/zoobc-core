// ZooBC Copyright (C) 2020 Quasisoft Limited - Hong Kong
// This file is part of ZooBC <https://github.com/zoobc/zoobc-core>
//
// ZooBC is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ZooBC is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ZooBC.  If not, see <http://www.gnu.org/licenses/>.
//
// Additional Permission Under GNU GPL Version 3 section 7.
// As the special exception permitted under Section 7b, c and e,
// in respect with the Author’s copyright, please refer to this section:
//
// 1. You are free to convey this Program according to GNU GPL Version 3,
//     as long as you respect and comply with the Author’s copyright by
//     showing in its user interface an Appropriate Notice that the derivate
//     program and its source code are “powered by ZooBC”.
//     This is an acknowledgement for the copyright holder, ZooBC,
//     as the implementation of appreciation of the exclusive right of the
//     creator and to avoid any circumvention on the rights under trademark
//     law for use of some trade names, trademarks, or service marks.
//
// 2. Complying to the GNU GPL Version 3, you may distribute
//     the program without any permission from the Author.
//     However a prior notification to the authors will be appreciated.
//
// ZooBC is architected by Roberto Capodieci & Barton Johnston
//             contact us at roberto.capodieci[at]blockchainzoo.com
//             and barton.johnston[at]blockchainzoo.com
//
// Core developers that contributed to the current implementation of the
// software are:
//             Ahmad Ali Abdilah ahmad.abdilah[at]blockchainzoo.com
//             Allan Bintoro allan.bintoro[at]blockchainzoo.com
//             Andy Herman
//             Gede Sukra
//             Ketut Ariasa
//             Nawi Kartini nawi.kartini[at]blockchainzoo.com
//             Stefano Galassi stefano.galassi[at]blockchainzoo.com
//
// IMPORTANT: The above copyright notice and this permission notice
// shall be included in all copies or substantial portions of the Software.
package api

import (
	"fmt"
	"html/template"
	"net"
	"net/http"

	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/feedbacksystem"

	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/api/handler"
	"github.com/zoobc/zoobc-core/api/service"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/interceptor"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"github.com/zoobc/zoobc-core/common/query"
	rpcService "github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/storage"
	"github.com/zoobc/zoobc-core/common/transaction"
	coreService "github.com/zoobc/zoobc-core/core/service"
	coreUtil "github.com/zoobc/zoobc-core/core/util"
	"github.com/zoobc/zoobc-core/observer"
	"github.com/zoobc/zoobc-core/p2p"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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
	ownerAccountAddress []byte,
	nodefilePath string,
	logger *log.Logger,
	isDebugMode bool,
	apiCertFile, apiKeyFile string,
	maxAPIRequestPerSecond uint32,
	nodePublicKey []byte,
	feedbackStrategy feedbacksystem.FeedbackStrategyInterface,
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
			queryExecutor,
			crypto.NewSignature(),
			actionTypeSwitcher,
			mempoolService,
			observer.NewObserver(),
			transactionUtil,
			feedbackStrategy,
			logger,
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
			query.NewSkippedBlocksmithQuery(&chaintype.MainChain{}),
			queryExecutor,
		),
	})

	// Set GRPC handler for liquid transactions
	rpcService.RegisterLiquidPaymentServiceServer(grpcServer, &handler.LiquidTransactionHandler{
		Service: service.NewLiquidTransactionService(
			queryExecutor,
			query.NewLiquidPaymentTransactionQuery(),
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
	grpcPort, httpPort int,
	ownerAccountAddress []byte,
	nodefilePath string,
	logger *log.Logger,
	isDebugMode bool,
	apiCertFile, apiKeyFile string,
	maxAPIRequestPerSecond uint32,
	nodePublicKey []byte,
	feedbackStrategy feedbacksystem.FeedbackStrategyInterface,
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
		feedbackStrategy,
	)
}
