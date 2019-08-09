package native

import (
	"github.com/zoobc/zoobc-core/common/model"
	coreService "github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/observer"
	"github.com/zoobc/zoobc-core/p2p"
	"github.com/zoobc/zoobc-core/p2p/native/service"

	nativeUtil "github.com/zoobc/zoobc-core/p2p/native/util"
)

type Service struct {
	HostService   *service.HostService
	BlockServices map[int32]coreService.BlockServiceInterface
	service.PeerServiceClient
	Observer *observer.Observer
}

var hostServiceInstance *service.HostService

// InitService to initialize services of the native strategy
func (s *Service) InitService(myAddress string, port uint32, wellknownPeers []string, obsr *observer.Observer) (p2p.P2pServiceInterface, error) {
	if s.HostService == nil {
		knownPeersResult, err := nativeUtil.ParseKnownPeers(wellknownPeers)
		if err != nil {
			return nil, err
		}
		host := nativeUtil.NewHost(myAddress, port, knownPeersResult)
		hostServiceInstance = service.CreateHostService(host)
		s.HostService = hostServiceInstance
		s.Observer = obsr
	}
	return s, nil
}

func (s *Service) SetBlockServices(blockServices map[int32]coreService.BlockServiceInterface) {
	s.BlockServices = blockServices
}

// GetHostInstance returns the host model
func (s *Service) GetHostInstance() *model.Host {
	return s.HostService.Host
}

// DisconnectPeer returns the host model
func (s *Service) DisconnectPeer(peer *model.Peer) {
	s.HostService.DisconnectPeer(peer)
}

// GetAnyResolvedPeer Get any random resolved peer
func (s *Service) GetAnyResolvedPeer() *model.Peer {
	return s.HostService.GetAnyResolvedPeer()
}

// GetResolvedPeer Get resolved peers
func (s *Service) GetResolvedPeers() map[string]*model.Peer {
	return s.HostService.GetResolvedPeers()
}

// StartP2P to run all p2p Thread service
func (s *Service) StartP2P() {
	startServer(s.BlockServices, s.Observer)

	// p2p thread
	go resolvePeersThread()
	go getMorePeersThread()
	go updateBlacklistedStatus()
}

// SendBlockListener setup listener for send block to the list peer
func (s *Service) SendBlockListener() observer.Listener {
	return observer.Listener{
		OnNotify: func(block interface{}, args interface{}) {
			b := block.(*model.Block)
			sendBlock(b)
		},
	}
}

// SendTransactionListener setup listener for transaction to the list peer
func (s *Service) SendTransactionListener() observer.Listener {
	return observer.Listener{
		OnNotify: func(transactionBytes interface{}, args interface{}) {
			t := transactionBytes.([]byte)
			sendTransactionBytes(t)
		},
	}
}
