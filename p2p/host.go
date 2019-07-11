package p2p

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zoobc/zoobc-core/common/contract"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/p2p/util"
	"google.golang.org/grpc"
)

/*
HostService is
*/
type HostService struct {
	Host       *model.Host
	GrpcServer *grpc.Server
	ChainType  contract.ChainType
}

var hostServiceInstance *HostService

const (
	resolvePeersGapSecond uint8 = 10
)

// InitHostService to start peer process
func InitHostService(myAddress string, port uint32, wellknownPeers []string) (*HostService, error) {
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

// NewHostService Get instance of intialized host service
func NewHostService(chaintype contract.ChainType) *HostService {
	fmt.Printf("chaintype %v\n", chaintype)
	host := new(HostService)
	host.ChainType = chaintype
	fmt.Println(host)

	hostServiceInstance.ChainType = chaintype

	return hostServiceInstance
}

// ResolvePeersThread to checking UnresolvedPeer
func (hs HostService) ResolvePeersThread() {
	go hs.resolvePeers()
	ticker := time.NewTicker(time.Duration(resolvePeersGapSecond) * time.Second)
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

func (hs HostService) resolvePeers() {
	// hs.Host.Log("resolving peers")

	for _, peer := range hs.Host.GetUnresolvedPeers() {
		go hs.resolvePeer(peer)
	}
}

func (hs HostService) resolvePeer(destPeer *model.Peer) {
	_, err := NewPeerService(0).GetPeerInfo(destPeer)
	if err != nil {
		return
	}
	// hs.Host.Log(util.GetFullAddressPeer(destPeer) + " success")
	updatedHost := util.ResolvedPeer(hs.Host, destPeer)
	hs.Host = updatedHost
}

// GetMorePeersThread to Get More Peers
func (hs HostService) GetMorePeersThread() {
	go hs.getMorePeers()
	ticker := time.NewTicker(time.Duration(resolvePeersGapSecond) * time.Second)
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

func (hs HostService) getMorePeers() {
	peer := util.GetAnyPeer(hs.Host)
	if peer != nil {
		go NewPeerService(0).GetMorePeers(peer)
	}
}
