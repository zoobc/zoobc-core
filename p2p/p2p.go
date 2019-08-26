package p2p

import (
	log "github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/model"
	coreService "github.com/zoobc/zoobc-core/core/service"
	"github.com/zoobc/zoobc-core/observer"
	nativeService "github.com/zoobc/zoobc-core/p2p/native/service"
)

type ServiceInterface interface {
	InitService(myAddress string, port uint32, wellknownPeers []string,
		obsr *observer.Observer, nodeSecretPhrase string) (ServiceInterface, error)
	SetBlockServices(blockServices map[int32]coreService.BlockServiceInterface)
	StartP2P()

	GetHostInstance() *model.Host
	DisconnectPeer(*model.Peer)

	// GetAnyResolvedPeer Get any random connected peer
	GetAnyResolvedPeer() *model.Peer
	// GetResolvedPeers returns resolved peers in thread-safe manner
	GetResolvedPeers() map[string]*model.Peer

	SendBlockListener() observer.Listener
	SendTransactionListener() observer.Listener
	nativeService.PeerServiceClientInterface
}

// InitP2P to initialize p2p strategy will used
// TODO: Add Switcer Interface
func InitP2P(
	myAddress string,
	port uint32,
	wellknownPeers []string,
	p2pType ServiceInterface,
	obsr *observer.Observer,
	nodeSecretPhrase string) ServiceInterface {
	p2pService, err := p2pType.InitService(
		myAddress, port, wellknownPeers, obsr, nodeSecretPhrase)
	if err != nil {
		log.Fatalf("Faild to initialize P2P service\nError : %v\n", err)
	}
	return p2pService
}
