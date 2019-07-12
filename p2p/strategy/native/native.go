package native

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/p2p/strategy/native/util"

	// nativeUtil "github.com/zoobc/zoobc-core/p2p/strategy/native/util"
	"google.golang.org/grpc"
)

// HostService is
type HostService struct {
	Host       *model.Host
	GrpcServer *grpc.Server
	ChainType  contract.ChainType
}

var hostServiceInstance *HostService

const (
	constantResolvePeersGapSecond uint32 = 10
)

// InitService to start peer process with
func (hs *HostService) InitService(myAddress string, port uint32, wellknownPeers []string) (contract.P2PType, error) {
	if hostServiceInstance == nil {
		knownPeersResult, err := util.ParseKnownPeers(wellknownPeers)
		if err != nil {
			return nil, err
		}

		host := util.NewHost(myAddress, port, knownPeersResult)
		hostServiceInstance = &HostService{
			Host: host,
		}
	}
	return hostServiceInstance, nil
}

// StartP2P Get instance of intialized host service
func (hs *HostService) StartP2P(chaintype contract.ChainType) {
	hostServiceInstance.ChainType = chaintype
	hostServiceInstance.startServer()

	// p2p thread
	go hostServiceInstance.resolvePeersThread()
	go hostServiceInstance.getMorePeersThread()
	go hostServiceInstance.updateBlacklistedStatus()
}

// ResolvePeersThread to checking UnresolvedPeer
func (hs *HostService) resolvePeersThread() {
	go hs.resolvePeers()
	ticker := util.GetTickerTime(constantResolvePeersGapSecond)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-ticker.C:
			go hs.resolvePeers()
		case <-sigs:
			ticker.Stop()
			return
		}
	}
}

func (hs *HostService) resolvePeers() {
	for _, peer := range hs.Host.GetUnresolvedPeers() {
		go hs.resolvePeer(peer)
	}
}

func (hs *HostService) resolvePeer(destPeer *model.Peer) {
	_, err := NewPeerService(hs.ChainType).GetPeerInfo(destPeer)
	if err != nil {
		return
	}
	updatedHost := util.AddToResolvedPeer(hs.Host, destPeer)
	hs.Host = updatedHost

	log.Info(util.GetFullAddressPeer(destPeer) + " success")
}

// getMorePeersThread to periodically request more peers from another node
func (hs *HostService) getMorePeersThread() {
	go hs.getMorePeers()
	ticker := util.GetTickerTime(constantResolvePeersGapSecond)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-ticker.C:
			go hs.getMorePeers()
		case <-sigs:
			ticker.Stop()
			return
		}
	}
}

func (hs *HostService) getMorePeers() {
	peer := util.GetAnyPeer(hs.Host)
	if peer != nil {
		newPeers, err := NewPeerService(hs.ChainType).GetMorePeers(peer)
		if err != nil {
			log.Warnf("getMorePeers Error accord %v\n", err)
		}
		hostServiceInstance.Host = util.AddToUnresolvedPeers(hostServiceInstance.Host, newPeers.GetPeers())
	}
}

// updateBlacklistedStatus go routine that checks, every 60sec if there are blacklisted peers to unblacklist
func (hs *HostService) updateBlacklistedStatus() {
	ticker := util.GetTickerTime(60)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	BlacklistingPeriodSeconds := uint32(5) // ---> draft
	go func() {
		for {
			select {
			case <-ticker.C:
				curTime := uint32(time.Now().Unix())
				for _, p := range hs.Host.GetKnownPeers() {
					if p.GetState() == model.PeerState_BLACKLISTED &&
						p.GetBlacklistingTime() > 0 &&
						p.GetBlacklistingTime()+BlacklistingPeriodSeconds <= curTime {
						hs.Host.KnownPeers[util.GetFullAddressPeer(p)] = util.PeerUnblacklist(p)
					}
				}
			case <-sigs:
				ticker.Stop()
				return
			}
		}
	}()
}
