package p2p

import (
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/model"
	coreService "github.com/zoobc/zoobc-core/core/service"
	nativeService "github.com/zoobc/zoobc-core/p2p/native/service"
)

type P2pServiceInterface interface {
	InitService(myAddress string, port uint32, wellknownPeers []string) (P2pServiceInterface, error)
	SetBlockServices(blockServices map[int32]coreService.BlockServiceInterface)
	StartP2P()

	GetHostInstance() *model.Host
	DisconnectPeer(*model.Peer)

	// GetAnyResolvedPeer Get any random connected peer
	GetAnyResolvedPeer() *model.Peer
	// GetResolvedPeers returns resolved peers in thread-safe manner
	GetResolvedPeers() map[string]*model.Peer

	nativeService.PeerServiceClientInterface
}

// InitP2P to initialize p2p strategy will used
func InitP2P(myAddress string, port uint32, wellknownPeers []string, p2pType P2pServiceInterface) P2pServiceInterface {
	p2pService, err := p2pType.InitService(myAddress, port, wellknownPeers)
	if err != nil {
		log.Fatalf("Faild to initialize P2P service\nError : %v\n", err)
	}
	return p2pService
}
