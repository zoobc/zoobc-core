package native

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zoobc/zoobc-core/common/constant"
	"github.com/zoobc/zoobc-core/common/model"
	coreService "github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/p2p"
	"github.com/zoobc/zoobc-core/p2p/native/service"

	nativeUtil "github.com/zoobc/zoobc-core/p2p/native/util"
)

type Service struct {
	HostService   *service.HostService
	BlockServices map[int32]coreService.BlockServiceInterface
	service.PeerServiceClient
}

var hostServiceInstance *service.HostService

// InitService to initialize services of the native strategy
func (s *Service) InitService(myAddress string, port uint32, wellknownPeers []string) (p2p.P2pServiceInterface, error) {
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
	s.startServer()

	// p2p thread
	go resolvePeersThread()
	go getMorePeersThread()
	go updateBlacklistedStatus()
}

// startServer to run p2p service as server
func (s *Service) startServer() {
	port := hostServiceInstance.Host.GetInfo().GetPort()
	listener := nativeUtil.ServerListener(int(port))
	go func() {
		_ = service.NewServerService(s.BlockServices).StartListening(listener)
	}()
}

// ResolvePeersThread to periodically try get response from peers in UnresolvedPeer list
func resolvePeersThread() {
	go hostServiceInstance.ResolvePeers()
	ticker := nativeUtil.GetTickerTime(constant.ResolvePeersGap)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-ticker.C:
			go hostServiceInstance.ResolvePeers()
			go hostServiceInstance.UpdateResolvedPeers()
		case <-sigs:
			ticker.Stop()
			return
		}
	}
}

// getMorePeersThread to periodically request more peers from another node in Peers list
func getMorePeersThread() {
	go hostServiceInstance.GetMorePeersHandler()
	ticker := nativeUtil.GetTickerTime(constant.ResolvePeersGap)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-ticker.C:
			go hostServiceInstance.GetMorePeersHandler()
		case <-sigs:
			ticker.Stop()
			return
		}
	}
}

// updateBlacklistedStatus to periodically check blacklisting time of black listed peer,
// every 60sec if there are blacklisted peers to unblacklist
func updateBlacklistedStatus() {
	ticker := nativeUtil.GetTickerTime(60)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for {
			select {
			case <-ticker.C:
				curTime := uint64(time.Now().Unix())
				for _, p := range hostServiceInstance.Host.GetBlacklistedPeers() {
					if p.GetBlacklistingTime() > 0 &&
						p.GetBlacklistingTime()+constant.BlacklistingPeriod <= curTime {
						hostServiceInstance.Host.KnownPeers[nativeUtil.GetFullAddressPeer(p)] = hostServiceInstance.PeerUnblacklist(p)
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
