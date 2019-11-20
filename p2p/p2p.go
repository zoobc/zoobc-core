package p2p

import (
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/interceptor"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/query"
	"github.com/zoobc/zoobc-core/common/service"
	"github.com/zoobc/zoobc-core/common/util"
	coreService "github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/observer"
	"github.com/zoobc/zoobc-core/p2p/client"
	"github.com/zoobc/zoobc-core/p2p/handler"
	p2pService "github.com/zoobc/zoobc-core/p2p/service"
	"github.com/zoobc/zoobc-core/p2p/strategy"
	p2pUtil "github.com/zoobc/zoobc-core/p2p/util"
	"google.golang.org/grpc"
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
		)
		// exposed api list
		GetHostInfo() *model.Host
		GetResolvedPeers() map[string]*model.Peer
		GetUnresolvedPeers() map[string]*model.Peer
		GetPriorityPeers() map[string]*model.Peer

		// event listener that relate to p2p communication
		SendBlockListener() observer.Listener
		SendTransactionListener() observer.Listener
	}
	Peer2PeerService struct {
		Host              *model.Host
		PeerExplorer      strategy.PeerExplorerStrategyInterface
		PeerServiceClient client.PeerServiceClientInterface
		Logger            *log.Logger
	}
)

// InitService to initialize peer to peer service wrapper
func NewP2PService(
	host *model.Host,
	peerServiceClient client.PeerServiceClientInterface,
	peerExplorer strategy.PeerExplorerStrategyInterface,
	logger *log.Logger,
) (Peer2PeerServiceInterface, error) {
	return &Peer2PeerService{
		Host:              host,
		PeerServiceClient: peerServiceClient,
		PeerExplorer:      peerExplorer,
		Logger:            logger,
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
) {
	// peer to peer service layer | under p2p handler
	p2pServerService := p2pService.NewP2PServerService(
		s.PeerExplorer,
		blockServices,
		mempoolServices,
		nodeSecretPhrase,
	)
	// start listening on peer port
	go func() { // register handlers and listening to incoming p2p request
		ownerAddress := util.GetAddressFromSeed(nodeSecretPhrase)
		grpcServer := grpc.NewServer(
			grpc.UnaryInterceptor(interceptor.NewServerInterceptor(s.Logger, ownerAddress)),
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
		OnNotify: func(block interface{}, args interface{}) {
			b := block.(*model.Block)
			peers := s.PeerExplorer.GetResolvedPeers()
			chainType := args.(chaintype.ChainType)
			for _, peer := range peers {
				p := peer
				go func() {
					_ = s.PeerServiceClient.SendBlock(p, b, chainType)
				}()
			}
		},
	}
}

// SendTransactionListener setup listener for transaction to the list peer
func (s *Peer2PeerService) SendTransactionListener() observer.Listener {
	return observer.Listener{
		OnNotify: func(transactionBytes interface{}, args interface{}) {
			t := transactionBytes.([]byte)
			chainType := args.(chaintype.ChainType)
			peers := s.PeerExplorer.GetResolvedPeers()
			for _, peer := range peers {
				p := peer
				go func() {
					_ = s.PeerServiceClient.SendTransaction(p, t, chainType)

				}()
			}
		},
	}
}
