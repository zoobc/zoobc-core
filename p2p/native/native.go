package native

import (
	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/observer"
	"github.com/zoobc/zoobc-core/p2p/native/service"

	nativeUtil "github.com/zoobc/zoobc-core/p2p/native/util"
)

type Service struct {
	HostService *service.HostService
}

var hostServiceInstance *service.HostService

// InitService to initialize services of the native strategy
func (s *Service) InitService(myAddress string, port uint32, wellknownPeers []string) (contract.P2PType, error) {
	if s.HostService == nil {
		knownPeersResult, err := nativeUtil.ParseKnownPeers(wellknownPeers)
		if err != nil {
			return nil, err
		}
		host := nativeUtil.NewHost(myAddress, port, knownPeersResult)
		hostServiceInstance = service.CreateHostService(host)
		s.HostService = hostServiceInstance
	}
	return s, nil
}

// GetHostInstance returns the host model
func (s *Service) GetHostInstance() *model.Host {
	return s.HostService.Host
}

// StartP2P to run all p2p Thread service
func (s *Service) StartP2P() {
	startServer()

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
		OnNotify: func(block interface{}, args interface{}) {
			t := block.(*model.Transaction)
			sendTransaction(t)
		},
	}
}
