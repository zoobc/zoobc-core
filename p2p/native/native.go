package native

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/p2p/native/service"

	nativeUtil "github.com/zoobc/zoobc-core/p2p/native/util"
)

type Service struct {
	HostService *service.HostService
}

var hostServiceInstance *service.HostService

const (
	constantResolvePeersGapSecond uint32 = 10
)

// InitService to initialize hostServiceInstance if not set
func (s *Service) InitService(myAddress string, port uint32, wellknownPeers []string) (contract.P2PType, error) {
	if hostServiceInstance == nil {
		knownPeersResult, err := nativeUtil.ParseKnownPeers(wellknownPeers)
		if err != nil {
			return nil, err
		}

		host := nativeUtil.NewHost(myAddress, port, knownPeersResult)
		hostServiceInstance = &service.HostService{
			Host: host,
		}
		s.HostService = hostServiceInstance
	}
	return s, nil
}

func (s *Service) GetHostInstance() *model.Host {
	return s.HostService.Host
}

// StartP2P to update  ChainType of hostServiceInstance and run all p2p Thread service
func (s *Service) StartP2P(chaintype contract.ChainType) {
	hostServiceInstance.ChainType = chaintype
	startServer()

	// p2p thread
	go resolvePeersThread()
	go getMorePeersThread()
	go updateBlacklistedStatus()
}

// startServer to run p2p service as server
func startServer() {
	go hostServiceInstance.StartListening()
}

// ResolvePeersThread to periodically try get response from peers in UnresolvedPeer list
func resolvePeersThread() {
	go hostServiceInstance.ResolvePeers()
	ticker := nativeUtil.GetTickerTime(constantResolvePeersGapSecond)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-ticker.C:
			go hostServiceInstance.ResolvePeers()
			go hostServiceInstance.UpdateConnectedPeers()
		case <-sigs:
			ticker.Stop()
			return
		}
	}
}

// getMorePeersThread to periodically request more peers from another node in Peers list
func getMorePeersThread() {
	go hostServiceInstance.GetMorePeersHandler()
	ticker := nativeUtil.GetTickerTime(constantResolvePeersGapSecond)
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
	BlacklistingPeriodSeconds := uint32(5) // ---> draft
	go func() {
		for {
			select {
			case <-ticker.C:
				curTime := uint32(time.Now().Unix())
				for _, p := range hostServiceInstance.Host.GetKnownPeers() {
					if p.GetState() == model.PeerState_BLACKLISTED &&
						p.GetBlacklistingTime() > 0 &&
						p.GetBlacklistingTime()+BlacklistingPeriodSeconds <= curTime {
						hostServiceInstance.Host.KnownPeers[nativeUtil.GetFullAddressPeer(p)] = nativeUtil.PeerUnblacklist(p)
					}
				}
			case <-sigs:
				ticker.Stop()
				return
			}
		}
	}()
}
