package p2p

import (
	"github.com/zoobc/zoobc-core/common/chaintype"
	"github.com/zoobc/zoobc-core/common/constant"
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
	"os"
	"os/signal"
	"syscall"
	"time"
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

		// event listener that relate to p2p communication
		SendBlockListener() observer.Listener
		SendTransactionListener() observer.Listener
	}
	Peer2PeerService struct {
		Host              *model.Host
		PeerExplorer      strategy.PeerExplorerStrategyInterface
		PeerServiceClient client.PeerServiceClientInterface
	}
)

// InitService to initialize peer to peer service wrapper
func NewP2PService(
	host *model.Host,
	peerServiceClient client.PeerServiceClientInterface,
	peerExplorer strategy.PeerExplorerStrategyInterface,
) (Peer2PeerServiceInterface, error) {
	return &Peer2PeerService{
		Host:              host,
		PeerServiceClient: peerServiceClient,
		PeerExplorer:      peerExplorer,
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
	// initialize log
	p2pLogger, err := util.InitLogger(".log/", "debug.log")
	if err != nil {
		panic(err)
	}

	// peer to peer service layer | under p2p handler
	p2pServerService := p2pService.NewP2PServerService(
		s.PeerExplorer,
		blockServices,
		mempoolServices,
		nodeSecretPhrase,
	)
	// start listening on peer port
	go func() { // register handlers and listening to incoming p2p request
		grpcServer := grpc.NewServer(
			grpc.UnaryInterceptor(interceptor.NewServerInterceptor(p2pLogger)),
		)
		service.RegisterP2PCommunicationServer(grpcServer, &handler.P2PServerHandler{
			Service: p2pServerService,
		})
		_ = grpcServer.Serve(p2pUtil.ServerListener(int(s.Host.GetInfo().GetPort())))
	}()
	// start p2p process threads
	go s.resolvePeersThread()
	go s.getMorePeersThread()
	go s.updateBlacklistedStatus()
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

// resolvePeersThread to periodically try get response from peers in UnresolvedPeer list
func (s *Peer2PeerService) resolvePeersThread() {
	go s.PeerExplorer.ResolvePeers()
	ticker := time.NewTicker(time.Duration(constant.ResolvePeersGap) * time.Second)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-ticker.C:
			go s.PeerExplorer.ResolvePeers()
			go s.PeerExplorer.UpdateResolvedPeers()
		case <-sigs:
			ticker.Stop()
			return
		}
	}
}

// getMorePeersThread to periodically request more peers from another node in Peers list
func (s *Peer2PeerService) getMorePeersThread() {
	go func() {
		peer, err := s.PeerExplorer.GetMorePeersHandler()
		if err != nil {
			return
		}
		var myPeers []*model.Node
		myResolvedPeers := s.PeerExplorer.GetResolvedPeers()
		for _, peer := range myResolvedPeers {
			myPeers = append(myPeers, peer.Info)
		}
		if peer == nil {
			return
		}
		myPeers = append(myPeers, s.Host.GetInfo())
		_, _ = s.PeerServiceClient.SendPeers(
			peer,
			myPeers,
		)
	}()
	ticker := time.NewTicker(time.Duration(constant.ResolvePeersGap) * time.Second)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-ticker.C:
			go func() {
				_, _ = s.PeerExplorer.GetMorePeersHandler()
			}()
		case <-sigs:
			ticker.Stop()
			return
		}
	}
}

// updateBlacklistedStatus to periodically check blacklisting time of black listed peer,
// every 60sec if there are blacklisted peers to unblacklist
func (s *Peer2PeerService) updateBlacklistedStatus() {
	ticker := time.NewTicker(time.Duration(60) * time.Second)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for {
			select {
			case <-ticker.C:
				curTime := uint64(time.Now().Unix())
				for _, p := range s.Host.GetBlacklistedPeers() {
					if p.GetBlacklistingTime() > 0 &&
						p.GetBlacklistingTime()+constant.BlacklistingPeriod <= curTime {
						s.Host.KnownPeers[p2pUtil.GetFullAddressPeer(p)] = s.PeerExplorer.PeerUnblacklist(p)
					}
				}
				break
			case <-sigs:
				ticker.Stop()
				return
			}
		}
	}()
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
					_, _ = s.PeerServiceClient.SendBlock(p, b, chainType)
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
					_, _ = s.PeerServiceClient.SendTransaction(p, t, chainType)
				}()
			}
		},
	}
}
