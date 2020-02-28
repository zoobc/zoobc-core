package p2p

import (
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/interceptor"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/transaction"
	coreService "github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/observer"
	"github.com/zoobc/zoobc-core/p2p/client"
	"github.com/zoobc/zoobc-core/p2p/handler"
	p2pService "github.com/zoobc/zoobc-core/p2p/service"
	"github.com/zoobc/zoobc-core/p2p/strategy"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type (
	Peer2PeerServiceInterface interface {
		StartP2P(
			myAddress string,
			peerPort uint32,
			nodeSecretPhrase string,
			queryExecutor query.ExecutorInterface,
			blockServices map[int32]coreService.BlockServiceInterface,
			mempoolServices map[int32]coreService.MempoolServiceInterface,
			observer *observer.Observer,
		)
		// exposed api list
		GetHostInfo() *model.Host
		GetResolvedPeers() map[string]*model.Peer
		GetUnresolvedPeers() map[string]*model.Peer
		GetPriorityPeers() map[string]*model.Peer

		// event listener that relate to p2p communication
		SendBlockListener() observer.Listener
		SendTransactionListener() observer.Listener
		RequestBlockTransactionsListener() observer.Listener
		SendBlockTransactionsListener() observer.Listener
	}
	Peer2PeerService struct {
		Host              *model.Host
		PeerExplorer      strategy.PeerExplorerStrategyInterface
		PeerServiceClient client.PeerServiceClientInterface
		Logger            *log.Logger
		TransactionUtil   transaction.UtilInterface
	}
)

// NewP2PService to initialize peer to peer service wrapper
func NewP2PService(
	host *model.Host,
	peerServiceClient client.PeerServiceClientInterface,
	peerExplorer strategy.PeerExplorerStrategyInterface,
	logger *log.Logger,
	transactionUtil transaction.UtilInterface,
) (Peer2PeerServiceInterface, error) {
	return &Peer2PeerService{
		Host:              host,
		PeerServiceClient: peerServiceClient,
		PeerExplorer:      peerExplorer,
		TransactionUtil:   transactionUtil,
	}, nil
}

// StartP2P initiate all p2p dependencies and run all p2p thread service
func (s *Peer2PeerService) StartP2P(
	myAddress string,
	peerPort uint32,
	nodeSecretPhrase string,
	queryExecutor query.ExecutorInterface,
	blockServices map[int32]coreService.BlockServiceInterface,
	mempoolServices map[int32]coreService.MempoolServiceInterface,
	observer *observer.Observer,
) {
	// peer to peer service layer | under p2p handler
	p2pServerService := p2pService.NewP2PServerService(
		s.PeerExplorer,
		blockServices,
		mempoolServices,
		nodeSecretPhrase,
		observer,
	)
	// start listening on peer port
	go func() { // register handlers and listening to incoming p2p request
		var (
			ownerAddress = crypto.NewEd25519Signature().GetAddressFromSeed(nodeSecretPhrase)
			grpcServer   = grpc.NewServer(
				grpc.UnaryInterceptor(interceptor.NewServerInterceptor(
					s.Logger,
					ownerAddress,
					map[codes.Code]string{
						codes.Unavailable:     "indicates the destination service is currently unavailable",
						codes.InvalidArgument: "indicates the argument request is invalid",
						codes.Unauthenticated: "indicates the request is unauthenticated",
					},
				)),
			)
		)

		service.RegisterP2PCommunicationServer(grpcServer, handler.NewP2PServerHandler(
			p2pServerService,
		))
		if err := grpcServer.Serve(p2pUtil.ServerListener(int(s.Host.GetInfo().GetPort()))); err != nil {
			s.Logger.Fatal(err.Error())
		}
	}()
	go s.PeerExplorer.Start()
}

// GetHostInfo exposed the p2p host information to the client
func (s *Peer2PeerService) GetHostInfo() *model.Host {
	return s.Host
}

// GetResolvedPeers exposed current node resolved peer list
func (s *Peer2PeerService) GetResolvedPeers() map[string]*model.Peer {
	return s.PeerExplorer.GetResolvedPeers()
}

// GetUnresolvedPeers exposed current node unresolved peer list.
func (s *Peer2PeerService) GetUnresolvedPeers() map[string]*model.Peer {
	return s.PeerExplorer.GetUnresolvedPeers()
}

// GetPriorityPeers exposed current node priority peer list.
func (s *Peer2PeerService) GetPriorityPeers() map[string]*model.Peer {
	return s.PeerExplorer.GetPriorityPeers()
}

// SendBlockListener setup listener for send block to the list peer
func (s *Peer2PeerService) SendBlockListener() observer.Listener {
	return observer.Listener{
		OnNotify: func(block interface{}, args ...interface{}) {
			var (
				b         *model.Block
				chainType chaintype.ChainType
				ok        bool
			)
			b, ok = block.(*model.Block)
			if !ok {
				s.Logger.Fatalln("Block casting failures in SendBlockListener")
			}

			chainType, ok = args[0].(chaintype.ChainType)
			if !ok {
				s.Logger.Fatalln("chainType casting failures in SendBlockListener")
			}

			peers := s.PeerExplorer.GetResolvedPeers()
			for _, peer := range peers {
				go func(p *model.Peer) {
					_ = s.PeerServiceClient.SendBlock(p, b, chainType)
				}(peer)
			}
		},
	}
}

// SendTransactionListener setup listener for transaction to the list peer
func (s *Peer2PeerService) SendTransactionListener() observer.Listener {
	return observer.Listener{
		OnNotify: func(transactionBytes interface{}, args ...interface{}) {
			var (
				t         []byte
				chainType chaintype.ChainType
				ok        bool
			)
			t, ok = transactionBytes.([]byte)
			if !ok {
				s.Logger.Fatalln("transactionBytes casting failures in SendTransactionListener")
			}

			chainType, ok = args[0].(chaintype.ChainType)
			if !ok {
				s.Logger.Fatalln("chainType casting failures in SendTransactionListener")
			}
			peers := s.PeerExplorer.GetResolvedPeers()
			for _, peer := range peers {
				go func(p *model.Peer) {
					_ = s.PeerServiceClient.SendTransaction(p, t, chainType)

				}(peer)
			}
		},
	}
}

func (s *Peer2PeerService) RequestBlockTransactionsListener() observer.Listener {
	return observer.Listener{
		OnNotify: func(transactionIDs interface{}, args ...interface{}) {
			var (
				txIDs     = transactionIDs.([]int64)
				peer      *model.Peer
				chainType chaintype.ChainType
				blockID   int64
				ok        bool
			)

			// check number of arguments before casting the argument type
			if len(args) < 3 {
				s.Logger.Fatalln("number of needed arguments too few in RequestBlockTransactionsListener")
				return
			}

			blockID, ok = args[0].(int64)
			if !ok {
				s.Logger.Fatalln("blockID casting failures in RequestBlockTransactionsListener")
			}

			chainType, ok = args[1].(chaintype.ChainType)
			if !ok {
				s.Logger.Fatalln("chainType casting failures in RequestBlockTransactionsListener")
			}

			peer, ok = args[2].(*model.Peer)
			if !ok {
				s.Logger.Fatalln("peer casting failures in RequestBlockTransactionsListener")
			}

			go func(p *model.Peer) {
				_ = s.PeerServiceClient.RequestBlockTransactions(p, txIDs, chainType, blockID)
			}(peer)
		},
	}
}

func (s *Peer2PeerService) SendBlockTransactionsListener() observer.Listener {
	return observer.Listener{
		OnNotify: func(transactionsInterface interface{}, args ...interface{}) {
			var (
				txsBytes  [][]byte
				txs       []*model.Transaction
				chainType chaintype.ChainType
				peer      *model.Peer
				ok        bool
			)

			txs, ok = transactionsInterface.([]*model.Transaction)
			if !ok {
				s.Logger.Fatalln("Transaction casting failures in SendBlockTransactionsListener")
			}

			chainType, ok = args[0].(chaintype.ChainType)
			if !ok {
				s.Logger.Fatalln("chainType casting failures in SendBlockTransactionsListener")
			}

			peer, ok = args[1].(*model.Peer)
			if !ok {
				s.Logger.Fatalln("Peer casting failures in SendBlockTransactionsListener")
			}

			for _, tx := range txs {
				txByte, err := s.TransactionUtil.GetTransactionBytes(tx, true)
				if err != nil {
					continue
				}
				txsBytes = append(txsBytes, txByte)
			}
			go func(p *model.Peer) {
				_ = s.PeerServiceClient.SendBlockTransactions(p, txsBytes, chainType)
			}(peer)
		},
	}
}
